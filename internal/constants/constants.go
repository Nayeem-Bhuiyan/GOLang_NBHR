package constants

const (
	// Context keys
	ContextKeyUserID    = "user_id"
	ContextKeyUserEmail = "user_email"
	ContextKeyRoles     = "roles"
	ContextKeyRequestID = "request_id"
	ContextKeyUser      = "current_user"

	// Token types
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	// Header keys
	HeaderRequestID    = "X-Request-ID"
	HeaderCorrelationID = "X-Correlation-ID"
	HeaderAuthorization = "Authorization"
	BearerPrefix        = "Bearer "

	// Pagination defaults
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100

	// Sort directions
	SortAsc  = "asc"
	SortDesc = "desc"

	// Status
	StatusActive   = "active"
	StatusInactive = "inactive"

	// Default roles
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"

	// Permission actions
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionList   = "list"

	// HTTP response messages
	MsgSuccess           = "success"
	MsgCreated           = "created successfully"
	MsgUpdated           = "updated successfully"
	MsgDeleted           = "deleted successfully"
	MsgUnauthorized      = "unauthorized"
	MsgForbidden         = "forbidden"
	MsgNotFound          = "not found"
	MsgValidationFailed  = "validation failed"
	MsgInternalError     = "internal server error"
	MsgBadRequest        = "bad request"
	MsgConflict          = "conflict"
)