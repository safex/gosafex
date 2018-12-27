package cmd

import (
	"fmt"

	"github.com/atanmarko/gosafex/pkg/safexdrpc"
	"github.com/spf13/cobra"
)

var daemonHost string
var daemonPort uint

// safexdRpcCmd represents the RPC daemon api test command
var safexdRpcCmd = &cobra.Command{
	Use:   "safexdrpc",
	Short: "Test cmd rpc client for safex daemon",
	Long:  `Cmd that talks to safexd rpc daemon and prints basic info about daemon`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("Connecting to host ", daemonHost, " port ", daemonPort)
		safexdClient := safexdrpc.InitClient(daemonHost, daemonPort)

		//Get total block count from node
		count, _ := safexdClient.GetBlockCount()
		fmt.Println("Retrieved block count is:", count)

		//Get hash of particular block from node
		blockNumber := uint64(50000)
		hash, _ := safexdClient.OnGetBlockHash(blockNumber)
		fmt.Println("Retrieved hash for block ", blockNumber, " is:", hash)

		//Get block template from node
		walletAddress := "SFXtzU6Azx3N61CBXBK2KZBGUw2U3XQXKEZkSvBrfeczNvn6yXeWk4wXkNajNNe7xv1eeuH4rrrFiJMC5Ed1uN3GXt5vuDJkV3B"
		reservedSize := uint64(60)
		blockTemplate, _ := safexdClient.GetBlockTemplate(walletAddress, reservedSize)
		fmt.Println("Block template difficulty:", blockTemplate.Difficulty, " expected_reward:", blockTemplate.ExpectedReward, " height:", blockTemplate.Height)

		//Submit mined block

		err := safexdClient.SubmitBlock([]byte(blockTemplate.BlockTemplateBlob))
		fmt.Println("Submit block result", err)

		safexdClient.Close()

	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&daemonHost, "daemon_host", "d", "", "Target safex daemon host")
	rootCmd.PersistentFlags().UintVar(&daemonPort, "daemon_port", 29393, "Target safex daemon port")
	rootCmd.AddCommand(safexdRpcCmd)
}
