package core

import (
	"fmt"
)

type ConfigManifest struct {
	Base   `yaml:",inline" mapstructure:",squash"`
	Spec   ConfigSpec `mapstructure:"spec" yaml:"spec" json:"spec"`
	Status Status     `yaml:"status,omitempty" mapstructure:"status.omitempty" json:"status"`
}

type ConfigSpec struct {
	EnableLogs     bool        `mapstructure:"enableLogs" yaml:"enableLogs" json:"enableLogs"`
	SelectedConfig string      `mapstructure:"selectedConfig" yaml:"selectedConfig" json:"selectedConfig"`
	Configs        []SubConfig `mapstructure:"configs" yaml:"configs" json:"configs"`
}

type SubConfig struct {
	Name      string `mapstructure:"name" yaml:"name" json:"name"`
	Type      string `mapstructure:"type" yaml:"type" json:"type"`
	Directory string `mapstructure:"directory" yaml:"directory" json:"directory"`
}

func (c *ConfigManifest) SelectConfig(configName string) error {
	for _, cfg := range c.Spec.Configs {
		if cfg.Name == configName {
			c.Spec.SelectedConfig = configName
			return nil
		}
	}

	return fmt.Errorf("config '%s' not found in spec.configs.*.name", configName)
}
