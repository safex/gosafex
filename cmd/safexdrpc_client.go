package cmd

import (
	"fmt"

	"github.com/safex/gosafex/pkg/safexdrpc"
	"github.com/spf13/cobra"
)

var daemonHost string
var daemonPort uint

// safexdRPCCmd represents the RPC daemon api test command
var safexdRPCCmd = &cobra.Command{
	Use:   "safexdrpc",
	Short: "Test cmd rpc client for safex daemon",
	Long:  `Cmd that talks to safexd rpc daemon and prints basic info about daemon`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("Connecting to host ", daemonHost, " port ", daemonPort)
		safexdClient := safexdrpc.InitClient(daemonHost, daemonPort)

		count, _ := safexdClient.GetBlockCount()
		fmt.Println("Retrieved block count is:", count)

		blockNumber := 50000
		hash, _ := safexdClient.OnGetBlockHash(50000)
		fmt.Println("Retrieved hash for block ", blockNumber, " is:", hash)

		safexdClient.Close()

	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&daemonHost, "daemon_host", "d", "", "Target safex daemon host")
	rootCmd.PersistentFlags().UintVar(&daemonPort, "daemon_port", 29393, "Target safex daemon port")
	rootCmd.AddCommand(safexdRPCCmd)
}
