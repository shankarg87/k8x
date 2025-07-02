package schemas

// Credentials represents the credentials file structure
// This struct is used for loading and storing LLM provider credentials
// and is shared between config and llm packages.
type Credentials struct {
	// SelectedProvider indicates which LLM provider is currently selected
	SelectedProvider string `yaml:"selected_provider,omitempty"`
	OpenAI           struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"openai"`
	Anthropic struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"anthropic"`
	Google struct {
		ApplicationCredentials string `yaml:"application_credentials"`
	} `yaml:"google"`
}

// HasAnyKey checks if the credentials contain any of the specified keys
func (c *Credentials) HasAnyKey(keys ...string) bool {
	for _, key := range keys {
		switch key {
		case "openai_api_key":
			if c.OpenAI.APIKey != "" {
				return true
			}
		case "anthropic_api_key":
			if c.Anthropic.APIKey != "" {
				return true
			}
		case "google_application_credentials":
			if c.Google.ApplicationCredentials != "" {
				return true
			}
		}
	}
	return false
}

// SetProviderAPIKey sets the API key for the given provider in the Credentials struct
// and updates the selected provider
func (c *Credentials) SetProviderAPIKey(provider, apiKey string) {
	// Set the provider as selected
	c.SelectedProvider = provider

	// Set the API key for the specified provider
	switch provider {
	case "openai":
		c.OpenAI.APIKey = apiKey
	case "anthropic":
		c.Anthropic.APIKey = apiKey
	case "google":
		c.Google.ApplicationCredentials = apiKey
	}
}
