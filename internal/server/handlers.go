package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type StartStreamRequest struct {
	IndicatorID string `json:"indicator_id" form:"indicator_id" binding:"required"`
	RTSPLink    string `json:"rtsp_link" form:"rtsp_link" binding:"required"`
}

type StopStreamRequest struct {
	IndicatorID string `json:"indicator_id" binding:"required"`
}

type HeartbeatRequest struct {
	IndicatorID string `json:"indicator_id" form:"indicator_id" binding:"required"`
}

func (s *Server) startStream(c *gin.Context) {
	var req StartStreamRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	streamPath, err := s.manager.StartStream(req.IndicatorID, req.RTSPLink)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stream_path": streamPath,
		"status":      "started",
	})
}

func (s *Server) stopStream(c *gin.Context) {
	var req StopStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.manager.StopStream(req.IndicatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "stopped",
	})
}

func (s *Server) getStreamStatus(c *gin.Context) {
	indicatorID := c.Query("indicator_id")
	if indicatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "indicator_id is required"})
		return
	}

	status := s.manager.GetStreamStatus(indicatorID)
	c.JSON(http.StatusOK, gin.H{
		"indicator_id": indicatorID,
		"status":       status,
	})
}

func (s *Server) heartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.manager.UpdateStreamAccess(req.IndicatorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "keepalive_updated",
		"indicator_id": req.IndicatorID,
	})
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"active_streams": s.manager.GetActiveStreamCount(),
	})
}