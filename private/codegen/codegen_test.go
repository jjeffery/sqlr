package codegen

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	var filenames []string
	{
	}
	fileInfos, err := ioutil.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, fileInfo := range fileInfos {
		filename := fileInfo.Name()
		if strings.HasSuffix(filename, "_sqlrow.go") {
			continue
		}
		if !strings.HasSuffix(filename, ".go") {
			continue
		}
		if !strings.HasPrefix(filename, "test") {
			continue
		}
		filenames = append(filenames, filepath.Join("testdata", filename))
	}

	for i, filename := range filenames {
		model, err := Parse(filename)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
		model.CommandLine = "sqlr-gen"
		output := DefaultOutput(filename)

		var buf bytes.Buffer
		if err := DefaultTemplate.Execute(&buf, model); err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		func() {
			outfile, err := os.Create(output)
			if err != nil {
				t.Errorf("%d %v", i, err)
				return
			}
			defer outfile.Close()
			if _, err := outfile.Write(formatted); err != nil {
				t.Errorf("%d %v", i, err)
				return
			}
		}()
	}
}
