package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/nicholastcs/alchemy/cmd"
	"github.com/nicholastcs/alchemy/internal/environment"
	"github.com/nicholastcs/alchemy/internal/utils"
)

//go:embed embed/*.yaml
var fs embed.FS

func main() {
	environment.PreloadEmbedFS(fs)

	err := cmd.Execute()
	if err != nil {
		wrappedErr := fmt.Errorf(`%w

Append "--help" to get the CLI usage`, err)

		utils.Error("Alchemy has returned error(s)", wrappedErr)

		os.Exit(1)
	}
}
