package middleware

import (
	"context"
	"errors"
	"financial-data-backend-2/internal/api/constant"
	"financial-data-backend-2/internal/api/dto"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func Error() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		ctxErr := c.Request.Context().Err()
		if ctxErr != nil {
			// Check if the context error is specifically a deadline exceeded.
			if errors.Is(ctxErr, context.DeadlineExceeded) {
				// If so, abort and write the Gateway Timeout response.
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, dto.Res{
					Success: false,
					Error:   "request timed out",
				})
				return
			}
		}

		// Check if there is no error
		if len(c.Errors) == 0 {
			return
		}

		// There is error; what error is it?
		err := c.Errors[0]

		// - Validation error from requests' JSON binding
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			validationErrors := make([]dto.ErrorType, 0)
			for _, fe := range ve {
				validationErrors = append(validationErrors, dto.ErrorType{
					Field:   fe.Field(),
					Message: fe.Error(),
				})
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.Res{
				Success: false,
				Error:   validationErrors,
			})
			return
		}

		// - Custom error from `constant` repo
		var ce constant.CustomError
		if errors.As(err, &ce) {
			c.AbortWithStatusJSON(ce.StatusCode, dto.Res{
				Success: false,
				Error:   ce.Error(),
			})
			return
		}

		// - Unknown error, likely internal server error
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.Res{
			Success: false,
			Error:   err.Error(),
		})
	}
}
