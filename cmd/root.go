package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nginxctl",
	Short: "nginx reverse-proxy and SSL setup wizard",
	Long: `nginxctl sets up nginx as a reverse proxy with WebSocket support,
configures SSL via Let's Encrypt certbot, and verifies DNS before attempting
certificate issuance.

Commands:
  setup   Full wizard — install nginx, write proxy config, optionally add SSL
  ssl     Add SSL to an existing nginx config`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(sslCmd)
}

func requireRoot() {
	if os.Getuid() != 0 {
		fmt.Fprintln(os.Stderr, "error: must be run as root — try: sudo nginxctl "+os.Args[1])
		os.Exit(1)
	}
}
