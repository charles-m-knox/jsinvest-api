package config

import (
	"fmt"

	viper "github.com/spf13/viper"
)

const (
	ConfigFile = "config.yml"
)

type Config struct {
	FusionAuthHost              string `yaml:"fusionAuthHost"`       // "http://fusionauth:9011"
	FusionAuthPublicHost        string `yaml:"fusionAuthPublicHost"` // "http://localhost:9011"
	FusionAuthAPIKey            string `yaml:"fusionAuthAPIKey"`
	FusionAuthAppID             string `yaml:"fusionAuthAppID"`
	FusionAuthOauthRedirectURL  string `yaml:"fusionAuthOauthRedirectURL"`
	FusionAuthOauthClientID     string `yaml:"fusionAuthOauthClientID"`
	FusionAuthOauthClientSecret string `yaml:"fusionAuthOauthClientSecret"`
	FusionAuthTenantID          string `yaml:"fusionAuthTenantID"`
	BindAddr                    string `yaml:"bindAddr"`
	BindPort                    int    `yaml:"bindPort"`
	BindPortExternal            int    `yaml:"bindPortExternal"`
	PostgresUser                string `yaml:"postgresUser"`
	PostgresPass                string `yaml:"postgresPass"`
	PostgresHost                string `yaml:"postgresHost"`
	PostgresPort                string `yaml:"postgresPort"`
	PostgresDBName              string `yaml:"postgresDbName"`
	PostgresOptions             string `yaml:"postgresOptions"`
	JWTCookieDomain             string `yaml:"jwtCookieDomain"`
	JWTCookieMaxAgeSeconds      int    `yaml:"jwtCookieMaxAgeSeconds"`
	JWTCookieSetSecure          bool   `yaml:"jwtCookieSetSecure"` // sets the "secure" flag for the jwt cookie
	JWTCookieName               string `yaml:"jwtCookieName"`      // sets the "secure" flag for the jwt cookie
}

// LoadConfig reads from a provided yaml-formatted configuration filename
func LoadConfig() (conf Config, err error) {
	err = viper.BindEnv("config")
	if err != nil {
		return conf, fmt.Errorf("failed to bind config env: %v", err.Error())
	}
	configName := viper.GetString("config")
	viper.SetConfigName(configName) // name of config file (without extension)
	viper.SetConfigType("yaml")     // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("./res")    // optionally look for config in the working directory
	viper.AddConfigPath(".")        // optionally look for config in the working directory
	err = viper.ReadInConfig()      // Find and read the config file
	if err != nil {                 // Handle errors reading the config file
		return conf, fmt.Errorf("error config file: %s", err)
	}
	err = viper.Unmarshal(&conf)
	if err != nil {
		return conf, fmt.Errorf("unable to decode into struct, %v", err)
	}

	return conf, nil
}
