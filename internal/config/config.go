package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	StreamDir     string
	CleanupDelay  time.Duration
	HLSSegmentTime int
	HLSListSize   int
}

func New() *Config {
	cleanupDelayStr := getEnv("CLEANUP_DELAY_SECONDS", "30")
	cleanupDelay, err := strconv.Atoi(cleanupDelayStr)
	if err != nil {
		log.Printf("Invalid CLEANUP_DELAY_SECONDS '%s', using default 30", cleanupDelayStr)
		cleanupDelay = 30
	}
	
	segmentTime, _ := strconv.Atoi(getEnv("HLS_SEGMENT_TIME", "2"))
	listSize, _ := strconv.Atoi(getEnv("HLS_LIST_SIZE", "5"))

	config := &Config{
		StreamDir:      getEnv("STREAM_DIR", "./streams"),
		CleanupDelay:   time.Duration(cleanupDelay) * time.Second,
		HLSSegmentTime: segmentTime,
		HLSListSize:    listSize,
	}
	
	log.Printf("Config loaded: StreamDir=%s, CleanupDelay=%v, HLSSegmentTime=%d, HLSListSize=%d", 
		config.StreamDir, config.CleanupDelay, config.HLSSegmentTime, config.HLSListSize)
		
	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}