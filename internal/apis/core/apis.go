package core

type RegisteredAPIManifest struct {
	Base   `yaml:",inline" mapstructure:",squash"`
	Spec   RegisteredAPISpec `mapstructure:"spec" yaml:"spec" json:"spec"`
	Status Status            `yaml:"status" mapstructure:"status" json:"status"`
}

type RegisteredAPISpec struct {
	Meta API `yaml:"meta" mapstructure:"meta"`
}

type API struct {
	APIVersion string   `yaml:"apiVersion" mapstructure:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind" mapstructure:"kind" json:"kind"`
	Aliases    []string `yaml:"aliases" mapstructure:"aliases" json:"aliases"`
}
