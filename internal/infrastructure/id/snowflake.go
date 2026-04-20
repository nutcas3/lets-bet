package id

import (
	"fmt"
	"sync"
	"time"
)

// SnowflakeIDGenerator implements a Twitter Snowflake-like ID generator
// Format: YYYYmmddHhmmss + WorkerID + Sequence
// Example: 202604200358509988001
// Where: 20260420035850 = timestamp
//        9988 = worker ID (user-specific)
//        001 = sequence number (resets every second)

const (
	// WorkerIDBits = 16 bits (0-65535)
	// SequenceBits = 3 bits (0-999, allows 1000 operations per second per worker)
	workerIDShift  = 3
	sequenceMask   = 0x3FF // 1023 sequences per second
	maxWorkerID    = 0xFFFF // 65535 max workers
)

// SnowflakeGenerator generates time-based deterministic IDs
type SnowflakeGenerator struct {
	mu           sync.Mutex
	workerID     int
	lastTime     int64
	sequence     int
	timeFormat   string
}

// NewSnowflakeGenerator creates a new Snowflake ID generator
func NewSnowflakeGenerator(workerID int) (*SnowflakeGenerator, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, fmt.Errorf("worker ID must be between 0 and %d", maxWorkerID)
	}
	
	return &SnowflakeGenerator{
		workerID:   workerID,
		lastTime:   0,
		sequence:   0,
		timeFormat: "20060102150405", // YYYYmmddHhmmss
	}, nil
}

// GenerateID generates a new time-based deterministic ID
func (s *SnowflakeGenerator) GenerateID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	timestamp := now.Unix()
	
	// Handle clock moving backwards
	if timestamp < s.lastTime {
		panic(fmt.Sprintf("clock moved backwards. Refusing to generate id for %d milliseconds", s.lastTime-timestamp))
	}
	
	// If we're still in the same second, increment sequence
	if timestamp == s.lastTime {
		s.sequence = (s.sequence + 1) & sequenceMask
		// If sequence overflows, wait until next second
		if s.sequence == 0 {
			timestamp = s.waitNextSecond(s.lastTime)
		}
	} else {
		// Reset sequence for new second
		s.sequence = 0
	}
	
	s.lastTime = timestamp
	
	// Format timestamp as YYYYmmddHhmmss
	timeStr := now.Format(s.timeFormat)
	
	// Combine: timestamp + workerID + sequence
	workerStr := fmt.Sprintf("%04d", s.workerID)
	sequenceStr := fmt.Sprintf("%03d", s.sequence)
	
	return timeStr + workerStr + sequenceStr
}

// waitNextSecond waits until the next second
func (s *SnowflakeGenerator) waitNextSecond(lastTime int64) int64 {
	timestamp := time.Now().Unix()
	for timestamp <= lastTime {
		timestamp = time.Now().Unix()
	}
	return timestamp
}

// ParseSnowflakeID parses a Snowflake ID back into its components
func ParseSnowflakeID(id string) (timestamp, workerID, sequence string, err error) {
	if len(id) != 19 { // 14 (timestamp) + 4 (worker) + 3 (sequence)
		return "", "", "", fmt.Errorf("invalid Snowflake ID length: expected 19, got %d", len(id))
	}
	
	timestamp = id[:14]
	workerID = id[14:18]
	sequence = id[18:20]
	
	return timestamp, workerID, sequence, nil
}

// GetTimestampFromID extracts the timestamp from a Snowflake ID
func GetTimestampFromID(id string) (time.Time, error) {
	timestamp, _, _, err := ParseSnowflakeID(id)
	if err != nil {
		return time.Time{}, err
	}
	
	return time.Parse("20060102150405", timestamp)
}

// UserSpecificGenerator creates a generator for a specific user
func UserSpecificGenerator(userID string) (*SnowflakeGenerator, error) {
	// Convert user ID to numeric worker ID (simple hash)
	workerID := 0
	for _, char := range userID {
		workerID = (workerID*31 + int(char)) % (maxWorkerID + 1)
	}
	
	return NewSnowflakeGenerator(workerID)
}

// CountrySpecificGenerator creates a generator for a specific country
func CountrySpecificGenerator(countryCode string) (*SnowflakeGenerator, error) {
	// Map country codes to specific worker ID ranges
	countryMap := map[string]int{
		"KE": 1000, // Kenya
		"NG": 2000, // Nigeria  
		"GH": 3000, // Ghana
	}
	
	workerID, exists := countryMap[countryCode]
	if !exists {
		workerID = 9999 // Default/unknown
	}
	
	return NewSnowflakeGenerator(workerID)
}

// ServiceTypeGenerator creates a generator for specific service types
func ServiceTypeGenerator(serviceType string) (*SnowflakeGenerator, error) {
	// Map service types to specific worker ID ranges
	serviceMap := map[string]int{
		"wallet":    5000,
		"betting":   6000,
		"jackpot":   7000,
		"payment":   8000,
		"compliance": 9000,
	}
	
	workerID, exists := serviceMap[serviceType]
	if !exists {
		workerID = 9998 // Default service
	}
	
	return NewSnowflakeGenerator(workerID)
}
