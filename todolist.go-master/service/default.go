package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
)

// Home renders index.html
func Home(ctx *gin.Context) {
	login := true
	if sessions.Default(ctx).Get(userkey) == nil{
		login = false
	}
	ctx.HTML(http.StatusOK, "index.html", gin.H{"Title": "HOME","Login":login})
}

// NotImplemented renders error.html with 501 Not Implemented
func NotImplemented(ctx *gin.Context) {
	msg := fmt.Sprintf("%s access to %s is not implemented yet", ctx.Request.Method, ctx.Request.URL)
	ctx.Header("Cache-Contrl", "no-cache")
	Error(http.StatusNotImplemented, msg)(ctx)
}

// Error returns a handler which renders error.html
func Error(code int, message string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(code, "error.html", gin.H{"Code": code, "Error": message})
	}
}
