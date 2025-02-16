package system

import (
	"fmt"
	"io/fs"
	"slices"

	"github.com/nicholastcs/alchemy/internal/apis/core"

	"github.com/goccy/go-yaml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type staticManifestSourceProvider struct {
	fs          fs.FS
	allowedAPIs []string
	log         *logrus.Entry
}

func NewFileLoader(fs fs.FS, allowedAPIs []string, log *logrus.Entry) (*staticManifestSourceProvider, error) {
	return &staticManifestSourceProvider{
		fs:          fs,
		allowedAPIs: allowedAPIs,
		log:         log.WithField("context", "readManifests"),
	}, nil
}

func (p *staticManifestSourceProvider) GetFiles(uri string) ([]core.AbstractedManifest, error) {
	iofs := afero.FromIOFS{
		FS: p.fs,
	}

	output := []core.AbstractedManifest{}

	// TODO: we take in "embed" as a standardized directory for now
	fileInfos, err := afero.ReadDir(iofs, uri)
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			in, err := afero.ReadFile(iofs, uri+"/"+fileInfo.Name())
			if err != nil {
				return nil, err
			}

			var out core.AbstractedManifest
			err = yaml.Unmarshal(in, &out)
			if err != nil {
				return nil, err
			}

			// check for API, discard silently if they are not relevant
			api := fmt.Sprintf("%s/%s", out.APIVersion, out.Kind)
			isAllowed := slices.Contains(p.allowedAPIs, api)
			if !isAllowed {
				p.log.WithField("illegalAPI", api).Warnf("illegal API '%s' found, skipping...", api)
				continue
			}

			out.SetFilePath(fileInfo.Name())

			output = append(output, out)
		}
	}

	return output, nil
}

func (p *staticManifestSourceProvider) Name() string {
	return "static"
}
