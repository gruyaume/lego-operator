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

func (c *ConfigOptions) LoadFromJuju() {
	email, err := goops.GetConfigString("email")
	if err != nil {
		email = ""
	}

	server, err := goops.GetConfigString("server")
	if err != nil {
		server = ""
	}

	plugin, err := goops.GetConfigString("plugin")
	if err != nil {
		plugin = ""
	}

	pluginConfigSecretID, err := goops.GetConfigString("plugin-config-secret-id")
	if err != nil {
		pluginConfigSecretID = ""
	}

	c.email = email
	c.server = server
	c.plugin = plugin
	c.pluginConfigSecretID = pluginConfigSecretID
}

func (c *ConfigOptions) Validate() error {
	if c.email == "" {
		return fmt.Errorf("email config is empty")
	}

	if c.server == "" {
		return fmt.Errorf("server config is empty")
	}

	if c.plugin == "" {
		return fmt.Errorf("plugin config is empty")
	}

	if c.pluginConfigSecretID == "" {
		return fmt.Errorf("plugin config secret ID is empty")
	}

	return nil
}
