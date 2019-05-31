package mnemonic

import (
	"reflect"
	"strings"
	"testing"

	"github.com/safex/gosafex/internal/mnemonic/dictionary"
)

type tVec struct {
	Words     string
	Positions []int
	Mnemonic  *Mnemonic
	Key       []byte
}

func newTVec(words string, positions []int, hexKey string) (result *tVec) {
	result = &tVec{words, positions, nil, nil}
	if positions != nil {
		result.Mnemonic = &Mnemonic{
			Words:     strings.Fields(words),
			dict:      &dictionary.CompiledEnglish,
			positions: positions,
		}
	}
	if hexKey != "" {
		result.Key = []byte(hexKey)
	}
	return result
}

// Test vectors:
var (
	emptyList = newTVec(
		"",
		nil,
		"",
	)
	tooShort = newTVec(
		"short",
		nil,
		"",
	)
	tooLong = newTVec(
		"long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long",
		nil,
		"",
	)
	invalidList = newTVec(
		"hubcaps jabbed derp kiosk gave actress eels eccentric splendid beer fictional kitchens doing unknown king enmity cycling derp pivot niece igloo hounded joining jump",
		nil,
		"",
	)
	noChecksum = newTVec(
		"hubcaps jabbed solved kiosk gave actress eels eccentric splendid beer fictional kitchens doing unknown king enmity cycling anxiety pivot niece igloo hounded joining jump",
		[]int{637, 700, 1296, 755, 527, 16, 386, 377, 1310, 156, 474, 757, 335, 1458, 754, 414, 280, 84, 1076, 931, 656, 634, 727, 739},
		"739ced5dbb69d8afb0882f93b4c6a02cc58267913e947ee15a450bd5dc6ce601",
	)
	invalidChecksum = newTVec(
		"hubcaps jabbed solved kiosk gave actress eels eccentric splendid beer fictional kitchens doing unknown king enmity cycling anxiety pivot niece igloo hounded joining jump splendid",
		[]int{637, 700, 1296, 755, 527, 16, 386, 377, 1310, 156, 474, 757, 335, 1458, 754, 414, 280, 84, 1076, 931, 656, 634, 727, 739, 1310},
		"",
	)
	validChecksum = newTVec(
		"hubcaps jabbed solved kiosk gave actress eels eccentric splendid beer fictional kitchens doing unknown king enmity cycling anxiety pivot niece igloo hounded joining jump pivot",
		[]int{637, 700, 1296, 755, 527, 16, 386, 377, 1310, 156, 474, 757, 335, 1458, 754, 414, 280, 84, 1076, 931, 656, 634, 727, 739, 1076},
		"739ced5dbb69d8afb0882f93b4c6a02cc58267913e947ee15a450bd5dc6ce601",
	)
)

func TestListDictionaries(t *testing.T) {
	tests := []struct {
		name       string
		wantResult []dictionary.Description
	}{
		{
			name: "passes, returns compiled english",
			wantResult: []dictionary.Description{{
				Name:     "English",
				NameEng:  "English",
				LangCode: dictionary.LangCodeEnglish,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := ListDictionaries(); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("ListDictionaries() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestFromString(t *testing.T) {
	type args struct {
		mnemonicStr string
	}
	tests := []struct {
		name    string
		tVec    *tVec
		wantErr bool
	}{
		{
			name:    "Fails, empty word list",
			tVec:    emptyList,
			wantErr: true,
		},
		{
			name:    "Fails, short word list",
			tVec:    tooShort,
			wantErr: true,
		},
		{
			name:    "Fails, long word list",
			tVec:    tooLong,
			wantErr: true,
		},
		{
			name:    "Fails, words do not match any dictionary",
			tVec:    invalidList,
			wantErr: true,
		},
		{
			name: "Passes, real english sequence",
			tVec: noChecksum,
		},
		{
			name: "Passes, real english sequence with invalid checksum",
			tVec: invalidChecksum,
		},
		{
			name: "Passes, real english sequence with valid checksum",
			tVec: validChecksum,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := FromString(tt.tVec.Words)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.tVec.Mnemonic) {
				t.Errorf("FromString() = %v, want %v", gotResult, tt.tVec.Mnemonic)
			}
		})
	}
}

func TestMnemonic_VerifyChecksum(t *testing.T) {
	tests := []struct {
		name      string
		tVec      *tVec
		wantErr   bool
		wantPanic bool
	}{
		{
			name:      "fails, missing checkusm word",
			tVec:      noChecksum,
			wantPanic: true,
		},
		{
			name:    "fails, invalid checksum word",
			tVec:    invalidChecksum,
			wantErr: true,
		},
		{
			name:    "passes, valid checksum",
			tVec:    validChecksum,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); (r != nil) != tt.wantPanic {
					t.Errorf("Mnemonic.VerifyChecksum() panic")
				}
			}()
			if err := tt.tVec.Mnemonic.VerifyChecksum(); (err != nil) != tt.wantErr {
				t.Errorf("Mnemonic.VerifyChecksum() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMnemonic_BackAndForth(t *testing.T) {
	tests := []struct {
		name            string
		tVec            *tVec
		langCode        int
		checksum        bool
		wantToSeedErr   bool
		wantFromSeedErr bool
	}{
		{
			name:          "fails, illegal lang code",
			tVec:          validChecksum,
			langCode:      9999,
			checksum:      true,
			wantToSeedErr: true,
		},
		{
			name:     "passes, english mnemonic with checksum",
			tVec:     validChecksum,
			langCode: dictionary.LangCodeEnglish,
			checksum: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSeed, err := tt.tVec.Mnemonic.ToSeed()
			if err != nil {
				if !tt.wantToSeedErr {
					t.Errorf("Mnemonic.ToSeed() err = %v, wantErr %v", err, tt.wantToSeedErr)
				}
				return
			}
			gotMnemonic, err := FromSeed(gotSeed, tt.langCode, tt.checksum)
			if err != nil {
				if !tt.wantToSeedErr {
					t.Errorf("FromSeed() err = %v, wantErr %v", err, tt.wantToSeedErr)
				}
				return
			}
			if !reflect.DeepEqual(tt.tVec.Mnemonic, gotMnemonic) {
				t.Errorf("Error in back and forth test, expected = %v, got = %v", tt.tVec.Mnemonic, gotMnemonic)
			}
		})
	}
}
