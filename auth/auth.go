package auth

// https://fusionauth.io/docs/v1/tech/client-libraries/go/

import (
	"fa-middleware/config"
	"fmt"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
)

func GetUserByJWT(conf config.App, jwt string) (user fusionauth.User, err error) {
	userResp, errs, err := conf.FusionAuth.Client.RetrieveUserUsingJWT(jwt)
	// userResp, errs, err := fa.RetrieveUserInfoFromAccessToken(token.AccessToken)
	if err != nil {
		return user, fmt.Errorf(
			"failed to retrieve user by token: %v",
			err.Error(),
		)
	}

	if errs != nil {
		if errs.Present() {
			return user, fmt.Errorf(
				"failed to retrieve user by token due to errors: %v",
				errs,
			)
		}

		return user, fmt.Errorf(
			"no errors were present while trying to retrieve user by token: %v",
			errs.Error(),
		)
	}

	return userResp.User, nil
}

func SetUserData(conf config.App, user fusionauth.User, key string, value interface{}) error {
	if user.Data == nil {
		user.Data = make(map[string]interface{})
	}

	user.Data[key] = value

	_, errs, err := conf.FusionAuth.Client.UpdateUser(
		user.Id,
		fusionauth.UserRequest{
			User: user,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %v", err.Error())
	}
	if errs != nil {
		return fmt.Errorf(
			"failed to update user due to errors: %v",
			errs.Error(),
		)
	}
	return nil
}
