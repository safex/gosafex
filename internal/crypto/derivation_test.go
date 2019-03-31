package crypto

import (
	"reflect"
	"testing"
)

// Test vectors:
var (
	testPubKeyStr     = "5195ac8b6933ca20d1f09114cffc89b61c6531b7c2228d03ea84dc5b944cbe8a"
	testPrivKeyStr    = "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e"
	testDerivationStr = "9a1bdc439bb8446b5a7cfbbc3279bee5777336d98ba70f5c5a6f6bbbfb07d1b0"
	testPubKey        = curve.HexToKey(testPubKeyStr)
	testPrivKey       = curve.HexToKey(testPrivKeyStr)
	testDerivation    = curve.HexToKey(testDerivationStr)
)

func TestDeriveKey(t *testing.T) {
	type args struct {
		pub  *Key
		priv *Key
	}
	tests := []struct {
		name string
		args args
		want Key
	}{
		{
			name: "passes, derivation",
			args: args{
				pub:  &testPubKey,
				priv: &testPrivKey,
			},
			want: testDerivation,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeriveKey(tt.args.pub, tt.args.priv); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeriveKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
