package bootstrap

import (
	"io/ioutil"

	"os"
	"path/filepath"

	"github.com/meddler-vault/cortex/logger"
)

func setupFileSystem() {

}

// Bootstrap...
func Bootstrap() (err error) {

	inputDir := *CONSTANTS.System.INPUTDIR
	outputDir := *CONSTANTS.System.OUTPUTDIR
	resultsJson := *CONSTANTS.System.RESULTSJSON

	// outputDir := filepath.Join(*BASEPATH, *OUTPUTDIR)
	// resultsSchema := filepath.Join(*BASEPATH, *RESULTSSCHEMA)
	logger.Println("Creating Dir Sync")

	logger.Println("inputDir", inputDir)
	logger.Println("outputDir", outputDir)
	logger.Println("resultsSchema", resultsJson)

	err = os.RemoveAll(inputDir)
	if err != nil {
		return
	}
	err = os.RemoveAll(outputDir)
	if err != nil {
		return
	}

	err = os.MkdirAll(inputDir, os.ModePerm)
	logger.Println("Creating Directory: inputDir", inputDir, err)
	if err != nil {
		return
	}

	err = os.MkdirAll(outputDir, os.ModePerm)
	logger.Println("Creating Directory: outputDir", outputDir, err)
	if err != nil {
		return
	}

	err = os.MkdirAll(filepath.Dir(resultsJson), os.ModePerm)
	logger.Println("Creating Directory: resultsSchema", resultsJson, filepath.Dir(resultsJson))
	if err != nil {
		return
	}

	return

}

func PrintDir(root string, tag string) {
	logger.Println("************DIR:", root, tag, "************")

	files, err := ioutil.ReadDir(root)
	if err != nil {
		logger.Println(err)
	}
	for _, f := range files {
		logger.Println("DIR:", root, tag, f.Name())
	}
	logger.Println("************DIR:", root, tag, "************")

}
