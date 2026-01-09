package spam

import (
	"strings"
	"sync"
	"time"
)

// SpamScore represents the spam analysis result
type SpamScore struct {
	Score     int      `json:"score"`     // 0-100, higher = more likely spam
	IsSpam    bool     `json:"is_spam"`   // true if score >= threshold
	Flags     []string `json:"flags"`     // Reasons for the score
	Threshold int      `json:"threshold"` // Score threshold used
}

// Config holds spam detection configuration
type Config struct {
	ScoreThreshold     int           // Score at which submission is marked spam (default: 50)
	RateLimitWindow    time.Duration // Time window for rate limiting (default: 1 minute)
	RateLimitMax       int           // Max submissions per IP in window (default: 10)
	HoneypotFieldNames []string      // Hidden field names to detect bots
}

// DefaultConfig returns sensible default configuration
func DefaultConfig() Config {
	return Config{
		ScoreThreshold:     50,
		RateLimitWindow:    time.Minute,
		RateLimitMax:       10,
		HoneypotFieldNames: []string{"_honeypot", "_hp", "website", "url", "fax"},
	}
}

// Detector handles spam detection logic
type Detector struct {
	config     Config
	rateLimits map[string][]time.Time // IP -> submission timestamps
	mu         sync.RWMutex
}

// NewDetector creates a new spam detector
func NewDetector(config Config) *Detector {
	return &Detector{
		config:     config,
		rateLimits: make(map[string][]time.Time),
	}
}

// Analyze checks submission for spam signals
func (d *Detector) Analyze(ip string, userAgent string, data map[string]interface{}, submissionTime time.Duration) SpamScore {
	var score int
	var flags []string

	// 1. Check honeypot fields
	for _, field := range d.config.HoneypotFieldNames {
		if val, ok := data[field]; ok {
			if str, isStr := val.(string); isStr && str != "" {
				score += 100 // Guaranteed spam
				flags = append(flags, "honeypot_filled:"+field)
			}
		}
	}

	// 2. Check user agent
	if userAgent == "" {
		score += 30
		flags = append(flags, "empty_user_agent")
	} else {
		lowerUA := strings.ToLower(userAgent)
		// Known bot patterns
		botPatterns := []string{"bot", "crawler", "spider", "curl", "wget", "python", "scrapy", "headless"}
		for _, pattern := range botPatterns {
			if strings.Contains(lowerUA, pattern) {
				score += 40
				flags = append(flags, "bot_user_agent:"+pattern)
				break
			}
		}
	}

	// 3. Check submission speed (if provided)
	if submissionTime > 0 && submissionTime < 2*time.Second {
		score += 25
		flags = append(flags, "fast_submission")
	}

	// 4. Check rate limit
	if d.isRateLimited(ip) {
		score += 30
		flags = append(flags, "rate_limited")
	}

	// 5. Check for suspicious content patterns
	for _, v := range data {
		if str, ok := v.(string); ok {
			// Links in message (common spam pattern)
			if strings.Contains(str, "http://") || strings.Contains(str, "https://") {
				linkCount := strings.Count(str, "http")
				if linkCount > 2 {
					score += 15
					flags = append(flags, "multiple_links")
				}
			}
		}
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return SpamScore{
		Score:     score,
		IsSpam:    score >= d.config.ScoreThreshold,
		Flags:     flags,
		Threshold: d.config.ScoreThreshold,
	}
}

// RecordSubmission tracks a submission for rate limiting
func (d *Detector) RecordSubmission(ip string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	d.rateLimits[ip] = append(d.rateLimits[ip], now)

	// Cleanup old entries
	d.cleanupRateLimits(ip)
}

// isRateLimited checks if IP has exceeded rate limit
func (d *Detector) isRateLimited(ip string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	timestamps, exists := d.rateLimits[ip]
	if !exists {
		return false
	}

	cutoff := time.Now().Add(-d.config.RateLimitWindow)
	count := 0
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			count++
		}
	}

	return count >= d.config.RateLimitMax
}

// cleanupRateLimits removes old entries for an IP
func (d *Detector) cleanupRateLimits(ip string) {
	cutoff := time.Now().Add(-d.config.RateLimitWindow * 2)
	timestamps := d.rateLimits[ip]

	var cleaned []time.Time
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			cleaned = append(cleaned, ts)
		}
	}

	if len(cleaned) == 0 {
		delete(d.rateLimits, ip)
	} else {
		d.rateLimits[ip] = cleaned
	}
}

// CheckHoneypot is a helper to check if honeypot field was filled
func CheckHoneypot(data map[string]interface{}, fieldNames []string) bool {
	for _, field := range fieldNames {
		if val, ok := data[field]; ok {
			if str, isStr := val.(string); isStr && str != "" {
				return true
			}
		}
	}
	return false
}
