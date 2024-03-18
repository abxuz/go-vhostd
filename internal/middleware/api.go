package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ApiResponse() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		resp, ok := ctx.Get("resp")
		if !ok {
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}
