package archive

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	tempDir string
)

func TestMain(m *testing.M) {
	var err error
	tempDir, err = ioutil.TempDir("", "gct-temp")
	if err != nil {
		fmt.Printf("failed to create temp file: %v", err)
		os.Exit(1)
	}
	t := m.Run()
	err = os.RemoveAll(tempDir)
	if err != nil {
		fmt.Printf("Failed to remove temp db file: %v", err)
	}
	os.Exit(t)
}

func TestZip(t *testing.T) {
	zipFile := filepath.Join("..", "..", "..", "testdata", "testdata.zip")
	files, err := Unzip(zipFile, tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files to be extracted received: %v ", len(files))
	}

	zipFile = filepath.Join("..", "..", "..", "testdata", "zip-slip.zip")
	_, err = Unzip(zipFile, tempDir)
	if err == nil {
		t.Fatal("Zip() expected to error due to ZipSlip detection but extracted successfully")
	}
}
func TestUnZip(t *testing.T) {
	singleFile := filepath.Join("..", "..", "..", "testdata", "configtest.json")
	outFile := filepath.Join(tempDir, "out.zip")
	err := Zip(singleFile, outFile)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Unzip(outFile, tempDir)
	if err != nil {
		t.Fatal(err)
	}

	folder := filepath.Join("..", "..", "..", "testdata", "http_mock")
	outFolderZip := filepath.Join(tempDir, "out_folder.zip")
	err = Zip(folder, outFolderZip)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Unzip(outFolderZip, tempDir)
	if err != nil {
		t.Fatal(err)
	}

	folder = filepath.Join("..", "..", "..", "testdata", "invalid_file.json")
	outFolderZip = filepath.Join(tempDir, "invalid.zip")
	err = Zip(folder, outFolderZip)
	if err == nil {
		t.Fatal("expected IsNotExistError on invalid file")
	}
}
