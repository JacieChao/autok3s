package cmd

import (
	"github.com/cnrancher/autok3s/pkg/cli/kubectl"

	"github.com/spf13/cobra"
)

// KubectlCommand returns kubectl command for autok3s cli
func KubectlCommand() *cobra.Command {
	return kubectl.EmbedCommand()
}
