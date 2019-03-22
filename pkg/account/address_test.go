package account

import (
	"bytes"
	"encoding/hex"
	"testing"
)

type testVector struct {
	prefix         int
	addressType    Type
	networkType    NetworkType
	spendingKeyHex string
	viewKeyHex     string
	adrStr         string
}

// Test vectors:
var (
	empty = testVector{}
	short = testVector{
		adrStr: "Safex616",
	}
	badNetworkID = testVector{
		spendingKeyHex: "8c1a9d5ff5aaf1c3cdeb2a1be62f07a34ae6b15fe47a254c8bc240f348271679",
		viewKeyHex:     "0a29b163e392eb9416a52907fd7d3b84530f8d02ff70b1f63e72fdcb54cf7fe1",
		adrStr:         "46w3n5EGhBeZkYmKvQRsd8UK9GhvcbYWQDobJape3NLMMFEjFZnJ3CnRmeKspubQGiP8iMTwFEX2QiBsjUkjKT4SSPd3fKp",
	}
	validMainnet = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "d645dd142d38c950d5d38c7d81c3fff2ff9c8e267169dc1c896ed9b84509ca84",
		viewKeyHex:     "0fadcae4d1c82d3cd3f62c5c27ce036cc2c7617cb19f4bcef17cf4650d949ff6",
		adrStr:         "Safex616cpc4NjrBS34ciXMzaJL6y8fUZ7RwqYVuUjdJXpeWbNXcr2ufGqTHWjiMKZGR2NcMFPap8Mrgp6z9Ndb9HuLnfUQMs2R11",
	}
)

func Test_AddressFromBase58(t *testing.T) {
	tests := []struct {
		name    string
		tVec    testVector
		wantErr bool
	}{
		// {
		// 	name:    "fails, address empty",
		// 	tVec:    empty,
		// 	wantErr: true,
		// },
		// {
		// 	name:    "fails, address too short",
		// 	tVec:    short,
		// 	wantErr: true,
		// },
		{
			name:    "fails, bad network id",
			tVec:    badNetworkID,
			wantErr: true,
		},
		{
			name: "passes, valid mainnet address",
			tVec: validMainnet,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			wantSpendKey, err := hex.DecodeString(tt.tVec.spendingKeyHex)
			if err != nil {
				t.Fatalf("Failed to decode test spend key, key = %s", tt.tVec.spendingKeyHex)
			}
			wantViewKey, err := hex.DecodeString(tt.tVec.viewKeyHex)
			if err != nil {
				t.Fatalf("Failed to decode test view key, key = %s", tt.tVec.viewKeyHex)
			}

			// Decode address
			adr, err := FromBase58(tt.tVec.adrStr)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("FromBase58 error = %v, wantErr = %v", err, tt.wantErr)
				}

				return
			}

			// Test network type and address type
			if (adr.NetworkType() != tt.tVec.networkType) != tt.wantErr {
				t.Errorf("Bad network type, want = %v, got = %v", tt.tVec.networkType, adr.NetworkType())
			}
			if (adr.Type() != tt.tVec.addressType) != tt.wantErr {
				t.Errorf("Bad address type, want = %v, got = %v", tt.tVec.addressType, adr.Type())
			}

			// Test if proper view/spend public keys were extracted
			if (bytes.Compare(adr.SpendKey, wantSpendKey) != 0) != tt.wantErr {
				t.Errorf("Bad spend key, want = %v, got = %v", wantSpendKey, adr.SpendKey)
			}
			if (bytes.Compare(adr.ViewKey, wantViewKey) != 0) != tt.wantErr {
				t.Errorf("Bad view key, want = %v, got = %v", wantViewKey, adr.ViewKey)
			}

			// Convert address back to base58
			if res := adr.String(); (res != tt.tVec.adrStr) != tt.wantErr {
				t.Errorf("Address.String() = %v, want %v", res, tt.tVec.adrStr)
			}
		})
	}
}
