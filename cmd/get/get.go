package get

import (
	"errors"
	"fmt"

	"github.com/nicholastcs/alchemy/internal/apis/core/experimentation"
	"github.com/nicholastcs/alchemy/internal/system"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var errArg error = errors.New(`the get command can receive either one of:
 * one arg (that you are query resources of same kind),
 * two args (that you are query a single resource) or,
 * zero if and only if all flag is toggled to true`,
)

func NewCommandV2(db *system.Db, log *logrus.Entry) *cobra.Command {
	var (
		//displayMode string
		all bool
	)

	getCmd := &cobra.Command{
		Use:   "get <alias | kind> [resource-name]",
		Short: "To list or view the underlying APIs resources.",
		Args: func(cmd *cobra.Command, args []string) error {
			argLen := len(args)

			querySingle := (argLen == 1 || argLen == 2) && !all
			queryAll := argLen == 0 && all
			validArg := querySingle || queryAll

			if validArg {
				return nil
			}

			return errArg
		},
		SilenceErrors: true,
		SilenceUsage:  true,
		Example: fmt.Sprintf(`Allowed kinds: 
  %s

To retrieve available API resources:
  alchemy get apis

To retrieve API of kind 'form' with name 'k8s-deployment':
  alchemy get forms k8s-deployment

To retrieve APIs of kind 'codeTemplate':
  alchemy get codetemplates`, experimentation.Kinds()),

		RunE: func(cmd *cobra.Command, args []string) error {
			// retrieve all resources
			if len(args) == 0 && all {
				err := experimentation.DisplayAllResourcesInTable(db)
				if err != nil {
					return err
				}

				return nil
			}

			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				return err
			}

			apiVersion, kind, err := experimentation.ToFormalApiVersionKind(args[0])
			if err != nil {
				return err
			}

			// retrieve list of same API
			if len(args) == 1 {
				manifests, err := db.GetByGVKNs(apiVersion, kind, namespace)
				if err != nil {
					return err
				}

				if len(manifests) == 0 {
					return errors.New("resource not found")
				}

				err = experimentation.DisplayTable(manifests)
				if err != nil {
					return err
				}
			}

			// retrieve single item
			if len(args) == 2 {
				name := args[1]

				manifest, err := db.Get(apiVersion, kind, name, namespace)
				if err != nil {
					return err
				}

				if manifest == nil {
					return fmt.Errorf("no resource available under the namespace `%s`", namespace)
				}

				err = experimentation.DisplaySingle(*manifest)
				if err != nil {
					return err
				}

				if manifest.Status.HasErr() {
					return fmt.Errorf("manifest has error: %w", manifest.Status.ToNativeErr())
				}
			}

			return nil
		},
	}

	// TODO: need to relook of the display mode
	//getCmd.Flags().StringVar(&displayMode, "display", "default", "display mode for the resource")

	getCmd.Flags().BoolVarP(&all, "all", "A", false, "retrieve all values")

	return getCmd
}
