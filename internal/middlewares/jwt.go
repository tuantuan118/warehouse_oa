package middlewares

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"warehouse_oa/utils"
)

func JWTAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Request.Header.Get("X-Token")
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, map[string]string{
				"message": "token is empty",
			})
			ctx.Abort()
			return
		}
		j := utils.NewJWT()

		claims, err := j.ParseToken(token)
		if err != nil {
			if errors.Is(err, utils.TokenExpired) {
				ctx.JSON(http.StatusUnauthorized, map[string]string{
					"message": "token is expired",
				})
				ctx.Abort()
				return
			}

			ctx.JSON(http.StatusUnauthorized, map[string]string{
				"message": "token is invalid",
			})
			ctx.Abort()
			return
		}

		logrus.Infoln(claims)
		logrus.Infoln(claims.Id)

		ctx.Set("claims", claims)
		ctx.Set("userId", claims.Id)
		ctx.Set("userName", claims.Name)
		ctx.Next()
	}
}
