// Directory docs contains detailed documentation.
package main

// copy-html.go copies from the _build/html directory to the
// directory specified in the SQLR_GHPAGES environment variable.

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(log.Lshortfile)
	destDir := os.Getenv("SQLR_GHPAGES")
	if destDir == "" {
		log.Fatal("Expect environment var SQLR_GHPAGES for Github pages destination dir")
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		log.Fatal(err)
	}
	srcDir := filepath.Join("_build", "html")
	copyFiles(destDir, srcDir)
}

func copyFiles(destDir, srcDir string) {
	srcFileInfos, err := ioutil.ReadDir(srcDir)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		log.Fatal(err)
	}
	for _, srcFileInfo := range srcFileInfos {
		if srcFileInfo.IsDir() {
			subDestDir := filepath.Join(destDir, srcFileInfo.Name())
			subSrcDir := filepath.Join(srcDir, srcFileInfo.Name())
			copyFiles(subDestDir, subSrcDir)
			continue
		}
		srcFile := filepath.Join(srcDir, srcFileInfo.Name())
		destFile := filepath.Join(destDir, srcFileInfo.Name())
		copyFile(destFile, srcFile)
	}
}

func copyFile(destFile, srcFile string) {
	data, err := ioutil.ReadFile(srcFile)
	if err != nil {
		log.Fatal(err)
	}
	if err = ioutil.WriteFile(destFile, data, 0644); err != nil {
		log.Fatal(err)
	}
}
