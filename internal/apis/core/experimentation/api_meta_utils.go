package experimentation

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/nicholastcs/alchemy/internal/apis/core"
	"github.com/nicholastcs/alchemy/internal/system"
	"github.com/nicholastcs/alchemy/internal/utils"
)

// DeepValidate validates manifest by convert input abstracted manifest to
// its underlying type.
func DeepValidate(m core.AbstractedManifest) (validationErr error, innerErr error) {
	accessor, err := getAccessor(m.APIVersion, m.Kind)
	if err != nil {
		return nil, err
	}

	return accessor.deepValidate(m)
}

// ToActualManifest is a generic function that converts abstracted manifest
// into actual manifest of the specific type, governed by its apiVersion and kind.
//
// This requires the type to be known at compile-time.
func ToActualManifest[T core.ManifestPattern](m core.AbstractedManifest) (out T, err error) {
	accessor, err := getAccessor(m.APIVersion, m.Kind)
	if err != nil {
		return *new(T), err
	}

	output, err := accessor.toActualManifestAny(m)
	if err != nil {
		return *new(T), err
	}

	out, ok := output.(T)
	if !ok {
		return *new(T), errors.New("unable to assert type to T")
	}

	return out, nil
}

// DisplaySingle is a function that displays the YAML of the resource to
// the terminal.
func DisplaySingle(m core.AbstractedManifest) error {
	accessor, err := getAccessor(m.APIVersion, m.Kind)
	if err != nil {
		return err
	}

	out, err := accessor.renderManifest(m)
	if err != nil {
		return err
	}

	fmt.Printf("%v", out)

	return nil
}

// DisplayTable is a function that displays lists of resource to the
// terminal in the form of table. The field of the rows is queried by the
// CEL expression of the accessor itself.
func DisplayTable(ms []core.AbstractedManifest) error {
	apiVersion := ms[0].APIVersion
	kind := ms[0].Kind
	accessor, err := getAccessor(apiVersion, kind)
	if err != nil {
		return err
	}

	err = accessor.displayTableV2(ms)
	if err != nil {
		return err
	}

	return nil
}

// DisplayMultipleManifests is a sugar function that prints multiple
// abstracted manifests in one go.
//
// Different YAML documents are delimited with --- separator.
func DisplayMultipleManifests(ms ...core.AbstractedManifest) error {
	contents := []string{}

	for _, m := range ms {
		accessor, err := getAccessor(m.APIVersion, m.Kind)
		if err != nil {
			return err
		}

		mString, err := accessor.renderManifest(m)
		if err != nil {
			return err
		}
		contents = append(contents, mString)
	}
	utils.ContentsPrinter("---", contents...)

	return nil
}

// Kinds return allowed kinds of first element in the aliases slice.
func Kinds() []string {
	kinds := []string{}

	for _, accessor := range registeredAccessors {
		kinds = append(kinds, accessor.aliases[0])
	}

	slices.Sort(kinds)

	return kinds
}

// ToFormalApiVersionKind is a function converts the alias into API
// Version and Kind.
//
// Invalid alias will return 2 empty string and a non-nil error.
func ToFormalApiVersionKind(alias string) (apiVersion, kind string, err error) {
	for _, accessor := range registeredAccessors {
		if slices.Contains(accessor.aliases, strings.ToLower(alias)) {
			return accessor.APIVersion, accessor.Kind, nil
		}
	}

	return "", "", fmt.Errorf("invalid resource alias called '%s'", alias)
}

// GetAPIsMetadata is a function that retrieves API metadata list from the
// versioned package underlying APIs.
func GetAPIsMetadata() ([]core.AbstractedManifest, error) {
	ref := registeredAccessors["alchemy.io/core/readonly/API"]

	apis := []core.AbstractedManifest{}
	for _, meta := range registeredAccessors {
		registeredApi := core.RegisteredAPIManifest{
			Base: core.Base{
				APIVersion: ref.APIVersion,
				Kind:       ref.Kind,
				Metadata: core.Metadata{
					Name:      fmt.Sprintf("mt-%s", strings.ToLower(meta.Kind)),
					Namespace: "default",
				},
			},
			Spec: core.RegisteredAPISpec{
				Meta: core.API{
					APIVersion: meta.APIVersion,
					Kind:       meta.Kind,
					Aliases:    meta.aliases,
				},
			},
		}

		m, err := core.ConvertToAbstractedManifest(registeredApi)
		if err != nil {
			return nil, err
		}

		apis = append(apis, *m)
	}

	return apis, nil
}

// AllowedAPIs is a function that returns list of allowed APIs.
func AllowedAPIs() []string {
	output := []string{}
	for apiUid := range registeredAccessors {
		if strings.HasPrefix(apiUid, "alchemy.io/core/readonly") {
			continue
		}
		output = append(output, apiUid)
	}

	return output
}

// Get is a generic function where it retrieves data from the system
// database. Used when the type of the output is known.
func Get[T core.ManifestPattern](
	db *system.Db, apiVersion, kind, name, namespace string,
) (out T, err error) {

	in, err := db.Get(apiVersion, kind, name, namespace)
	if err != nil {
		return out, err
	}

	err = mapstructure.Decode(in, &out)
	if err != nil {
		return out, err
	}

	return out, nil
}

// DisplayAllResourcesInTable is a function that will dump all resources
// into the terminal in a table form.
func DisplayAllResourcesInTable(db *system.Db) error {
	for _, accessor := range registeredAccessors {
		apiVersion := accessor.APIVersion
		kind := accessor.Kind

		abstractedResources, err := db.GetByGVK(apiVersion, kind)
		if err != nil {
			return err
		}

		err = accessor.displayTableV2(abstractedResources)
		if err != nil {
			return err
		}
	}

	return nil
}
