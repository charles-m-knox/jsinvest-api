package main

import (
	"fa-middleware/auth"
	"fa-middleware/config"
	h "fa-middleware/helpers"
	"fa-middleware/models"
	"fa-middleware/payments"
	"fa-middleware/routes"

	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"github.com/gin-gonic/gin"
)

func main() {
	payments.InitializeSubscribedUserCache()

	conf, err := config.LoadConfigYaml()
	if err != nil {
		log.Fatalf("failed to load config2: %v", err.Error())
	}

	for i, app := range conf.Apps {
		faURL, err := url.Parse(app.FusionAuth.InternalHostURL)
		if err != nil {
			log.Fatalf("failed to parse fusionauth url: %v", err.Error())
		}

		// http client with custom options for usage with fusionauth
		hc := &http.Client{Timeout: time.Second * 10}

		// get the fusionauth client
		conf.Apps[i].FusionAuth.Client = fusionauth.NewClient(
			hc,
			faURL,
			app.FusionAuth.APIKey,
		)
	}

	// start up the api server
	r := gin.Default()
	r.GET("/mw/ping", func(c *gin.Context) {
		_, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}
		c.JSON(200, gin.H{"message": "pong"})
	})
	r.OPTIONS("/mw/create-checkout-session", func(c *gin.Context) {
		_, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}
		h.SetCORSMethods(c)
		h.Simple200OK(c)
	})
	r.POST("/mw/create-checkout-session", func(c *gin.Context) {
		app, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}

		user, err := routes.GetUserFromGinJWT(c, app)
		if err != nil {
			return
		}

		h.SetCORSMethods(c)

		err = payments.CreateCheckoutSession(c, app, user)
		if err != nil {
			log.Printf(
				"failed to create checkout session for user %v: %v",
				user.Id,
				err.Error(),
			)
			h.Simple500(c)
			return
		}
	})
	r.OPTIONS("/mw/substatus", func(c *gin.Context) {
		_, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}
		h.Simple200OK(c)
	})
	r.GET("/mw/substatus", func(c *gin.Context) {
		// alllows a logged-in user to check to see if they are subscribed
		// to a product
		app, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}
		user, err := routes.GetUserFromGinJWT(c, app) // will set the gin response if there's an error
		if err != nil {
			return
		}
		productID := c.Query("p")
		if productID == "" { // TODO: add validation that this app contains this product ID
			c.Data(400, "text/plain", []byte("invalid p value"))
			return
		}
		subscribed, err := payments.IsUserSubscribed(app, user, productID)
		if err != nil {
			log.Printf(
				"failed to check app id %v if user id %v is subscribed to product ID %v: %v",
				app.FusionAuth.AppID,
				user.Id,
				productID,
				err.Error(),
			)
			h.Simple500(c)
			return
		}
		c.Data(200, "text/plain", []byte(fmt.Sprintf("%v", subscribed)))
	})
	r.OPTIONS("/mw/private/substatus", func(c *gin.Context) {
		h.Simple200OK(c)
	})
	r.POST("/mw/private/substatus", func(c *gin.Context) {
		// enables other api's to check if a user is subscribed
		sBody := models.SubscriptionStatusCheckBody{} // "value" will hold the product id
		err := c.Bind(&sBody)
		if err != nil {
			h.Simple404(c)
			return
		}

		if sBody.APIKey == "" {
			c.Data(401, "text/plain", []byte("unauthorized"))
			return
		}

		for _, app := range conf.Apps {
			if sBody.APIKey == app.APIKey {
				user := fusionauth.User{}
				// if the jwt isn't specified, attempt to retrieve the user via the other params
				if sBody.JWT == "" {
					if sBody.UserID == "" {
						c.Data(400, "text/plain", []byte("not all required fields were specified for substatus"))
						return
					}
					// TODO: properly handler the "errors" return value
					qUser, _, err := app.FusionAuth.Client.RetrieveUser(sBody.UserID)
					if err != nil {
						log.Printf("failed to find user for substatus: %v", err.Error())
						c.Data(400, "text/plain", []byte("failed to find user"))
						return
					}
					if qUser.User.Id != sBody.UserID {
						c.Data(400, "text/plain", []byte("failed to find user"))
						return
					}
					// if errs.Present() {
					// 	log.Printf("errs finding user for substatus: %v", errs)
					// 	c.Data(400, "text/plain", []byte("failed to find user"))
					// 	return
					// }
					user = qUser.User
				}

				if user.Id == "" {
					qUser, err := auth.GetUserByJWT(app, sBody.JWT)
					if err != nil {
						c.Data(400, "text/plain", []byte("jwt doesn't correspond to any user"))
						return
					}
					user = qUser
				}

				// check if the user is subscribed now
				result, err := payments.IsUserSubscribed(app, user, sBody.ProductID)
				if err != nil {
					log.Printf(
						"failed to check if user is subscribed to product %v: %v",
						sBody.ProductID,
						err.Error(),
					)
					c.Data(400, "text/plain", []byte("failed to check if user is subscribed"))
					return
				}
				c.Data(200, "text/plain", []byte(fmt.Sprintf("%v", result)))
				return
			}
		}
		c.Data(401, "text/plain", []byte("unauthorized"))
	})
	r.OPTIONS("/mw/login", func(c *gin.Context) {
		_, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}
		h.Simple200OK(c)
	})
	r.POST("/mw/login", func(c *gin.Context) {
		app, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}
		// check if the user is already logged in
		jwt := routes.GetJWTFromGin(c, app)
		if jwt != "" {
			log.Printf("user is already logged in")
			user, err := auth.GetUserByJWT(app, jwt)
			if err != nil {
				log.Printf("user is already logged in, but failed to get user: %v", err.Error())
				return
			}

			if user.Id != "" {
				c.Data(200, "text/plain", []byte("already logged in"))
				return
			}
		}
		// user is not logged in, so redirect
		routes.Login(c, app)
	})
	r.OPTIONS("/mw/register", func(c *gin.Context) {
		_, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}
		h.Simple200OK(c)
	})
	r.POST("/mw/register", func(c *gin.Context) {
		app, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}
		// check if the user is already logged in
		jwt := routes.GetJWTFromGin(c, app)
		if jwt != "" {
			log.Printf("user is already logged in")
			user, err := auth.GetUserByJWT(app, jwt)
			if err != nil {
				log.Printf("user is already logged in, but failed to get user: %v", err.Error())
				return
			}

			if user.Id != "" {
				c.Data(200, "text/plain", []byte("already logged in"))
				return
			}
		}
		routes.Register(c, app)
	})
	r.OPTIONS("/mw/loggedin", func(c *gin.Context) {
		_, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}
		h.Simple200OK(c)
	})
	r.GET("/mw/loggedin", func(c *gin.Context) {
		app, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}
		routes.LoggedIn(c, app, app.FusionAuth.Client)
	})
	r.OPTIONS("/mw/products", func(c *gin.Context) {
		_, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}
		h.Simple200OK(c)
	})
	r.GET("/mw/products", func(c *gin.Context) {
		app, ok := routes.GetConfigViaRouteOrigin(c, conf)
		if !ok {
			h.Simple404(c)
			return
		}
		products, err := payments.GetProducts(app)
		if err != nil {
			log.Printf("/mw/products failure: %v", err.Error())
			h.Simple500(c)
			return
		}
		c.JSON(200, products)
	})
	err = r.Run(
		fmt.Sprintf(
			"%v:%v",
			conf.Global.BindAddr,
			conf.Global.BindPort,
		),
	)
	if err != nil {
		log.Fatalf("error running gin: %v", err.Error())
	}
}
