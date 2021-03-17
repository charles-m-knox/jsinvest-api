package helpers

import "github.com/gin-gonic/gin"

const (
	NotFound                      = "not found"
	ServerError                   = "server error"
	Unauthorized                  = "unauthorized"
	Forbidden                     = "forbidden"
	BadRequest                    = "bad request"
	OK                            = "OK"
	AccessControlAllowMethods     = "Access-Control-Allow-Methods"
	AccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	AccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	CORSMethodsOptPost            = "OPTIONS, POST"
)

// Simple400 sets a quick and easy 400 gin response
func Simple400(c *gin.Context) {
	c.Data(400, "text/plain", []byte(BadRequest))
}

// Simple401 sets a quick and easy 400 gin response
func Simple401(c *gin.Context) {
	c.Data(401, "text/plain", []byte(Unauthorized))
}

// Simple403 sets a quick and easy 400 gin response
func Simple403(c *gin.Context) {
	c.Data(403, "text/plain", []byte(Forbidden))
}

// Simple404 sets a quick and easy 400 gin response
func Simple404(c *gin.Context) {
	c.Data(404, "text/plain", []byte(NotFound))
}

// Simple500 sets a quick and easy 500 gin response
func Simple500(c *gin.Context) {
	c.Data(500, "text/plain", []byte(ServerError))
}

// Simple200OK sets a quick and easy gin response, typically used for Options
// preflight CORS requests
func Simple200OK(c *gin.Context) {
	c.Data(200, "text/plain", []byte(OK))
}

// SetCORSMethods sets Options and Post headers for CORS
func SetCORSMethods(c *gin.Context) {
	c.Header(AccessControlAllowMethods, CORSMethodsOptPost)
}
