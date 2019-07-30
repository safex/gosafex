package balance

type digitSplitStrategyHandler func(uint64)

func DecomposeAmountIntoDigits(
	amount uint64, 
	dustThreshold uint64, 
	chunkHandler digitSplitStrategyHandler, 
	dustHandler digitSplitStrategyHandler) {

	if amount == 0 {
		return
	}

	isDustHandled := false
	var dust uint64 = 0
	var order uint64 = 1
	for amount != 0 {
		chunk := (amount % 10) * order
		amount /= 10
		order *= 10
		
		if (dust + chunk) <= dustThreshold {
			dust += chunk
		} else {
			if !isDustHandled && dust != 0 {
				dustHandler(dust)
				isDustHandled = true
			}
			if chunk != 0 {
				chunkHandler(chunk)
			}
		}
	}

	if !isDustHandled && dust != 0 {
		dustHandler(dust)
	}
}

func DigitSplitStrategy(
	dsts 			*[]DestinationEntry,
	changeDst 		*DestinationEntry,
	changeDstToken 	*DestinationEntry,
	dustTrehshold 	uint64,
	splittedDsts 	*[]DestinationEntry,
	dustDsts 	 	*[]DestinationEntry) {

	*splittedDsts = nil
	*dustDsts = nil

	for _,val := range(*dsts) {
		if val.TokenTransaction {
			DecomposeAmountIntoDigits(val.TokenAmount, 0, 
			func(input uint64) {
				*splittedDsts = append(*splittedDsts, DestinationEntry{0, input, val.Address, false, true})
			}, func(input uint64){
				*dustDsts = append(*dustDsts, DestinationEntry{0, input, val.Address, false, true})
			})
		} else {
			DecomposeAmountIntoDigits(val.Amount, 0, 
				func(input uint64) {
					*splittedDsts = append(*splittedDsts, DestinationEntry{input, 0, val.Address, false, false})
				}, func(input uint64){
					*dustDsts = append(*dustDsts, DestinationEntry{input, 0, val.Address, false, false})
				})
		}
		
	}
}
