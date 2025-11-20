package middleware

import (
	"context"
	"financial-data-backend-2/internal/api/dto"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// The handler finished in time. We can just return. The handler
			// or a subsequent middleware has already written the response.
			return
		case <-ctx.Done():
			// The context's deadline was exceeded. The timeout was hit.
			// THIS MIDDLEWARE MUST WRITE THE RESPONSE.
			// We use c.Abort() to prevent any subsequent handlers from writing.
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, dto.Res{
				Success: false,
				Error:   "request timed out",
			})
		}
	}
}
