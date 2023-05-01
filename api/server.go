package api

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
)

type ServerConfig struct {
	Logger     log.Logger
	ListenAddr string
}

type Server struct {
	ServerConfig
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{
		ServerConfig: cfg,
	}
}

func (s *Server) Start() error {
	e := echo.New()

	e.GET("/block/:hashorid", s.handleGetBlock)

	return e.Start(s.ListenAddr)
}

func (s *Server) handleGetBlock(c echo.Context) error {
	hashOrID := c.Param("hashorid")

	return c.JSON(http.StatusOK, map[string]any{"mdg": "server working"})
}
