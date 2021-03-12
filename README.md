# fa-middleware

This repository contains the building blocks for a FusionAuth-based middleware layer, written in Go.

## Getting started

For the first time:

1. Run `make run-fa`
2. Set up FusionAuth by visiting `http://localhost:9011`
3. Set up the steps in https://fusionauth.io/docs/v1/tech/5-minute-setup-guide/
  1. Make sure to set up your application to not require API keys on login:
    1. Add application -> Security -> Require an API Key (off)
  2. Set up your application to allow self-service registration:
    1. Add application -> Registration
  3. Create an api key:
    1. Left sidebar -> Settings -> API Keys
    2. Click GET/POST/PUT/PATCH/DELETE to enable the API key to have global access
4. Set the bind ports and other options in `.env` and in `./res/config.yml` based on their `.example` variants.
5. Set the secrets from the application in `./res/config.yml` accordingly
6. Run `make run-middleware` to run the middleware now that it's been properly configured
7. Visit `http://localhost:8080/login` in your browser, or whatever the port/bind address is for your setup
8. Follow the login/registration workflow, brokered by FusionAuth
9. Observe that the callback URL has been hit after logging in

These are the basic steps to get this middleware up and running. Up next is to get frontend interactions working with this layer.

## References

* https://github.com/FusionAuth/go-client
* https://fusionauth.io/blog/2020/03/10/securely-implement-oauth-in-react/
* https://fusionauth.io/docs/v1/tech/oauth/endpoints/
* https://fusionauth.io/docs/v1/tech/example-apps/go/
* https://fusionauth.io/blog/2020/06/17/building-cli-app-with-device-grant-and-golang/

## Fusion Auth Example

### Login directly

https://fusionauth.io/docs/v1/tech/apis/login/#authenticate-a-user

```bash
curl -vvv -X POST -H "Content-Type: application/json" -d '{"loginId":"test2@site.com","password":"password1234","applicationId":"4ef26681-80e6-4c6e-895f-3ef45321a2cd","noJWT":false,"ipAddress":"192.168.0.10"}' http://localhost:9011/api/login
```
