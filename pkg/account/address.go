package account

import (
	"bytes"

	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/internal/tools/base58"
	"github.com/safex/gosafex/pkg/key"
)

// Address is the full safex address
// - NetworkID - the network prefix
// - Public spend key
// - Public view key
// - Payment ID
type Address struct {
	NetworkID NetworkID
	SpendKey  PublicKey
	ViewKey   PublicKey
	PaymentID PaymentID
}

// Checksum is a slice representing the checksum of the address.
// The checksum is computed by taking the first ChecksumSize bytes from a Keccak256 hash of the raw address bytes
type Checksum []byte

// PaymentID is the byte slice containing the unique payment id
type PaymentID []byte

func getPaymentID(nid *NetworkID, raw []byte) (PaymentID, error) {
	if nid.AddressType() == IntegratedAddressType {
		offset := nid.Size + 2*KeyLength // Offset network ID and spend and view keys
		pid := raw[offset:]
		size := len(pid)
		if (EncryptedPaymentIDSize == size) || (UnencryptedPaymentIDSize == size) {
			return pid, nil
		}

		return nil, ErrInvalidPaymentID
	}

	return nil, nil
}

// Return the first ChecksumSize bytes of the digest as the checksum.
func computeChecksum(raw []byte) []byte {
	digest := crypto.NewDigest(raw)
	return digest[:ChecksumSize]
}

func verifyChecksum(raw []byte) error {
	checksum := crypto.NewDigest(raw[:len(raw)-ChecksumSize])
	if bytes.Compare(checksum[:], raw[len(raw)-ChecksumSize:]) != 0 {
		return ErrInvalidChecksum
	}
	return nil
}

func decodeBase58(b58string string) (result *Address, err error) {
	raw, err := base58.Decode(b58string)
	if err != nil {
		return nil, err
	}

	// TODO: should max size be checked as well?
	// TODO - edo: isn't it a fixed size?
	if len(raw) < MinRawAddressSize {
		return nil, ErrRawAddressTooShort
	}

	verifyChecksum(raw)
	raw = raw[0 : len(raw)-ChecksumSize]

	networkID, err := bytesToNetworkID(raw)
	if err != nil {
		return nil, err
	}

	paymentID, err := getPaymentID(networkID, raw)
	if err != nil {
		return nil, err
	}

	var bytes [KeyLength]byte

	spendKeyOffset := networkID.Size
	viewKeyOffset := spendKeyOffset + KeyLength
	paymentIDOffset := viewKeyOffset + KeyLength

	copy(bytes[:], raw[spendKeyOffset:viewKeyOffset])
	spendKey := key.NewPublicKeyFromBytes(bytes)

	copy(bytes[:], raw[viewKeyOffset:paymentIDOffset])
	viewKey := key.NewPublicKeyFromBytes(bytes)

	result = &Address{
		NetworkID: *networkID,
		SpendKey:  *spendKey,
		ViewKey:   *viewKey,
		PaymentID: paymentID,
	}

	return
}

func (adr *Address) encodeBase58() string {

	raw := networkIDToBytes(adr.NetworkID)
	bytes := adr.SpendKey.ToBytes()
	raw = append(raw, bytes[:]...)
	bytes = adr.ViewKey.ToBytes()
	raw = append(raw, bytes[:]...)

	raw = append(raw, adr.PaymentID...)
	raw = append(raw, computeChecksum(raw)...)
	return base58.Encode(raw)
}

// New constructs an empty address.
func New() *Address { return &Address{} }

// NewAddress forms an Address with given keys, network ID and payment ID.
func NewAddress(nid NetworkID, spendKey, viewKey PublicKey, paymentID PaymentID) *Address {
	return &Address{
		NetworkID: nid,
		SpendKey:  spendKey,
		ViewKey:   viewKey,
		PaymentID: paymentID,
	}
}

// NewRegularTestnetAddress forms a regular testnet address.
// Payment ID is not set at this point.
func NewRegularTestnetAddress(spendKey, viewKey PublicKey) *Address {
	return NewAddress(*TestnetRegularNetworkID, spendKey, viewKey, nil)
}

// NewRegularMainnetAdress forms a regular mainnet address.
// Payment ID is not set at this point.
func NewRegularMainnetAdress(spendKey, viewKey PublicKey) *Address {
	return NewAddress(*MainnetRegularNetworkID, spendKey, viewKey, nil)
}

// FromBase58 will decode an address from a raw base 58 format. Returns an error if the address size is too short or if the network/payment ID's are not matching up
func FromBase58(str string) (result *Address, err error) { return decodeBase58(str) }

// ToBase58 will encode an address as a string of symbols in base 58 encoding
func (adr *Address) ToBase58() string { return adr.encodeBase58() }

// String implements the Stringer interface
func (adr *Address) String() string { return adr.ToBase58() }

// NetworkType returns the NetworkType of the address
func (adr *Address) NetworkType() NetworkType { return adr.NetworkID.NetworkType() }

// Type returns the Type of the address
func (adr *Address) Type() Type { return adr.NetworkID.AddressType() }

// IsMainnet will return true for a mainnet address
func (adr *Address) IsMainnet() bool {
	return adr.NetworkID.NetworkType() == MainnetNetworkType
}

// IsTestnet will return true for a testnet address
func (adr *Address) IsTestnet() bool {
	return adr.NetworkID.NetworkType() == TestnetNetworkType
}

// IsIntegrated will return true for all integrated address types
func (adr *Address) IsIntegrated() bool {
	return adr.NetworkID.AddressType() == IntegratedAddressType
}

// IsSameNetwork will return true if the given address has the exact same network type.
func (adr *Address) IsSameNetwork(b *Address) bool {
	return adr.NetworkID.NetworkType() == b.NetworkID.NetworkType()
}

func equalPaymentIDs(a, b PaymentID) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Equals will return true if the Address is the same.
func (adr *Address) Equals(other *Address) bool {
	return adr.NetworkID == other.NetworkID &&
		adr.SpendKey.Equal(&other.SpendKey) &&
		adr.ViewKey.Equal(&other.ViewKey) &&
		equalPaymentIDs(adr.PaymentID, other.PaymentID)
}
