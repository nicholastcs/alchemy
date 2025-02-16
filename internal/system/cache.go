package system

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"

	"github.com/maypok86/otter"
	"github.com/nicholastcs/alchemy/internal/apis/core"

	"github.com/sirupsen/logrus"
)

type Db struct {
	cache otter.Cache[string, core.AbstractedManifest]
	log   *logrus.Entry
}

func NewLocalDB(log *logrus.Entry) (*Db, error) {
	entry := log.WithField("context", "cache")

	cache, err := otter.MustBuilder[string, core.AbstractedManifest](100).Build()
	if err != nil {
		return nil, err
	}

	entry.Trace("cache is ready")

	return &Db{
		cache: cache,
		log:   log.WithField("context", "cache"),
	}, nil
}

func (db *Db) Set(manifest core.AbstractedManifest) error {
	hash, err := db.manifestHash(manifest.Base)
	if err != nil {
		return err
	}

	if len(manifest.Metadata.Annotations) == 0 {
		manifest.Metadata.Annotations = map[string]string{
			"alchemy.io/resourcehash": hash,
		}
	}

	// retains the previous annotation if there are any annotation(s) defined.
	if manifest.Metadata.Annotations["alchemy.io/resourcehash"] == "" {
		manifest.Metadata.Annotations["alchemy.io/resourcehash"] = hash
	}

	if len(manifest.Metadata.Namespace) == 0 {
		manifest.Metadata.Namespace = "default"
	}

	entry := db.log.WithFields(logrus.Fields{
		"action":   "add",
		"hash":     hash,
		"identity": manifest.Base,
	})

	outcome := db.cache.Set(hash, manifest)
	if !outcome {
		return errors.New("unable to add into cache")
	}

	entry.Tracef("added '%s' under namespace '%s' into cache", manifest.Metadata.Name, manifest.Metadata.Namespace)

	return nil
}

func (db *Db) SetAll(manifests []core.AbstractedManifest) error {
	for _, manifest := range manifests {
		err := db.Set(manifest)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Db) Get(apiVersion string, kind string, name string, namespace string) (*core.AbstractedManifest, error) {
	base := core.Base{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata: core.Metadata{
			Name:      name,
			Namespace: namespace,
		},
	}

	hash, err := db.manifestHash(base)
	if err != nil {
		return nil, err
	}

	entry := db.log.WithFields(logrus.Fields{
		"action":   "get",
		"hash":     hash,
		"identity": base,
	})

	value, outcome := db.cache.Get(hash)
	if !outcome {
		return nil, nil
	}

	entry.Tracef("got '%s' under namespace '%s' from cache", name, namespace)

	return &value, nil
}

func (db *Db) GetByGVKNs(apiVersion string, kind string, namespace string) ([]core.AbstractedManifest, error) {
	output := []core.AbstractedManifest{}

	db.cache.Range(
		func(k string, v core.AbstractedManifest) bool {
			if v.APIVersion == apiVersion &&
				v.Metadata.Namespace == namespace &&
				v.Kind == kind {

				output = append(output, v)
			}
			return true
		},
	)

	return output, nil
}

func (db *Db) GetByGVK(apiVersion string, kind string) ([]core.AbstractedManifest, error) {
	output := []core.AbstractedManifest{}

	db.cache.Range(
		func(k string, v core.AbstractedManifest) bool {
			if v.APIVersion == apiVersion &&
				v.Kind == kind {

				output = append(output, v)
			}
			return true
		},
	)

	return output, nil
}

func (db *Db) Dump() ([]core.AbstractedManifest, error) {
	output := []core.AbstractedManifest{}

	db.cache.Range(func(_ string, v core.AbstractedManifest) bool {
		output = append(output, v)

		return true
	})

	return output, nil
}

func (db *Db) manifestHash(manifest core.Base) (string, error) {
	if manifest.APIVersion == "" {
		return "", errors.New("apiVersion is empty")
	}
	if manifest.Kind == "" {
		return "", errors.New("kind is empty")
	}
	if manifest.Metadata.Name == "" {
		return "", errors.New("metadata.name is empty")
	}
	if manifest.Metadata.Namespace == "" {
		manifest.Metadata.Namespace = "default"
	}

	list := []string{
		manifest.APIVersion,
		manifest.Kind,
		manifest.Metadata.Namespace,
		manifest.Metadata.Name,
	}

	h := sha1.New()
	for _, str := range list {
		h.Write([]byte(str))
	}

	hashBytes := h.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString, nil
}
