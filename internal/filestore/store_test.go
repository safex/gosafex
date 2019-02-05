package filestore

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var (
	testDirPath       = filepath.Join(os.TempDir(), "go_safex_test_temp")
	testDummyFileName = "dummy.dat"
	testDummyFilePath = testDirPath + string(os.PathSeparator) + testDummyFileName
	testDummyFile     *os.File
	testContent       = []byte("TESTTESTTESTTEST")
)

func setUp(t *testing.T) {
	if err := os.RemoveAll(testDirPath); err != nil {
		t.Fatalf("Failed to remove test dir, error = %v", err)
	}
	if err := os.Mkdir(testDirPath, NewDirectoryPermissions); err != nil {
		t.Fatalf("Failed to create test dir, error = %v", err)
	}
	f, err := os.Create(testDummyFilePath)
	if err != nil {
		t.Fatalf("Failed to create dummy file, error = %v", err)
	}
	f.Write(testContent)
	testDummyFile = f
}

func tearDown(t *testing.T) {
	testDummyFile.Close()
	if err := os.RemoveAll(testDirPath); err != nil {
		t.Fatalf("Failed to remove test dir, error = %v", err)
	}
}

func TestNewWithCustomPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *TempFileStore
		wantErr bool
	}{
		{
			name:    "passes",
			args:    args{testDirPath},
			want:    &TempFileStore{dirPath: testDirPath},
			wantErr: false,
		},
		{
			name:    "fails, path does not exist",
			args:    args{"/test124/test23124"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fails, is file",
			args:    args{testDummyFilePath},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setUp(t)
			got, err := NewWithCustomPath(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWithCustomPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWithCustomPath() = %v, want %v", got, tt.want)
			}
			tearDown(t)
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *TempFileStore
	}{
		{
			name: "passes",
			want: &TempFileStore{os.TempDir()},
		},
	}
	for _, tt := range tests {
		setUp(t)
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
		tearDown(t)
	}
}

func TestTempFileStore_Create(t *testing.T) {
	type fields struct {
		dirPath string
	}
	type args struct {
		name    string
		content []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		// wantFile bool
		wantErr bool
	}{
		{
			name:    "passes",
			fields:  fields{testDirPath},
			args:    args{name: "test.tmp", content: testContent},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		setUp(t)
		t.Run(tt.name, func(t *testing.T) {
			s := &TempFileStore{
				dirPath: tt.fields.dirPath,
			}
			if err := s.Create(tt.args.name, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("TempFileStore.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		tearDown(t)
	}
}

func TestTempFileStore_Read(t *testing.T) {
	type fields struct {
		dirPath string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "passes",
			fields:  fields{testDirPath},
			args:    args{testDummyFileName},
			want:    testContent,
			wantErr: false,
		},
		{
			name:    "fails, name not found",
			fields:  fields{testDirPath},
			args:    args{"TESTERRORNOTFOUND"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		setUp(t)
		t.Run(tt.name, func(t *testing.T) {
			s := &TempFileStore{
				dirPath: tt.fields.dirPath,
			}
			got, err := s.Read(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("TempFileStore.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TempFileStore.Read() = %v, want = %v", got, tt.want)
			}
		})
		tearDown(t)
	}
}
