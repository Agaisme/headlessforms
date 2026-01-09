// Package options provides functional options for service constructors
package options

// FormOptions contains optional configuration for creating a form
type FormOptions struct {
	RedirectURL   string
	NotifyEmails  []string
	WebhookURL    string
	WebhookSecret string
	AccessMode    string // "public", "with_key", "private"
	SubmissionKey string
}

// FormOption is a functional option for form creation
type FormOption func(*FormOptions)

// NewFormOptions creates FormOptions with defaults
func NewFormOptions(opts ...FormOption) *FormOptions {
	o := &FormOptions{
		AccessMode: "public",
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithRedirectURL sets the redirect URL after form submission
func WithRedirectURL(url string) FormOption {
	return func(o *FormOptions) {
		o.RedirectURL = url
	}
}

// WithNotifyEmails sets emails to notify on new submissions
func WithNotifyEmails(emails ...string) FormOption {
	return func(o *FormOptions) {
		o.NotifyEmails = emails
	}
}

// WithWebhook sets the webhook URL and optional secret
func WithWebhook(url, secret string) FormOption {
	return func(o *FormOptions) {
		o.WebhookURL = url
		o.WebhookSecret = secret
	}
}

// WithAccessMode sets the access mode for the form
func WithAccessMode(mode string) FormOption {
	return func(o *FormOptions) {
		o.AccessMode = mode
	}
}

// WithSubmissionKey sets the submission key for access_mode="with_key"
func WithSubmissionKey(key string) FormOption {
	return func(o *FormOptions) {
		o.SubmissionKey = key
	}
}

// WithPrivateAccess sets up private form access
func WithPrivateAccess() FormOption {
	return func(o *FormOptions) {
		o.AccessMode = "private"
	}
}

// WithKeyAccess sets up key-protected form access
func WithKeyAccess(key string) FormOption {
	return func(o *FormOptions) {
		o.AccessMode = "with_key"
		o.SubmissionKey = key
	}
}
