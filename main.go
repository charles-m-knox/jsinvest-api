package main

import (
	"fa-middleware/config"
	"fa-middleware/htmltemplates"
	"fa-middleware/routes"

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
	r.GET("/pages/makepayment", func(c *gin.Context) {
		htmlstr, err := htmltemplates.GetPaymentTemplate()
		if err != nil {
			log.Printf("error getting template: %v", err.Error())
			c.Data(500, "text/plain", []byte("server error"))
			return
		}

		c.Data(200, "text/html", []byte(htmlstr))
	})
	r.GET("/auth/login", func(c *gin.Context) {
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
	// Not needed - instead, use /logout (directly on the fusionauth host)
	// r.GET("/auth/logout", func(c *gin.Context) {
	// 	routes.GetAuthLogout(c, conf, fa)
	// })
	r.GET("/api/currentuser/email", func(c *gin.Context) {
		routes.GetAPICurrentUserEmail(c, conf, fa)
	})
	r.GET("/pages/welcome", func(c *gin.Context) {
		routes.LoggedIn(c, conf, fa)
	})
	r.GET("/auth/oauth-cb", func(c *gin.Context) {
		routes.OauthCallback(c, conf, fa, codeVerif.String())
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
