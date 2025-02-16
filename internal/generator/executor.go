package generator

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/nicholastcs/alchemy/internal/apis/core"
	"github.com/nicholastcs/alchemy/internal/apis/v1alpha"
	"github.com/nicholastcs/alchemy/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type v1alphaTemplateExecutor struct {
	log *logrus.Entry
}

func NewExecutor(log *logrus.Entry) (*v1alphaTemplateExecutor, error) {
	return &v1alphaTemplateExecutor{
		log: log.WithField("context", "executor/v1alpha"),
	}, nil
}

var errCodeTemplateConsumtptionNotReady = errors.New("found condition CodeTemplateConsumptionReady is false")
var errResourceNotReady = errors.New("found condition ResourceReady is false")

func (g *v1alphaTemplateExecutor) Generate(r *v1alpha.FormResultManifest, t *v1alpha.CodeTemplateManifest) error {
	log := g.log.WithFields(logrus.Fields{
		"targets": []interface{}{
			r, t,
		},
	})

	if !r.Status.GetCondition(v1alpha.CodeTemplateConsumptionReady) {
		log.Error("code template is not ready")
		return errCodeTemplateConsumtptionNotReady
	}

	if !t.Status.GetCondition(core.ResourceReady) {
		log.Error("resource is not ready")
		return errResourceNotReady
	}

	t.Status.GeneratedCodeFiles = []v1alpha.CodeTemplateStatusResult{}

	// TODO: do CEL preflight check!
	tmpl := template.New("alchemy-main")

	templateKind := t.Spec.Kind
	switch templateKind {
	case "go-template":
		for _, opt := range t.Spec.Options {
			switch opt {
			case "funcs=sprig":
				tmpl = tmpl.Funcs(sprig.FuncMap())
			case "missingkey=error":
				tmpl = tmpl.Option("missingkey=error")
			default:
				log.Panicf("invalid opt found %s, looks like it is a bug and it must be vetted early", opt)
			}

			log.Tracef("template option '%s' added", opt)
		}
	default:
		log.Panicf("invalid template kind found %s, looks like it is a bug and it must be vetted early", templateKind)
	}

	// TODO: normalise value before hand, based on type hints
	//
	// should normalise *bool into bool if not it will not be
	// registered properly in go-template
	for name := range r.Spec.Result {
		// if r.Spec.TypeHintByResult[name] == "boolean" {
		// 	// g, ok := r.Spec.Result[name].(*bool)
		// 	// if ok {
		// 	// 	r.Spec.Result[name] = *g
		// 	// } else {
		// 	// 	log.WithField("hint", r.Spec.TypeHintByResult).
		// 	// 		WithField("field_name", name).
		// 	// 		WithField("field_value", v).
		// 	// 		Panic("unable to cast type on type hint, it is a bug")
		// 	// }
		// }

		val := reflect.Indirect(reflect.ValueOf(r.Spec.Result[name])).Interface()
		r.Spec.Result[name] = val
	}

	// TODO: support multiple templating engine
	for _, g := range t.Spec.GenerateFiles {
		output, err := tmpl.Parse(g.Template)
		if err != nil {
			return err
		}

		var code bytes.Buffer

		err = output.Execute(&code, r.Spec.Result)
		if err != nil {
			return err
		}

		t.Status.GeneratedCodeFiles = append(t.Status.GeneratedCodeFiles, v1alpha.CodeTemplateStatusResult{
			File: g.File,
			Code: code.String(),
		})
	}

	t.Status.SetCondition(v1alpha.CodeTemplateConsumptionDone, true)

	return nil
}

var errCodeTemplateConsumptionNotDone = errors.New("found condition CodeTemplateConsumptionDone is false")

func (g *v1alphaTemplateExecutor) MakeFiles(dir string, s *v1alpha.CodeTemplateStatus) error {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	log := g.log.WithFields(logrus.Fields{
		"absPath": absPath,
	})

	if !s.GetCondition(v1alpha.CodeTemplateConsumptionDone) {
		return errCodeTemplateConsumptionNotDone
	}

	fs := afero.NewOsFs()
	for _, f := range s.GeneratedCodeFiles {
		fileDirectory := path.Join(absPath, filepath.Dir(f.File))

		err := fs.MkdirAll(fileDirectory, os.ModePerm)
		if err != nil {
			s.SetError(fmt.Errorf("unable to make directory '%s': %w", fileDirectory, err))

			return err
		}

		log.WithField("fileDirectory", fileDirectory).Tracef("make directory '%s' done", fileDirectory)
		absDir := filepath.Join(fileDirectory, filepath.Base(f.File))

		// redo everything
		_ = fs.Remove(absDir)

		// should the code block is empty - silently ignore
		if strings.TrimSpace(f.Code) == "" {
			continue
		}

		file, err := fs.Create(absDir)
		if err != nil {
			return err
		}

		defer file.Close()

		_, err = file.WriteString(f.Code)
		if err != nil {
			return err
		}
	}

	utils.Tell("✨ Code Generated", fmt.Sprintf("Alchemy has created code into '%s' from forms without any issues.", dir))

	return nil
}
