package request

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remoteIP string
		wantIP   string
	}{
		{
			name:     "CF-Connecting-IP takes priority",
			headers:  map[string]string{"CF-Connecting-IP": "1.2.3.4", "X-Forwarded-For": "5.6.7.8"},
			remoteIP: "9.9.9.9:12345",
			wantIP:   "1.2.3.4",
		},
		{
			name:     "X-Forwarded-For fallback",
			headers:  map[string]string{"X-Forwarded-For": "5.6.7.8, 9.9.9.9"},
			remoteIP: "10.10.10.10:12345",
			wantIP:   "5.6.7.8",
		},
		{
			name:     "X-Real-IP fallback",
			headers:  map[string]string{"X-Real-IP": "2.2.2.2"},
			remoteIP: "10.10.10.10:12345",
			wantIP:   "2.2.2.2",
		},
		{
			name:     "RemoteAddr fallback",
			headers:  map[string]string{},
			remoteIP: "192.168.1.1:12345",
			wantIP:   "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			req.RemoteAddr = tt.remoteIP

			got := GetClientIP(req)
			if got != tt.wantIP {
				t.Errorf("got %v, want %v", got, tt.wantIP)
			}
		})
	}
}

func TestGetTimezoneFromCountry(t *testing.T) {
	tests := []struct {
		country string
		wantTZ  string
	}{
		{"ID", "Asia/Jakarta"},
		{"US", "America/New_York"},
		{"GB", "Europe/London"},
		{"JP", "Asia/Tokyo"},
		{"AU", "Australia/Sydney"},
		{"XX", ""}, // Unknown country
		{"", ""},   // Empty country
	}

	for _, tt := range tests {
		t.Run(tt.country, func(t *testing.T) {
			got := getTimezoneFromCountry(tt.country)
			if got != tt.wantTZ {
				t.Errorf("got %v, want %v", got, tt.wantTZ)
			}
		})
	}
}

func TestGetServerMeta(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/submissions/test123", nil)
	req.Header.Set("CF-Connecting-IP", "1.2.3.4")
	req.Header.Set("CF-IPCountry", "ID")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Referer", "https://example.com/contact")

	meta := GetServerMeta(req)

	if meta.IP != "1.2.3.4" {
		t.Errorf("IP: got %v, want 1.2.3.4", meta.IP)
	}
	if meta.Country != "ID" {
		t.Errorf("Country: got %v, want ID", meta.Country)
	}
	if meta.EstimatedTZ != "Asia/Jakarta" {
		t.Errorf("EstimatedTZ: got %v, want Asia/Jakarta", meta.EstimatedTZ)
	}
	if meta.Origin != "https://example.com" {
		t.Errorf("Origin: got %v, want https://example.com", meta.Origin)
	}
}
