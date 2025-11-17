package middleware

import (
	"context"
	"errors"
	"financial-data-backend-2/internal/api/constant"
	"financial-data-backend-2/internal/api/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Error() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there is no error
		if len(c.Errors) == 0 {
			return
		}

		// There is error; what error is it?
		err := c.Errors[0]

		// - Custom error from `constant` repo
		var ce constant.CustomError
		if errors.As(err, &ce) {
			c.AbortWithStatusJSON(ce.StatusCode, dto.Res{
				Success: false,
				Error:   ce.Error(),
			})
			return
		}

		// - Timeout error
		if errors.Is(err, context.DeadlineExceeded) {
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, dto.Res{
				Success: false,
				Error:   err.Error(),
			})
		}

		// - Unknown error, likely internal server error
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.Res{
			Success: false,
			Error:   err.Error(),
		})
	}
}
