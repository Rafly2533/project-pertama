package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"intan-florist-api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestAuthAcceptsValidToken(t *testing.T) {
	secret := "01234567890123456789012345678901"
	token, err := utils.GenerateToken(secret, time.Hour, 3, "staff", "staff")
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.GET("/", Auth(secret), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("got %d", response.Code)
	}
}

func TestStaffAuthorization(t *testing.T) {
	tests := []struct {
		method, resource string
		want             int
	}{{http.MethodGet, "products", 204}, {http.MethodPost, "products", 204}, {http.MethodPut, "categories", 204}, {http.MethodDelete, "products", 403}, {http.MethodGet, "users", 403}, {http.MethodPost, "settings", 403}}
	for _, test := range tests {
		t.Run(test.method+test.resource, func(t *testing.T) {
			r := gin.New()
			r.Handle(test.method, "/", func(c *gin.Context) { c.Set("role", "staff") }, Authorize(test.resource), func(c *gin.Context) { c.Status(http.StatusNoContent) })
			request := httptest.NewRequest(test.method, "/", nil)
			response := httptest.NewRecorder()
			r.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("got %d want %d", response.Code, test.want)
			}
		})
	}
}

func TestCORS(t *testing.T) {
	r := gin.New()
	r.Use(CORS([]string{"http://localhost:3000"}))
	r.GET("/", func(c *gin.Context) { c.Status(204) })
	request := httptest.NewRequest(http.MethodOptions, "/", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("unexpected CORS response: %d %q", response.Code, response.Header().Get("Access-Control-Allow-Origin"))
	}
}
