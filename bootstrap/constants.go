package bootstrap

import "flag"

// Configurable Constants
var (
	BASEPATH          = PopulateStr("BASEPATH", "/Users/meddler/Office/Workspaces/Secoflex/secoflex/modules/watchdog/tmp", "Base Path")
	INPUTDIR          = PopulateStr("INPUTDIR", "input", "Specify output directory")
	OUTPUTDIR         = PopulateStr("OUTPUTDIR", "ouput", "Specify output directory")
	RESULTSJSON       = PopulateStr("RESULTSJSON", "ouput", "Specify output directory")
	RESULTSSCHEMA     = PopulateStr("RESULTSSCHEMA", "schema", "Specify output directory")
	LOGTOFILE         = PopulateBool("LOGTOFILE", false, "Specify output directory")
	STDOUTFILE        = PopulateStr("STDOUTFILE", "schema", "Specify output directory")
	STDERRFILE        = PopulateStr("STDERRFILE", "schema", "Specify output directory")
	ENABLELOGGING     = PopulateBool("ENABLELOGGING", true, "Enable Logging")
	MWXOUTPUTFILESIZE = PopulateInt("MWXOUTPUTFILESIZE", 500, "Enable Logging")
	SAMPLEINPUTFILE   = PopulateStr("SAMPLEINPUTFILE", "PopulateStr", "Enable Logging")
	SAMPLEOUTPUTFILE  = PopulateStr("SAMPLEOUTPUTFILE", "PopulateStr", "Enable Logging")
)

// Watchdog: Constants
var (
	INPUTAPI      = PopulateStr("INPUTAPI", "input", "Specify output directory")
	INPUTAPITOKEN = PopulateStr("INPUTAPITOKEN", "input", "Specify output directory")
	FILEUPLOADAPI = PopulateStr("FILEUPLOADAPI", "input", "Specify output directory")
	OUTPUTAPI     = PopulateStr("OUTPUTAPI", "input", "Specify output directory")
)

func init() {
	flag.Parse()
}
