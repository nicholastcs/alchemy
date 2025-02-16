package system

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var file1 string = `
apiVersion: alchemy.io/v1alpha
kind: CodeTemplate
metadata:
  name: test-template
  namespace: terraform
spec:
  kind: go-template
  options:
    - missingkey=error
    - funcs=sprig
  generateFiles:
    - file: terraform/version.tf
      template: |
        terraform {
          required_providers {
            aws = {
              source  = "hashicorp/aws"
              version = "5.72.1"
            }
          }
        }

        provider "aws" {
          region = "{{ .region }}"
        }
    - file: terraform/main.tf
      template: |
        module "iam_acc" {
          source  = "terraform-aws-modules/iam/aws//modules/iam-account"
          version = "5.46.0"

          account_alias = "{{ .account_alias }}"
        }
`
var file2 string = `
apiVersion: alchemy.io/v1alpha
kind: IllegalKind
metadata:
  name: test-form
spec:
  confirmationRequired: true
  fields:
    - name: name
      title: Name
      description: Name of service
      inputType: text
      constraint:
        cel:
          expressions:
            - message: length of name must be greater than 0.
              value: this.size() > 0
`

func TestNewFileLoaderFull(t *testing.T) {
	uFs := afero.NewMemMapFs()

	err := uFs.Mkdir("embed/", 0755)
	require.NoError(t, err)

	err = afero.WriteFile(uFs, "embed/code-template.yaml", []byte(file1), 0644)
	require.NoError(t, err)

	err = afero.WriteFile(uFs, "embed/user-form.yaml", []byte(file2), 0644)
	require.NoError(t, err)

	provider, err := NewFileLoader(afero.NewIOFS(uFs), []string{
		"alchemy.io/v1alpha/CodeTemplate",
		"alchemy.io/v1alpha/Form",
	}, logT)
	require.NoError(t, err)
	require.NotNil(t, provider, "mock test should sufficient to test the struct")

	manifests, err := provider.GetFiles("embed")
	require.NoError(t, err)

	// file2 has illegal Kind, it will be silently discarded, thus the
	// actual manifest count is 1
	assert.Equal(t, len(manifests), 1)
}
