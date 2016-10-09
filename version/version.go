package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yesnault/mclui/internal"
)

// Cmd version
var Cmd = &cobra.Command{
	Use:     "version",
	Short:   "Display Version of mclui",
	Long:    `mclui version`,
	Aliases: []string{"v"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(internal.VERSION)
	},
}
