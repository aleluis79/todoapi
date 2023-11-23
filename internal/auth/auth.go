package auth

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("todos")

func Verify(c *gin.Context) bool {
	_, span := tracer.Start(c.Request.Context(), "auth")
	defer span.End()

	isValid := false
	apikey := c.Request.Header.Get("X-API-KEY")
	if apikey == "123456" {
		isValid = true
		span.SetAttributes(attribute.String("auth.valid", "OK"))
		slog.Info("Auth Ok - Url: " + c.Request.Method + " " + c.Request.URL.Path)
	} else {
		isValid = false
		span.SetStatus(codes.Error, "authentication failed")
		slog.Error("Auth Unauthorized - Url: " + c.Request.Method + " " + c.Request.URL.Path + " - apikey: " + apikey)
		c.String(http.StatusUnauthorized, "Unauthorized")
	}
	return isValid
}
