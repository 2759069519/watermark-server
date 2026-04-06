package modules

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

type BaseController struct{}

func jsonNoEscape(c *gin.Context, code int, obj interface{}) {
	c.Status(code)
	c.Header("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(c.Writer)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(obj)
}

func (con *BaseController) Success(c *gin.Context, data interface{}, msg string) {
	jsonNoEscape(c, http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": data,
		"msg":  msg,
	})
}

func (con *BaseController) Err(c *gin.Context, msg string) {
	jsonNoEscape(c, http.StatusOK, gin.H{
		"code": http.StatusBadRequest,
		"msg":  msg,
		"data": nil,
	})
}

func (con *BaseController) Unauthorized(c *gin.Context, msg string) {
	jsonNoEscape(c, http.StatusOK, gin.H{
		"code": http.StatusUnauthorized,
		"msg":  msg,
		"data": nil,
	})
}

func (con *BaseController) Failed(c *gin.Context, msg string) {
	jsonNoEscape(c, http.StatusInternalServerError, gin.H{
		"code": http.StatusUnauthorized,
		"msg":  msg,
		"data": nil,
	})
}
