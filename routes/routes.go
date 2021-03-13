package routes

import (
	"fa-middleware/auth"
	"fa-middleware/config"
	"fa-middleware/htmltemplates"
	"fa-middleware/models"
	"fa-middleware/userdata"
	"net/http"

	"log"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"github.com/gin-gonic/gin"
)

func LoggedIn(c *gin.Context, conf config.Config, fa *fusionauth.FusionAuthClient) {
	cookies := c.Request.Cookies()
	jwt := ""
	for _, cookie := range cookies {
		if cookie.Name == conf.JWTCookieName {
			jwt = cookie.Value
			break
		}
	}

	if jwt == "" {
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}

	// check if the user has a valid jwt
	user, err := auth.GetUserByJWT(conf, fa, jwt)
	if err != nil {
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}

	// render the loggedin.html template
	htmlstr, err := htmltemplates.GetLoggedInTemplate(user)
	if err != nil {
		log.Printf("error getting template: %v", err.Error())
		c.Data(500, "text/plain", []byte("server error!"))
		return
	}
	c.Data(200, "text/html", []byte(htmlstr))
}

func OauthCallback(c *gin.Context, conf config.Config, fa *fusionauth.FusionAuthClient, codeVerif string) {
	// https://github.com/gin-gonic/examples/blob/master/basic/main.go
	err := c.Request.ParseForm()
	if err != nil {
		log.Printf("oauth-callback failed to process form: %v", err.Error())
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}

	receivedOauthState, ok := c.Request.Form["state"]
	if !ok {
		log.Printf("login: no state")
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}
	receivedOauthCode, ok := c.Request.Form["code"]
	if !ok {
		log.Printf("login: no code")
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}

	if len(receivedOauthState) != 1 || len(receivedOauthCode) != 1 {
		log.Printf("login: didn't receive 1 state and 1 code")
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}

	oauths := models.OauthState{
		Code:     receivedOauthCode[0],
		State:    receivedOauthState[0],
		Verifier: codeVerif,
	}

	// log.Printf("post form array: %v", c.Request.Form)

	user, jwt, err := auth.Login(conf, fa, oauths)
	if err != nil {
		log.Printf("err login: %v", err.Error())
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}

	err = userdata.SetUserData(
		conf,
		user.Id,
		struct {
			TestVal string `yaml:"testVal"`
		}{
			"test1234!",
		},
	)

	if err != nil {
		log.Printf("error setting user data: %v", err.Error())
		c.JSON(500, models.OauthState{})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		conf.JWTCookieName,
		jwt,
		conf.JWTCookieMaxAgeSeconds, // TODO: change this to a bigger value than 1 hour?
		"/",
		conf.JWTCookieDomain,    // TODO: integrate with multi-tenancy?
		conf.JWTCookieSetSecure, // TODO: use secure only
		true,
	)

	c.Redirect(301, "/pages/welcome")

	// c.JSON(200, user)
}

func GetAuthLogout(c *gin.Context, conf config.Config, fa *fusionauth.FusionAuthClient) {
	cookies := c.Request.Cookies()
	jwt := ""
	for _, cookie := range cookies {
		if cookie.Name == conf.JWTCookieName {
			jwt = cookie.Value
			break
		}
	}

	if jwt == "" {
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}

	// get the current user
	u, errs, err := fa.RetrieveUserUsingJWT(jwt)
	if err != nil {
		log.Printf("currentuser/email: %v", err.Error())
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}
	if errs != nil {
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}

	c.Data(200, "text/plain", []byte(u.User.Email))
}

func GetAPICurrentUserEmail(c *gin.Context, conf config.Config, fa *fusionauth.FusionAuthClient) {
	cookies := c.Request.Cookies()
	jwt := ""
	for _, cookie := range cookies {
		if cookie.Name == conf.JWTCookieName {
			jwt = cookie.Value
			break
		}
	}

	if jwt == "" {
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}

	// get the current user
	u, errs, err := fa.RetrieveUserUsingJWT(jwt)
	if err != nil {
		log.Printf("currentuser/email: %v", err.Error())
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}
	if errs != nil {
		c.Data(403, "text/plain", []byte("Unauthorized"))
		return
	}

	c.Data(200, "text/plain", []byte(u.User.Email))
}
