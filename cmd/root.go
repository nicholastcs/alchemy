package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/nicholastcs/alchemy/cmd/docs"
	"github.com/nicholastcs/alchemy/cmd/get"
	"github.com/nicholastcs/alchemy/cmd/run"
	"github.com/nicholastcs/alchemy/internal/apis/core"
	"github.com/nicholastcs/alchemy/internal/apis/core/experimentation"
	"github.com/nicholastcs/alchemy/internal/environment"

	"github.com/nicholastcs/alchemy/internal/system"
	"github.com/nicholastcs/alchemy/internal/utils"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	log *logrus.Entry
	db  system.Db

	dump bool

	cfgFile   string
	logLevel  string
	namespace string
)

var ascii = ` 
      _/_/    _/          _/_/_/  _/    _/  _/_/_/_/  _/      _/  _/      _/
   _/    _/  _/        _/        _/    _/  _/        _/_/  _/_/    _/  _/   
  _/_/_/_/  _/        _/        _/_/_/_/  _/_/_/    _/  _/  _/      _/      
 _/    _/  _/        _/        _/    _/  _/        _/      _/      _/       
_/    _/  _/_/_/_/    _/_/_/  _/    _/  _/_/_/_/  _/      _/      _/        `

var rootCmd = &cobra.Command{
	Use: "alchemy",
	Short: ascii + "\n\nAlchemy, an platform-agnostic template CLI.\n\n" +
		"Alchemy is a command-line interface that enables handoff of the\n" +
		"Golden Pattern codes from the Platform engineers to the\n" +
		"software developers.\n\n" +
		"It is to empowers Platform engineers to deliver no-frills\n" +
		"fill-the-form generated IAC without managing a full fledged\n" +
		"Internal Developer Portal.",

	DisableSuggestions: false,
	CompletionOptions:  cobra.CompletionOptions{DisableDefaultCmd: true},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		d, err := environment.New(log)

		if err != nil {
			return err
		}
		db = *d

		return nil
	},
}

func Execute() error {
	log = utils.NewLogger()

	rootCmd.AddCommand(get.NewCommandV2(&db, log))
	rootCmd.AddCommand(run.NewCommandV2(&db, log))
	rootCmd.AddCommand(docs.NewCommand())

	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initLogLevel, initConfig)
	cobra.OnFinalize(
		func() {
			if !dump {
				return
			}

			dir, err := dumpEnv()
			if err != nil {
				utils.Error("Error occurred during dumping", err)

				return
			}

			utils.Warning(
				"Dump enabled",
				fmt.Sprintf("dump of environment resources available in '%s'", dir),
			)
		},
	)

	rootCmd.DisableSuggestions = false
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
	rootCmd.SilenceUsage = true

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.alchemy.yaml)")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "namespace of manifest")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "text logging level")
	rootCmd.PersistentFlags().BoolVarP(&dump, "dump", "d", false, "dump object to CHDIR after CLI completion (if applicable)")

}

func initConfig() {
	log.WithField("args", os.Args).Tracef("running with raw arguments '%s'", strings.Join(os.Args, " "))

	// For now it will be disabled for this MVP
	log.Warn("config file API is not loaded for now")

	// dir, err := os.UserHomeDir()
	// if err != nil {
	// 	log.WithError(err).Panic("unable to determine home directory")
	// }
	// if cfgFile == "" {
	// 	cfgFile = dir + "/.alchemy.yaml"
	// }

	// log.Debugf("the directory pointed to %s", cfgFile)
}

func initLogLevel() {
	// CLI flag --log-level would take the first precedence, else would
	// take Environment Variable ALCHEMY_LOG_LEVEL for log level.
	//
	// If both are unset, it will be ultimately in ERROR log level.
	l := lo.CoalesceOrEmpty(logLevel, os.Getenv("ALCHEMY_LOG_LEVEL"))

	switch l {
	case "WARN", "WRN":
		log.Logger.SetLevel(logrus.WarnLevel)
	case "INFO", "INF":
		log.Logger.SetLevel(logrus.InfoLevel)
	case "DEBUG", "DBG":
		log.Logger.SetLevel(logrus.DebugLevel)
	case "TRACE":
		log.Logger.SetLevel(logrus.TraceLevel)
	default:
		log.Logger.SetLevel(logrus.ErrorLevel)
	}
}

func dumpEnv() (dir string, err error) {
	dirName := fmt.Sprintf("./alchemy-dump-%s", time.Now().Format(time.RFC3339))

	fs := afero.NewOsFs()

	err = fs.Mkdir(dirName, os.ModePerm)
	if err != nil {
		return "", err
	}

	manifests, err := db.Dump()
	for _, manifest := range manifests {
		c, err := experimentation.ToActualManifest[core.ManifestPattern](manifest)
		if err != nil {
			return "", err
		}

		var b bytes.Buffer
		encoder := yaml.NewEncoder(&b,
			yaml.UseLiteralStyleIfMultiline(true),

			// for readability
			yaml.IndentSequence(true),
			yaml.Indent(2),
		)

		err = encoder.Encode(c)
		if err != nil {
			return "", err
		}

		filePath := filepath.Join(dirName, fmt.Sprintf("%s_%s_%s_%s.yaml",
			strings.ReplaceAll(manifest.APIVersion, "/", "-"),
			manifest.Kind,
			manifest.Metadata.Namespace,
			manifest.Metadata.Name),
		)

		file, err := fs.Create(filePath)
		if err != nil {
			return "", err
		}

		defer file.Close()

		_, err = file.Write(b.Bytes())
		if err != nil {
			return "", err
		}
	}

	return dirName, nil
}
