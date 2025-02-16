package system

import (
	"testing"

	"github.com/nicholastcs/alchemy/internal/apis/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteCELOnManifest(t *testing.T) {
	testCases := []struct {
		input         core.AbstractedManifest
		celExpression string
		output        any
	}{
		{
			input: core.AbstractedManifest{
				Base: core.Base{
					APIVersion: "alchemy.io/mock",
					Kind:       "Test",
					Metadata: core.Metadata{
						Name: "testing-hello-world",
					},
				},
			},
			celExpression: "metadata.name",
			output:        "testing-hello-world",
		},
		{
			input: core.AbstractedManifest{
				Spec: map[string]interface{}{
					"hello": "world",
				},
			},
			celExpression: "spec.hello",
			output:        "world",
		},
		{
			input: core.AbstractedManifest{
				Spec: map[string]interface{}{
					"classes": []string{
						"onion knight",
						"black mage",
						"white mage",
					},
				},
			},
			celExpression: "spec.classes",
			output: []string{
				"onion knight",
				"black mage",
				"white mage",
			},
		},
	}

	for _, u := range testCases {
		val, err := ExecuteCELOnManifest(u.input, u.celExpression)

		require.NoError(t, err, "evaluate CEL expression on the input must pass")
		assert.Equal(t, u.output, val.Value())
	}
}

func TestExecuteCELOnFormValidation(t *testing.T) {
	testCases := []struct {
		input         map[string]interface{}
		celExpression string
		output        bool
	}{
		{
			input: map[string]interface{}{
				"this": []int{
					8080,
					8443,
					443,
					80,
				},
			},
			celExpression: "this.size() == 4",
			output:        true,
		},
		{
			input: map[string]interface{}{
				"this": "50Mi",
			},
			celExpression: `quantity(this) == quantity("50Mi")`,
			output:        true,
		},
		{
			input: map[string]interface{}{
				"this": "1000m",
			},
			celExpression: `quantity(this) < quantity("500m")`,
			output:        false,
		},

		// evaluates based on previously filled value based on
		// `result` key
		{
			// mimics for `this` is `maximum_concurrency`
			input: map[string]interface{}{
				"this": 20000,
				"result": map[string]interface{}{
					"minimum_concurrency": 30000,
				},
			},
			celExpression: `this > result.minimum_concurrency`,
			output:        false,
		},
		{
			// mimics for `this` is `cpu_limit`.
			input: map[string]interface{}{
				"this": "750m",
				"result": map[string]interface{}{
					"cpu_request": "500m",
				},
			},
			celExpression: `quantity(this) > quantity(result.cpu_request)`,
			output:        true,
		},
		{
			// mimics for `this` is `memory_limit`.
			input: map[string]interface{}{
				"this": "512M",
				"result": map[string]interface{}{
					"memory_request": "512M",
				},
			},
			celExpression: `quantity(this) == quantity(result.memory_request)`,
			output:        true,
		},
		{
			// mimics for `this` is `memory_limit`.
			input: map[string]interface{}{
				"this": "512M",
				"result": map[string]interface{}{
					"memory_request": "512M",
				},
			},
			celExpression: `quantity(this) == quantity(result.memory_request)`,
			output:        true,
		},
	}

	for _, u := range testCases {
		ok, err := ExecuteCELOnFormValidation(u.input, u.celExpression)

		require.NoError(t, err)
		assert.Equal(t, u.output, ok)
	}
}
