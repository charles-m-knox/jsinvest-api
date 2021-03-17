package routes

import (
	"fa-middleware/auth"
	"fa-middleware/config"
	h "fa-middleware/helpers"
	"fa-middleware/models"
	"fa-middleware/payments"

	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"github.com/gin-gonic/gin"
)

// GetConfigViaRouteOrigin sets the CORS headers that will allow HttpOnly
// cookies to work when requests are made via the web browser, as well as
// automatically retrieving the app config that corresponds to the request
// origin
func GetConfigViaRouteOrigin(c *gin.Context, conf config.Config) (app config.App, success bool) {
	originHeader := c.Request.Header.Get("Origin")
	if originHeader == "" {
		referer := c.Request.Header.Get("Referer")
		if referer == "" {
			h.Simple404(c)
			return
		}
		originHeader = referer
	}
	parsedURL, err := url.Parse(originHeader)
	if err != nil {
		h.Simple404(c)
		return
	}
	origin := parsedURL.Host
	log.Printf("origin: %v", origin)
	app, ok := conf.GetAppByOrigin(origin)
	if !ok {
		return app, false
	}
	c.Header(h.AccessControlAllowOrigin, app.FullDomainURL)
	c.Header(h.AccessControlAllowCredentials, "true")
	return app, true
}

// GetUserFromGinJWT extracts the user via the JWT HttpOnly cookie and will
// set the gin response if there's an error
func GetUserFromGinJWT(c *gin.Context, app config.App) (user fusionauth.User, err error) {
	jwt := GetJWTFromGin(c, app)
	if jwt == "" {
		h.Simple401(c)
		return user, fmt.Errorf(h.Unauthorized)
	}

	// check if the user has a valid jwt
	user, err = auth.GetUserByJWT(app, jwt)
	if err != nil {
		h.Simple401(c)
		return user, fmt.Errorf(h.Unauthorized)
	}

	if user.Id == "" {
		h.Simple401(c)
		return user, fmt.Errorf(h.Unauthorized)
	}

	return user, nil
}

// GetJWTFromGin allows for quick retrieval of a JWT HttpOnly cookie from
// a Gin context
func GetJWTFromGin(c *gin.Context, app config.App) string {
	cookies := c.Request.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == app.JWT.CookieName {
			return cookie.Value
		}
	}
	return ""
}

func Register(c *gin.Context, app config.App) {
	register := models.RegisterBody{}
	err := c.Bind(&register)
	if err != nil {
		h.Simple400(c)
		return
	}
	if register.Email == "" || register.Password == "" || register.ConfirmedPassword != register.Password {
		h.Simple400(c)
		return
	}
	credentials := fusionauth.RegistrationRequest{
		// GenerateAuthenticationToken: true, // requires application to have this enabled
		Registration: fusionauth.UserRegistration{
			ApplicationId: app.FusionAuth.AppID,
		},
		User: fusionauth.User{
			Email: register.Email,
			SecureIdentity: fusionauth.SecureIdentity{
				Password: register.Password,
			},
		},
	}
	// Use FusionAuth Go client to log in the user
	authResponse, errors, err := app.FusionAuth.Client.Register("", credentials)
	if err != nil {
		log.Printf("err on register: %v", err.Error())
		h.Simple401(c)
		return
	}

	if errors != nil {
		log.Printf("errors on register: %v", errors.Error())
		h.Simple401(c)
		return
	}
	log.Printf("auth response: %v", authResponse)
	if authResponse.Token == "" {
		log.Printf("empty register token")
		h.Simple401(c)
		return
	}
	// test out the token
	userResp, errs, err := app.FusionAuth.Client.RetrieveUserUsingJWT(authResponse.Token)
	if err != nil {
		log.Printf("failed to retrieve user by token: %v", err.Error())
		h.Simple401(c)
		return
	}
	if errs != nil {
		log.Printf("errors on register: %v", errors.Error())
		h.Simple401(c)
		return
	}
	if userResp.User.Id == "" {
		log.Printf("empty user id after register")
		h.Simple401(c)
		return
	}

	user := userResp.User
	resp := models.LoggedInResponse{}

	resp.LoggedIn = true
	resp.UserID = user.Id
	resp.UserEmail = user.Email
	resp.UserFullName = user.FullName

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		app.JWT.CookieName,
		authResponse.Token,
		app.JWT.CookieMaxAgeSeconds,
		"/",
		app.JWT.CookieDomain,
		app.JWT.CookieSetSecure,
		true,
	)

	customerID, err := payments.PropagateUserToStripe(app, user)
	if err != nil {
		log.Printf(
			"failed to push user %v to stripe: %v",
			user.Id,
			err.Error(),
		)
	}

	log.Printf("new customer id: %v", customerID)

	c.JSON(200, resp)
}

func Login(c *gin.Context, app config.App) {
	login := models.LoginBody{}
	err := c.Bind(&login)
	if err != nil {
		h.Simple400(c)
		return
	}
	if login.Email == "" || login.Password == "" {
		h.Simple400(c)
		return
	}
	credentials := fusionauth.LoginRequest{
		BaseLoginRequest: fusionauth.BaseLoginRequest{
			ApplicationId: app.FusionAuth.AppID,
			IpAddress:     c.ClientIP(),
			NoJWT:         false,
		},
		LoginId:  login.Email,
		Password: login.Password,
	}
	// Use FusionAuth Go client to log in the user
	authResponse, errors, err := app.FusionAuth.Client.Login(credentials)
	if err != nil {
		log.Printf("err on login: %v", err.Error())
		h.Simple401(c)
		return
	}

	if errors != nil {
		log.Printf("errors on login: %v", errors.Error())
		h.Simple401(c)
		return
	}
	log.Printf("auth response: %v", authResponse)
	if authResponse.Token == "" {
		log.Printf("empty login token")
		h.Simple401(c)
		return
	}
	// test out the token
	userResp, errs, err := app.FusionAuth.Client.RetrieveUserUsingJWT(authResponse.Token)
	if err != nil {
		log.Printf("failed to retrieve user by token: %v", err.Error())
		h.Simple401(c)
		return
	}
	if errs != nil {
		log.Printf("errors on login: %v", errors.Error())
		h.Simple401(c)
		return
	}
	if userResp.User.Id == "" {
		log.Printf("empty user id after login")
		h.Simple401(c)
		return
	}

	user := userResp.User
	resp := models.LoggedInResponse{}

	resp.LoggedIn = true
	resp.UserID = user.Id
	resp.UserEmail = user.Email
	resp.UserFullName = user.FullName

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		app.JWT.CookieName,
		authResponse.Token,
		app.JWT.CookieMaxAgeSeconds,
		"/",
		app.JWT.CookieDomain,
		app.JWT.CookieSetSecure,
		true,
	)

	customerID, err := payments.PropagateUserToStripe(app, user)
	if err != nil {
		log.Printf(
			"failed to push user %v to stripe: %v",
			user.Id,
			err.Error(),
		)
	}

	log.Printf("new customer id: %v", customerID)

	c.JSON(200, resp)
}

// LoggedIn allows the frontend to quickly check if the user is logged in
func LoggedIn(c *gin.Context, app config.App, fa *fusionauth.FusionAuthClient) {
	jwt := GetJWTFromGin(c, app)
	resp := models.LoggedInResponse{}

	if jwt == "" {
		log.Printf("loggedin: empty jwt")
		c.JSON(200, resp)
		return
	}

	// check if the user has a valid jwt
	user, err := auth.GetUserByJWT(app, jwt)
	if err != nil {
		log.Printf("loggedin: couldn't get user")
		c.JSON(200, resp)
		return
	}

	resp.LoggedIn = true
	resp.UserID = user.Id
	resp.UserEmail = user.Email
	resp.UserFullName = user.FullName

	c.JSON(200, resp)
}
