package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Cors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method

		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type,x-token")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
		}
	}
}
