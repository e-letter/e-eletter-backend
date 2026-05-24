package repository

import (
	"database/sql"
	"fmt"
	"time"
)

var romanMonths = [...]string{
	"", "I", "II", "III", "IV", "V", "VI",
	"VII", "VIII", "IX", "X", "XI", "XII",
}

// generateRequestNumber generates an atomic request number within the given transaction.
// Format: {PREFIX}/{COUNTER:03d}/{SCHOOL_CODE}/{MONTH_ROMAN}/{YEAR}
func generateRequestNumber(tx *sql.Tx, requestTypeID int, schoolCode string) (string, int, error) {
	// Get active academic year
	var academicYearID int
	err := tx.QueryRow(`SELECT id FROM academic_years WHERE is_active = 1 LIMIT 1`).Scan(&academicYearID)
	if err != nil {
		return "", 0, fmt.Errorf("no active academic year: %w", err)
	}

	// Get letter prefix from request_types
	var prefix string
	err = tx.QueryRow(`SELECT letter_prefix FROM request_types WHERE id = ?`, requestTypeID).Scan(&prefix)
	if err != nil {
		return "", 0, fmt.Errorf("invalid request_type_id: %w", err)
	}

	// Ensure a counter row exists for this academic year and request type.
	_, err = tx.Exec(
		`INSERT IGNORE INTO letter_number_counters (academic_year_id, request_type_id, last_counter)
		 VALUES (?, ?, 0)`,
		academicYearID, requestTypeID,
	)
	if err != nil {
		return "", 0, fmt.Errorf("counter initialization failed: %w", err)
	}

	// Seed the counter from already existing requests so legacy/manual data does not
	// cause the next generated number to collide with uq_request_number.
	var existingMaxCounter int
	err = tx.QueryRow(
		`SELECT COALESCE(MAX(CAST(SUBSTRING_INDEX(SUBSTRING_INDEX(request_number, '/', 2), '/', -1) AS UNSIGNED)), 0)
		 FROM requests
		 WHERE academic_year_id = ? AND request_type_id = ?`,
		academicYearID, requestTypeID,
	).Scan(&existingMaxCounter)
	if err != nil {
		return "", 0, fmt.Errorf("existing counter lookup failed: %w", err)
	}

	// Atomically bump the stored counter to the next available number.
	_, err = tx.Exec(
		`UPDATE letter_number_counters
		 SET last_counter = GREATEST(last_counter, ?) + 1
		 WHERE academic_year_id = ? AND request_type_id = ?`,
		existingMaxCounter, academicYearID, requestTypeID,
	)
	if err != nil {
		return "", 0, fmt.Errorf("counter update failed: %w", err)
	}

	var counter int
	err = tx.QueryRow(
		`SELECT last_counter FROM letter_number_counters WHERE academic_year_id = ? AND request_type_id = ?`,
		academicYearID, requestTypeID,
	).Scan(&counter)

	if err != nil && err != sql.ErrNoRows {
		return "", 0, fmt.Errorf("failed to get counter: %w", err)
	}

	if err == sql.ErrNoRows {
		// Insert new counter row
		counter = 1
		_, err = tx.Exec(
			`INSERT INTO letter_number_counters (academic_year_id, request_type_id, last_counter) VALUES (?, ?, 1)`,
			academicYearID, requestTypeID,
		)
		if err != nil {
			return "", 0, fmt.Errorf("failed to insert counter: %w", err)
		}
	} else {
		counter++
		_, err = tx.Exec(
			`UPDATE letter_number_counters SET last_counter = ? WHERE academic_year_id = ? AND request_type_id = ?`,
			counter, academicYearID, requestTypeID,
		)
		if err != nil {
			return "", 0, fmt.Errorf("failed to update counter: %w", err)
		}
	}

	now := time.Now()
	requestNumber := fmt.Sprintf("%s/%03d/%s/%s/%d", prefix, counter, schoolCode, romanMonths[now.Month()], now.Year())
	for i := 0; i < 100; i++ {
		var exists bool
		err = tx.QueryRow(`SELECT 1 FROM requests WHERE request_number = ? LIMIT 1`, requestNumber).Scan(&exists)
		if err == sql.ErrNoRows {
			break // Number is unique, proceed
		}
		counter++
		requestNumber = fmt.Sprintf("%s/%03d/%s/%s/%d", prefix, counter, schoolCode, romanMonths[now.Month()], now.Year())
		_, err = tx.Exec(
			`UPDATE letter_number_counters SET last_counter = ? WHERE academic_year_id = ? AND request_type_id = ?`,
			counter, academicYearID, requestTypeID,
		)
		if err != nil {
			return "", 0, fmt.Errorf("counter update failed: %w", err)
		}
	}

	return requestNumber, academicYearID, nil
}
