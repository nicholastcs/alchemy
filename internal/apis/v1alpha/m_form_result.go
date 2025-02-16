package v1alpha

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/nicholastcs/alchemy/internal/apis/core"
)

type FormResultManifest struct {
	core.Base `yaml:",inline" mapstructure:",squash"`
	Spec      FormResultSpec   `yaml:"spec" mapstructure:"spec" json:"spec"`
	Status    FormResultStatus `yaml:"status" mapstructure:"status" json:"status,omitempty"`
}

type FormResultSpec struct {
	FormReference core.Base `yaml:"formReference" mapstructure:"formReference" json:"formReference"`

	TypeHintByResult map[string]string `yaml:"-" mapstructure:"-" json:"-"`
	Result           map[string]any    `yaml:"result" mapstructure:"result" json:"result"`
}

type FormResultStatus struct {
	core.Status  `yaml:",inline" mapstructure:",squash"`
	MappedResult map[string]any `yaml:"mappedResult,omitempty" mapstructure:"mappedResult,omitempty" json:"mappedResult,omitempty"`
}

func NewFormResult(name string, namespace string, formReference core.Base, result map[string]any) (*FormResultManifest, error) {
	if namespace == "" {
		namespace = "default"
	}

	base := core.Base{
		APIVersion: "alchemy.io/v1alpha/internal",
		Kind:       "FormResult",
		Metadata: core.Metadata{
			Name:      name,
			Namespace: namespace,
		},
	}

	output := &FormResultManifest{
		Base: base,
		Spec: FormResultSpec{
			Result:           result,
			FormReference:    formReference,
			TypeHintByResult: map[string]string{},
		},
	}

	err := output.Validate()
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (m *FormResultManifest) Validate() error {
	var errs error
	errs = errors.Join(errs, m.Base.Validate())
	errs = errors.Join(errs, validateFormResultSpec(m.Spec))

	return errs
}

func validateFormResultSpec(spec FormResultSpec) error {
	var errs error

	errs = errors.Join(errs, spec.FormReference.Validate())

	if spec.Result == nil {
		errs = errors.Join(errs, core.NewPathError("spec.result", errors.New("result value cannot be null")))
	}

	return errs
}

func (m *FormResultSpec) NewEmptyResult(key string, value any, fieldType string) {
	m.TypeHintByResult[key] = fieldType
	m.Result[key] = value
}

// ConvertResultToNative converts certain string types which are actually
// should be numerical into 64-bit float.
func (m *FormResultSpec) ConvertResultToNative() error {
	m.indirectResults()

	// convert to native for certain types that are handled as strings
	// literal in charmbracelet/huh API
	for name, typeHint := range m.TypeHintByResult {
		if strings.Contains(typeHint, "numerical") {
			val := m.Result[name]

			valueLiteral := fmt.Sprintf("%v", val)

			if strings.TrimSpace(valueLiteral) == "" {
				m.Result[name] = 0

				continue
			}

			c, err := strconv.ParseFloat(valueLiteral, 64)
			if err != nil {
				return err
			}

			m.Result[name] = c
		}
	}

	return nil
}

// indirectResults simply indirects all values under m.Results into value
// type.
func (m *FormResultSpec) indirectResults() {
	for name := range m.Result {
		val := reflect.Indirect(reflect.ValueOf(m.Result[name])).Interface()
		m.Result[name] = val
	}
}
