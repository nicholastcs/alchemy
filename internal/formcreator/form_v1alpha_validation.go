package formcreator

import (
	"errors"

	"github.com/mitchellh/copystructure"
	"github.com/nicholastcs/alchemy/internal/apis/v1alpha"
	"github.com/nicholastcs/alchemy/internal/system"
)

// validationOutcome is the outcome of the CEL expression evaluation
type validationOutcome struct {
	// UserDefinedError is an error defined by user which can be set by
	// user in the `messages` field inside the expression value pair
	// like so:-
	//
	//   cel:
	//     expressions:
	//       - message: length of name must be greater than 0. <<-
	//         value: this.size() > 0
	UserDefinedError error

	// RuntimeError is an error emitted by the CEL evaluation itself,
	// whereby it is emitted inside CEL expression.
	//
	// It should be logged into the manifest for troubleshooting.
	//
	// However, not all runtime error are erratic, it could be due to
	// wrong field input value. For example,
	//
	//   quantity(this) > quantity("500m")
	//
	// Expression above will return runtime error when `this` is not
	// well formed.
	RuntimeError error
}

func (v *validationOutcome) HasRuntimeError() bool {
	return v.RuntimeError != nil
}

func (v *validationOutcome) HasUserDefinedError() bool {
	return v.UserDefinedError != nil
}

func (p *v1alphaFormCreator) validate(
	field v1alpha.Field,
	resultManifest *v1alpha.FormResultManifest,
) func(any) error {
	return func(input any) error {
		// copying the struct as we don't want to mess with the
		// struct under consumption by charmbracelet/huh API.
		resultToValidate, err := copystructure.Copy(resultManifest.Spec)
		if err != nil {
			return err
		}
		r, ok := resultToValidate.(v1alpha.FormResultSpec)
		if !ok {
			return errors.New("unable to assert type to v1alpha.FormResultSpec")
		}

		err = r.ConvertResultToNative()
		if err != nil {
			return err
		}

		valueUnderCheck := map[string]interface{}{
			"this":   input,
			"result": r.Result,
		}

		validationOutcome := p.validationHarness(valueUnderCheck, field)
		if validationOutcome.HasRuntimeError() {
			resultManifest.Status.SetError(validationOutcome.RuntimeError)
		}

		return errors.Join(validationOutcome.RuntimeError, validationOutcome.UserDefinedError)
	}
}

func (p *v1alphaFormCreator) validationHarness(valueUnderCheck map[string]any, form v1alpha.Field) *validationOutcome {
	entry := p.log.WithField("input", valueUnderCheck)

	var v validationOutcome

	if form.Constraint == nil {
		entry.Trace(".constraint not found, skipping")
		return &v
	}
	if form.Constraint.Cel == nil {
		entry.Trace(".constraint.cel not found, skipping")
		return &v
	}
	if form.Constraint.Cel.Expressions == nil {
		entry.Trace(".constraint.cel.expressions not found, skipping")
		return &v
	}

	for _, expression := range form.Constraint.Cel.Expressions {
		ok, err := system.ExecuteCELOnFormValidation(valueUnderCheck, expression.Value)
		if err != nil {
			v.RuntimeError = errors.Join(v.RuntimeError, err)
		}

		// The user defined error only valid if there are no syntax or type errors.
		if !ok && err == nil {
			v.UserDefinedError = errors.Join(v.UserDefinedError, errors.New(expression.Message))
		}
	}

	return &v
}
