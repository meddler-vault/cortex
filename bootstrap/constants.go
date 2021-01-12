package bootstrap

import (
	"flag"
	"path/filepath"
)

type ReservedConstants struct {
	MESSAGEQUEUE string `json:"message_queue_topic"`
}
type ProcessConstants struct {
	INPUTAPI      string `json:"input_api"`
	INPUTAPITOKEN string `json:"input_api_token"`
	OUTPUTAPI     string `json:"output_api"`
	FILEUPLOADAPI string `json:"file_upload_api"`
}

// SystemConstants
type SystemConstants struct {
	BASEPATH          string `json:"base_path"`
	INPUTDIR          string `json:"input_dir"`
	OUTPUTDIR         string `json:"output_dir"`
	RESULTSJSON       string `json:"results_json"`
	RESULTSSCHEMA     string `json:"results_schema"`
	LOGTOFILE         bool   `json:"log_to_file"`
	STDOUTFILE        string `json:"stdout_file"`
	STDERRFILE        string `json:"stderr_file"`
	ENABLELOGGING     bool   `json:"enable_logging"`
	MAXOUTPUTFILESIZE int    `json:"max_output_filesize"`
	SAMPLEINPUTFILE   string `json:"sample_inputfile"`
	SAMPLEOUTPUTFILE  string `json:"sample_outputfile"`
}

// Constants
type Constants struct {
	processConstants *ProcessConstants
	ProcessConstants *ProcessConstants

	reservedConstants *ReservedConstants
	ReservedConstants *ReservedConstants

	systemConstants *SystemConstants
	SystemConstants *SystemConstants
}

func initialize() Constants {

	reservedConstants := &ReservedConstants{
		MESSAGEQUEUE: *PopulateStr("message_queue_topic", "tasks", "Message Queue Topic"),
	}

	systemConstants := &SystemConstants{
		BASEPATH:          *PopulateStr("base_path", "/Users/meddler/Office/Workspaces/Secoflex/secoflex/modules/watchdog/tmp", "Base Path"),
		INPUTDIR:          *PopulateStr("input_dir", "input", "Specify output directory"),
		OUTPUTDIR:         *PopulateStr("output_dir", "output", "Specify output directory"),
		RESULTSJSON:       *PopulateStr("results_json", "results.json", "Specify output directory"),
		RESULTSSCHEMA:     *PopulateStr("results_schema", "schema.json", "Specify output directory"),
		LOGTOFILE:         *PopulateBool("log_to_file", false, "Specify output directory"),
		STDOUTFILE:        *PopulateStr("stdout_file", "schema", "Specify output directory"),
		STDERRFILE:        *PopulateStr("stderr_file", "schema", "Specify output directory"),
		ENABLELOGGING:     *PopulateBool("enable_logging", true, "Enable Logging"),
		MAXOUTPUTFILESIZE: *PopulateInt("max_output_filesize", 500, "Enable Logging"),
		SAMPLEINPUTFILE:   *PopulateStr("sample_inputfile", "PopulateStr", "Enable Logging"),
		SAMPLEOUTPUTFILE:  *PopulateStr("sample_outputfile", "PopulateStr", "Enable Logging"),
	}

	// Relative to Absolute Path
	systemConstants.INPUTDIR = filepath.Join(systemConstants.BASEPATH, systemConstants.INPUTDIR)
	systemConstants.OUTPUTDIR = filepath.Join(systemConstants.BASEPATH, systemConstants.OUTPUTDIR)
	systemConstants.RESULTSSCHEMA = filepath.Join(systemConstants.BASEPATH, systemConstants.RESULTSSCHEMA)

	processConstants := &ProcessConstants{

		INPUTAPI:      *PopulateStr("input_api", "input", "Specify output directory"),
		INPUTAPITOKEN: *PopulateStr("input_api_token", "input", "Specify output directory"),
		FILEUPLOADAPI: *PopulateStr("output_api", "input", "Specify output directory"),
		OUTPUTAPI:     *PopulateStr("file_upload_api", "input", "Specify output directory"),
	}

	return Constants{

		reservedConstants: reservedConstants,
		ReservedConstants: reservedConstants,

		systemConstants: systemConstants,
		SystemConstants: systemConstants,

		processConstants: processConstants,
		ProcessConstants: processConstants,
	}
}

func (constants *Constants) reset() {

	constants.ProcessConstants = constants.processConstants
	constants.ReservedConstants = constants.reservedConstants
	constants.SystemConstants = constants.systemConstants

}

// Reset
func (constants *Constants) Reset() {
	constants.reset()
}

// Exprted CONSTANTS
var (
	CONSTANTS Constants
)

func init() {
	flag.Parse()
	CONSTANTS = initialize()

}
