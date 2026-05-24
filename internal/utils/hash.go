package utils

import "golang.org/x/crypto/bcrypt"

// BcryptCostDefault is the default cost for bcrypt hashing (12 is industry standard)
const BcryptCostDefault = 12

// HashPassword hashes a password using bcrypt with the default cost (12)
func HashPassword(p string) (string, error) {
	return HashPasswordWithCost(p, BcryptCostDefault)
}

// HashPasswordWithCost hashes a password using bcrypt with a specified cost
func HashPasswordWithCost(p string, cost int) (string, error) {
	// Validate cost is within acceptable range (4-31 per bcrypt spec)
	if cost < 4 {
		cost = 4
	} else if cost > 31 {
		cost = 31
	}
	b, err := bcrypt.GenerateFromPassword([]byte(p), cost)
	return string(b), err
}

// VerifyPassword verifies a password against a bcrypt hash
func VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
