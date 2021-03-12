package auth

// https://fusionauth.io/docs/v1/tech/client-libraries/go/

import (
	"fa-middleware/config"
	"fa-middleware/models"
	"fmt"

	"log"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
)

// Login logs in the user using the FusionAuth Go client library
func Login(conf config.Config, fa *fusionauth.FusionAuthClient, oauthState models.OauthState) (user fusionauth.User, err error) {
	token, oauthError, err := fa.ExchangeOAuthCodeForAccessTokenUsingPKCE(
		oauthState.Code,
		conf.FusionAuthOauthClientID,
		conf.FusionAuthOauthClientSecret,
		conf.FusionAuthOauthRedirectURL,
		oauthState.Verifier,
	)

	if err != nil {
		return user, fmt.Errorf(
			"failed to exchange oauth code for access token via pkce: %v",
			err.Error(),
		)
	}

	if oauthError != nil {
		return user, fmt.Errorf(
			"failed to exchange oauth code for access token via pkce oauth error: %v",
			oauthError,
		)
	}

	userResp, errs, err := fa.RetrieveUserUsingJWT(token.AccessToken)
	// userResp, errs, err := fa.RetrieveUserInfoFromAccessToken(token.AccessToken)
	if err != nil {
		log.Fatalf("failed to retrieve user by token: %v", err.Error())
	}
	if errs != nil {
		log.Fatalf(
			"failed to retrieve user by token due to errors: %v",
			errs,
		)
	}
	log.Printf("%v", userResp.User.Data)

	return userResp.User, nil
}
