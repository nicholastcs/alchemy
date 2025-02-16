package formcreator

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/nicholastcs/alchemy/internal/apis/v1alpha"
	"github.com/sirupsen/logrus"
)

type v1alphaFormCreator struct {
	log *logrus.Entry
}

func NewFormCreatorV1Alpha(log *logrus.Entry) (*v1alphaFormCreator, error) {
	return &v1alphaFormCreator{
		log: log.WithField("context", "FormCreator/v1alpha"),
	}, nil
}

func (p *v1alphaFormCreator) Run(m v1alpha.FormManifest) (*v1alpha.FormResultManifest, error) {
	formResult, err := v1alpha.NewFormResult(
		fmt.Sprintf("form-%s", m.Metadata.Name),
		m.Metadata.Namespace,
		m.Base,
		make(map[string]any),
	)
	if err != nil {
		return nil, err
	}

	fields, err := p.initFieldsV1Alpha(m, formResult)
	if err != nil {
		return nil, err
	}

	grp := []*huh.Group{
		huh.NewGroup(fields...),
	}

	if m.Spec.ConfirmationRequired {
		confirmation := huh.NewConfirm().
			Description("Warning, the generated code could override file(s) on your current directory.").
			Title("Proceed to generate IAC")

		grp = append(grp, huh.NewGroup(confirmation))
	}

	form := huh.NewForm(grp...)

	form.WithTheme(themeFP())

	err = form.Run()
	if err != nil {
		return nil, err
	}

	formResult.Status.SetCondition(v1alpha.CodeTemplateConsumptionReady, true)

	return formResult, nil
}

func (p *v1alphaFormCreator) initFieldsV1Alpha(m v1alpha.FormManifest, resultManifest *v1alpha.FormResultManifest) ([]huh.Field, error) {
	fds := []huh.Field{}

	for _, form := range m.Spec.Fields {
		entry := p.log.
			WithFields(
				logrus.Fields{
					"formManifest": m.Base,
					"form":         form,
				},
			)

		switch form.InputType {
		case v1alpha.TextInputType:
			var value string
			resultManifest.Spec.NewEmptyResult(form.Name, &value, form.InputType)

			fd := huh.NewInput().
				Title(form.Title).
				Description(form.Description).
				Validate(func(s string) error {
					f := p.validate(form, resultManifest)

					return f(s)
				}).
				Value(&value)

			fds = append(fds, fd)

		case v1alpha.NumericalInputType:
			var value string
			resultManifest.Spec.NewEmptyResult(form.Name, &value, form.InputType)

			fd := huh.NewInput().
				Title(form.Title).
				Description(form.Description).
				Validate(func(s string) error {
					floatVal, err := strconv.ParseFloat(s, 64)
					if err != nil {
						return err
					}

					f := p.validate(form, resultManifest)

					return f(floatVal)
				}).
				Value(&value)

			fds = append(fds, fd)

		case v1alpha.MultilineTextInputType:
			var value string
			resultManifest.Spec.NewEmptyResult(form.Name, &value, form.InputType)

			fd := huh.NewText().
				Title(form.Title).
				Description(form.Description).
				Validate(func(s string) error {
					f := p.validate(form, resultManifest)

					return f(s)
				}).
				Value(&value)

			fds = append(fds, fd)

		case v1alpha.SingleSelectTextInputType:
			if len(form.Choices) < 2 {
				return nil, fmt.Errorf("selection type forms must have at least 2 choices")
			}

			opts := []huh.Option[string]{}

			for _, choice := range form.Choices {
				c, ok := choice.(string)
				if !ok {
					return nil, fmt.Errorf("unable to assert type string for value `%v`", choice)
				}
				opts = append(opts, huh.NewOption(c, c))
			}

			var value string
			resultManifest.Spec.NewEmptyResult(form.Name, &value, form.InputType)

			fd := huh.NewSelect[string]().
				Title(form.Title).
				Description(form.Description).
				Validate(func(s string) error {
					f := p.validate(form, resultManifest)

					return f(s)
				}).
				Options(opts...).
				Value(&value)

			fds = append(fds, fd)

		case v1alpha.SingleSelectNumericalInputType:
			if len(form.Choices) < 2 {
				return nil, fmt.Errorf("selection type forms must have at least 2 choices")
			}

			opts := []huh.Option[float64]{}

			for _, choice := range form.Choices {
				rawString := fmt.Sprintf("%v", choice)

				c, err := strconv.ParseFloat(rawString, 64)
				if err != nil {
					return nil, err
				}

				opts = append(opts, huh.NewOption(rawString, c))
			}

			var value float64
			resultManifest.Spec.NewEmptyResult(form.Name, &value, form.InputType)

			fd := huh.NewSelect[float64]().
				Title(form.Title).
				Description(form.Description).
				Validate(func(s float64) error {
					f := p.validate(form, resultManifest)

					return f(s)
				}).
				Options(opts...).
				Value(&value)

			fds = append(fds, fd)

		case v1alpha.MultiSelectTextInputType:
			if len(form.Choices) < 2 {
				return nil, fmt.Errorf("selection type forms must have at least 2 choices")
			}

			opts := []huh.Option[string]{}

			for _, choice := range form.Choices {
				c, ok := choice.(string)
				if !ok {
					return nil, fmt.Errorf("unable to assert type string for value `%v`", choice)
				}
				opts = append(opts, huh.NewOption(c, c))
			}

			var value []string
			resultManifest.Spec.NewEmptyResult(form.Name, &value, form.InputType)

			fd := huh.NewMultiSelect[string]().
				Title(form.Title).
				Description(form.Description).
				Validate(func(s []string) error {
					f := p.validate(form, resultManifest)

					return f(s)
				}).
				Options(opts...).
				Value(&value)

			fds = append(fds, fd)

		case v1alpha.MultiSelectNumericalInputType:
			if len(form.Choices) < 2 {
				return nil, fmt.Errorf("selection type forms must have at least 2 choices")
			}

			opts := []huh.Option[float64]{}

			for _, choice := range form.Choices {
				rawString := fmt.Sprintf("%f", choice)

				c, err := strconv.ParseFloat(rawString, 64)
				if err != nil {
					return nil, err
				}

				opts = append(opts, huh.NewOption(rawString, c))
			}

			var value []float64
			resultManifest.Spec.NewEmptyResult(form.Name, &value, form.InputType)

			fd := huh.NewMultiSelect[float64]().
				Title(form.Title).
				Description(form.Description).
				Validate(func(s []float64) error {
					f := p.validate(form, resultManifest)

					return f(s)
				}).
				Options(opts...).
				Value(&value)

			fds = append(fds, fd)

		case v1alpha.BooleanInputType:

			var value bool
			resultManifest.Spec.NewEmptyResult(form.Name, &value, form.InputType)

			fd := huh.NewSelect[bool]().Title(form.Title).
				Description(form.Description).
				Validate(func(s bool) error {
					f := p.validate(form, resultManifest)

					return f(s)
				}).
				Options(huh.NewOption("Yes", true), huh.NewOption("No", false)).
				Value(&value)

			fds = append(fds, fd)

		default:
			err := fmt.Errorf("unsupported type '%s'", form.InputType)
			entry.WithError(err).Debug("unable to generate field for form group")
			return nil, err
		}

		entry.Debugf("generate field '%s' for manifest '%s' successful", form.Name, m.Metadata.Name)
	}

	return fds, nil

}
