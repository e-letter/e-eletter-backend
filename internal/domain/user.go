package domain

type User struct {
	ID       int     `json:"id" db:"id"`
	Username *string `json:"username,omitempty" db:"username"`
	Email    *string `json:"email,omitempty" db:"email"`
	Role     string  `json:"role" db:"role"`
	Status   string  `json:"status" db:"status"`

	// Computed from profile JOINs — not direct columns in users table
	FullName             *string `json:"full_name,omitempty"`
	StudentCode          *string `json:"student_code,omitempty"`
	EmployeeCode         *string `json:"employee_code,omitempty"`
	Gender               *string `json:"gender,omitempty"`
	PhoneNumber          *string `json:"phone_number,omitempty"`
	ClassID              *int    `json:"class_id,omitempty"`
	CanRequestDispensasi bool    `json:"can_request_dispensasi"`
	ProfileCompleted     bool    `json:"profile_completed"`

	PasswordHash string `json:"-" db:"password_hash"`
}
