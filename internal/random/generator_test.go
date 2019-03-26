package random

import (
	"reflect"
	"testing"
)

func setUp(wantPanic bool, t *testing.T) func() {
	t.Helper()
	return func() {
		rec := recover()
		if rec == nil {
			// Did not recover from panic
			if wantPanic {
				// Panic expected, error
				t.Errorf("Test did not panic, wantPanic = true")
			}
			// Panic not expected, OK
		} else {
			// Did recover from panic
			if !wantPanic {
				// Panic not expected, error
				t.Errorf("Test did panic, wantPanic = false")
			}
			// Panic expected, OK
		}
	}
}

func TestNewGenerator(t *testing.T) {
	type args struct {
		isCaching bool
		maxCache  int
	}
	tests := []struct {
		name       string
		args       args
		wantPanic  bool
		wantResult *Generator
	}{
		{
			name: "passes, no caching",
			args: args{
				isCaching: false,
			},
			wantResult: &Generator{0, nil},
		},
		{
			name: "panics, cache size exceeds MaxCacheSize",
			args: args{
				isCaching: true,
				maxCache:  9999999,
			},
			wantPanic: true,
		},
		{
			name: "passes, valid cahce size",
			args: args{
				isCaching: true,
				maxCache:  64,
			},
			wantResult: &Generator{
				cacheSize: 64,
				cache:     make([][]byte, 64),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer setUp(tt.wantPanic, t)()
			if gotResult := NewGenerator(tt.args.isCaching, tt.args.maxCache); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("NewGenerator() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestGenerator_NewSequence(t *testing.T) {
	type fields struct {
		cacheSize int
		cache     [][]byte
	}
	tests := []struct {
		name           string
		fields         fields
		wantResultSize int
		wantCacheSize  int
	}{
		{
			name:           "passes, generates a valid length sequence without caching",
			wantResultSize: RandomSliceByteSize,
		},
		{
			name:           "passes, generate and caches a valid length sequence",
			fields:         fields{cacheSize: 1},
			wantResultSize: RandomSliceByteSize,
			wantCacheSize:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				cacheSize: tt.fields.cacheSize,
				cache:     tt.fields.cache,
			}
			gotResult := g.NewSequence()
			gotResultSize := len(gotResult)
			if gotResultSize != tt.wantResultSize {
				t.Errorf("Generator.NewSequence() sequence length = %v, want %v",
					gotResultSize, tt.wantResultSize)
			}
			postCacheSize := len(g.cache)
			if !reflect.DeepEqual(postCacheSize, tt.wantCacheSize) {
				t.Errorf("Generator.NewSequence() bad cache size: got = %v, want = %v",
					postCacheSize, tt.wantCacheSize)
			}
		})
	}
}

func TestGenerator_GetCachedSequence(t *testing.T) {
	type fields struct {
		cacheSize int
		cache     [][]byte
	}
	type args struct {
		idx int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantResult []byte
		wantErr    bool
	}{
		{
			name: "passes, returns cached entry",
			fields: fields{
				cacheSize: 1,
				cache:     [][]byte{[]byte("TESTDATA")},
			},
			args: args{
				idx: 0,
			},
			wantResult: []byte("TESTDATA"),
		},
		{
			name: "fails, out of cache range",
			args: args{
				idx: 999,
			},
			fields: fields{
				cacheSize: 1,
				cache:     [][]byte{[]byte("TESTDATA")},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				cacheSize: tt.fields.cacheSize,
				cache:     tt.fields.cache,
			}
			gotResult, err := g.GetCachedSequence(tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generator.GetCachedSequence() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("Generator.GetCachedSequence() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestGenerator_Flush(t *testing.T) {
	type fields struct {
		cacheSize int
		cache     [][]byte
	}
	tests := []struct {
		name      string
		fields    fields
		wantCache [][]byte
	}{
		{
			name:      "passes, cache flushed",
			fields:    fields{cacheSize: 12},
			wantCache: make([][]byte, 12),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				cacheSize: tt.fields.cacheSize,
				cache:     tt.fields.cache,
			}
			g.Flush()
			if ok := reflect.DeepEqual(g.cache, tt.wantCache); !ok {
				t.Errorf("Generator.Flush() error in cache: got = %v, want = %v",
					g.cache, tt.wantCache)
			}
		})
	}
}
