package middleware

import (
	"errors"
	"financial-data-backend-2/internal/api/constant"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert"
	"github.com/go-playground/validator/v10"
)

func TestMiddlewareError(t *testing.T) {
	testCases := []struct {
		name           string
		handle         func(c *gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "no error",
			handle: func(c *gin.Context) {
			},
			expectedStatus: http.StatusOK,
			expectedBody:   ``,
		},
		{
			name: "validation errors - empty",
			handle: func(c *gin.Context) {
				c.Error(validator.ValidationErrors{})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":[],"data":null}`,
		},
		{
			name: "validation errors - non-empty 1",
			handle: func(c *gin.Context) {
				type request struct {
					Field string `json:"field" binding:"required"`
				}

				var r request
				errorArg := c.ShouldBindQuery(&r)

				if errorArg != nil {
					c.Error(errorArg)
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{"success":false,` +
				`"error":[{"field":"Field",` +
				`"message":"` +
				`Key: 'request.Field' Error:Field validation for 'Field' failed on the 'required' tag` +
				`"}],` +
				`"data":null}`,
		},
		{
			name: "validation errors - non-empty 2",
			handle: func(c *gin.Context) {
				type request struct {
					Field string `json:"field" binding:"required"`
					Var   string `json:"variable" binding:"required"`
				}

				var r request
				errorArg := c.ShouldBindQuery(&r)

				if errorArg != nil {
					c.Error(errorArg)
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{"success":false,` +
				`"error":[{"field":"Field",` +
				`"message":"` +
				`Key: 'request.Field' Error:Field validation for 'Field' failed on the 'required' tag` +
				`"},{"field":"Var",` +
				`"message":"` +
				`Key: 'request.Var' Error:Field validation for 'Var' failed on the 'required' tag` +
				`"}],` +
				`"data":null}`,
		},
		{
			name: "custom error 1",
			handle: func(c *gin.Context) {
				c.Error(constant.CustomError{
					StatusCode: http.StatusBadRequest,
					Message:    "custom error message",
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{"success":false,` +
				`"error":"custom error message","data":null}`,
		},
		{
			name: "custom error 2",
			handle: func(c *gin.Context) {
				c.Error(constant.CustomError{
					StatusCode: http.StatusBadGateway,
					Message:    "custom error message",
				})
			},
			expectedStatus: http.StatusBadGateway,
			expectedBody: `{"success":false,` +
				`"error":"custom error message","data":null}`,
		},
		{
			name: "interval server error",
			handle: func(c *gin.Context) {
				c.Error(errors.New("unknown error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{"success":false,` +
				`"error":"unknown error","data":null}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			//given
			recorder := httptest.NewRecorder()
			_, engine := gin.CreateTestContext(recorder)

			engine.GET("/", Error(), tt.handle)
			r := httptest.NewRequest("", "/", nil)

			//when
			engine.ServeHTTP(recorder, r)

			//then
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			assert.Equal(t, tt.expectedBody, recorder.Body.String())
		})
	}
}
func TestTimeoutMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	r.Use(Error())
	r.Use(Timeout(50 * time.Millisecond))

	// Create a handler that is deliberately slower than the timeout.
	r.GET("/slow", func(c *gin.Context) {
		// This handler will sleep for 100ms. The timeout is 50ms.
		time.Sleep(100 * time.Millisecond)

		// The timeout will fire while this handler is sleeping.
		// After the sleep, we check if the context was cancelled.
		// If it was, the middleware chain has already been aborted, and
		// this response will never be written.
		if c.Request.Context().Err() != nil {
			return
		}

		// This should never be reached in a successful test.
		c.JSON(http.StatusOK,
			gin.H{
				"message": "OK",
				"error":   nil,
				"data":    nil,
			})
	})

	// Create a fake HTTP request and a response recorder.
	req, _ := http.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()

	// Act: Serve the HTTP request.
	r.ServeHTTP(w, req)

	// Assert: The Error middleware should have detected the timeout and written the correct response.
	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Equal(t, `{"success":false,"error":"request timed out","data":null}`, w.Body.String())
}
