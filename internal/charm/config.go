package charm

import (
	"fmt"

	"github.com/gruyaume/goops"
)

type ConfigOptions struct {
	email                string
	server               string
	plugin               string
	pluginConfigSecretID string
}

func (c *ConfigOptions) LoadFromJuju() error {
	email, err := goops.GetConfigString("email")
	if err != nil {
		return fmt.Errorf("failed to get email config: %w", err)
	}

	server, err := goops.GetConfigString("server")
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}

	plugin, err := goops.GetConfigString("plugin")
	if err != nil {
		return fmt.Errorf("failed to get plugin config: %w", err)
	}

	pluginConfigSecretID, err := goops.GetConfigString("plugin-config-secret-id")
	if err != nil {
		return fmt.Errorf("failed to get plugin-config-secret-id config: %w", err)
	}

	c.email = email
	c.server = server
	c.plugin = plugin
	c.pluginConfigSecretID = pluginConfigSecretID

	return nil
}

func (c *ConfigOptions) Validate() error {
	if c.email == "" {
		return fmt.Errorf("email config is empty")
	}

	if c.server == "" {
		return fmt.Errorf("server config is empty")
	}

	return nil
}
