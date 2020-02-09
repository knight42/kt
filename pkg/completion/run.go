package completion

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Generate(cmd *cobra.Command, shell string) error {
	switch shell {
	case "zsh":
		return cmd.GenZshCompletion(os.Stdout)
	case "bash":
		return cmd.GenBashCompletion(os.Stdout)
	default:
		return fmt.Errorf("unknown shell: %s", shell)
	}
}
