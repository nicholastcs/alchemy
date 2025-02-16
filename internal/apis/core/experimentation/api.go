package experimentation

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/goccy/go-yaml"
	"github.com/nicholastcs/alchemy/internal/apis/core"
	"github.com/nicholastcs/alchemy/internal/system"
	"github.com/nicholastcs/alchemy/internal/utils"
)

var registeredAccessors map[string]*accessor = map[string]*accessor{}

// accessor is the struct for accessing some low-level resource metadatas.
type accessor struct {
	core.Base
	tableHeader      []string
	tableRowCelQuery []string
	typeIdentity     func() core.ManifestPattern
	aliases          []string
}

func registerAPI(
	apiVersion, kind string,
	tableHeader, tableRowCelQuery []string,
	identity func() core.ManifestPattern,
	aliases []string,
) error {
	if apiVersion == "" {
		return errors.New("`apiVersion` cannot be empty")
	}
	if kind == "" {
		return errors.New("`kind` cannot be empty")
	}
	if len(tableHeader) != len(tableRowCelQuery) {
		return errors.New("`tableRowCelQuery` must be equal to tableHeader")
	}

	if identity == nil {
		return errors.New("`typeIdentity` function cannot be nil")
	}
	if len(aliases) == 0 {
		return errors.New("`aliases` cannot be empty")
	}

	accessor := &accessor{
		Base: core.Base{
			APIVersion: apiVersion,
			Kind:       kind,
		},
		tableHeader:      tableHeader,
		tableRowCelQuery: tableRowCelQuery,
		typeIdentity:     identity,
		aliases:          aliases,
	}
	registeredAccessors[accessor.key()] = accessor

	return nil
}

func getAccessor(apiVersion, kind string) (*accessor, error) {
	k := fmt.Sprintf("%s/%s", apiVersion, kind)
	accessor, ok := registeredAccessors[k]
	if !ok {
		return nil, fmt.Errorf("`%s/%s` is not a member of registered API accessors",
			apiVersion, kind)
	}

	return accessor, nil
}

func (r *accessor) key() string {
	return fmt.Sprintf("%s/%s", r.APIVersion, r.Kind)
}

func (r *accessor) hasDisplayAPI() bool {
	return len(r.tableHeader) > 0
}

func (r *accessor) deepValidate(m core.AbstractedManifest) (validationErr error, conversionErr error) {
	actual := r.typeIdentity()
	err := mapstructure.Decode(m, actual)
	if err != nil {
		return nil, err
	}

	return m.Validate(), nil
}

func (r *accessor) toActualManifestAny(m core.AbstractedManifest) (out any, err error) {
	out = r.typeIdentity()

	err = mapstructure.Decode(m, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *accessor) displayTableV2(ms []core.AbstractedManifest) error {
	caption := fmt.Sprintf(" * kind %s of %s has %v resource(s)", r.Kind, r.APIVersion, len(ms))
	if !r.hasDisplayAPI() {

		return nil
	}

	contents := [][]string{}
	for _, m := range ms {
		content := []string{}

		for _, query := range r.tableRowCelQuery {
			out, err := system.ExecuteCELOnManifest(m, query)
			if err != nil {
				return err
			}
			content = append(content, fmt.Sprintf("%v", out))
		}
		contents = append(contents, content)
	}

	utils.PrintTableV2(r.tableHeader, contents, caption)

	return nil
}

func (r *accessor) renderManifest(m core.AbstractedManifest) (string, error) {
	actualManifest, err := r.toActualManifestAny(m)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	encoder := yaml.NewEncoder(&b,
		yaml.UseLiteralStyleIfMultiline(true),
		yaml.IndentSequence(true),
		yaml.Indent(2),
	)
	err = encoder.Encode(actualManifest)
	if err != nil {
		return "", err
	}

	return b.String(), err
}
