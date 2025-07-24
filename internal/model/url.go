package model

// URLRequest represents the incoming JSON for shortening a URL
type URLRequest struct {
	URL        string  `json:"url"`
	CustomSlug *string `json:"custom_slug,omitempty"`
}

// URL represents the structure of a URL record in the database
type URL struct {
	ShortCode   string  `json:"short_code"`
	OriginalURL string  `json:"original_url"`
	ClickCount  int     `json:"click_count"`
	CustomSlug  *string `json:"custom_slug,omitempty"`
}

// URLResponse represents the outgoing JSON after shortening a URL
type URLResponse struct {
	ShortURL string `json:"short_url"`
}
