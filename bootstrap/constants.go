package bootstrap

import "flag"

// Container Constants
var (
	BASEPATH          = PopulateStr("BASEPATH", "/Users/meddler/Office/Workspaces/Secoflex/secoflex/modules/watchdog/tmp", "Base Path")
	INPUTDIR          = PopulateStr("INPUTDIR", "input", "Specify output directory")
	OUTPUTDIR         = PopulateStr("OUTPUTDIR", "ouput", "Specify output directory")
	SCHEMADIR         = PopulateStr("SCHEMADIR", "schema", "Specify output directory")
	LOGDIR            = PopulateStr("LOGDIR", "schema", "Specify output directory")
	STDOUTFILE        = PopulateStr("STDOUTFILE", "schema", "Specify output directory")
	STDERRFILE        = PopulateStr("STDERRFILE", "schema", "Specify output directory")
	ENABLELOGGING     = PopulateBool("ENABLELOGGING", true, "Enable Logging")
	MWXOUTPUTFILESIZE = PopulateInt("MWXOUTPUTFILESIZE", 500, "Enable Logging")
)

// Watchdog Constants
var (
	INPUTAPI      = PopulateStr("INPUTAPI", "input", "Specify output directory")
	INPUTAPITOKEN = PopulateStr("INPUTAPITOKEN", "input", "Specify output directory")
	FILEUPLOADAPI = PopulateStr("FILEUPLOADAPI", "input", "Specify output directory")
	OUTPUTAPI     = PopulateStr("OUTPUTAPI", "input", "Specify output directory")
)

func init() {
	flag.Parse()
}
