package dictionary

import (
	"reflect"
	"testing"
)

// Test dictionaries
var (
	TestSimpleDict = Dictionary{
		LangCode:  0,
		Name:      "English Test",
		NameEng:   "English Test",
		PrefixLen: 3,
		Entries: []string{
			"one",
			"two",
			"three",
			"four",
			"five",
		},
	}
)

func TestDictionary_FindAll(t *testing.T) {
	type args struct {
		words []string
	}
	tests := []struct {
		name          string
		dict          Dictionary
		args          args
		wantPositions []int
		wantErr       bool
	}{
		{
			name:    "Fails, empty word list",
			dict:    TestSimpleDict,
			wantErr: true,
		},
		{
			name: "Fails, word mismatch",
			dict: TestSimpleDict,
			args: args{[]string{
				"one",
				"two",
				"three",
				"Five",
			}},
			wantErr: true,
		},
		{
			name: "Passes, all words match",
			dict: TestSimpleDict,
			args: args{[]string{
				"one",
				"two",
				"three",
				"four",
			}},
			wantPositions: []int{
				0,
				1,
				2,
				3,
			},
		},
		{
			name: "Passes, real dictionary",
			dict: CompiledEnglish,
			args: args{[]string{
				"sequence",
				"atlas",
				"unveil",
				"summon",
				"pebbles",
				"tuesday",
				"beer",
				"rudely",
				"snake",
				"rockets",
				"different",
				"fuselage",
				"woven",
				"tagged",
				"bested",
				"dented",
				"vegan",
				"hover",
				"rapid",
				"fawns",
				"obvious",
				"muppet",
				"randomly",
				"seasons"}},
			wantPositions: []int{
				1242,
				115,
				1469,
				1331,
				1047,
				1425,
				156,
				1196,
				1284,
				1182,
				319,
				514,
				1596,
				1350,
				165,
				309,
				1511,
				635,
				1136,
				457,
				972,
				897,
				1135,
				1232,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPositions, err := tt.dict.FindAll(tt.args.words)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dictionary.FindAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPositions, tt.wantPositions) {
				t.Errorf("Dictionary.FindAll() = %v, want %v", gotPositions, tt.wantPositions)
			}
		})
	}
}

func TestDictionary_GetDescription(t *testing.T) {
	tests := []struct {
		name string
		dict *Dictionary
		want Description
	}{
		{
			name: "passes",
			dict: &CompiledEnglish,
			want: Description{
				Name:     "English",
				NameEng:  "English",
				LangCode: LangCodeEnglish,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dict.GetDescription(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dictionary.GetDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDictionary(t *testing.T) {
	type args struct {
		langCode int
	}
	tests := []struct {
		name       string
		args       args
		wantResult *Dictionary
		wantErr    bool
	}{
		{
			name:       "fails, illegal lang code",
			args:       args{LangCodeEnglish},
			wantResult: &CompiledEnglish,
		},
		{
			name:    "passess, compiled English dictionary found",
			args:    args{9999999},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := GetDictionary(tt.args.langCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDictionary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("GetDictionary() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
