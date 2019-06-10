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
	testV1 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "f656cbe5b6f2665e737394e36a2ce379230c183dc0aed94fa265a992d988428c",
		viewKeyHex:     "b6a3abae9d3e9fed15b2123916bb94391af09369b8eeb41eb07280712dd8d7ed",
		adrStr:         "Safex61HPX8fRX9MKkJMAbf3D4SytCWV5BL5YBPJbKcLRZdLADmWw7pWCyf3r5LWWMAYqJmse6zMpJge6kDFQqaFKvynBKcLZM22K",
	}

	testV2 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "72cc74f8bc2185b6906310a36f6981bc6b37d900173996482bbaf678283cec8f",
		viewKeyHex:     "068b8d2842b53dff277abc529410f9abe809d6ba05ca816fd41c0bae0ed082b5",
		adrStr:         "Safex5zXCuZic3KV1yYxm5ULXX9Tv9FEY11t3jGL9RFTM6gAEJqNTQ87jaXXRYya51Ep7biHhAsH7Y7eqqKnsz66W7abyufA9B146",
	}

	testV4 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "4c33129c4f70b09acfed9dc08c7ec60dd8fde99a62d19e0098fff3af2816341c",
		viewKeyHex:     "2ba7ed2180a8a08724308926205073755b9af0a621280a477115f94389036c6e",
		adrStr:         "Safex5zJEwBT9QYFRjxNGCZCxx5MGu6cxSpjncqLyno4WJEpiNJUNNG6c1ywP4SJ4Y7Nsay8tVZJfUnfY7AyUG7iCJBDUCeQ3dn3o",
	}

	testV5 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "ede195ba5830ef31ac7b144c023fcd0d4b9ea4e90231766b337d7776e04aa9be",
		viewKeyHex:     "716c62b48c46474a5cb72f80a936799871aeedc36a746d700f6ec4336fdf8f46",
		adrStr:         "Safex61EYkLYAmxF54r4i7DiP4novhS1HfyUYqRA2XcSLtFMTehATKXXCYPmxDrLpNNXB27mxG1WLZgn9dMop3ST9c1CNhHvkLv3Y",
	}

	testV6 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "8f222a11c9803cb2b4bd358afa1878f6cd851892767f7fd1254638ef3fd10c74",
		viewKeyHex:     "954ffd039df7fb4eee15039b2c23254c33fe586cb3742273ea29cd8ec89e8874",
		adrStr:         "Safex5zgiuw3yZKgSnBom6QFFXEof9REfRVsfE2cujPdh222hZkLYBv1c6DyMAkYJ6SxNGoWBA2JsKBY5MCv4DikQtBU2VbCrkz3X",
	}

	testV7 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "ac9677c93f70bec3fcf890ef0e3f8b6dbe8ef0912dd9fe9dafee3f1e21c56a4e",
		viewKeyHex:     "1b71716004f58901db922c1b0470e1afb85620216a7c8db95850a9815d3fde0e",
		adrStr:         "Safex5zrcjGafMeDRaRJuMgz95hRjvBH1RHRKCQ9J89k63KQakFTcqNH4WUXbdD14X5X6vWZJQgYw6bBFRHpBcKiNdzjHiG8PjY1Y",
	}

	testV8 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "413f38d2e481a97b2c8ed7608950d3468747b66c899fa591bae64596efd921a2",
		viewKeyHex:     "ce63c58e7d2d41cdf542142aef71fb5b406e8e3f5aa9844ea03cca9a85a6652a",
		adrStr:         "Safex5zEZZZcGvW3tnNp6aH9XPR9kSzEhK9x8qPDXNMSSFH54MJ9gWgQqKeaYfgpyq8BXYByUd1FKBbceYDuFLNmSr4Luwkayyu2P",
	}

	testV9 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "c549a452419e8eee8c890e7002726ca12d2c23243804aae2254e8967401c7995",
		viewKeyHex:     "f2581c7a4d4512594f69dcbd50c6ac8a92b50b9e77ba875e69da88f4f01c3858",
		adrStr:         "Safex5zzuvTEkzPLJ5XGWHKjdiVzK5qGW74NLjcEy25EJGfCNGXU3f1MTUu6Su7CdRYfbg1MbWxtSTWLSGVbP4fqhyCsmuWhHsf2T",
	}

	testV10 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "d29e28949e9fb3e355ce2857366eb0d0676f79833279bf9680fdbe0aee46e3f1",
		viewKeyHex:     "5792e04fdbe9ff74830d5501758cf41e243eabc6dede252aba93c1cf09c65be2",
		adrStr:         "Safex615PcbRroCWZk2audFb5AHEGZkStNwn4Udxx15F2q3Soh7SUX1EMjVrmrDauz1FA3RhhP2W6aGJ347RGuAGbdXiyzULsRE1V",
	}

	testV11 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "63c40cfd01f20df6cda7e0b667ddd617c0f772295061b2685b894fe494ec161e",
		viewKeyHex:     "757c48331574db599b2e71e3b1c838899c75871ed1a74141bd66ccc9c0f3d434",
		adrStr:         "Safex5zSA3VjKVTYC5AcHDXWZiVhnwvUZ7uoFfSFAARtfEXgraSrhh19YaTzYLk7ccf5vSv8WTx8v69yz8psQggbakGJUhWhY3R2j",
	}

	testV12 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "f6a88f32fd0254667f6bdb8e8a6da082ef17a06c7b3cdf8ee40b41e2d76e848a",
		viewKeyHex:     "7ed9ad30b18b5de688ed78c7446ab1e07faf8818d2b2c53a19a87a1f4d407611",
		adrStr:         "Safex61HVk29XejRqxSWLaQqpmt1kew4bK9QWvPkuqvpeweXVZd72V299PX82pt72XaL9EaQxTVXm59pFGSuy1vR6EfeXq3aQdm1c",
	}

	testV13 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "322aa25558f5c5ec6b60e2c83e94cd14423d406b73c61c925554183fcda1a03d",
		viewKeyHex:     "87d39cde36a6712cec7a11faebab139cd1e47df51bb3a8df152ea9d722007ef3",
		adrStr:         "Safex5z9VmjFGyhrhXpW6uaVd5g8CLPV9JyRSQP9wwGjBfyLsQUB3SBeAkSvS7don8iyF9bRWHpvGhzrgrFi3Zfncz4MaU6EAbd5F",
	}

	testV14 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "3c76d8b4b9dcc8aca7422a066fefb9fb6555dc9c4ad48eca2d1f3c6f2b0991e3",
		viewKeyHex:     "91ffbc5b116ecbf79e35a9a1481fb97165ced550b7733df9254525168e025eed",
		adrStr:         "Safex5zCxPRXEGbM4WwG9X25T4oJhDR7ZT9EQaVSiUFDKbUFyRjRuzbGEUXcKUTbmiTydpduEkV72EW42SJpLz9A4movipf8YrH2v",
	}

	testV15 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "6c34868fb70656086226996ddd1e30873ee57e4b569568a8cb91eec669947a51",
		viewKeyHex:     "d146e3a50f1f892f2343b2ee6c9e86b5d9221d275eeca01330d0fb841591e968",
		adrStr:         "Safex5zUzTPR3DU5huKs9rKNpKkH3gs7bDbskmfR7Ky7aBrF9be1yVUUcHED5NtJcRgt1q9iegKsz7awsYaWfueiP6PCb6hWCnV3v",
	}

	testV16 = testVector{
		prefix:         MainnetRegularAddressPrefix,
		addressType:    RegularAddressType,
		networkType:    MainnetNetworkType,
		spendingKeyHex: "424c3012cc3be32ae58af5e4b9a494d9c7632580c35168b0bd7d5448547678a8",
		viewKeyHex:     "9805d12d82b9c2a8af41e87594a368403630108369cdcd8a4a694d4b22fdde41",
		adrStr:         "Safex5zEv2j49N16mNEvLUfFvQ2YyXyc8NYAQ2KVpAuqD6h5MxJnPZ28cWbbf98JxwLfgQ7qK5gbZNysfmCFEAAxDZvMoS3bmqB29",
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
		{
			name: "passes, valid mainnet address",
			tVec: testV1,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV2,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV4,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV5,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV6,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV7,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV8,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV9,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV10,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV11,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV12,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV13,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV14,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV15,
		},
		{
			name: "passes, valid mainnet address",
			tVec: testV16,
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
