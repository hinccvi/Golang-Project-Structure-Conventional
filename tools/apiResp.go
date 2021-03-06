package tools

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type message string

const (
	Success message = "success"
	Created message = "created"
	Updated message = "updated"
	Deleted message = "deleted"
)

type response struct {
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func RespOkWithData[I any](c *gin.Context, msg message, i I) {
	c.JSON(http.StatusOK, response{
		Message: string(msg),
		Data:    i,
	})
}

func RespOk(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, response{
		Message: msg,
	})
}
