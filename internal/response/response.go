package response

import "github.com/gin-gonic/gin"

// Standard response structure for all API endpoints
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Success returns a successful response with optional data
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error returns an error response
func Error(c *gin.Context, statusCode int, errorMessage string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   errorMessage,
	})
}

// ErrorWithMessage returns an error response with both error and message fields
func ErrorWithMessage(c *gin.Context, statusCode int, errorMessage string, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   errorMessage,
		Message: message,
	})
}

// Raw sends a raw gin.H response (for backward compatibility during migration)
func Raw(c *gin.Context, statusCode int, payload gin.H) {
	c.JSON(statusCode, payload)
}
