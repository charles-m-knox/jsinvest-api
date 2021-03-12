package main

import (
	"fa-middleware/auth"
	"fa-middleware/config"
	"fa-middleware/models"
	"fa-middleware/userdata"

	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"github.com/gin-gonic/gin"
	cv "github.com/nirasan/go-oauth-pkce-code-verifier"
	"github.com/thanhpk/randstr"
	"golang.org/x/oauth2"
)

func main() {
	// load config
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err.Error())
	}

	// initialize oauth state
	oauthstr := randstr.Hex(16)
	log.Printf("oauthstr: %v", oauthstr)

	// initialize the code verifier for pkce
	codeVerif, err := cv.CreateCodeVerifier()
	if err != nil {
		log.Fatalf("failed to initialize code verifier: %v", err.Error())
	}
	// Create code_challenge with S256 method
	codeChallenge := codeVerif.CodeChallengeS256()
	log.Printf("code challenge: %v", codeChallenge)

	// http client with custom options for usage with fusionauth
	hc := &http.Client{
		Timeout: time.Second * 10,
	}

	faURL, err := url.Parse(conf.FusionAuthHost)
	if err != nil {
		log.Fatalf("failed to parse fusionauth url: %v", err.Error())
	}

	// get the fusionauth client
	fa := fusionauth.NewClient(
		hc,
		faURL,
		conf.FusionAuthAPIKey,
	)

	// build out the oauth2 config
	oauthc := &oauth2.Config{
		RedirectURL:  conf.FusionAuthOauthRedirectURL,
		ClientID:     conf.FusionAuthOauthClientID,
		ClientSecret: conf.FusionAuthOauthClientSecret,
		Scopes:       []string{"openid"},
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("%v/oauth2/authorize", conf.FusionAuthPublicHost),
			TokenURL:  fmt.Sprintf("%v/oauth2/token", conf.FusionAuthPublicHost),
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}

	// start up the api server
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/login", func(c *gin.Context) {
		url := oauthc.AuthCodeURL(
			oauthstr,
			oauth2.SetAuthURLParam("response_type", "code"),
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
		// https://github.com/gin-gonic/examples/blob/master/basic/main.go
		// redirect to the login page
		c.Redirect(
			301,
			url,
		)
	})
	r.GET("/oauth-callback", func(c *gin.Context) {
		// https://github.com/gin-gonic/examples/blob/master/basic/main.go
		err = c.Request.ParseForm()
		if err != nil {
			log.Printf("oauth-callback failed to process form: %v", err.Error())
			c.JSON(403, models.OauthState{})
			return
		}

		receivedOauthState, ok := c.Request.Form["state"]
		if !ok {
			c.JSON(403, models.OauthState{})
			return
		}
		receivedOauthCode, ok := c.Request.Form["code"]
		if !ok {
			c.JSON(403, models.OauthState{})
			return
		}

		if len(receivedOauthState) != 1 || len(receivedOauthCode) != 1 {
			c.JSON(403, models.OauthState{})
			return
		}

		oauths := models.OauthState{
			Code:     receivedOauthCode[0],
			State:    receivedOauthState[0],
			Verifier: codeVerif.String(),
		}

		// log.Printf("post form array: %v", c.Request.Form)

		user, err := auth.Login(conf, fa, oauths)
		if err != nil {
			log.Printf("err login: %v", err.Error())
			c.JSON(403, models.OauthState{})
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
			log.Fatalf("error setting user data: %v", err.Error())
		}

		c.JSON(200, user)
	})
	err = r.Run(
		fmt.Sprintf(
			"%v:%v",
			conf.BindAddr,
			conf.BindPort,
		),
	)
	if err != nil {
		log.Fatalf("error running gin: %v", err.Error())
	}
}
