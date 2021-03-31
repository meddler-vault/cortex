package bootstrap

import (
	"flag"

	"github.com/meddler-io/watchdog/logger"

	"path/filepath"
	"strconv"
	"strings"

	"github.com/jinzhu/copier"
)

type BaseConstants struct {
}

type ReservedConstants struct {
	BaseConstants
	MESSAGEQUEUE        string `json:"message_queue_topic"`
	PUBLISHMESSAGEQUEUE string `json:"publish_message_queue_topic"`
}
type ProcessConstants struct {
	BaseConstants
	INPUTAPI      *string `json:"input_api"`
	INPUTAPITOKEN *string `json:"input_api_token"`
	OUTPUTAPI     *string `json:"output_api"`
	FILEUPLOADAPI *string `json:"file_upload_api"`
}

// SystemConstants
type ConfigConstants struct {
	BaseConstants
	Process   *string   `json:"process"`
	Arguments []*string `json:"args"`
}

// SystemConstants
type SystemConstants struct {
	BaseConstants
	BASEPATH          *string `json:"base_path"`
	INPUTDIR          *string `json:"input_dir"`
	OUTPUTDIR         *string `json:"output_dir"`
	RESULTSJSON       *string `json:"results_json"`
	RESULTSSCHEMA     *string `json:"results_schema"`
	LOGTOFILE         *bool   `json:"log_to_file"`
	STDOUTFILE        *string `json:"stdout_file"`
	STDERRFILE        *string `json:"stderr_file"`
	ENABLELOGGING     *bool   `json:"enable_logging"`
	MAXOUTPUTFILESIZE *int    `json:"max_output_filesize"`
	SAMPLEINPUTFILE   *string `json:"sample_inputfile"`
	SAMPLEOUTPUTFILE  *string `json:"sample_outputfile"`
	TRACEID           *string `json:"trace_id"`
	EXECTIMEOUT       *string `json:"exec_timeout"`
}

// Override
func (current *Constants) Override(new *Constants) {

	logger.Println("Override", new.System.BASEPATH)
	if new.System.BASEPATH != nil {

		*current.System.BASEPATH = *new.System.BASEPATH
	}

	if new.System.INPUTDIR != nil {
		*current.System.INPUTDIR = *new.System.INPUTDIR
	}

	if new.System.OUTPUTDIR != nil {
		*current.System.OUTPUTDIR = *new.System.OUTPUTDIR
	}

	if new.System.RESULTSJSON != nil {
		*current.System.RESULTSJSON = *new.System.RESULTSJSON

		// Add extension .json
		if !strings.HasSuffix(*current.System.RESULTSJSON, ".json") {
			*current.System.RESULTSJSON += ".json"
		}
	}

	if new.System.RESULTSSCHEMA != nil {
		*current.System.RESULTSSCHEMA = *new.System.RESULTSSCHEMA
	}

	if new.System.LOGTOFILE != nil {
		*current.System.LOGTOFILE = *new.System.LOGTOFILE
	}

	if new.System.STDOUTFILE != nil {
		*current.System.STDOUTFILE = *new.System.STDOUTFILE
	}

	if new.System.STDERRFILE != nil {
		*current.System.STDERRFILE = *new.System.STDERRFILE
	}

	if new.System.ENABLELOGGING != nil {
		current.System.ENABLELOGGING = new.System.ENABLELOGGING
	}

	if new.System.MAXOUTPUTFILESIZE != nil {
		current.System.MAXOUTPUTFILESIZE = new.System.MAXOUTPUTFILESIZE
	}

	if new.System.SAMPLEINPUTFILE != nil {
		current.System.SAMPLEINPUTFILE = new.System.SAMPLEINPUTFILE
	}

	if new.System.TRACEID != nil {
		current.System.TRACEID = new.System.TRACEID
	}

	if new.System.EXECTIMEOUT != nil {
		current.System.EXECTIMEOUT = new.System.EXECTIMEOUT
	}
	current.resolveRelativePaths()
}

func (current *Constants) resolveRelativePaths() {
	// Relative to Absolute Path
	*current.System.INPUTDIR = filepath.Join(*current.System.BASEPATH, *current.System.INPUTDIR)
	*current.System.OUTPUTDIR = filepath.Join(*current.System.BASEPATH, *current.System.OUTPUTDIR)
	*current.System.RESULTSJSON = filepath.Join(*current.System.BASEPATH, *current.System.RESULTSJSON)
	*current.System.RESULTSSCHEMA = filepath.Join(*current.System.BASEPATH, *current.System.RESULTSSCHEMA)

}

// Constants
type Constants struct {
	_process ProcessConstants `json:"-"`
	Process  ProcessConstants `json:"process"`

	_reserved ReservedConstants `json:"-"`
	Reserved  ReservedConstants `json:"reserved"`

	_system SystemConstants `json:"-"`
	System  SystemConstants `json:"system"`
}

func (constants Constants) GenerateMapForProcessEnv() map[string]string {

	dataMap := make(map[string]string)
	if constants.Process.INPUTAPI != nil {
		dataMap["input_api"] = *constants.Process.INPUTAPI
	}
	if constants.Process.INPUTAPITOKEN != nil {

		dataMap["input_api_token"] = *constants.Process.INPUTAPITOKEN
	}
	if constants.Process.OUTPUTAPI != nil {

		dataMap["output_api"] = *constants.Process.OUTPUTAPI
	}
	if constants.Process.FILEUPLOADAPI != nil {

		dataMap["file_upload_api"] = *constants.Process.FILEUPLOADAPI
	}

	return dataMap

}

func (constants Constants) GenerateMapForSystemEnv() map[string]string {

	dataMap := make(map[string]string)

	if constants.System.BASEPATH != nil {
		dataMap["base_path"] = *constants.System.BASEPATH
	}
	if constants.System.INPUTDIR != nil {
		dataMap["input_dir"] = *constants.System.INPUTDIR

	}

	if constants.System.OUTPUTDIR != nil {
		dataMap["output_dir"] = *constants.System.OUTPUTDIR

	}

	if constants.System.RESULTSJSON != nil {
		dataMap["results_json"] = *constants.System.RESULTSJSON

	}

	if constants.System.RESULTSSCHEMA != nil {
		dataMap["results_schema"] = *constants.System.RESULTSSCHEMA

	}

	if constants.System.LOGTOFILE != nil {
		dataMap["log_to_file"] = strconv.FormatBool(*constants.System.LOGTOFILE)

	}

	if constants.System.STDOUTFILE != nil {
		dataMap["stdout_file"] = *constants.System.STDOUTFILE

	}

	if constants.System.STDERRFILE != nil {
		dataMap["stderr_file"] = *constants.System.STDERRFILE

	}

	if constants.System.ENABLELOGGING != nil {
		dataMap["enable_logging"] = strconv.FormatBool(*constants.System.ENABLELOGGING)

	}

	if constants.System.MAXOUTPUTFILESIZE != nil {
		dataMap["max_output_filesize"] = strconv.Itoa(*constants.System.MAXOUTPUTFILESIZE)

	}

	if constants.System.SAMPLEINPUTFILE != nil {
		dataMap["sample_inputfile"] = *constants.System.SAMPLEINPUTFILE

	}

	if constants.System.SAMPLEOUTPUTFILE != nil {
		dataMap["sample_outputfile"] = *constants.System.SAMPLEOUTPUTFILE

	}

	if constants.System.TRACEID != nil {
		dataMap["trace_id"] = *constants.System.TRACEID

	}

	if constants.System.EXECTIMEOUT != nil {
		dataMap["exec_timeout"] = *constants.System.EXECTIMEOUT

	}

	return dataMap

}

func initialize() *Constants {

	reservedConstants := ReservedConstants{
		MESSAGEQUEUE:        *PopulateStr("message_queue_topic", "tasks_test", "Message Queue Topic"),
		PUBLISHMESSAGEQUEUE: *PopulateStr("publish_message_queue_topic", "tasks_publish", "Publish Message Queue Topic"),
	}

	systemConstants := SystemConstants{
		BASEPATH:          PopulateStr("base_path", "/Users/meddler/Office/Workspaces/Secoflex/secoflex/modules/watchdog/tmp", "Base Path"),
		INPUTDIR:          PopulateStr("input_dir", "input", "Specify output directory"),
		OUTPUTDIR:         PopulateStr("output_dir", "output", "Specify output directory"),
		RESULTSJSON:       PopulateStr("results_json", "results.json", "Specify output directory"),
		RESULTSSCHEMA:     PopulateStr("results_schema", "schema.json", "Specify output directory"),
		LOGTOFILE:         PopulateBool("log_to_file", false, "Specify output directory"),
		STDOUTFILE:        PopulateStr("stdout_file", "schema", "Specify output directory"),
		STDERRFILE:        PopulateStr("stderr_file", "schema", "Specify output directory"),
		ENABLELOGGING:     PopulateBool("enable_logging", true, "Enable Logging"),
		MAXOUTPUTFILESIZE: PopulateInt("max_output_filesize", 500, "Enable Logging"),
		SAMPLEINPUTFILE:   PopulateStr("sample_inputfile", "PopulateStr", "Enable Logging"),
		SAMPLEOUTPUTFILE:  PopulateStr("sample_outputfile", "PopulateStr", "Enable Logging"),
		TRACEID:           PopulateStr("trace_id", "default_trace_id", "Trace Id"),
	}

	processConstants := ProcessConstants{

		INPUTAPI:      PopulateStr("input_api", "input", "Specify output directory"),
		INPUTAPITOKEN: PopulateStr("input_api_token", "input", "Specify output directory"),
		FILEUPLOADAPI: PopulateStr("output_api", "input", "Specify output directory"),
		OUTPUTAPI:     PopulateStr("file_upload_api", "input", "Specify output directory"),
	}

	_processConstants := &ProcessConstants{}
	_reservedConstants := &ReservedConstants{}
	_systemConstants := &SystemConstants{}

	copier.Copy(&_processConstants, &processConstants)
	copier.Copy(&_reservedConstants, &reservedConstants)
	copier.Copy(&_systemConstants, &systemConstants)

	return &Constants{

		_reserved: *_reservedConstants,
		Reserved:  reservedConstants,

		_system: *_systemConstants,
		System:  systemConstants,

		_process: *_processConstants,
		Process:  processConstants,
	}
}

func (constants *Constants) reset() {

	copier.Copy(&constants.System, &constants._system)
	copier.Copy(&constants.Process, &constants._process)

}

// Reset
func (constants *Constants) Reset() {
	constants.reset()
}

// Exprted CONSTANTS
var (
	CONSTANTS *Constants
)

func init() {
	flag.Parse()
	CONSTANTS = initialize()

}
