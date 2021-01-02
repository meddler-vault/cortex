package bootstrap

import (
	"log"
	"os"
	"path/filepath"
)

//
func setupFileSystem() {

}

// Bootstrap...
func Bootstrap() (err error) {

	inputDir := filepath.Join(*BASEPATH, *INPUTDIR)
	outputDir := filepath.Join(*BASEPATH, *OUTPUTDIR)
	schemaDir := filepath.Join(*BASEPATH, *SCHEMADIR)

	log.Println("inputDir", inputDir)
	log.Println("outputDir", outputDir)
	log.Println("schemaDir", schemaDir)

	err = os.RemoveAll(inputDir)
	if err != nil {
		return
	}
	err = os.RemoveAll(outputDir)
	if err != nil {
		return
	}
	err = os.RemoveAll(schemaDir)
	if err != nil {
		return
	}

	err = os.MkdirAll(inputDir, os.ModePerm)
	if err != nil {
		return
	}
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return
	}
	err = os.MkdirAll(schemaDir, os.ModePerm)
	if err != nil {
		return
	}

	return

}
