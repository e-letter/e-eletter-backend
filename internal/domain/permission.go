package domain

import "time"

type PermissionRequest struct {
	RequestID       int        `json:"request_id" db:"id"`
	TypeID          int        `json:"type_id" db:"request_type_id"`
	RequestNumber   string     `json:"request_number" db:"request_number"`
	RequesterUserID int        `json:"requester_user_id" db:"requester_user_id"`
	Reason          *string    `json:"reason,omitempty" db:"reason"`
	RequestDate     *string    `json:"request_date,omitempty" db:"request_date"`
	StartTime       *string    `json:"start_time,omitempty" db:"start_time"`
	EndTime         *string    `json:"end_time,omitempty" db:"end_time"`
	Status          string     `json:"status" db:"status"`
	CurrentStep     int        `json:"current_step" db:"current_step"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type PermissionClass struct {
	ClassID   int    `json:"class_id" db:"class_id"`
	ClassName string `json:"class_name" db:"class_name"`
}

type PermissionMajor struct {
	MajorID    int    `json:"major_id" db:"major_id"`
	MajorName  string `json:"major_name" db:"major_name"`
	MajorShort string `json:"major_short" db:"major_short"`
}

type ApprovalRequest struct {
	RequestID    int     `json:"request_id"`
	StageID      int     `json:"stage_id"`
	Status       string  `json:"status"`
	Notes        *string `json:"notes"`
	SignatureURL *string `json:"signature_url"`
}

type CreatePermissionRequest struct {
	IDSiswa      int     `json:"id_siswa"`
	TypeID       int     `json:"type_id"`
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	EventLocation *string `json:"event_location"`
	RequestDate  string  `json:"request_date"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
}

type UpdatePermissionRequest struct {
	RequestID     int     `json:"request_id"`
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	EventLocation *string `json:"event_location"`
	StartTime     *string `json:"start_time"`
	EndTime       *string `json:"end_time"`
	Status        *string `json:"status"`
}

type PermissionRepository interface {
	ListAll() ([]PermissionRequest, error)
	ListByUser(userID int) ([]PermissionRequest, error)
	ListClasses() ([]PermissionClass, error)
	ListMajors() ([]PermissionMajor, error)
	GetUserByNISN(nisn string) (*User, error)
	GetUserByID(userID int) (*User, error)
	Create(req CreatePermissionRequest) (int, error)
	Update(req UpdatePermissionRequest) error
	Delete(requestID int) error
	Approve(req ApprovalRequest, approverID int) error
	ListRegistrationTokens() ([]TokenRecord, error)
	CreateOrUpdateRegistrationToken(token string, roleID int, usageLimit *int, expiresAt *time.Time) error
	GetRegistrationTokenByValue(token string) (*TokenRecord, error)
	CancelRequest(requestID, userID int, reason string) error
	GetRequestDetail(requestID int) (any, error)
	GetTeacherRoles(userID int) (any, error)
	RequestTeacherRole(userID int, roleName string) error
	CreateDelegation(userID, delegateUserID int, validFrom, validUntil, reason string) error
	ListDelegations(userID int) (any, error)
	DeleteDelegation(id, userID int) error
}

type PermissionService interface {
	Get(action, idSiswa, nisn string, userID int, roleID int) (any, error)
	Create(req CreatePermissionRequest) (int, error)
	Update(req UpdatePermissionRequest) error
	Delete(id int) error
	Approve(req ApprovalRequest, approverID int) error
	ListRegistrationTokens() ([]TokenRecord, error)
	UpsertRegistrationToken(token string, roleID int, usageLimit *int, expiresAt *time.Time) (*TokenRecord, error)
	CancelRequest(requestID, userID int, reason string) error
	GetRequestDetail(requestID int) (any, error)
	GetTeacherRoles(userID int) (any, error)
	RequestTeacherRole(userID int, roleName string) error
	CreateDelegation(userID, delegateUserID int, validFrom, validUntil, reason string) error
	ListDelegations(userID int) (any, error)
	DeleteDelegation(id, userID int) error
}
