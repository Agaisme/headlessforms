package spam

import (
	"testing"
	"time"
)

func TestDetector_Honeypot(t *testing.T) {
	detector := NewDetector(DefaultConfig())

	tests := []struct {
		name     string
		data     map[string]interface{}
		wantSpam bool
	}{
		{
			name:     "clean submission",
			data:     map[string]interface{}{"name": "John", "email": "john@example.com"},
			wantSpam: false,
		},
		{
			name:     "honeypot filled",
			data:     map[string]interface{}{"name": "John", "_honeypot": "gotcha"},
			wantSpam: true,
		},
		{
			name:     "empty honeypot is fine",
			data:     map[string]interface{}{"name": "John", "_honeypot": ""},
			wantSpam: false,
		},
		{
			name:     "website field filled (common honeypot)",
			data:     map[string]interface{}{"name": "John", "website": "http://spam.com"},
			wantSpam: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.Analyze("1.2.3.4", "Mozilla/5.0", tt.data, 0)
			if result.IsSpam != tt.wantSpam {
				t.Errorf("got IsSpam=%v, want %v (score: %d, flags: %v)",
					result.IsSpam, tt.wantSpam, result.Score, result.Flags)
			}
		})
	}
}

func TestDetector_BotUserAgent(t *testing.T) {
	detector := NewDetector(DefaultConfig())
	data := map[string]interface{}{"name": "Test"}

	tests := []struct {
		name      string
		userAgent string
		wantFlags bool
	}{
		{"normal browser", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0", false},
		{"empty agent", "", true},
		{"curl", "curl/7.64.1", true},
		{"python requests", "python-requests/2.25.1", true},
		{"wget", "Wget/1.20.3", true},
		{"googlebot", "Googlebot/2.1", true},
		{"scrapy", "Scrapy/2.5.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.Analyze("1.2.3.4", tt.userAgent, data, 0)
			hasFlags := len(result.Flags) > 0
			if hasFlags != tt.wantFlags {
				t.Errorf("got hasFlags=%v, want %v (flags: %v)", hasFlags, tt.wantFlags, result.Flags)
			}
		})
	}
}

func TestDetector_RateLimiting(t *testing.T) {
	config := Config{
		ScoreThreshold:     50,
		RateLimitWindow:    time.Minute,
		RateLimitMax:       3,
		HoneypotFieldNames: []string{},
	}
	detector := NewDetector(config)
	data := map[string]interface{}{"name": "Test"}
	ip := "192.168.1.1"

	// First 3 submissions should not trigger rate limit
	for i := 0; i < 3; i++ {
		detector.RecordSubmission(ip)
		result := detector.Analyze(ip, "Mozilla/5.0", data, 0)
		if result.IsSpam {
			t.Errorf("submission %d should not be spam", i+1)
		}
	}

	// 4th submission should trigger rate limit flag
	detector.RecordSubmission(ip)
	result := detector.Analyze(ip, "Mozilla/5.0", data, 0)

	hasRateLimitFlag := false
	for _, flag := range result.Flags {
		if flag == "rate_limited" {
			hasRateLimitFlag = true
			break
		}
	}

	if !hasRateLimitFlag {
		t.Errorf("expected rate_limited flag, got: %v", result.Flags)
	}
}

func TestDetector_FastSubmission(t *testing.T) {
	detector := NewDetector(DefaultConfig())
	data := map[string]interface{}{"name": "Test"}

	// Fast submission (1 second) should be flagged
	fast := detector.Analyze("1.2.3.4", "Mozilla/5.0", data, time.Second)
	if !containsFlag(fast.Flags, "fast_submission") {
		t.Error("expected fast_submission flag for 1s submission")
	}

	// Slow submission (5 seconds) should not be flagged
	slow := detector.Analyze("5.6.7.8", "Mozilla/5.0", data, 5*time.Second)
	if containsFlag(slow.Flags, "fast_submission") {
		t.Error("5s submission should not have fast_submission flag")
	}
}

func TestDetector_MultipleLinks(t *testing.T) {
	detector := NewDetector(DefaultConfig())

	tests := []struct {
		name     string
		message  string
		wantFlag bool
	}{
		{"no links", "Hello, how are you?", false},
		{"one link", "Check out https://example.com", false},
		{"two links", "Visit https://a.com and https://b.com", false},
		{"many links", "Buy now: https://a.com https://b.com https://c.com https://d.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{"message": tt.message}
			result := detector.Analyze("1.2.3.4", "Mozilla/5.0", data, 0)
			hasFlag := containsFlag(result.Flags, "multiple_links")
			if hasFlag != tt.wantFlag {
				t.Errorf("got hasFlag=%v, want %v", hasFlag, tt.wantFlag)
			}
		})
	}
}

func TestDetector_ScoreCapping(t *testing.T) {
	detector := NewDetector(DefaultConfig())

	// Multiple spam indicators should still cap at 100
	data := map[string]interface{}{
		"_honeypot": "filled",
		"website":   "http://spam.com",
		"fax":       "more spam",
		"message":   "https://a.com https://b.com https://c.com https://d.com",
	}

	result := detector.Analyze("", "", data, time.Millisecond)
	if result.Score > 100 {
		t.Errorf("score should be capped at 100, got %d", result.Score)
	}
	if result.Score != 100 {
		t.Logf("score: %d, flags: %v", result.Score, result.Flags)
	}
}

func containsFlag(flags []string, target string) bool {
	for _, f := range flags {
		if f == target {
			return true
		}
	}
	return false
}
