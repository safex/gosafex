package account

import (
	"encoding/binary"
)

// NetworkID is the varint network identifier
type NetworkID struct {
	Val  uint64
	Size int
}

// NetworkType is the network identifier
type NetworkType uint8

// Type is the type of the address: regular, integrated, subaddress
type Type uint8

// Address Type:
const (
	RegularAddressType    = 0
	IntegratedAddressType = iota
	SubaddressType        = iota
	UndefinedAddressType  = 255
)

// Network Type:
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
	MainnetRegularAddressPrefix    = 0x10003798  // should map to "Safex" in base58
	MainnetIntegratedAddressPrefix = 0xa90a03798 // should map to "Safexi" in base58
	MainnetSubaddressPrefix        = 0x10e03798  // should map to "Safexs" in base58
	// Testnet:
	TestnetRegularAddressPrefix    = 0x263b16   // should map to "SFXt" in base58
	TestnetIntegratedAddressPrefix = 0xe05fb16  // should map to "SFXi" in base58
	TestnetSubaddressPrefix        = 0x1905fb16 // should map to "SfXts" in base58
	// Stagenet:
	StageRegularAddressPrefix   = 0x25bb16   // should map to "SFXs" in base58
	StageIntegratedAddresPrefix = 0xdc57b16  // should map to "SFXsi" in base58
	StageSubadressPrefix        = 0x18c57b16 // should map to "SFXss" in base58
)

// MinNetworkIDSize is the minimal size of the NetworkID (in bytes)
const MinNetworkIDSize = 1

func bytesToNetworkID(raw []byte) (result *NetworkID, err error) {
	val, size := binary.Uvarint(raw)
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
	buf := make([]byte, nid.Size+1, nid.Size+1)
	size := binary.PutUvarint(buf, uint64(nid.Val))

	return buf[:size]
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
