package stream

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"go-rtsp-streamer/internal/config"
)

type StreamInfo struct {
	IndicatorID string
	RTSPLink    string
	Process     *exec.Cmd
	StartTime   time.Time
	LastAccess  time.Time
	StreamPath  string
	IsActive    bool
}

type Manager struct {
	config  *config.Config
	streams map[string]*StreamInfo
	mutex   sync.RWMutex
}

func NewManager(cfg *config.Config) *Manager {
	m := &Manager{
		config:  cfg,
		streams: make(map[string]*StreamInfo),
	}

	if err := os.MkdirAll(cfg.StreamDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create streams directory: %v", err))
	}

	go m.cleanupWorker()
	return m
}

func (m *Manager) StartStream(indicatorID, rtspLink string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if stream, exists := m.streams[indicatorID]; exists && stream.IsActive {
		stream.LastAccess = time.Now()
		return stream.StreamPath, nil
	}

	streamDir := filepath.Join(m.config.StreamDir, indicatorID)
	if err := os.MkdirAll(streamDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create stream directory: %v", err)
	}

	streamPath := fmt.Sprintf("/streams/%s/stream.m3u8", indicatorID)
	outputPath := filepath.Join(streamDir, "stream.m3u8")

	cmd := exec.Command("ffmpeg",
		"-rtsp_transport", "tcp",
		"-i", rtspLink,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", m.config.HLSSegmentTime),
		"-hls_list_size", fmt.Sprintf("%d", m.config.HLSListSize),
		"-hls_flags", "delete_segments",
		"-y",
		outputPath,
	)

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	streamInfo := &StreamInfo{
		IndicatorID: indicatorID,
		RTSPLink:    rtspLink,
		Process:     cmd,
		StartTime:   time.Now(),
		LastAccess:  time.Now(),
		StreamPath:  streamPath,
		IsActive:    true,
	}

	m.streams[indicatorID] = streamInfo

	go func() {
		if err := cmd.Wait(); err != nil {
			m.mutex.Lock()
			if stream, exists := m.streams[indicatorID]; exists {
				stream.IsActive = false
			}
			m.mutex.Unlock()
		}
	}()

	return streamPath, nil
}

func (m *Manager) StopStream(indicatorID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	stream, exists := m.streams[indicatorID]
	if !exists {
		return fmt.Errorf("stream not found: %s", indicatorID)
	}

	if stream.Process != nil && stream.IsActive {
		if err := stream.Process.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %v", err)
		}
		stream.IsActive = false
	}

	m.cleanupStreamFiles(indicatorID)
	delete(m.streams, indicatorID)

	return nil
}

func (m *Manager) GetStreamStatus(indicatorID string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stream, exists := m.streams[indicatorID]
	if !exists {
		return "not_found"
	}

	if !stream.IsActive {
		return "stopped"
	}

	return "active"
}

func (m *Manager) UpdateStreamAccess(indicatorID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	stream, exists := m.streams[indicatorID]
	if !exists {
		return fmt.Errorf("stream not found: %s", indicatorID)
	}

	if !stream.IsActive {
		return fmt.Errorf("stream is not active: %s", indicatorID)
	}

	oldTime := stream.LastAccess
	stream.LastAccess = time.Now()
	log.Printf("Heartbeat: Updated LastAccess for %s from %v to %v", indicatorID, oldTime.Format("15:04:05"), stream.LastAccess.Format("15:04:05"))
	return nil
}

func (m *Manager) GetActiveStreamCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	count := 0
	for _, stream := range m.streams {
		if stream.IsActive {
			count++
		}
	}
	return count
}

func (m *Manager) cleanupWorker() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()
		now := time.Now()
		log.Printf("Cleanup worker: Checking %d active streams at %v", len(m.streams), now.Format("15:04:05"))

		for indicatorID, stream := range m.streams {
			timeSinceAccess := now.Sub(stream.LastAccess)
			log.Printf("Stream %s: LastAccess=%v, TimeSince=%v, CleanupDelay=%v, IsActive=%v", 
				indicatorID, 
				stream.LastAccess.Format("15:04:05"), 
				timeSinceAccess, 
				m.config.CleanupDelay, 
				stream.IsActive)
				
			if stream.IsActive && timeSinceAccess > m.config.CleanupDelay {
				log.Printf("Cleanup: Stopping stream %s (inactive for %v)", indicatorID, timeSinceAccess)
				if stream.Process != nil {
					stream.Process.Process.Kill()
				}
				stream.IsActive = false
				m.cleanupStreamFiles(indicatorID)
				delete(m.streams, indicatorID)
			}
		}
		m.mutex.Unlock()
	}
}

func (m *Manager) cleanupStreamFiles(indicatorID string) {
	streamDir := filepath.Join(m.config.StreamDir, indicatorID)
	os.RemoveAll(streamDir)
}