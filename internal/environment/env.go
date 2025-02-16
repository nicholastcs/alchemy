package environment

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/nicholastcs/alchemy/internal/apis/core"
	"github.com/nicholastcs/alchemy/internal/apis/core/experimentation"
	"github.com/nicholastcs/alchemy/internal/system"
	"github.com/sirupsen/logrus"
)

var embed fs.FS

func PreloadEmbedFS(fs fs.FS) error {
	if fs == nil {
		return errors.New("embed fs cannot be empty")
	}

	embed = fs

	return nil
}

func New(log *logrus.Entry) (*system.Db, error) {
	c := log.WithField("context", "init")

	db, err := system.NewLocalDB(log)
	if err != nil {
		return nil, err
	}

	// retrieve allowed APIs
	allowedApis := experimentation.AllowedAPIs()

	loader, err := system.NewFileLoader(embed, allowedApis, log)
	if err != nil {
		return nil, err
	}

	manifests, err := loader.GetFiles("embed")
	if err != nil {
		return nil, err
	}

	// load API metadata too...
	metas, err := experimentation.GetAPIsMetadata()
	if err != nil {
		return nil, err
	}
	manifests = append(manifests, metas...)

	for i, m := range manifests {
		mLog := c.WithField("resource", m.Base)

		mErr, conversionErr := experimentation.DeepValidate(m)
		if conversionErr != nil {
			return nil, fmt.Errorf("%s/%s %s of namespace '%s' :%w",
				m.APIVersion, m.Kind, m.Metadata.Name, m.Metadata.Namespace, conversionErr)
		}
		if mErr != nil {
			manifests[i].Status.SetError(mErr)

			mLog.WithError(mErr).Infof("resource %s under namespace %s has error", m.Metadata.Name, m.Metadata.Namespace)
		}
		manifests[i].Status.SetCondition(core.ResourceReady, mErr == nil)

		mLog.Infof("resource %s under namespace %s is ready", m.Metadata.Name, m.Metadata.Namespace)
	}

	err = db.SetAll(manifests)
	if err != nil {
		return nil, err
	}

	return db, nil
}
