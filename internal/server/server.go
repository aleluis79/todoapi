package server

import (
	"context"
	"log/slog"
	"slices"

	"net/http"
	"os"
	"os/signal"

	"syscall"
	"time"

	"example.com/todoapi/internal/api"
	"example.com/todoapi/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	router *gin.Engine
	config config.Config
}

// NOT traces for:
var notToLogEndpoints = []string{"/api/ping", "/metrics"}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}

func New(config config.Config) *Server {

	s := Server{config: config}

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	s.router = gin.New()

	// Add trazas
	s.router.Use(
		otelgin.Middleware("simp.alias", otelgin.WithFilter(filterTraces)),
	)

	s.router.Use(gin.Recovery())

	if config.Environment == "local" {
		c := cors.DefaultConfig()
		c.AllowOrigins = []string{"http://localhost:4200"}
		c.AllowHeaders = []string{"Origin", "Authorization", "X-API-KEY"}
		s.router.Use(cors.New(c))
	}

	return &s
}

// PingExample   godoc
// @Summary      Ping
// @Description  Healt Check
// @Tags         ping
// @Accept       json
// @Produce      plain
// @Success      200  {string} string  "Ok - alias api versión: x.y.z"
// @Failure      500  {string} string  "Server Error"
// @Router       /api/ping [get]
// @Security 	 OAuth2AccessCode
func (s *Server) AddHealtCheck() {
	s.router.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Ok - versión: 0.1.0"})
	})
}

func (s *Server) AddSwagger() {
	s.router.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (s *Server) AddRoutes() {

	api := api.NewTodosAPI(s.config)

	r := s.router.Group("/api/todos")
	{
		r.GET("", api.Get)
	}
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) Run() {

	srv := &http.Server{
		Addr:    s.config.ServerAddress,
		Handler: s.router,
	}

	slog.Info("Levantando API server en http://" + s.config.ServerAddress + "/api/swagger/index.html")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// iniciando Http server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Error en " + err.Error())
			os.Exit(1)
		}
	}()

	// Esperando interrupción
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	slog.Info("Cerrando servidor...")

	// 5 segundos para termminar la sesiones activas
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forzando cierre del servidor: " + err.Error())
		os.Exit(1)
	}

	slog.Info("Servidor finalizado.")
}
