package core

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
)

// Base is a struct that should be embedded by all APIs in Alchemy project.
type Base struct {
	APIVersion string   `yaml:"apiVersion" mapstructure:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind" mapstructure:"kind" json:"kind"`
	Metadata   Metadata `yaml:"metadata" mapstructure:"metadata" json:"metadata"`
}

// Metadata is a struct that contains identity information of an manifest,
// except annotations.
type Metadata struct {
	Name        string            `yaml:"name" mapstructure:"name" json:"name"`
	Namespace   string            `yaml:"namespace" mapstructure:"namespace" json:"namespace"`
	Annotations map[string]string `yaml:"annotations,omitempty" mapstructure:"annotations" json:"annotations,omitempty"`
}

// AbstractedManifest is an manifest that has abstracted types in Spec and
// Status.
//
// The types of these two fields are not known during runtime.
type AbstractedManifest struct {
	Base   `yaml:",inline" mapstructure:",squash"`
	Spec   any              `yaml:"spec" mapstructure:"spec" json:"spec"`
	Status AbstractedStatus `yaml:"status" mapstructure:"status" json:"status"`
}

type AbstractedStatus struct {
	Status `yaml:",inline" mapstructure:",squash"`

	// AdditionalInfo is a object that retains supplementary info on
	// top of standard status.
	AdditionalInfo map[string]interface{} `yaml:"additionalInfo,omitempty" mapstructure:",remain" json:"additionalInfo"`
}

func (m AbstractedManifest) ConvertToActualManifest(output any) error {
	return mapstructure.Decode(m, output)
}

func ConvertToAbstractedManifest(input any) (*AbstractedManifest, error) {
	var output AbstractedManifest

	err := mapstructure.Decode(input, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func NewManifestBase(name string, namespace string, apiVersion string, kind string) (*Base, error) {
	base := Base{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata: Metadata{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := base.Validate()
	if err != nil {
		return nil, err
	}

	return &base, nil
}

// Validate validates the base of the manifest.
func (m *Base) Validate() error {
	if m.APIVersion == "" {
		return fmt.Errorf("apiVersion cannot be empty")
	}

	if m.Kind == "" {
		return fmt.Errorf("kind cannot be empty")
	}

	if m.Metadata.Name == "" {
		return fmt.Errorf("metadata.name cannot be empty")
	}

	if m.Metadata.Namespace == "" {
		m.Metadata.Namespace = "default"
	}

	for key, value := range m.Metadata.Annotations {
		if key == "" {
			return fmt.Errorf("metadata.annotations key cannot be empty")
		}
		if value == "" {
			return fmt.Errorf("metadata.annotations value cannot be empty")
		}
	}

	return nil
}

func (b *Base) SetFilePath(path string) {
	if b.Metadata.Annotations == nil {
		b.Metadata.Annotations = map[string]string{}
	}

	b.Metadata.Annotations["alchemy.io/filepath"] = path
}

func (b *Base) GetFilePath() string {
	if b.Metadata.Annotations == nil {
		return "N/A"
	}

	if b.Metadata.Annotations["alchemy.io/filepath"] == "" {
		return "N/A"
	}

	return b.Metadata.Annotations["alchemy.io/filepath"]
}
