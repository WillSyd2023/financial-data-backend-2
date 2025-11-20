package middleware

import (
	"errors"
	"financial-data-backend-2/internal/api/constant"
	"net/http"
	"net/http/httptest"
	"testing"

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
