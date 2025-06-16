package charm

import (
	"fmt"
)

type ConfigOptions struct {
	Email                string `json:"email"`
	Server               string `json:"server"`
	Plugin               string `json:"plugin"`
	PluginConfigSecretID string `json:"plugin-config-secret-id"`
}

func (c *ConfigOptions) Validate() error {
	if c.Email == "" {
		return fmt.Errorf("email config is empty")
	}

	if c.Server == "" {
		return fmt.Errorf("server config is empty")
	}

	if c.Plugin == "" {
		return fmt.Errorf("plugin config is empty")
	}

	if c.PluginConfigSecretID == "" {
		return fmt.Errorf("plugin config secret ID is empty")
	}

	return nil
}
