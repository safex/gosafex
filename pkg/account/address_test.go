package account

import (
	"encoding/hex"
	"reflect"
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
		spendingKeyHex: "322aa25558f5c5ec6b60e2c83e94cd14423d406b73c61c925554183fcda1a03d",
		viewKeyHex:     "91ffbc5b116ecbf79e35a9a1481fb97165ced550b7733df9254525168e025eed",
		adrStr:         "5zCxPRXEGbM4WwG9X25T4oJhDR7ZT9EQaVSiUFDKbUFyRjRuzbGEUXcKUTbmiTydpduEkV72EW42SJpLz9A4movipf8YrH2v",
	}
	validMainnet = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "6c34868fb70656086226996ddd1e30873ee57e4b569568a8cb91eec669947a51",
		viewKeyHex:     "d146e3a50f1f892f2343b2ee6c9e86b5d9221d275eeca01330d0fb841591e968",
		adrStr:         "Safex5zUzTPR3DU5huKs9rKNpKkH3gs7bDbskmfR7Ky7aBrF9be1yVUUcHED5NtJcRgt1q9iegKsz7awsYaWfueiP6PCb6hWCnV3v",
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
			var wantSpendKey [32]byte
			var wantViewKey [32]byte
			wantSpendKeyTemp, err := hex.DecodeString(tt.tVec.spendingKeyHex)
			if err != nil {
				t.Fatalf("Failed to decode test spend key, key = %s", tt.tVec.spendingKeyHex)
			}
			copy(wantSpendKey[:], wantSpendKeyTemp)
			wantViewKeyTemp, err := hex.DecodeString(tt.tVec.viewKeyHex)
			if err != nil {
				t.Fatalf("Failed to decode test view key, key = %s", tt.tVec.viewKeyHex)
			}
			copy(wantViewKey[:], wantViewKeyTemp)

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
			if reflect.DeepEqual(adr.SpendKey.ToBytes(), wantSpendKey) == tt.wantErr {
				t.Errorf("Bad spend key, want = %v, got = %v", wantSpendKey, adr.SpendKey.ToBytes())
			}

			if reflect.DeepEqual(adr.ViewKey.ToBytes(), wantViewKey) == tt.wantErr {
				t.Errorf("Bad view key, want = %v, got = %v", wantViewKey, adr.ViewKey.ToBytes())
			}

			// Convert address back to base58
			if res := adr.String(); reflect.DeepEqual(res, tt.tVec.adrStr) == tt.wantErr {
				t.Errorf("Address.String() = %v, want %v", res, tt.tVec.adrStr)
			}
		})
	}
}

func TestAddress_Equals(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "passes, true",
			a:    validMainnet.adrStr,
			b:    validMainnet.adrStr,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := FromBase58(tt.a)
			b, _ := FromBase58(tt.b)
			if got := a.Equals(b); got != tt.want {
				t.Errorf("Address.Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}
