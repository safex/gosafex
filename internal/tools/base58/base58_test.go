package base58

import (
	"testing"
)

// Test Data
var (
	testEncodedBlocks = []string{ // Block size == 11B == 11 symbols
		"4495qPNxDGA",
	}
	testValidAddr = []string{
		"4495qPNxDGAT241zya1WdwG5YU6RQ6s7ZgGeQryFtooAMMxoqx2N2oNTjP5NTqDf9GMaZ52iS2Q6xhnWyW16cdi47MCsVRg",
		"47Mov77LGqgRoRh6K6XVheSagWVRS7jkQLCR9jPQxTa8g2SrnwbWuMzKWRLyyBFsxn7gHJv15987MDMkYXCXGGvhKA7Qsx4",
		"48fj5P3zky9FETVG144GWh2oxnEdBc45VFHLKgKQfZ7UdyJ5M7mDFxuEA12eBuD55RAwgX2jzFYfwjhukHavcLHW9vKn1VG",
		"48vTj54ZtU7e6sqwcJY9uq2LApd3Zi6H23vmYFc3wMteS2QzJwi2Z1xCLVwMac55h2HnQAiYwZTceJbwMZJRrm3uNh76Hci",
		"48oYzqzeGqY3Nfg6LG8HwS3uF1Y3vV2gfRH6ZMcnhhEmUgkL2mPSjtuSekenrYGkbp8RNvAvrtq3r7Ze4iPoBH3kFK9vbgP",
	}
	testDecodedAddr = [][]byte{
		[]byte("12426a2b555065c79b8d6293c5dd18c25a25bae1a8b8c67ceac7484133e6798579bba3444bd48d5d9fcffa64d805e3977b07e2d420a2212df3d612a5dbcc67653844ded707"),
		[]byte("12975e989ae39b7b9445ac7384edb7a598efe3fbfab6c0bd72c5372fadd86071e95096d3b5eedd396ea5c521456640fb27ebb5a222269eac49e1ddac7134735ea0efb2b899"),
		[]byte("12b9e8cd1f42a48c55166f75ead8293e0ad1c420f566b9c85562572936207557dd08613f96d197024ea651e8f226feb03b71aa82f487cb6eff518a30a3b6a2514f0eb176af"),
		[]byte("12c09d10f3c5f580ddd0765063d9246007f45ef025a76c7d117fe4e811fa78f3959c66f7487c1bef43c64ee0ace763116456666a389eea3b693cd7670c3515a0c043794fbf"),
		[]byte("12bd785822c5e8330e30cc7e6e7abd3d11579da04e4131d091255172583059aea58501a7d7657332995b54357cc02c972c5cf5b2d1804d4d273c6f214854c9cf7edd34d73c"),
	}
	// Encoding is too short
	testMalformedBase58Encoding = "4495qPNxDGAT241zya1WdwG5YU6RQ6s7ZgGeQryFtooAMMf9GMaZ52iS2Q6xhnWyW16cdi47MCsVRg"
	// Encoding contains non-base58 symbols
	testIllegalBase58SymbolEncoding = "00OOj54ZtU7e6sqwcJY9uq2LApd3Zi6H23vmYFc3wMteS2QzJwi2Z1xCLVwMac55h2HnQAiYwZTceJbwMZJRrm3uNh76Hci"
)

func TestLookupTable(t *testing.T) {
	for i := 0; i < 58; i++ {
		curChar := string(alphabet[i])
		curNum := charLookup[curChar]
		if curNum != i {
			t.Errorf("Malformed lookup table at %d, got %s for %d", i, curChar, curNum)
		}
	}
}

func TestBackAndForth(t *testing.T) {
	tests := []struct {
		desc    string
		data    string
		wantErr bool
	}{
		{"passes, empty", "", false},
		{"passes, valid addr #1", testValidAddr[0], false},
		{"passes, valid addr #2", testValidAddr[1], false},
		{"passes, valid addr #3", testValidAddr[2], false},
		{"passes, valid addr #4", testValidAddr[3], false},
		{"passes, valid addr #5", testValidAddr[4], false},
		{"fails, encoded value too short", testMalformedBase58Encoding, true},
		{"fails, illegal symbol", testIllegalBase58SymbolEncoding, true},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			decoded, err := Decode(tt.data)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Error in back-and-forth encoding/decoding, err = %v", err)
				}
				return
			}
			if gotResult := Encode(decoded); gotResult != tt.data {
				t.Errorf("Failed back-and-forth encoding/decoding, got = %v should equal %v", gotResult, tt.data)
			}
		})
	}
}
