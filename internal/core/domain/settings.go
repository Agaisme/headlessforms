package domain

import "time"

// SiteSettings represents global site configuration
type SiteSettings struct {
	ID       string `json:"id"`
	SiteName string `json:"site_name"`
	SiteURL  string `json:"site_url"`

	// SMTP Configuration
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password,omitempty"` // Never expose in GET responses
	SMTPFrom     string `json:"smtp_from"`
	SMTPFromName string `json:"smtp_from_name"`
	SMTPSecure   bool   `json:"smtp_secure"` // TLS

	// System Info (read-only)
	Version   string    `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by,omitempty"`
}

// SiteSettingsUpdate is used for PATCH/PUT requests (without password if empty)
type SiteSettingsUpdate struct {
	SiteName     *string `json:"site_name,omitempty"`
	SiteURL      *string `json:"site_url,omitempty"`
	SMTPHost     *string `json:"smtp_host,omitempty"`
	SMTPPort     *int    `json:"smtp_port,omitempty"`
	SMTPUser     *string `json:"smtp_user,omitempty"`
	SMTPPassword *string `json:"smtp_password,omitempty"`
	SMTPFrom     *string `json:"smtp_from,omitempty"`
	SMTPFromName *string `json:"smtp_from_name,omitempty"`
	SMTPSecure   *bool   `json:"smtp_secure,omitempty"`
}

// SMTPConfig returns SMTP configuration for email service
func (s *SiteSettings) SMTPConfig() (host string, port int, user, pass, from, fromName string, secure bool) {
	return s.SMTPHost, s.SMTPPort, s.SMTPUser, s.SMTPPassword, s.SMTPFrom, s.SMTPFromName, s.SMTPSecure
}

// HasSMTPConfig returns true if SMTP is configured
func (s *SiteSettings) HasSMTPConfig() bool {
	return s.SMTPHost != "" && s.SMTPPort > 0
}

// ToPublic returns settings safe for API response (masks password)
func (s *SiteSettings) ToPublic() *SiteSettings {
	copy := *s
	if copy.SMTPPassword != "" {
		copy.SMTPPassword = "********" // Mask password
	}
	return &copy
}
