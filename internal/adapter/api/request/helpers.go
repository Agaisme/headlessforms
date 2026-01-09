package request

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ServerMeta contains metadata auto-collected from HTTP request
// This data is TRUSTED (collected server-side, cannot be spoofed by client)
type ServerMeta struct {
	// Core identification
	IP        string    `json:"ip"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`

	// Browser/Client info
	UserAgent  string `json:"user_agent"`
	Language   string `json:"language,omitempty"`    // Accept-Language
	IsMobile   string `json:"is_mobile,omitempty"`   // Sec-CH-UA-Mobile
	Platform   string `json:"platform,omitempty"`    // Sec-CH-UA-Platform
	ClientHint string `json:"client_hint,omitempty"` // Sec-CH-UA

	// Request context
	Referer     string `json:"referer,omitempty"`
	Origin      string `json:"origin,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Protocol    string `json:"protocol,omitempty"` // http or https

	// Cloudflare-specific (only present if behind Cloudflare)
	Country     string `json:"country,omitempty"`      // CF-IPCountry
	EstimatedTZ string `json:"estimated_tz,omitempty"` // Timezone estimated from country
	CFRay       string `json:"cf_ray,omitempty"`       // CF-Ray (request ID)

	// Privacy
	DoNotTrack string `json:"dnt,omitempty"` // DNT header
}

// GetClientIP extracts the real client IP, handling proxies and Cloudflare
// Priority: CF-Connecting-IP > True-Client-IP > X-Real-IP > X-Forwarded-For > RemoteAddr
func GetClientIP(r *http.Request) string {
	// Cloudflare headers (highest priority)
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("True-Client-IP"); ip != "" {
		return ip
	}

	// Standard proxy headers
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// X-Forwarded-For can contain multiple IPs: client, proxy1, proxy2
	// The first IP is the original client
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Fallback to direct connection (remove port if present)
	ip := r.RemoteAddr
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		// Handle IPv6 addresses like [::1]:8080
		if bracketIdx := strings.LastIndex(ip, "]"); bracketIdx != -1 && bracketIdx < colonIdx {
			ip = ip[:colonIdx]
		} else if strings.Count(ip, ":") == 1 {
			// IPv4 like 127.0.0.1:8080
			ip = ip[:colonIdx]
		}
	}
	return ip
}

// getProtocol detects if request came via HTTP or HTTPS
func getProtocol(r *http.Request) string {
	// Check Cloudflare header first
	if cfVisitor := r.Header.Get("CF-Visitor"); cfVisitor != "" {
		if strings.Contains(cfVisitor, "https") {
			return "https"
		}
		return "http"
	}

	// Check standard proxy header
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}

	// Check TLS
	if r.TLS != nil {
		return "https"
	}

	return "http"
}

// getTimezoneFromCountry estimates timezone from country code
// This is an approximation - countries with multiple zones use the main/capital timezone
var countryTimezones = map[string]string{
	// Southeast Asia
	"ID": "Asia/Jakarta",
	"MY": "Asia/Kuala_Lumpur",
	"SG": "Asia/Singapore",
	"TH": "Asia/Bangkok",
	"VN": "Asia/Ho_Chi_Minh",
	"PH": "Asia/Manila",
	// East Asia
	"JP": "Asia/Tokyo",
	"KR": "Asia/Seoul",
	"CN": "Asia/Shanghai",
	"TW": "Asia/Taipei",
	"HK": "Asia/Hong_Kong",
	// South Asia
	"IN": "Asia/Kolkata",
	"PK": "Asia/Karachi",
	"BD": "Asia/Dhaka",
	// Middle East
	"AE": "Asia/Dubai",
	"SA": "Asia/Riyadh",
	"TR": "Europe/Istanbul",
	// Europe
	"GB": "Europe/London",
	"DE": "Europe/Berlin",
	"FR": "Europe/Paris",
	"NL": "Europe/Amsterdam",
	"IT": "Europe/Rome",
	"ES": "Europe/Madrid",
	"RU": "Europe/Moscow",
	// Americas
	"US": "America/New_York",
	"CA": "America/Toronto",
	"MX": "America/Mexico_City",
	"BR": "America/Sao_Paulo",
	"AR": "America/Buenos_Aires",
	// Oceania
	"AU": "Australia/Sydney",
	"NZ": "Pacific/Auckland",
}

func getTimezoneFromCountry(country string) string {
	if tz, ok := countryTimezones[country]; ok {
		return tz
	}
	return "" // Unknown country
}

// GetServerMeta collects all server-detectable metadata from the HTTP request
// This returns TRUSTED data that cannot be manipulated by the client
func GetServerMeta(r *http.Request) ServerMeta {
	country := r.Header.Get("CF-IPCountry")

	return ServerMeta{
		// Core
		IP:        GetClientIP(r),
		RequestID: uuid.New().String(),
		Timestamp: time.Now().UTC(),

		// Browser info
		UserAgent:  r.Header.Get("User-Agent"),
		Language:   r.Header.Get("Accept-Language"),
		IsMobile:   r.Header.Get("Sec-CH-UA-Mobile"),
		Platform:   r.Header.Get("Sec-CH-UA-Platform"),
		ClientHint: r.Header.Get("Sec-CH-UA"),

		// Context
		Referer:     r.Header.Get("Referer"),
		Origin:      r.Header.Get("Origin"),
		ContentType: r.Header.Get("Content-Type"),
		Protocol:    getProtocol(r),

		// Cloudflare + Timezone estimate
		Country:     country,
		EstimatedTZ: getTimezoneFromCountry(country),
		CFRay:       r.Header.Get("CF-Ray"),

		// Privacy
		DoNotTrack: r.Header.Get("DNT"),
	}
}
