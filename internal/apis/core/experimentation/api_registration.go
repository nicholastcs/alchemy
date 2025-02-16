package experimentation

import (
	"github.com/nicholastcs/alchemy/internal/apis/core"
	"github.com/nicholastcs/alchemy/internal/apis/v1alpha"
)

func init() {
	registerAPI(
		"alchemy.io/core/readonly",
		"API",
		[]string{
			"apiVersion", "kind", "aliases",
		},
		[]string{
			"spec.meta.apiVersion", "spec.meta.kind", "spec.meta.aliases",
		},
		func() core.ManifestPattern {
			return &core.RegisteredAPIManifest{}
		},
		[]string{
			"apis", "api",
		},
	)
	registerAPI(
		"alchemy.io/v1alpha",
		"Form",
		[]string{
			"namespace", "name", "fields", "confirmation-required",
		},
		[]string{
			"metadata.namespace", "metadata.name", "spec.fields.size()", "spec.confirmationRequired",
		},
		func() core.ManifestPattern {
			return &v1alpha.FormManifest{}
		},
		[]string{
			"forms", "form",
		},
	)
	registerAPI(
		"alchemy.io/v1alpha/internal",
		"FormResult",
		[]string{},
		[]string{},
		func() core.ManifestPattern {
			return &v1alpha.FormResultManifest{}
		},
		[]string{
			"formresults",
			"formresult",
			"result",
			"results",
		},
	)
	registerAPI(
		"alchemy.io/v1alpha",
		"CodeTemplate",
		[]string{
			"namespace", "name", "template-kind", "opts", "total-files",
		},
		[]string{
			"metadata.namespace", "metadata.name", "spec.kind", "spec.options", "spec.generateFiles.size()",
		},
		func() core.ManifestPattern {
			return &v1alpha.CodeTemplateManifest{}
		},
		[]string{
			"codetemplates",
			"codetemplate",
		},
	)
}
