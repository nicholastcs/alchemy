package run

import (
	"errors"

	"github.com/nicholastcs/alchemy/internal/apis/core"
	"github.com/nicholastcs/alchemy/internal/apis/core/experimentation"
	"github.com/nicholastcs/alchemy/internal/apis/v1alpha"
	"github.com/nicholastcs/alchemy/internal/formcreator"
	"github.com/nicholastcs/alchemy/internal/generator"
	"github.com/nicholastcs/alchemy/internal/system"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCommandV2(db *system.Db, log *logrus.Entry) *cobra.Command {
	var (
		codeTemplateName string
		preview          bool
		dir              string
	)

	runCmd := &cobra.Command{
		Use:           "run <form-name> -t|--codetemplate=<code-template-name>",
		Short:         "To execute the user form to generate IAC from code templates.",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		Aliases:       []string{"transmute"},

		RunE: func(cmd *cobra.Command, args []string) error {
			// flags and args
			formName := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				return err
			}
			codeTemplateName, err := cmd.Flags().GetString("codetemplate")
			if err != nil {
				return err
			}

			// canonicalize aliases to actual apiversion & upKind
			upApiVersion, upKind, err := experimentation.ToFormalApiVersionKind("forms")
			if err != nil {
				return err
			}
			ctApiVersion, ctKind, err := experimentation.ToFormalApiVersionKind("codetemplates")
			if err != nil {
				return err
			}

			formManifestActual, err := experimentation.Get[*v1alpha.FormManifest](
				db, upApiVersion, upKind, formName, namespace)
			if err != nil {
				return err
			}
			if formManifestActual == nil {
				return errors.New("form not found")
			}
			if formManifestActual.Status.HasErr() {
				return formManifestActual.Status.ToNativeErr()
			}

			// form
			p, err := formcreator.NewFormCreatorV1Alpha(log)
			if err != nil {
				return err
			}
			result, err := p.Run(*formManifestActual)
			if err != nil {
				return err
			}

			err = result.Spec.ConvertResultToNative()
			if err != nil {
				return err
			}

			// retrieve and store to environment db
			resultAbstracted, err := core.ConvertToAbstractedManifest(result)
			if err != nil {
				return err
			}
			err = db.Set(*resultAbstracted)
			if err != nil {
				return err
			}

			// generate the code
			g, err := generator.NewExecutor(log)
			if err != nil {
				return err
			}
			ctManifestActual, err := experimentation.Get[*v1alpha.CodeTemplateManifest](
				db, ctApiVersion, ctKind, codeTemplateName, namespace)
			if err != nil {
				return err
			}

			// TODO: do a dry run before allow form!
			err = g.Generate(result, ctManifestActual)
			if err != nil {
				return err
			}

			if preview {
				ctManifest, err := core.ConvertToAbstractedManifest(ctManifestActual)
				if err != nil {
					return err
				}

				err = experimentation.DisplayMultipleManifests(*resultAbstracted, *ctManifest)
				if err != nil {
					return err
				}
			} else {
				err = g.MakeFiles(dir, &ctManifestActual.Status)
				if err != nil {
					return err
				}
			}

			// persist everything into environment for
			// troubleshooting, whereas it can be dumped (using
			// --dump flag) for analysis.
			ctManifest, err := core.ConvertToAbstractedManifest(ctManifestActual)
			if err != nil {
				return err
			}

			resultAbstracted, err = core.ConvertToAbstractedManifest(result)
			if err != nil {
				return err
			}

			err = db.SetAll(
				[]core.AbstractedManifest{
					*ctManifest, *resultAbstracted,
				},
			)
			if err != nil {
				return err
			}

			return err
		},
	}

	runCmd.Flags().StringVarP(&codeTemplateName, "codetemplate", "t", "", "display mode for the resource")
	runCmd.Flags().BoolVarP(&preview, "preview", "p", false, "preview the outcome in YAML form only")
	runCmd.Flags().StringVar(&dir, "dir", "./", "directory of the code generated")
	runCmd.MarkFlagRequired("codetemplate")

	return runCmd
}
