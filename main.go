package main

import (
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"strconv"
)

func main() {
	err := loadConfig()
	if err != nil {
		panic(err)
	}
	cert, err := os.ReadFile(_config.Certificate)
	if err != nil {
		panic(err)
	}
	casdoorsdk.InitConfig(_config.Endpoint, _config.ClientId, _config.ClientSecret,
		string(cert), _config.OrganizationName, _config.ApplicationName)
	en := gin.New()
	en.GET("auth/casdoor", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")
		token, err := casdoorsdk.GetOAuthToken(code, state)
		if err != nil {
			c.JSON(http.StatusForbidden, err.Error())
			return
		}
		c.SetCookie("access-token", token.AccessToken, 0, "/",
			"", false, false)
		c.Redirect(http.StatusMovedPermanently, "/")
	})
	en.Use(authMiddle)
	en.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, c.MustGet("user"))
	})
	en.GET("user", func(c *gin.Context) {
		c.JSON(http.StatusOK, c.MustGet("user"))
	})
	err = en.Run(":" + strconv.Itoa(_config.Port))
	if err != nil {
		panic(err)
	}
}

func authMiddle(c *gin.Context) {
	cookie, _ := c.Cookie("access-token")
	if cookie == "" {
		c.Abort()
		c.Redirect(http.StatusMovedPermanently, _config.RedirectUri)
		return
	}
	cla, err := casdoorsdk.ParseJwtToken(cookie)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, err.Error())
		return
	}
	c.Set("user", cla)
}

type Config struct {
	Port             int
	Endpoint         string
	ClientId         string
	ClientSecret     string
	Certificate      string
	OrganizationName string
	ApplicationName  string
	RedirectUri      string
}

var _config Config

func loadConfig() error {
	viper.SetConfigType("yml")
	viper.SetConfigName("app")
	viper.AddConfigPath("./")
	viper.AddConfigPath("conf")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	_config.Port = viper.GetInt("server.port")
	_config.Endpoint = viper.GetString("casdoor.endpoint")
	_config.ClientId = viper.GetString("casdoor.clientId")
	_config.ClientSecret = viper.GetString("casdoor.clientSecret")
	_config.Certificate = viper.GetString("casdoor.certificate")
	_config.OrganizationName = viper.GetString("casdoor.organizationName")
	_config.ApplicationName = viper.GetString("casdoor.applicationName")
	_config.RedirectUri = viper.GetString("casdoor.redirectUri")
	return nil
}
