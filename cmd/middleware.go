package cmd

import (
	"ewallet-transaction/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (d *Dependency) MiddlewareValidateToken(c *gin.Context) {
	var (
		log = helpers.Logger
	)
	auth := c.Request.Header.Get("authorization")
	if auth == "" {
		log.Println("authorization empty")
		helpers.SendResponseHTTP(c, http.StatusUnauthorized, "Unauthorized", nil)
		c.Abort()
		return
	}

	tokenData, err := d.External.ValidateToken(c, auth)
	if err != nil {
		log.Error(err)
		helpers.SendResponseHTTP(c, http.StatusUnauthorized, "Unauthorized", nil)
		c.Abort()
		return
	}

	tokenData.Token = auth

	c.Set("token", tokenData)

	c.Next()
}
