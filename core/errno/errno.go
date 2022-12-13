package errno

// Common err code
const (
	CodeOk           = iota // Success
	CodeServerErr           // Server default err
	CodeNotLogin            // User not login
	CodeNoPermission        // User no permission
)
