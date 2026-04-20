package id

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSnowflakeGenerator(t *testing.T) {
	gen, err := NewSnowflakeGenerator(1234)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Test basic generation
	id1 := gen.GenerateID()
	id2 := gen.GenerateID()
	
	if id1 == id2 {
		t.Error("Generated IDs should be different")
	}

	// Test ID format (should be 19 characters: 14 timestamp + 4 worker + 3 sequence)
	if len(id1) != 19 {
		t.Errorf("Expected ID length 19, got %d", len(id1))
	}

	// Test that IDs are chronologically sortable
	id3 := gen.GenerateID()
	if id2 > id3 {
		t.Error("IDs should be chronologically sortable")
	}

	// Test parsing
	timestamp, workerID, sequence, err := ParseSnowflakeID(id1)
	if err != nil {
		t.Fatalf("Failed to parse ID: %v", err)
	}

	if workerID != "1234" {
		t.Errorf("Expected worker ID 1234, got %s", workerID)
	}

	// Test timestamp extraction
	parsedTime, err := GetTimestampFromID(id1)
	if err != nil {
		t.Fatalf("Failed to get timestamp from ID: %v", err)
	}

	if time.Since(parsedTime) > time.Minute {
		t.Error("Parsed timestamp should be recent")
	}

	fmt.Printf("Generated ID: %s\n", id1)
	fmt.Printf("Timestamp: %s, Worker: %s, Sequence: %s\n", timestamp, workerID, sequence)
	fmt.Printf("Parsed time: %v\n", parsedTime)
}

func TestConcurrentGeneration(t *testing.T) {
	gen, err := NewSnowflakeGenerator(5678)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	var wg sync.WaitGroup
	ids := make(chan string, 1000)
	
	// Generate IDs concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ids <- gen.GenerateID()
			}
		}()
	}

	wg.Wait()
	close(ids)

	// Check for duplicates
	seen := make(map[string]bool)
	for id := range ids {
		if seen[id] {
			t.Errorf("Duplicate ID generated: %s", id)
			return
		}
		seen[id] = true
	}

	if len(seen) != 1000 {
		t.Errorf("Expected 1000 unique IDs, got %d", len(seen))
	}
}

func TestServiceTypeGenerators(t *testing.T) {
	testCases := []struct {
		serviceType string
		expectedWorkerID string
	}{
		{"wallet", "5000"},
		{"betting", "6000"},
		{"jackpot", "7000"},
		{"payment", "8000"},
		{"compliance", "9000"},
		{"unknown", "9998"},
	}

	for _, tc := range testCases {
		t.Run(tc.serviceType, func(t *testing.T) {
			gen, err := ServiceTypeGenerator(tc.serviceType)
			if err != nil {
				t.Fatalf("Failed to create %s generator: %v", tc.serviceType, err)
			}

			id := gen.GenerateID()
			_, workerID, _, err := ParseSnowflakeID(id)
			if err != nil {
				t.Fatalf("Failed to parse ID: %v", err)
			}

			if workerID != tc.expectedWorkerID {
				t.Errorf("Expected worker ID %s for %s, got %s", tc.expectedWorkerID, tc.serviceType, workerID)
			}
		})
	}
}

func TestCountrySpecificGenerators(t *testing.T) {
	testCases := []struct {
		countryCode string
		expectedWorkerID string
	}{
		{"KE", "1000"},
		{"NG", "2000"},
		{"GH", "3000"},
		{"unknown", "9999"},
	}

	for _, tc := range testCases {
		t.Run(tc.countryCode, func(t *testing.T) {
			gen, err := CountrySpecificGenerator(tc.countryCode)
			if err != nil {
				t.Fatalf("Failed to create %s generator: %v", tc.countryCode, err)
			}

			id := gen.GenerateID()
			_, workerID, _, err := ParseSnowflakeID(id)
			if err != nil {
				t.Fatalf("Failed to parse ID: %v", err)
			}

			if workerID != tc.expectedWorkerID {
				t.Errorf("Expected worker ID %s for %s, got %s", tc.expectedWorkerID, tc.countryCode, workerID)
			}
		})
	}
}

func TestUserSpecificGenerators(t *testing.T) {
	userIDs := []string{"user123", "user456", "user789"}
	
	generators := make([]*SnowflakeGenerator, len(userIDs))
	for i, userID := range userIDs {
		gen, err := UserSpecificGenerator(userID)
		if err != nil {
			t.Fatalf("Failed to create generator for user %s: %v", userID, err)
		}
		generators[i] = gen
	}

	// Generate IDs and verify different worker IDs for different users
	workerIDs := make(map[string]string)
	for i, gen := range generators {
		id := gen.GenerateID()
		_, workerID, _, err := ParseSnowflakeID(id)
		if err != nil {
			t.Fatalf("Failed to parse ID: %v", err)
		}
		workerIDs[userIDs[i]] = workerID
	}

	// Different users should have different worker IDs
	if workerIDs["user123"] == workerIDs["user456"] || 
	   workerIDs["user456"] == workerIDs["user789"] || 
	   workerIDs["user123"] == workerIDs["user789"] {
		t.Error("Different users should have different worker IDs")
	}
}

func TestSequenceNumberHandling(t *testing.T) {
	gen, err := NewSnowflakeGenerator(9999)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Generate many IDs quickly to test sequence number handling
	var lastTimestamp string
	var lastSequence string
	
	for i := 0; i < 50; i++ {
		id := gen.GenerateID()
		timestamp, _, sequence, err := ParseSnowflakeID(id)
		if err != nil {
			t.Fatalf("Failed to parse ID: %v", err)
		}

		if lastTimestamp == timestamp {
			// Same timestamp, sequence should increment
			if sequence <= lastSequence {
				t.Errorf("Sequence should increment within same second: %s -> %s", lastSequence, sequence)
			}
		}

		lastTimestamp = timestamp
		lastSequence = sequence
	}
}

func BenchmarkSnowflakeGeneration(b *testing.B) {
	gen, err := NewSnowflakeGenerator(1234)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gen.GenerateID()
		}
	})
}
