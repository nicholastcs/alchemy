package v1alpha

import (
	"errors"
	"fmt"

	"github.com/dustin/go-humanize/english"
	"github.com/nicholastcs/alchemy/internal/apis/core"
	"github.com/samber/lo"
)

var (
	optskindMap = map[string][]string{
		"go-template": {
			"funcs=sprig",
			"missingkey=error",
		},
	}
)

const CodeTemplateConsumptionReady string = "CodeTemplateConsumptionReady"
const CodeTemplateConsumptionDone string = "CodeTemplateConsumptionDone"

type CodeTemplateManifest struct {
	core.Base `yaml:",inline" mapstructure:",squash"`
	Spec      CodeTemplateSpec   `mapstructure:"spec" yaml:"spec" json:"spec"`
	Status    CodeTemplateStatus `mapstructure:"status,omitempty" yaml:"status,omitempty" json:"status,omitempty"`
}

type CodeTemplateSpec struct {
	Kind          string         `mapstructure:"kind" yaml:"kind" json:"kind"`
	Options       []string       `mapstructure:"options" yaml:"options" json:"options"`
	GenerateFiles []GenerateFile `mapstructure:"generateFiles" yaml:"generateFiles" json:"generateFiles"`
}

type GenerateFile struct {
	File     string `mapstructure:"file" yaml:"file" json:"file"`
	Template string `mapstructure:"template" yaml:"template" json:"template"`
}

type CodeTemplateStatus struct {
	core.Status        `mapstructure:",squash" yaml:",inline"`
	GeneratedCodeFiles []CodeTemplateStatusResult `mapstructure:"result" yaml:"result" json:"result"`
}

type CodeTemplateStatusResult struct {
	File string `mapstructure:"file" yaml:"file" json:"file"`
	Code string `mapstructure:"code" yaml:"code" json:"code"`
}

func (m *CodeTemplateManifest) Validate() error {
	var errs error
	errs = errors.Join(errs, m.Base.Validate())
	errs = errors.Join(errs, validateCodeTemplateManifest(m.Spec))

	return errs
}

func validateCodeTemplateManifest(spec CodeTemplateSpec) error {
	var errs error

	hashes := map[string]bool{}
	for i, f := range spec.GenerateFiles {
		ok := hashes[f.File]
		if ok {
			errs = errors.Join(
				errs, core.NewPathError(
					fmt.Sprintf("spec.generatedFiles[%v].file", i),
					errors.New("generated file has duplicate file name"),
				),
			)
		}
		hashes[f.File] = true
	}

	isUniqueOpts := len(lo.Uniq(spec.Options)) == len(spec.Options)
	if !isUniqueOpts {
		errs = errors.Join(errs, core.NewPathError("spec.options", errors.New("template options are not unique")))
	}

	allowedOpts, err := getAllowedOpts(spec.Kind)
	if err != nil {
		errs = errors.Join(errs, core.NewPathError("spec.kind", err))

		// return early due to no allowed options for the invalid
		// templating kind
		return errs
	}

	validOpts := len(spec.Options) == 0 || lo.Every(allowedOpts, spec.Options)
	if !validOpts {
		pathedErr := core.NewPathError("spec.options",
			fmt.Errorf("template options are not valid, please select template options under %s",
				english.OxfordWordSeries(allowedOpts, "or"),
			),
		)

		errs = errors.Join(errs, pathedErr)
	}

	return errs
}

func getAllowedOpts(templateKind string) ([]string, error) {
	opts, ok := optskindMap[templateKind]
	if !ok {
		return nil, fmt.Errorf("illegal template kind, please select template kind under %s",
			english.OxfordWordSeries(lo.Keys(optskindMap), "or"),
		)
	}

	return opts, nil
}
