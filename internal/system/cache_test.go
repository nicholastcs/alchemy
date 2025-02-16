package system

import (
	"fmt"
	"testing"

	"github.com/nicholastcs/alchemy/internal/apis/core"
	"github.com/nicholastcs/alchemy/internal/utils"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var logT *logrus.Entry = utils.NewLogger()

func TestCacheNewLocalDB(t *testing.T) {
	u, err := NewLocalDB(logT)

	require.NoError(t, err, "must not emit error")
	assert.NotNil(t, u, "cache must not be nil")
}

func TestCacheSetThenGetWithCheckingFurnishedFields(t *testing.T) {
	u, _ := NewLocalDB(logT)

	testManifest := core.AbstractedManifest{
		Base: core.Base{
			APIVersion: "alchemy.io/mock",
			Kind:       "Test",
			Metadata: core.Metadata{
				Name: "testing",
			},
		},
	}

	err := u.Set(testManifest)
	require.NoError(t, err, "must can set data")

	out, err := u.Get("alchemy.io/mock", "Test", "testing", "default")
	require.NoError(t, err, "must get set data")

	assert.Equal(t, out.Metadata.Namespace, "default", "namespace must be default")
	assert.NotEmpty(t, out.Metadata.Annotations["alchemy.io/resourcehash"], "must annotated with resource hash")
}

func TestCacheMultipleSetGetWithoutCheckingFurnishedFields(t *testing.T) {
	u, _ := NewLocalDB(logT)

	count := 50

	manifests := []core.AbstractedManifest{}

	for i := 0; i < count; i++ {
		name := fmt.Sprintf("testing-%s", lo.RandomString(8, lo.AlphanumericCharset))
		manifests = append(manifests, core.AbstractedManifest{
			Base: core.Base{
				APIVersion: "alchemy.io/mock",
				Kind:       "Test",
				Metadata: core.Metadata{
					Name: name,
				},
			},
		})
	}

	err := u.SetAll(manifests)

	require.NoError(t, err, "must can set data")

	out, err := u.GetByGVKNs("alchemy.io/mock", "Test", "default")
	require.NoError(t, err, "must can get data by GVK")

	assert.Equal(t, len(out), count, "count must be equal")
}
