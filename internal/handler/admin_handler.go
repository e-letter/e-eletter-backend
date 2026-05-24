package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Refliqx/backend-eletter/internal/response"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) GetStats(c *gin.Context) {
	stats := map[string]int{}
	rows := []struct{ key, query string }{
		{"total_students", `SELECT COUNT(*) FROM users WHERE role='student' AND deleted_at IS NULL`},
		{"total_teachers", `SELECT COUNT(*) FROM users WHERE role IN ('teacher','kepala_sekolah') AND deleted_at IS NULL`},
		{"pending_requests", `SELECT COUNT(*) FROM requests WHERE status='pending'`},
		{"active_tokens", `SELECT COUNT(*) FROM registration_tokens WHERE used_count < usage_limit AND (expires_at IS NULL OR expires_at > NOW())`},
	}
	for _, r := range rows {
		var count int
		_ = h.db.QueryRow(r.query).Scan(&count)
		stats[r.key] = count
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "data": stats})
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	role := c.Query("role")
	status := c.Query("status")
	search := c.Query("search")

	query := `SELECT u.id, u.email, u.role, u.status, COALESCE(tp.full_name, sp.full_name, ap.full_name, pp.full_name, '') as full_name
		FROM users u 
		LEFT JOIN teacher_profiles tp ON tp.user_id = u.id 
		LEFT JOIN student_profiles sp ON sp.user_id = u.id
		LEFT JOIN admin_profiles ap ON ap.user_id = u.id
		LEFT JOIN principal_profiles pp ON pp.user_id = u.id
		WHERE u.deleted_at IS NULL`
	args := []any{}

	if role != "" {
		query += " AND u.role = ?"
		args = append(args, role)
	}
	if status != "" {
		query += " AND u.status = ?"
		args = append(args, status)
	}
	if search != "" {
		query += " AND (u.email LIKE ? OR tp.full_name LIKE ? OR sp.full_name LIKE ? OR ap.full_name LIKE ? OR pp.full_name LIKE ?)"
		s := "%" + search + "%"
		args = append(args, s, s, s, s, s)
	}
	query += " ORDER BY u.id DESC LIMIT 100"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close() 	

	type User struct {
		ID       int    `json:"id"`
		Email    *string `json:"email"`
		Role     string `json:"role"`
		Status   string `json:"status"`
		FullName string `json:"full_name"`
	}
	var users []User
	for rows.Next() {
		var u User
		rows.Scan(&u.ID, &u.Email, &u.Role, &u.Status, &u.FullName)
		users = append(users, u)
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "data": users})
}

func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.db.Exec(`UPDATE users SET status = ? WHERE id = ?`, body.Status, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Status berhasil diperbarui"})
}

func (h *AdminHandler) GetRegistrationTokens(c *gin.Context) {
	rows, err := h.db.Query(`SELECT token_id, token, role_id, usage_limit, used_count, expires_at, created_at FROM registration_tokens ORDER BY created_at DESC`)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type Token struct {
		ID         int     `json:"id"`
		Token      string  `json:"token"`
		RoleID     int     `json:"role_id"`
		UsageLimit int     `json:"usage_limit"`
		UsedCount  int     `json:"used_count"`
		ExpiresAt  *string `json:"expires_at"`
		CreatedAt  string  `json:"created_at"`
	}
	var tokens []Token
	for rows.Next() {
		var t Token
		rows.Scan(&t.ID, &t.Token, &t.RoleID, &t.UsageLimit, &t.UsedCount, &t.ExpiresAt, &t.CreatedAt)
		tokens = append(tokens, t)
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "data": tokens})
}

func (h *AdminHandler) CreateRegistrationToken(c *gin.Context) {
	var body struct {
		Token      string  `json:"token" binding:"required"`
		RoleID     int     `json:"role_id" binding:"required"`
		UsageLimit int     `json:"usage_limit"`
		ExpiresAt  *string `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if body.UsageLimit == 0 {
		body.UsageLimit = 100
	}
	_, err := h.db.Exec(`INSERT INTO registration_tokens (token, role_id, usage_limit, expires_at) VALUES (?, ?, ?, ?)`,
		body.Token, body.RoleID, body.UsageLimit, body.ExpiresAt)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Raw(c, http.StatusCreated, gin.H{"success": true, "message": "Token berhasil dibuat"})
}

func (h *AdminHandler) DeleteRegistrationToken(c *gin.Context) {
	id := c.Param("id")
	_, err := h.db.Exec(`DELETE FROM registration_tokens WHERE token_id = ?`, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Token berhasil dihapus"})
}

func (h *AdminHandler) VerifyTeacherRole(c *gin.Context) {
	id := c.Param("id")
	_, err := h.db.Exec(`UPDATE teacher_roles SET status = 'active', verified_at = NOW(), verified_by = 0 WHERE id = ?`, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Peran guru berhasil diverifikasi"})
}

func (h *AdminHandler) GetAcademicYears(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, year_name, semester, is_active, start_date, end_date FROM academic_years ORDER BY id DESC`)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	type AY struct {
		ID        int    `json:"id"`
		YearName  string `json:"year_name"`
		Semester  int    `json:"semester"`
		IsActive  bool   `json:"is_active"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	var items []AY
	for rows.Next() {
		var a AY
		rows.Scan(&a.ID, &a.YearName, &a.Semester, &a.IsActive, &a.StartDate, &a.EndDate)
		items = append(items, a)
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "data": items})
}

func (h *AdminHandler) CreateAcademicYear(c *gin.Context) {
	var body struct {
		YearName  string `json:"year_name" binding:"required"`
		Semester  int    `json:"semester" binding:"required"`
		StartDate string `json:"start_date" binding:"required"`
		EndDate   string `json:"end_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.db.Exec(`INSERT INTO academic_years (year_name, semester, start_date, end_date, is_active) VALUES (?, ?, ?, ?, 0)`,
		body.YearName, body.Semester, body.StartDate, body.EndDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Raw(c, http.StatusCreated, gin.H{"success": true, "message": "Tahun ajaran berhasil dibuat"})
}

func (h *AdminHandler) UpdateAcademicYear(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		YearName  *string `json:"year_name"`
		Semester  *int    `json:"semester"`
		IsActive  *bool   `json:"is_active"`
		StartDate *string `json:"start_date"`
		EndDate   *string `json:"end_date"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if body.IsActive != nil && *body.IsActive {
		h.db.Exec(`UPDATE academic_years SET is_active = 0`)
	}
	h.db.Exec(`UPDATE academic_years SET year_name = COALESCE(?, year_name), semester = COALESCE(?, semester), is_active = COALESCE(?, is_active), start_date = COALESCE(?, start_date), end_date = COALESCE(?, end_date) WHERE id = ?`,
		body.YearName, body.Semester, body.IsActive, body.StartDate, body.EndDate, id)
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Tahun ajaran berhasil diperbarui"})
}

func (h *AdminHandler) DeleteAcademicYear(c *gin.Context) {
	id := c.Param("id")
	h.db.Exec(`DELETE FROM academic_years WHERE id = ?`, id)
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Tahun ajaran berhasil dihapus"})
}

func (h *AdminHandler) GetClasses(c *gin.Context) {
	rows, err := h.db.Query(`SELECT c.id, c.class_name, c.major_id, COALESCE(m.name,'') FROM classes c LEFT JOIN majors m ON m.id = c.major_id ORDER BY c.class_name`)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	type Class struct {
		ID        int    `json:"id"`
		ClassName string `json:"class_name"`
		MajorID   int    `json:"major_id"`
		MajorName string `json:"major_name"`
	}
	var items []Class
	for rows.Next() {
		var item Class
		rows.Scan(&item.ID, &item.ClassName, &item.MajorID, &item.MajorName)
		items = append(items, item)
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "data": items})
}

func (h *AdminHandler) CreateClass(c *gin.Context) {
	var body struct {
		ClassName string `json:"class_name" binding:"required"`
		MajorID   int    `json:"major_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	h.db.Exec(`INSERT INTO classes (class_name, major_id) VALUES (?, ?)`, body.ClassName, body.MajorID)
	response.Raw(c, http.StatusCreated, gin.H{"success": true, "message": "Kelas berhasil dibuat"})
}

func (h *AdminHandler) UpdateClass(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		ClassName string `json:"class_name"`
		MajorID   int    `json:"major_id"`
	}
	c.ShouldBindJSON(&body)
	h.db.Exec(`UPDATE classes SET class_name = ?, major_id = ? WHERE id = ?`, body.ClassName, body.MajorID, id)
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Kelas berhasil diperbarui"})
}

func (h *AdminHandler) DeleteClass(c *gin.Context) {
	h.db.Exec(`DELETE FROM classes WHERE id = ?`, c.Param("id"))
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Kelas berhasil dihapus"})
}

func (h *AdminHandler) GetMajors(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, name, code FROM majors ORDER BY name`)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	type Major struct {
		ID         int    `json:"id"`
		MajorName  string `json:"major_name"`
		MajorShort string `json:"major_short"`
	}
	var items []Major
	for rows.Next() {
		var item Major
		rows.Scan(&item.ID, &item.MajorName, &item.MajorShort)
		items = append(items, item)
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "data": items})
}

func (h *AdminHandler) CreateMajor(c *gin.Context) {
	var body struct {
		MajorName  string `json:"major_name" binding:"required"`
		MajorShort string `json:"major_short"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.db.Exec(`INSERT INTO majors (name, code) VALUES (?, ?)`, body.MajorName, body.MajorShort)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Raw(c, http.StatusCreated, gin.H{"success": true, "message": "Jurusan berhasil dibuat"})
}

func (h *AdminHandler) UpdateMajor(c *gin.Context) {
	var body struct {
		MajorName  string `json:"major_name"`
		MajorShort string `json:"major_short"`
	}
	c.ShouldBindJSON(&body)
	_, err := h.db.Exec(`UPDATE majors SET name = ?, code = ? WHERE id = ?`, body.MajorName, body.MajorShort, c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Jurusan berhasil diperbarui"})
}

func (h *AdminHandler) DeleteMajor(c *gin.Context) {
	h.db.Exec(`DELETE FROM majors WHERE id = ?`, c.Param("id"))
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Jurusan berhasil dihapus"})
}

func (h *AdminHandler) GetSubjects(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, name, code FROM subjects ORDER BY name`)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	type Subject struct {
		ID          int    `json:"id"`
		SubjectName string `json:"subject_name"`
		SubjectCode string `json:"subject_code"`
	}
	var items []Subject
	for rows.Next() {
		var item Subject
		rows.Scan(&item.ID, &item.SubjectName, &item.SubjectCode)
		items = append(items, item)
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "data": items})
}

func (h *AdminHandler) CreateSubject(c *gin.Context) {
	var body struct {
		SubjectName string `json:"subject_name" binding:"required"`
		SubjectCode string `json:"subject_code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.db.Exec(`INSERT INTO subjects (name, code) VALUES (?, ?)`, body.SubjectName, body.SubjectCode)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Raw(c, http.StatusCreated, gin.H{"success": true, "message": "Mata pelajaran berhasil dibuat"})
}

func (h *AdminHandler) UpdateSubject(c *gin.Context) {
	var body struct {
		SubjectName string `json:"subject_name"`
		SubjectCode string `json:"subject_code"`
	}
	c.ShouldBindJSON(&body)
	_, err := h.db.Exec(`UPDATE subjects SET name = ?, code = ? WHERE id = ?`, body.SubjectName, body.SubjectCode, c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Mata pelajaran berhasil diperbarui"})
}

func (h *AdminHandler) DeleteSubject(c *gin.Context) {
	h.db.Exec(`DELETE FROM subjects WHERE id = ?`, c.Param("id"))
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Mata pelajaran berhasil dihapus"})
}

func (h *AdminHandler) GetEnrollments(c *gin.Context) {
	classID := c.Query("class_id")
	query := `SELECT sce.id, sce.student_id, sce.class_id, sce.academic_year_id, COALESCE(sp.full_name,'') as student_name, COALESCE(sp.student_code,'') as student_code
		FROM student_class_enrollments sce
		JOIN student_profiles sp ON sp.id = sce.student_id
		WHERE sce.is_active = 1`
	args := []any{}
	if classID != "" {
		query += " AND sce.class_id = ?"
		args = append(args, classID)
	}
	query += " ORDER BY sp.full_name LIMIT 200"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	type Enrollment struct {
		ID             int    `json:"id"`
		StudentID      int    `json:"student_id"`
		ClassID        int    `json:"class_id"`
		AcademicYearID int    `json:"academic_year_id"`
		StudentName    string `json:"student_name"`
		StudentCode    string `json:"student_code"`
	}
	var items []Enrollment
	for rows.Next() {
		var item Enrollment
		rows.Scan(&item.ID, &item.StudentID, &item.ClassID, &item.AcademicYearID, &item.StudentName, &item.StudentCode)
		items = append(items, item)
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "data": items})
}

func (h *AdminHandler) CreateEnrollment(c *gin.Context) {
	var body struct {
		StudentID      int `json:"student_id" binding:"required"`
		ClassID        int `json:"class_id" binding:"required"`
		AcademicYearID int `json:"academic_year_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if body.AcademicYearID == 0 {
		h.db.QueryRow(`SELECT id FROM academic_years WHERE is_active = 1 LIMIT 1`).Scan(&body.AcademicYearID)
	}
	h.db.Exec(`INSERT INTO student_class_enrollments (student_id, class_id, academic_year_id, is_active) VALUES (?, ?, ?, 1)`,
		body.StudentID, body.ClassID, body.AcademicYearID)
	response.Raw(c, http.StatusCreated, gin.H{"success": true, "message": "Enrollment berhasil dibuat"})
}

func (h *AdminHandler) DeleteEnrollment(c *gin.Context) {
	h.db.Exec(`DELETE FROM student_class_enrollments WHERE id = ?`, c.Param("id"))
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Enrollment berhasil dihapus"})
}

func (h *AdminHandler) GetSchoolConfig(c *gin.Context) {
	rows, err := h.db.Query(`SELECT config_key, config_value FROM school_config`)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	config := map[string]string{}
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		config[k] = v
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "data": config})
}

func (h *AdminHandler) UpdateSchoolConfig(c *gin.Context) {
	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	for k, v := range body {
		h.db.Exec(`INSERT INTO school_config (config_key, config_value) VALUES (?, ?) ON DUPLICATE KEY UPDATE config_value = ?`, k, v, v)
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "message": "Konfigurasi berhasil diperbarui"})
}

func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := 50
	offset := (page - 1) * limit

	rows, err := h.db.Query(`SELECT id, user_id, action, details, ip_address, created_at FROM audit_logs ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	type Log struct {
		ID        int     `json:"id"`
		UserID    int     `json:"user_id"`
		Action    string  `json:"action"`
		Details   *string `json:"details"`
		IPAddress *string `json:"ip_address"`
		CreatedAt string  `json:"created_at"`
	}
	var logs []Log
	for rows.Next() {
		var l Log
		rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Details, &l.IPAddress, &l.CreatedAt)
		logs = append(logs, l)
	}
	response.Raw(c, http.StatusOK, gin.H{"success": true, "data": logs})
}
