package auth

import (
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/hinccvi/Golang-Project-Structure-Conventional/internal/errors"

	"github.com/gin-gonic/gin"
)

func Handler(verificationKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(200, errors.Unauthorized("empty token"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.AbortWithStatusJSON(200, errors.Unauthorized("wrong token format"))
			return
		}

		mc, err := decodeToken(verificationKey, parts[1])
		if err != nil {
			c.AbortWithStatusJSON(200, errors.Unauthorized("invalid token"))
			return
		}

		userId := mc.(MyCustomClaims).Data.UserId

		c.Set("UserId", userId)
		c.Next()
	}
}

func decodeToken(verificationKey, tokenString string) (jwt.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return []byte(verificationKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		return *claims, nil
	}
	return nil, errors.Unauthorized("invalid token")
}
