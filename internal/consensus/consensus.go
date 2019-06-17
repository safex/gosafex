package consensus

import "github.com/safex/gosafex/pkg/safexdrpc"

func UseForkRules(version uint32, earlyBlocks uint64) bool {
	// @TODO Consider using singleton pattern for client communication
	client := safexdrpc.InitClient("127.0.0.1", 38001)

	info, _ := client.GetDaemonInfo()
	hfInfo, _ := client.GetHardForkInfo(version)

	// @TODO Log stuff
	return info.Height >= hfInfo.EarliestHeight-earlyBlocks
}

func GetUpperTransactionSizeLimit(forkVersion uint32, earlyBlocks uint64) int {
	var fullRewardZone uint64
	if UseForkRules(forkVersion, earlyBlocks) {
		fullRewardZone = BlockGrantedFullRewardZoneV2
	} else {
		fullRewardZone = BlockGrantedFullRewardZoneV1
	}

	return int(fullRewardZone - CoinbaseBlobReservedSize)
}

func GetFeeAlgorithm() uint32 {
	return 0
}

func GetPerKBFee() uint64 {
	// @TODO Consider using singleton pattern for client communication
	client := safexdrpc.InitClient("127.0.0.1", 38001)
	fee, err := client.GetDynamicFeeEstimate()

	if err == nil {
		return FeePerKB
	} else {
		return fee
	}
}

// wallet::get_fee_multiplier
func GetFeeMultiplier(priority uint32, feeAlgorithm uint32) uint64 {
	if priority == 0 {
		priority = 1
	}

	var maxPriority uint32 = 3
	if priority >= 1 && priority < maxPriority {
		if feeAlgorithm == 0 {
			return uint64(priority)
		} else {
			// @TODO: Handle error
			panic("This should not be case!!")
			return 0
		}
	}

	return 1
}

func CalculateFee(feePerKb uint64, bytes int, feeMultiplier uint64) uint64 {
	var kB uint64 = uint64(bytes+1023) / 1024
	return kB * feePerKb * feeMultiplier
}