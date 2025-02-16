package docs

import (
	"fmt"
	"strings"

	_ "embed"

	"github.com/charmbracelet/glamour"
	"github.com/dustin/go-humanize/english"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

//go:embed docs.overview.MD
var overview string

//go:embed docs.getting-started.MD
var gettingStarted string

func NewCommand() *cobra.Command {

	var (
		docsByName = map[string]string{
			"overview":        overview,
			"getting-started": gettingStarted,
		}
	)

	allowedDocs := strings.Join(lo.Keys(docsByName), "|")

	usage := fmt.Sprintf("docs (%s)", allowedDocs)

	docsCmd := &cobra.Command{
		Use:   usage,
		Short: "To view documentations.",
		Args: func(cmd *cobra.Command, args []string) error {
			var errWithSuggestion = fmt.Errorf("supported arg(s) either %s", english.OxfordWordSeries(lo.Keys(docsByName), "or"))

			err := cobra.ExactArgs(1)(cmd, args)
			if err != nil {
				return fmt.Errorf("%w\n\n%w", err, errWithSuggestion)
			}

			_, ok := docsByName[args[0]]
			if !ok {
				return errWithSuggestion
			}

			return nil
		},
		SilenceErrors:         true,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			in := docsByName[args[0]]

			r, _ := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(90),
				glamour.WithPreservedNewLines(),
			)
			out, err := r.Render(in)
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}

	return docsCmd
}
