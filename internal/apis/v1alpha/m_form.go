package v1alpha

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/nicholastcs/alchemy/internal/apis/core"
	"github.com/samber/lo"
)

const (
	TextInputType                  string = "text"
	NumericalInputType             string = "numerical"
	MultilineTextInputType         string = "multiline-text"
	SingleSelectNumericalInputType string = "single-select-numerical"
	MultiSelectNumericalInputType  string = "multi-select-numerical"
	SingleSelectTextInputType      string = "single-select-text"
	MultiSelectTextInputType       string = "multi-select-text"
	BooleanInputType               string = "boolean"
)

var inputTypes = map[string]bool{
	TextInputType:                  true,
	NumericalInputType:             true,
	MultilineTextInputType:         true,
	SingleSelectNumericalInputType: true,
	MultiSelectNumericalInputType:  true,
	SingleSelectTextInputType:      true,
	MultiSelectTextInputType:       true,
	BooleanInputType:               true,
}

type FormManifest struct {
	core.Base `yaml:",inline" mapstructure:",squash"`
	Spec      FormSpec    `yaml:"spec" mapstructure:"spec" json:"spec"`
	Status    core.Status `yaml:"status" mapstructure:"status" json:"status"`
}

type FormSpec struct {
	ConfirmationRequired bool    `yaml:"confirmationRequired" mapstructure:"confirmationRequired" json:"confirmationRequired"`
	Fields               []Field `yaml:"fields" mapstructure:"fields" json:"fields"`
}

type FormStatus struct {
	core.Status `yaml:"status" mapstructure:",squash"`
}

type Field struct {
	Name        string      `yaml:"name" mapstructure:"name" json:"name"`
	Title       string      `yaml:"title" mapstructure:"title" json:"title"`
	Description string      `yaml:"description" mapstructure:"description" json:"description"`
	Choices     []any       `yaml:"choices,omitempty" mapstructure:"choices" json:"choices,omitempty"`
	InputType   string      `yaml:"inputType" mapstructure:"inputType" json:"inputType"`
	Constraint  *Constraint `yaml:"constraint" mapstructure:"constraint" json:"constraint"`
}

type Constraint struct {
	Cel *Cel `yaml:"cel,omitempty" mapstructure:"cel" json:"cel,omitempty"`
}

type Cel struct {
	Expressions []CelExpression `yaml:"expressions" mapstructure:"expressions" json:"expressions"`
}

type CelExpression struct {
	Message string `yaml:"message" mapstructure:"message" json:"message"`
	Value   string `yaml:"value" mapstructure:"value" json:"value"`
}

func (m *FormManifest) Validate() error {
	var errs error
	errs = errors.Join(errs, m.Base.Validate())
	errs = errors.Join(errs, validateFormSpec(m.Spec))

	return errs

}

func validateFormSpec(spec FormSpec) error {
	var errs error

	for i, form := range spec.Fields {
		path := fmt.Sprintf("spec.forms[%d]", i)

		if len(form.Name) < 1 || len(form.Name) > 50 {
			errs = errors.Join(errs, core.NewPathError(fmt.Sprintf("%s.name", path),
				fmt.Errorf("form name '%s' must be between 1 and 50 characters", form.Name)))
		}

		if len(form.Title) < 1 || len(form.Title) > 50 {
			errs = errors.Join(errs, core.NewPathError(fmt.Sprintf("%s.title", path),
				fmt.Errorf("form title '%s' must be between 1 and 50 characters", form.Title)))
		}

		if len(form.Description) < 1 || len(form.Description) > 200 {
			errs = errors.Join(errs, core.NewPathError(fmt.Sprintf("%s.description", path),
				fmt.Errorf("form description '%s' must be between 1 and 200 characters", form.Description)))
		}

		if allowed := inputTypes[form.InputType]; !allowed {
			errs = errors.Join(errs, core.NewPathError(fmt.Sprintf("%s.inputType", path),
				fmt.Errorf("form type must be either %s", english.OxfordWordSeries(lo.Keys(inputTypes), "or"))))
		}

		if strings.Contains(form.InputType, "select") {
			if len(form.Choices) <= 1 {
				errs = errors.Join(errs, core.NewPathError(fmt.Sprintf("%s.choices", path),
					fmt.Errorf("form with 'select' inputType must have more than 1 choice")))
			}
		}
	}
	return errs
}
