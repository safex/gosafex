package consensus

// @TODO: Expose endpoint on node for getting consensus data

const BlockGrantedFullRewardZoneV1 uint64 = 60000
const BlockGrantedFullRewardZoneV2 uint64 = 20000
const CoinbaseBlobReservedSize uint64 = 600

// Fee related stuff
const FeePerKB uint64 = 100000000
const DynamicFeePerKBBaseFee uint64 = 100000000
const DynamicFeePerKBBaseBlockReward uint64 = 600000000000
const HFVersionDynamic uint64 = 1

const RecentOutputRatio float64 = 0.5 // 50% of outputs are from the recent zone
const RecentOutputDays float64 = 1.8 // last 1.8 day makes up the recent zone (taken from monerolink.pdf, Miller et al)
const RecentOutputZone uint64 = uint64(RecentOutputDays * 86400)
const RecentOutputBlocks uint64 = uint64(RecentOutputDays * 720)
