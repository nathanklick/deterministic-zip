package zip

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"github.com/timo-reymann/deterministic-zip/pkg/cli"
	"io"
	"math/rand"
	"os"
	"testing"
)

func checksum(file string) string {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		panic(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

type expectedFile struct {
	name  string
	isDir bool
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		config      cli.Configuration
		sha256      string
		compression uint16
		zipFiles    []expectedFile
	}{
		{
			config: cli.Configuration{
				SourceFiles: []string{
					"testdata/file.txt",
				},
			},
			sha256:      "32198c12721f7bc3b0fffad9df16c3e9fa56c4b698d5390f74dd1e7e74fbb915",
			compression: zip.Store,
			zipFiles: []expectedFile{
				{
					name: "testdata/file.txt",
				},
			},
		},
		{
			config: cli.Configuration{
				SourceFiles: []string{
					"testdata/folder",
				},
			},
			sha256:      "8739c76e681f900923b900c9df0ef75cf421d39cabb54650c4b9ad19b6a76d85",
			compression: zip.Store,
			zipFiles:    []expectedFile{},
		},
		{
			config: cli.Configuration{
				SourceFiles: []string{
					"testdata/folder/file.txt",
				},
			},
			sha256:      "dd97707d68eda2563e0686e29934e4a7cd0437e761e9d02fdc6456cb3fd91eb7",
			compression: zip.Store,
			zipFiles: []expectedFile{
				{
					name: "testdata/folder/file.txt",
				},
			},
		},
		{
			config: cli.Configuration{
				SourceFiles: []string{
					"testdata/folder/file.txt",
					"testdata/file.txt",
				},
			},
			sha256:      "b18ca34af3f15c04ec624e286412f44b2ed5c83e83d93b4b5b148aa03477ee9f",
			compression: zip.Store,
			zipFiles: []expectedFile{
				{
					name: "testdata/folder/file.txt",
				},
				{
					name: "testdata/file.txt",
				},
			},
		},
		{
			config: cli.Configuration{
				SourceFiles: []string{
					"testdata",
					"testdata/file.txt",
					"testdata/folder",
					"testdata/folder/file.txt",
				},
			},
			sha256:      "b18ca34af3f15c04ec624e286412f44b2ed5c83e83d93b4b5b148aa03477ee9f",
			compression: zip.Store,
			zipFiles: []expectedFile{
				{
					name: "testdata/file.txt",
				},
				{
					name: "testdata/folder/file.txt",
				},
			},
		},
		{
			config: cli.Configuration{
				SourceFiles: []string{
					"testdata",
					"testdata/file.txt",
					"testdata/folder",
					"testdata/folder/file.txt",
				},
			},
			sha256:      "8b3eeacdd0c5c265a67bf465d9fc7d3ed0c041fc27534fb3f14b34d5a2b0b518",
			compression: zip.Deflate,
			zipFiles: []expectedFile{
				{
					name: "testdata/file.txt",
				},
				{
					name: "testdata/folder/file.txt",
				},
			},
		},
		{
			config: cli.Configuration{
				SourceFiles: []string{
					"testdata/folder/file.txt",
					"testdata/file.txt",
				},
			},
			sha256:      "8b3eeacdd0c5c265a67bf465d9fc7d3ed0c041fc27534fb3f14b34d5a2b0b518",
			compression: zip.Deflate,
			zipFiles: []expectedFile{
				{
					name: "testdata/file.txt",
				},
				{
					name: "testdata/folder/file.txt",
				},
			},
		},
	}

	for _, tc := range testCases {
		rand.Shuffle(len(tc.config.SourceFiles), func(i, j int) {
			tc.config.SourceFiles[i], tc.config.SourceFiles[j] = tc.config.SourceFiles[j], tc.config.SourceFiles[i]
		})
		for i := 0; i < 20; i++ {
			tempFile := createTmpFile()
			tempFileZip := tempFile + extension
			// Create tempfile
			tc.config.ZipFile = tempFile

			_ = Create(&tc.config, tc.compression)

			sha256sum := checksum(tc.config.ZipFile)

			if tc.sha256 != sha256sum {
				t.Fatalf("Run #%d Expected checksum %s, but got %s, file: %s", i, tc.sha256, sha256sum, tc.config.ZipFile)
			}

			if tc.config.ZipFile != tempFileZip {
				t.Fatalf("Expected final zip name to be overriden")
			}

			r, err := zip.OpenReader(tempFileZip)
			if err != nil {
				t.Fatal(err)
			}

			var foundFile *zip.File = nil

			if len(tc.zipFiles) == 0 && len(tc.zipFiles) != 0 {
				t.Fatalf("Expected no files in zip, but got %v", tc.zipFiles)
			}

			for _, expectedFile := range tc.zipFiles {
				foundFile = nil
				for _, file := range r.File {
					if expectedFile.name == file.Name {
						foundFile = file
						break
					}
				}

				if foundFile == nil {
					t.Fatalf("Expected file %s to be in archive", expectedFile.name)
				}

				if foundFile.Modified.Sub(ModifiedTimestamp) != 0 {
					t.Fatalf("Modified timestamp not reset for file %s", expectedFile.name)
				}

				if (foundFile.FileHeader.Mode() == os.ModeDir) != expectedFile.isDir {
					t.Fatalf("Expected file %s to be directory == %v", expectedFile.name, expectedFile.isDir)
				}
			}

			_ = r.Close()
			_ = os.Remove(tempFileZip)
		}
	}
}

func createTmpFile() string {
	f, _ := os.CreateTemp(os.TempDir(), "go_test_")
	_ = f.Close()
	_ = os.Remove(f.Name())
	return f.Name()
}
