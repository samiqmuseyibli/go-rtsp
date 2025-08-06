package server

import (
	"go-rtsp-streamer/internal/config"
	"go-rtsp-streamer/internal/stream"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config  *config.Config
	manager *stream.Manager
	router  *gin.Engine
}

func New() *Server {
	cfg := config.New()
	manager := stream.NewManager(cfg)

	s := &Server{
		config:  cfg,
		manager: manager,
		router:  gin.Default(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Enable CORS for all routes
	s.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Serve static HLS files
	s.router.Static("/streams", s.config.StreamDir)
	
	// API routes
	api := s.router.Group("/api")
	{
		api.GET("/stream/start", s.startStream)
		api.POST("/stream/stop", s.stopStream)
		api.GET("/stream/status", s.getStreamStatus)
		api.GET("/stream/heartbeat", s.heartbeat)
		api.GET("/health", s.healthCheck)
	}
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}