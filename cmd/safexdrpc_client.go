package cmd

import (
	"fmt"

	"github.com/safex/gosafex/pkg/safex"
	"github.com/safex/gosafex/pkg/safexdrpc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logger = log.StandardLogger()
var logFile = "SafexSDK.log"

var daemonHost string
var daemonPort uint

// safexdRPCCmd represents the RPC daemon api test command
var safexdRPCCmd = &cobra.Command{
	Use:   "safexdrpc",
	Short: "Test cmd rpc client for safex daemon",
	Long:  `Cmd that talks to safexd rpc daemon and prints basic info about daemon`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		safexdClient := safexdrpc.InitClient("127.0.0.1", 29393, nil)

		count, _ := safexdClient.GetBlockCount()
		fmt.Println("Retrieved block count is:", count)

		blockNumber := 50000
		hash, _ := safexdClient.OnGetBlockHash(50000)
		fmt.Println("Retrieved hash for block ", blockNumber, " is:", hash)

		var gInfo safex.DaemonInfo
		var hInfo safex.HardForkInfo

		gInfo, _ = safexdClient.GetDaemonInfo()
		fmt.Println(gInfo)

		hInfo, _ = safexdClient.GetHardForkInfo(1)
		fmt.Println(hInfo)

		var txs safex.Transactions
		txs, _ = safexdClient.GetTransactions([]string{"7fdae840fa22793df69d197048b439c1d0b69711a31edd97701b9072e6c0c9fe"})
		fmt.Println(txs)

		var blocks safex.Blocks
		blocks, _ = safexdClient.GetBlocks(1, 4)
		fmt.Println(blocks)

		safexdClient.Close()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&daemonHost, "daemon_host", "d", "", "Target safex daemon host")
	rootCmd.PersistentFlags().UintVar(&daemonPort, "daemon_port", 29393, "Target safex daemon port")
	rootCmd.AddCommand(safexdRPCCmd)
}
