package bootstrap

import (
	"log"
	"os"
)

//
func setupFileSystem() {

}

// Bootstrap...
func Bootstrap() (err error) {

	inputDir := *CONSTANTS.System.INPUTDIR
	outputDir := *CONSTANTS.System.OUTPUTDIR
	resultsSchema := *CONSTANTS.System.RESULTSSCHEMA

	// outputDir := filepath.Join(*BASEPATH, *OUTPUTDIR)
	// resultsSchema := filepath.Join(*BASEPATH, *RESULTSSCHEMA)

	log.Println("inputDir", inputDir)
	log.Println("outputDir", outputDir)
	log.Println("resultsSchema", resultsSchema)

	err = os.RemoveAll(inputDir)
	if err != nil {
		return
	}
	err = os.RemoveAll(outputDir)
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

	return

}
