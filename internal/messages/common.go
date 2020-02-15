package messages

import (
	"time"

	"project/internal/logger"
)

// Log is the Node or Beacon log.
type Log struct {
	Time   time.Time
	Level  logger.Level
	Source string

	// reduce one copy about plain text log
	Log []byte
}