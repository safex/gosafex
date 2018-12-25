package cmd

import (
	"fmt"

	"github.com/atanmarko/gosafex/pkg/safexdrpc"
	"github.com/spf13/cobra"
)

var daemonHost string
var daemonPort int

// safexdRpcCmd represents the RPC daemon api test command
var safexdRpcCmd = &cobra.Command{
	Use:   "safexdrpc",
	Short: "Test cmd rpc client for safex daemon",
	Long:  `Cmd that talks to safexd rpc daemon and prints basic info about daemon`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("Connecting to host ", daemonHost, " port ", daemonPort)
		fmt.Println("Safex Node Info:", safexdrpc.GetInfo)

	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&daemonHost, "daemon_host", "d", "", "Target safex daemon host")
	rootCmd.PersistentFlags().IntVar(&daemonPort, "daemon_port", 29393, "Target safex daemon port")
	rootCmd.AddCommand(safexdRpcCmd)
}
