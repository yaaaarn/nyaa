package cmd

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <id>",
	Short: "open id in browser",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		u, err := url.JoinPath(rootUrl, "view", args[0])
		if err != nil {
			fmt.Printf("URL Error: %v\n", err)
			return
		}

		var command *exec.Cmd

		switch runtime.GOOS {
		case "darwin":
			command = exec.Command("open", u)
		case "linux":
			command = exec.Command("xdg-open", u)
		default:
			fmt.Printf("%s is not supported\n", runtime.GOOS)
			return
		}

		if command != nil {
			if err := command.Start(); err != nil {
				fmt.Printf("Failed to start command: %v\n", err)
				return
			}

			fmt.Println("Browser trigger sent successfully.")
		}
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
