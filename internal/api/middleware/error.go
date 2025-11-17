package middleware

import (
	"context"
	"errors"
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
