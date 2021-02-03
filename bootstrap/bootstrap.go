package bootstrap

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

//
func setupFileSystem() {

}

// Bootstrap...
func Bootstrap() (err error) {

	inputDir := *CONSTANTS.System.INPUTDIR
	outputDir := *CONSTANTS.System.OUTPUTDIR
	resultsJson := *CONSTANTS.System.RESULTSJSON

	// outputDir := filepath.Join(*BASEPATH, *OUTPUTDIR)
	// resultsSchema := filepath.Join(*BASEPATH, *RESULTSSCHEMA)
	log.Println("Creating Dir Sync")

	log.Println("inputDir", inputDir)
	log.Println("outputDir", outputDir)
	log.Println("resultsSchema", resultsJson)

	err = os.RemoveAll(inputDir)
	if err != nil {
		return
	}
	err = os.RemoveAll(outputDir)
	if err != nil {
		return
	}

	err = os.MkdirAll(inputDir, os.ModePerm)
	log.Println("Creating Directory: inputDir", inputDir, err)
	if err != nil {
		return
	}

	err = os.MkdirAll(outputDir, os.ModePerm)
	log.Println("Creating Directory: outputDir", outputDir, err)
	if err != nil {
		return
	}

	err = os.MkdirAll(filepath.Dir(resultsJson), os.ModePerm)
	log.Println("Creating Directory: resultsSchema", resultsJson, filepath.Dir(resultsJson))
	if err != nil {
		return
	}

	return

}

func PrintDir(root string, tag string) {
	log.Println("************DIR:", root, tag, "************")

	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.Println(err)
	}
	for _, f := range files {
		log.Println("DIR:", root, tag, f.Name())
	}
	log.Println("************DIR:", root, tag, "************")

}
