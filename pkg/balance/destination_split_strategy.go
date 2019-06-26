package balance

type digitSplitStrategyHandler func(uint64) void

func DecomposeAmountIntoDigits(
	amount uint64, 
	dustThreshold uint64, 
	splitHandler digitSplitStrategyHandler, 
	dustHandler digitSplitStrategyHandler) 
{

	if amount == 0 {
		return
	}

	isDustHandled := false
	var dust uint64 = 0
	var order uint64 = 0
	for amount != 0 {
		chunk := (amount % 10) * order
		amount /= 10
		order *= 10
		
		if (dust + chunk) <= dustThreshold {
			dust += chunk
		}
		else {
			if !isDustHandled && dust != 0 {
				dustHandler(dust)
				isDustHandled = true
			}
			if chunk != 0 {
				chunkHandler(chunk)
			}
		}
	}

	if !isDustHandler && dust != 0 {
		dustHandler(dustHandler)
	}
}

func DigitSplitStrategy(
	dsts 			*[]DestinationEntry,
	changeDst 		*DestinationEntry,
	changeDstToken 	*DestinationEntry,
	dustTrehshold 	uint64,
	splittedDsts 	*[]DestinationEntry,
	dustDsts 	 	*[]DestinationEntry
) {
	*splittedDsts = nil
	*dustDsts = nil

	for _,val := range(*dsts) {
		DecomposeAmountIntoDigits)
	}
}
