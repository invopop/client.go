package invopop

// Config is a configuration object for the Invopop client.
type Config struct {
	BaseURL      string `json:"base_url"`      // Base URL to access the Invopop API.
	ClientID     string `json:"client_id"`     // OAuth Client ID, if required.
	ClientSecret string `json:"client_secret"` // OAuth Client Secret, if required.
}
