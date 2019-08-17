package account

import (
	"bytes"
	"encoding/binary"
)

// NetworkID is the varint network identifier.
type NetworkID struct {
	Val  uint64
	Size int
}

// NetworkType is the network identifier.
type NetworkType uint8

// Type is the type of the address: regular, integrated, subaddress.
type Type uint8

// Address Types:
const (
	RegularAddressType    = 0
	IntegratedAddressType = iota
	SubaddressType        = iota
	UndefinedAddressType  = 255
)

// Network Types:
const (
	MainnetNetworkType   = 0
	TestnetNetworkType   = iota
	StagenetNetworkType  = iota
	FakeNetworkType      = iota
	UndefinedNetworkType = 255
)

// Network Prefixes:
const (
	// Mainnet:
	MainnetRegularAddressPrefix    = 268449688   // should map to "Safex" in base58
	MainnetIntegratedAddressPrefix = 45376092056 // should map to "Safexi" in base58
	MainnetSubaddressPrefix        = 0x10e03798  // should map to "Safexs" in base58
	// Testnet:
	TestnetRegularAddressPrefix    = 2505494    // should map to "SFXt" in base58
	TestnetIntegratedAddressPrefix = 235272982  // should map to "SFXi" in base58
	TestnetSubaddressPrefix        = 0x1905fb16 // should map to "SfXts" in base58
	// Stagenet:
	StageRegularAddressPrefix   = 0x25bb16   // should map to "SFXs" in base58
	StageIntegratedAddresPrefix = 0xdc57b16  // should map to "SFXsi" in base58
	StageSubadressPrefix        = 0x18c57b16 // should map to "SFXss" in base58
)

// MinNetworkIDSize is the minimal size of the NetworkID (in bytes).
const MinNetworkIDSize = 1

// TODO: store this as constants?

// Precomputed network IDs:
var (
	MainnetRegularNetworkID, _ = uint64ToNetworkID(MainnetRegularAddressPrefix)
	TestnetRegularNetworkID, _ = uint64ToNetworkID(TestnetRegularAddressPrefix)
)

func uint64ToNetworkID(rawInt uint64) (result *NetworkID, err error) {
	size := binary.Size(rawInt)
	if size < MinNetworkIDSize {
		return nil, ErrInvalidNetworkID
	}
	result = &NetworkID{
		Val:  rawInt,
		Size: size,
	}

	return
}

func bytesToNetworkID(raw []byte) (result *NetworkID, err error) {
	val, size := binary.Uvarint(raw)
	val, err = ReadVarInt(bytes.NewReader(raw))
	if size < MinNetworkIDSize {
		return nil, ErrInvalidNetworkID
	}
	result = &NetworkID{
		Val:  val,
		Size: size,
	}

	return
}

func networkIDToBytes(nid NetworkID) []byte {
	return Uint64ToBytes(nid.Val)
}

// AddressType will return the address Type
func (nid NetworkID) AddressType() Type {
	switch nid.Val {
	case MainnetRegularAddressPrefix:
		fallthrough
	case TestnetRegularAddressPrefix:
		fallthrough
	case StageRegularAddressPrefix:
		return RegularAddressType
	case MainnetIntegratedAddressPrefix:
		fallthrough
	case TestnetIntegratedAddressPrefix:
		fallthrough
	case StageIntegratedAddresPrefix:
		return IntegratedAddressType
	case MainnetSubaddressPrefix:
		fallthrough
	case TestnetSubaddressPrefix:
		fallthrough
	case StageSubadressPrefix:
		return StagenetNetworkType
	default:
		return UndefinedAddressType
	}
}

// Bytes implements the Bytes interface
func (nid NetworkID) Bytes() []byte { return networkIDToBytes(nid) }

// NetworkType will return the NetworkType
func (nid NetworkID) NetworkType() NetworkType {
	switch nid.Val {
	case MainnetRegularAddressPrefix:
		fallthrough
	case MainnetIntegratedAddressPrefix:
		fallthrough
	case MainnetSubaddressPrefix:
		return MainnetNetworkType
	case TestnetRegularAddressPrefix:
		fallthrough
	case TestnetIntegratedAddressPrefix:
		fallthrough
	case TestnetSubaddressPrefix:
		return TestnetNetworkType
	case StageRegularAddressPrefix:
		fallthrough
	case StageIntegratedAddresPrefix:
		fallthrough
	case StageSubadressPrefix:
		return StagenetNetworkType
	default:
		return UndefinedNetworkType
	}
}
