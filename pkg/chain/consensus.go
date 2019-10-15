package chain

func (w *Wallet) UseForkRules(version uint32, earlyBlocks uint64) bool {
	// @TODO Consider using singleton pattern for client communication
	if w.client != nil {
		if w.latestInfo == nil {
			return false
		}
		hfInfo, _ := w.client.GetHardForkInfo(version)

		// @TODO Log stuff
		return w.latestInfo.Height >= hfInfo.EarliestHeight-earlyBlocks
	}
	return false
}

func GetUpperTransactionSizeLimit(forkVersion uint32, earlyBlocks uint64) int {
	var fullRewardZone uint64
	if forkVersion == 1 {
		fullRewardZone = BlockGrantedFullRewardZoneV1
	} else if forkVersion == 2 {
		fullRewardZone = BlockGrantedFullRewardZoneV2
	}

	return int(fullRewardZone - CoinbaseBlobReservedSize)
}

func GetFeeAlgorithm() uint32 {
	return 0
}

func (w *Wallet) GetPerKBFee() uint64 {
	// @TODO Consider using singleton pattern for client communication
	if w.client != nil {
		fee, err := w.client.GetDynamicFeeEstimate()

		if err == nil {
			return FeePerKB
		} else {
			return fee
		}
	}
	return 0
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
