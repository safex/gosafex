package account

import (
	"bytes"

	"github.com/safex/gosafex/internal/crypto"
	"github.com/safex/gosafex/internal/tools/base58"
)

const (
	// ChecksumSize is the size of the address checksum (in bytes)
	ChecksumSize = 4
	// EncryptedPaymentIDSize is the size of the encrypted paymentID (in bytes)
	EncryptedPaymentIDSize = 8
	// UnencryptedPaymentIDSize is the size of the unencrypted paymentID (in bytes)
	UnencryptedPaymentIDSize = 32
)

// MinRawAddressSize is the minimal size of the raw address (in bytes).
const MinRawAddressSize = MinNetworkIDSize + 2*KeySize + ChecksumSize

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
		offset := nid.Size + 2*KeySize // Offset network ID and spend and view keys
		pid := raw[offset:]
		size := len(pid)
		if (EncryptedPaymentIDSize == size) || (UnencryptedPaymentIDSize == size) {
			return pid, nil
		}

		return nil, ErrInvalidPaymentID
	}

	return nil, nil
}

func computeChecksum(raw []byte) []byte {
	// Return the first ChecksumSize bytes of the Keccak hash as checksum
	return crypto.Keccak256(raw)[:ChecksumSize]
}

func verifyChecksum(raw []byte) error {
	checksum := crypto.Keccak256(raw[:len(raw)-ChecksumSize])
	if bytes.Compare(checksum, raw[len(raw)-ChecksumSize:]) != 0 {
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

	spendKeyOffset := networkID.Size
	viewKeyOffset := spendKeyOffset + KeySize
	paymentIDOffset := viewKeyOffset + KeySize
	result = &Address{
		NetworkID: *networkID,
		SpendKey:  raw[spendKeyOffset:viewKeyOffset],
		ViewKey:   raw[viewKeyOffset:paymentIDOffset],
		PaymentID: paymentID,
	}

	return
}

func (adr *Address) encodeBase58() string {
	raw := networkIDToBytes(adr.NetworkID)
	raw = append(raw, adr.SpendKey...)
	raw = append(raw, adr.ViewKey...)
	raw = append(raw, adr.PaymentID...)

	raw = append(raw, computeChecksum(raw)...)

	return base58.Encode(raw)
}

// New constructs an empty address
func New() *Address { return &Address{} }

// FromKeys forms an address with given keys, network ID and payment ID
func FromKeys(nid NetworkID, spendKey PublicKey, viewKey PublicKey, paymentID PaymentID) *Address {
	return &Address{
		NetworkID: nid,
		SpendKey:  spendKey,
		ViewKey:   viewKey,
		PaymentID: paymentID,
	}
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
