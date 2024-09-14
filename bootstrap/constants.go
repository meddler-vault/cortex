package bootstrap

import (
	"flag"
	"fmt"
	"log"

	"github.com/meddler-vault/cortex/logger"

	"path/filepath"
	"strconv"
	"strings"

	"github.com/jinzhu/copier"
)

type BaseConstants struct {
}

var DEBUG = false

// Cortex Mode : task_worker , image_builder , task_result_processor , image_builder_result_processor
// CortexMode is a custom type to represent different modes as strings.
type CortexMode string

// Define the possible values for CortexMode.
const (
	CortexModeTaskWorker                  CortexMode = "task_worker"
	CortexModeImageBuilder                CortexMode = "image_builder"
	CortexModeTaskResultProcessor         CortexMode = "task_result_processor"
	CortexModeImageBuilderResultProcessor CortexMode = "image_builder_result_processor"
	CortexModeResultProcessor             CortexMode = "result_processor"

	// Define the default mode
	DefaultCortexMode CortexMode = CortexModeTaskWorker
)

// getCortexModeFromEnv retrieves the CortexMode based on an environment variable.
func getCortexMode(mode string, defaultCortexMode CortexMode) CortexMode {

	mode = populateStringFromEnv(mode, "")
	log.Println("populateStringFromEnv", mode)
	switch mode {
	case string(CortexModeTaskWorker):
		return CortexModeTaskWorker
	case string(CortexModeImageBuilder):
		return CortexModeImageBuilder
	case string(CortexModeTaskResultProcessor):
		return CortexModeTaskResultProcessor
	case string(CortexModeImageBuilderResultProcessor):
		return CortexModeImageBuilderResultProcessor
	case string(CortexModeResultProcessor):
		return CortexModeResultProcessor
	default:
		fmt.Println("Invalid or unset CORTEX_MODE, defaulting to:", defaultCortexMode)
		return defaultCortexMode
	}
}

const TASK_MESSAGE_QUEUE_SUBJECT_PREFIX = "task"
const TASK_MESSAGE_QUEUE_SUBJECT_PREFIX_CONSUMER_GROUP_NAME = "task_worker"

const RESULT_MESSAGE_QUEUE_SUBJECT_PREFIX = "result.task.*"
const RESULT_MESSAGE_QUEUE_SUBJECT_PREFIX_CONSUMER_GROUP_NAME = "task_result_worker"

const RESULT_MESSAGE_QUEUE_SUBJECT_PREFIX_COMMON_PARENT = "result.>"

const BUILD_ASSEMBLER_MESSAGE_QUEUE_SUBJECT_PREFIX = "build.*"
const BUILD_ASSEMBLER_MESSAGE_QUEUE_SUBJECT_PREFIX_CONSUMER_GROUP_NAME = "build_worker"

const RESULT_BUILD_ASSEMBLER_MESSAGE_QUEUE_SUBJECT_PREFIX = "result.build.*"
const RESULT_BUILD_ASSEMBLER_MESSAGE_QUEUE_SUBJECT_PREFIX_CONSUMER_GROUP_NAME = "build_result_worker"

const COMMON_MESSAGE_QUEUE_SUBJECT_PREFIX_CONSUMER_GROUP_NAME = "result_worker"

// Publishing subjects
const TASKS_MESSAGE_QUEUE_SUBJECT_PUBLISH = "result.task"
const BUILD_MESSAGE_QUEUE_SUBJECT_PUBLISH = "result.build"

func GetCortexMode(mode string, defaultCortexMode CortexMode) CortexMode {

	cortex_mode := getCortexMode(mode, defaultCortexMode)

	if cortex_mode == CortexModeTaskWorker {
		CORTEX_MQ_PUBLISHER_SUBJECT = TASK_MESSAGE_QUEUE_SUBJECT_PREFIX
		CORTEX_MQ_CONSUMER_SUBJECT = TASKS_MESSAGE_QUEUE_SUBJECT_PUBLISH + "." + "result"
		CORTEX_MQ_CONSUMER_NAME = TASK_MESSAGE_QUEUE_SUBJECT_PREFIX_CONSUMER_GROUP_NAME
	} else if cortex_mode == CortexModeImageBuilder {
		CORTEX_MQ_CONSUMER_SUBJECT = BUILD_ASSEMBLER_MESSAGE_QUEUE_SUBJECT_PREFIX
		CORTEX_MQ_PUBLISHER_SUBJECT = BUILD_MESSAGE_QUEUE_SUBJECT_PUBLISH
		CORTEX_MQ_CONSUMER_NAME = BUILD_ASSEMBLER_MESSAGE_QUEUE_SUBJECT_PREFIX_CONSUMER_GROUP_NAME
	} else if cortex_mode == CortexModeTaskResultProcessor {
		CORTEX_MQ_CONSUMER_SUBJECT = RESULT_MESSAGE_QUEUE_SUBJECT_PREFIX
		CORTEX_MQ_PUBLISHER_SUBJECT = ""
		CORTEX_MQ_CONSUMER_NAME = RESULT_MESSAGE_QUEUE_SUBJECT_PREFIX_CONSUMER_GROUP_NAME
	} else if cortex_mode == CortexModeImageBuilderResultProcessor {
		CORTEX_MQ_CONSUMER_SUBJECT = RESULT_BUILD_ASSEMBLER_MESSAGE_QUEUE_SUBJECT_PREFIX
		CORTEX_MQ_PUBLISHER_SUBJECT = ""
		CORTEX_MQ_CONSUMER_NAME = RESULT_BUILD_ASSEMBLER_MESSAGE_QUEUE_SUBJECT_PREFIX_CONSUMER_GROUP_NAME
	} else if cortex_mode == CortexModeResultProcessor {
		CORTEX_MQ_CONSUMER_SUBJECT = RESULT_MESSAGE_QUEUE_SUBJECT_PREFIX_COMMON_PARENT
		CORTEX_MQ_PUBLISHER_SUBJECT = ""
		CORTEX_MQ_CONSUMER_NAME = COMMON_MESSAGE_QUEUE_SUBJECT_PREFIX_CONSUMER_GROUP_NAME
	}

	log.Println("cortex_mode", cortex_mode)

	return cortex_mode

}

// CortexConstants{
var (
	CORTEX_MQ_CONSUMER_SUBJECT  string = ""
	CORTEX_MQ_PUBLISHER_SUBJECT string = ""
	CORTEX_MQ_CONSUMER_NAME     string = ""
)

type ReservedConstants struct {
	BaseConstants

	CORTEXUUID string `json:"cortex_uuid"`

	CORTEXMODE CortexMode `json:"cortex_mode"`

	CORTEXPINGURL      string `json:"cortex_ping_url"`
	CORTEXPINGINTERVAL int    `json:"cortex_ping_interval"`

	MESSAGEQUEUE        string `json:"message_queue_topic"`
	PUBLISHMESSAGEQUEUE string `json:"publish_message_queue_topic"`
	PUBLISHSUBJECT      string `json:"publish_subject"`
	MOCKMESSAGE         string `json:"mock_message"`
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

	CORTEXID *string `json:"cortex_id"`

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
	// Adding new configurable parameters for GIT Cloner
	// Git Mode: true | True | TRUE | yes | 1 ; else False
	GITMODE           *bool   `json:"git_mode" `
	GITAUTHMODE       *string `json:"git_auth_mode" `
	GITAUTHUSERNAME   *string `json:"git_auth_username" `
	GITAUTHPASSWORD   *string `json:"git_auth_password" `
	GITREMOTE         *string `json:"git_remote" `
	GITPATH           *string `json:"git_path" ` // Output path for git-repo
	GITREF            *string `json:"git_ref" `  // Output path for git-repo
	GITBASECOMMITID   *string `json:"git_base_commit_id"`
	GITTARGETCOMMITID *string `json:"git_target_commit_id"`
	GITDEPTH          *int    `json:"git_depth"`

	// Job Output result publishing
	JOBMODE *string `json:"job_mode" `

	// Result sync params

	RESULT_FILE_PATH      *string `json:"result_file_path" `      // Path to file containing the result. TODO: If pattern .*.json , etc...parse all json files
	RESULT_PARSER_TYPE    *string `json:"result_parser_type" `    // common dojo parsers
	RESULT_PARSER_NAME    *string `json:"result_parser_name" `    // dojo sbom etc
	RESULT_TYPE           *string `json:"result_type" `           // Output result type
	RESULT_SYNC_DIRECTORY *string `json:"result_sync_directory" ` // Sync result directory to minio storage
	RESULT_PARSE          *bool   `json:"result_parse" `          // true , false : Needs parsing: Yes / no..if no...storage mounted directory files will be listed, else parsing results will b e

	// Read only volume mount from minio / s3
	MOUNT_VOLUME               *bool   `json:"mount_volume" `                // To mount the volume path or not. It is mandatoru to successfuly mount if true else the process fails
	MOUNT_VOLUME_PATH          *string `json:"mount_volume_path" `           // Relative volume mount point on base_path
	MOUNT_VOLUME_FOLDER_PATH   *string `json:"mount_volume_s3_folder_path" ` // if empty..go to object path to sunc the file
	MOUNT_VOLUME_OBJECT_PATH   *string `json:"mount_volume_s3_object_path" ` // if empty..the folder is synced else the object is synced
	MOUNT_VOLUME_S3_ACCESS_KEY *string `json:"mount_volume_s3_access_key" `
	MOUNT_VOLUME_BUCKET        *string `json:"mount_volume_s3_bucket" `
	MOUNT_VOLUME_S3_SECRET_KEY *string `json:"mount_volume_s3_secret_key" `
	MOUNT_VOLUME_S3_SECURE     *bool   `json:"mount_volume_s3_secure" ` // To mount the volume path or not. It is mandatoru to successfuly mount if true else the process fails
	MOUNT_VOLUME_S3_HOST       *string `json:"mount_volume_s3_host" `
	MOUNT_VOLUME_S3_REGION     *string `json:"mount_volume_s3_region" `

	// Write only volume mount to minio / s3
	EXPORT_VOLUME               *bool   `json:"export_volume" `                // To mount the volume path or not. It is mandatoru to successfuly mount if true else the process fails
	EXPORT_VOLUME_PATH          *string `json:"export_volume_path" `           // Relative volume mount point on base_path
	EXPORT_VOLUME_FOLDER_PATH   *string `json:"export_volume_s3_folder_path" ` // if empty..go to object path to sunc the file
	EXPORT_VOLUME_OBJECT_PATH   *string `json:"export_volume_s3_object_path" ` // if empty..the folder is synced else the object is synced
	EXPORT_VOLUME_S3_ACCESS_KEY *string `json:"export_volume_s3_access_key" `
	EXPORT_VOLUME_BUCKET        *string `json:"export_volume_s3_bucket" `
	EXPORT_VOLUME_S3_SECRET_KEY *string `json:"export_volume_s3_secret_key" `
	EXPORT_VOLUME_S3_SECURE     *bool   `json:"export_volume_s3_secure" ` // To mount the volume path or not. It is mandatoru to successfuly mount if true else the process fails
	EXPORT_VOLUME_S3_HOST       *string `json:"export_volume_s3_host" `
	EXPORT_VOLUME_S3_REGION     *string `json:"export_volume_s3_region" `

	// Host as variable
	HOST                    *string `json:"host" `                    // Host name Ex: example.com
	IP_ADDRESS              *string `json:"ip" `                      // Host name Ex: example.com
	IP_ADDRESS_V4           *string `json:"ip_v4" `                   // Host name Ex: example.com
	IP_ADDRESS_V6           *string `json:"ip_v6" `                   // Host name Ex: example.com
	URL                     *string `json:"url" `                     // Host name Ex: example.com
	FQDN                    *string `json:"fqdn" `                    // Host name Ex: example.com
	ANDROID_APK             *string `json:"android_apk_path" `        // Host name Ex: example.com
	IOS_IPA                 *string `json:"ios_ipa_path" `            // Host name Ex: example.com
	POSTMAN_COLLECTION_JSON *string `json:"postman_collection_json" ` // Host name Ex: example.com
	SWAGGER_COLLECTION_JSON *string `json:"swagger_json" `            // Host name Ex: example.com

}

// Git Constants: Auth Mode
const (
	NOAUTH     string = "none"
	BASICAUTH  string = "basicauth"
	TOKEN      string = "token"
	PRIVATEKEY string = "privatekey"
)

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
	// Git modification
	if new.System.GITMODE != nil {
		current.System.GITMODE = new.System.GITMODE
	}
	if new.System.GITAUTHMODE != nil {
		current.System.GITAUTHMODE = new.System.GITAUTHMODE
	}
	if new.System.GITREMOTE != nil {
		current.System.GITREMOTE = new.System.GITREMOTE
	}
	if new.System.GITPATH != nil {
		current.System.GITPATH = new.System.GITPATH
	}
	if new.System.GITAUTHUSERNAME != nil {
		current.System.GITAUTHUSERNAME = new.System.GITAUTHUSERNAME
	}
	if new.System.GITAUTHPASSWORD != nil {
		current.System.GITAUTHPASSWORD = new.System.GITAUTHPASSWORD
	}
	if new.System.GITREF != nil {
		current.System.GITREF = new.System.GITREF
	}
	if new.System.GITBASECOMMITID != nil {
		current.System.GITBASECOMMITID = new.System.GITBASECOMMITID
	}
	if new.System.GITTARGETCOMMITID != nil {
		current.System.GITTARGETCOMMITID = new.System.GITTARGETCOMMITID
	}
	if new.System.GITDEPTH != nil {
		current.System.GITDEPTH = new.System.GITDEPTH
	}

	// Mount volume configs
	if new.System.MOUNT_VOLUME != nil {
		current.System.MOUNT_VOLUME = new.System.MOUNT_VOLUME
	}
	if new.System.MOUNT_VOLUME_PATH != nil {
		current.System.MOUNT_VOLUME_PATH = new.System.MOUNT_VOLUME_PATH
	}
	if new.System.MOUNT_VOLUME_BUCKET != nil {
		current.System.MOUNT_VOLUME_BUCKET = new.System.MOUNT_VOLUME_BUCKET
	}
	if new.System.MOUNT_VOLUME_OBJECT_PATH != nil {
		current.System.MOUNT_VOLUME_OBJECT_PATH = new.System.MOUNT_VOLUME_OBJECT_PATH
	}
	if new.System.MOUNT_VOLUME_FOLDER_PATH != nil {
		current.System.MOUNT_VOLUME_FOLDER_PATH = new.System.MOUNT_VOLUME_FOLDER_PATH
	}
	if new.System.MOUNT_VOLUME_S3_ACCESS_KEY != nil {
		current.System.MOUNT_VOLUME_S3_ACCESS_KEY = new.System.MOUNT_VOLUME_S3_ACCESS_KEY
	}
	if new.System.MOUNT_VOLUME_S3_SECRET_KEY != nil {
		current.System.MOUNT_VOLUME_S3_SECRET_KEY = new.System.MOUNT_VOLUME_S3_SECRET_KEY
	}
	if new.System.MOUNT_VOLUME_S3_HOST != nil {
		current.System.MOUNT_VOLUME_S3_HOST = new.System.MOUNT_VOLUME_S3_HOST
	}
	if new.System.MOUNT_VOLUME_S3_SECURE != nil {
		current.System.MOUNT_VOLUME_S3_SECURE = new.System.MOUNT_VOLUME_S3_SECURE
	}

	// Export volume to  configs
	if new.System.EXPORT_VOLUME != nil {
		current.System.EXPORT_VOLUME = new.System.EXPORT_VOLUME
	}
	if new.System.EXPORT_VOLUME_PATH != nil {
		current.System.EXPORT_VOLUME_PATH = new.System.EXPORT_VOLUME_PATH
	}
	if new.System.EXPORT_VOLUME_FOLDER_PATH != nil {
		current.System.EXPORT_VOLUME_FOLDER_PATH = new.System.EXPORT_VOLUME_FOLDER_PATH
	}
	if new.System.EXPORT_VOLUME_OBJECT_PATH != nil {
		current.System.EXPORT_VOLUME_OBJECT_PATH = new.System.EXPORT_VOLUME_OBJECT_PATH
	}
	if new.System.EXPORT_VOLUME_S3_ACCESS_KEY != nil {
		current.System.EXPORT_VOLUME_S3_ACCESS_KEY = new.System.EXPORT_VOLUME_S3_ACCESS_KEY
	}
	if new.System.EXPORT_VOLUME_BUCKET != nil {
		current.System.EXPORT_VOLUME_BUCKET = new.System.EXPORT_VOLUME_BUCKET
	}
	if new.System.EXPORT_VOLUME_S3_SECRET_KEY != nil {
		current.System.EXPORT_VOLUME_S3_SECRET_KEY = new.System.EXPORT_VOLUME_S3_SECRET_KEY
	}

	if new.System.EXPORT_VOLUME_S3_HOST != nil {
		current.System.EXPORT_VOLUME_S3_HOST = new.System.EXPORT_VOLUME_S3_HOST
	}

	if new.System.EXPORT_VOLUME_S3_SECURE != nil {
		current.System.EXPORT_VOLUME_S3_SECURE = new.System.EXPORT_VOLUME_S3_SECURE
	}

	// scanner variables
	if new.System.HOST != nil {
		current.System.HOST = new.System.HOST
	}
	if new.System.IP_ADDRESS != nil {
		current.System.IP_ADDRESS = new.System.IP_ADDRESS
	}
	if new.System.IP_ADDRESS_V4 != nil {
		current.System.IP_ADDRESS_V4 = new.System.IP_ADDRESS_V4
	}
	if new.System.URL != nil {
		current.System.URL = new.System.URL
	}

	current.resolveRelativePaths()
}

func (dependencySchema *DependencySchema) ResolveRelativePathsInDependencies(basePath string) {
	// Relative to Absolute Path
	*dependencySchema.MOUNT_VOLUME_PATH = filepath.Join(basePath, *dependencySchema.MOUNT_VOLUME_PATH)
}

func (current *Constants) resolveRelativePaths() {
	// Relative to Absolute Path
	*current.System.INPUTDIR = filepath.Join(*current.System.BASEPATH, *current.System.INPUTDIR)
	*current.System.OUTPUTDIR = filepath.Join(*current.System.BASEPATH, *current.System.OUTPUTDIR)
	*current.System.RESULTSJSON = filepath.Join(*current.System.BASEPATH, *current.System.RESULTSJSON)
	*current.System.RESULTSSCHEMA = filepath.Join(*current.System.BASEPATH, *current.System.RESULTSSCHEMA)
	*current.System.GITPATH = filepath.Join(*current.System.BASEPATH, *current.System.GITPATH)
	*current.System.RESULT_FILE_PATH = filepath.Join(*current.System.OUTPUTDIR, *current.System.RESULT_FILE_PATH)

	// Volume mount path
	*current.System.MOUNT_VOLUME_PATH = filepath.Join(*current.System.BASEPATH, *current.System.MOUNT_VOLUME_PATH)
	// Export volume path
	*current.System.EXPORT_VOLUME_PATH = filepath.Join(*current.System.BASEPATH, *current.System.EXPORT_VOLUME_PATH)

	// Scanner Inputs Path
	// This will go in dependency part resolution: Just above function

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

	if constants.System.GITAUTHMODE != nil {
		dataMap["git_auth_mode"] = *constants.System.GITAUTHMODE

	}
	if constants.System.GITAUTHUSERNAME != nil {
		dataMap["git_auth_username"] = *constants.System.GITAUTHUSERNAME

	}
	if constants.System.GITAUTHPASSWORD != nil {
		dataMap["git_auth_password"] = *constants.System.GITAUTHPASSWORD

	}
	if constants.System.GITPATH != nil {
		dataMap["git_path"] = *constants.System.GITPATH

	}

	// TODO: this si reserved variable should be passsed in to all the pipeline
	if constants.System.GITPATH != nil {
		dataMap["git_remote"] = *constants.System.GITREMOTE

	}

	if constants.System.GITREF != nil {
		dataMap["git_ref"] = *constants.System.GITREF

	}
	if constants.System.GITBASECOMMITID != nil {
		dataMap["git_base_commit_id"] = *constants.System.GITBASECOMMITID

	}
	if constants.System.GITTARGETCOMMITID != nil {
		dataMap["git_target_commit_id"] = *constants.System.GITTARGETCOMMITID

	}
	if constants.System.GITDEPTH != nil {
		dataMap["git_depth"] = strconv.Itoa(*constants.System.GITDEPTH)

	}

	// For minio
	if constants.System.MOUNT_VOLUME_PATH != nil {
		dataMap["mount_volume_path"] = *constants.System.MOUNT_VOLUME_PATH

	}

	if constants.System.MOUNT_VOLUME_FOLDER_PATH != nil {
		dataMap["mount_volume_s3_folder_path"] = *constants.System.MOUNT_VOLUME_FOLDER_PATH

	}

	if constants.System.MOUNT_VOLUME_OBJECT_PATH != nil {
		dataMap["mount_volume_s3_object_path"] = *constants.System.MOUNT_VOLUME_OBJECT_PATH

	}

	// For export

	if constants.System.EXPORT_VOLUME_PATH != nil {
		dataMap["export_volume_path"] = *constants.System.EXPORT_VOLUME_PATH
	}

	if constants.System.EXPORT_VOLUME_FOLDER_PATH != nil {
		dataMap["export_volume_s3_folder_path"] = *constants.System.EXPORT_VOLUME_FOLDER_PATH

	}

	if constants.System.EXPORT_VOLUME_OBJECT_PATH != nil {
		dataMap["export_volume_s3_object_path"] = *constants.System.EXPORT_VOLUME_OBJECT_PATH

	}

	// For scanner related variables
	if constants.System.HOST != nil {
		dataMap["host"] = *constants.System.HOST
	}

	if constants.System.IP_ADDRESS != nil {
		dataMap["ip"] = *constants.System.IP_ADDRESS
	}

	if constants.System.IP_ADDRESS_V4 != nil {
		dataMap["ip_v4"] = *constants.System.IP_ADDRESS_V4
	}

	if constants.System.IP_ADDRESS_V6 != nil {
		dataMap["ip_v6"] = *constants.System.IP_ADDRESS_V6
	}

	if constants.System.URL != nil {
		dataMap["url"] = *constants.System.URL
	}

	if constants.System.FQDN != nil {
		dataMap["fqdn"] = *constants.System.FQDN
	}

	if constants.System.ANDROID_APK != nil {
		dataMap["android_apk_path"] = *constants.System.ANDROID_APK
	}

	if constants.System.IOS_IPA != nil {
		dataMap["ios_ipa_path"] = *constants.System.IOS_IPA
	}

	if constants.System.POSTMAN_COLLECTION_JSON != nil {
		dataMap["postman_collection_json"] = *constants.System.POSTMAN_COLLECTION_JSON
	}

	if constants.System.SWAGGER_COLLECTION_JSON != nil {
		dataMap["swagger_json"] = *constants.System.SWAGGER_COLLECTION_JSON
	}

	return dataMap

}

func initialize() *Constants {

	DEBUG = *PopulateBool("DEBUG", false, "Publish Subject")
	log.Println("BUGMODE", DEBUG)

	reservedConstants := ReservedConstants{

		CORTEXUUID: *PopulateStr("CORTEX_UUID", "", ""),

		CORTEXMODE: GetCortexMode("CORTEX_MODE", DefaultCortexMode),

		CORTEXPINGURL:      *PopulateStr("CORTEX_PING_URL", "", ""),
		CORTEXPINGINTERVAL: *PopulateInt("CORTEX_PING_INTERVAL", 10, ""),

		MESSAGEQUEUE:        *PopulateStr("message_queue_topic", "tasks_test", "Message Queue Topic"),
		PUBLISHMESSAGEQUEUE: *PopulateStr("publish_message_queue_topic", "task_result", "Publish Message Queue Topic"),
		PUBLISHSUBJECT:      *PopulateStr("publish_subject", "jobs", "Publish Subject"),

		MOCKMESSAGE: *PopulateStr("mock_message", "", "Test message to mock on init."),
	}

	systemConstants := SystemConstants{
		BASEPATH:          PopulateStr("base_path", "/tmp", "Base Path"),
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
		// Git Operations
		GITMODE:           PopulateBool("git_mode", false, "Enable GIT Mounting"),
		GITAUTHMODE:       PopulateStr("git_auth_mode", "no_auth", "Git auth mode"),
		GITAUTHUSERNAME:   PopulateStr("git_auth_username", "", "Git Auth Username"),
		GITAUTHPASSWORD:   PopulateStr("git_auth_password", "", "Git Auth Password"),
		GITREMOTE:         PopulateStr("git_remote", "", "Git Remote"),
		GITPATH:           PopulateStr("git_path", "/git-repo", "Git Path"),
		GITDEPTH:          PopulateInt("git_depth", 0, "Git depth while cloning"),
		GITREF:            PopulateStr("git_ref", "", "Git ref (tag/ branch). Use fully qualified git refernece"),
		GITBASECOMMITID:   PopulateStr("git_base_commit_id", "", "Git based commit id. (top / latest)"),
		GITTARGETCOMMITID: PopulateStr("git_target_commit_id", "", "Git historcal commit it to recurse to!"),

		// RESULT params
		RESULT_FILE_PATH:      PopulateStr("result_file_path", "", "Path to file containing the result. TODO: If pattern .*.json , etc...parse all json !"),
		RESULT_PARSER_TYPE:    PopulateStr("result_parser_type", "", "common dojo parsers!"),
		RESULT_PARSER_NAME:    PopulateStr("result_parser_name", "", "dojo sbom etc!"),
		RESULT_TYPE:           PopulateStr("result_type", "", "Result Type!"),
		RESULT_SYNC_DIRECTORY: PopulateStr("result_sync_directory", "", "To Sync  results directory to minio or not!"),
		RESULT_PARSE:          PopulateBool("result_parse", false, "To parse results or not!"),

		// Mount volume constants
		MOUNT_VOLUME:               PopulateBool("mount_volume", false, "To mount the volume"),
		MOUNT_VOLUME_PATH:          PopulateStr("mount_volume_path", "mount", "Result Type!"),
		MOUNT_VOLUME_BUCKET:        PopulateStr("mount_volume_s3_bucket", "", "Result Type!"),
		MOUNT_VOLUME_OBJECT_PATH:   PopulateStr("mount_volume_s3_object_path", "", "Result Type!"),
		MOUNT_VOLUME_FOLDER_PATH:   PopulateStr("mount_volume_s3_folder_path", "", "Result Type!"),
		MOUNT_VOLUME_S3_ACCESS_KEY: PopulateStr("mount_volume_s3_access_key", "", "Result Type!"),
		MOUNT_VOLUME_S3_SECRET_KEY: PopulateStr("mount_volume_s3_secret_key", "", "Result Type!"),
		MOUNT_VOLUME_S3_SECURE:     PopulateBool("mount_volume_s3_secure", true, "To mount the volume"),
		MOUNT_VOLUME_S3_HOST:       PopulateStr("mount_volume_s3_host", "", "Result Type!"),
		MOUNT_VOLUME_S3_REGION:     PopulateStr("mount_volume_s3_region", "auto", "Region!"),

		// Export to minio constants
		EXPORT_VOLUME:               PopulateBool("export_volume", false, "To mount the volume"),
		EXPORT_VOLUME_PATH:          PopulateStr("export_volume_path", "export", "Result Type!"),
		EXPORT_VOLUME_BUCKET:        PopulateStr("export_volume_s3_bucket", "", "Result Type!"),
		EXPORT_VOLUME_OBJECT_PATH:   PopulateStr("export_volume_s3_object_path", "", "Result Type!"),
		EXPORT_VOLUME_FOLDER_PATH:   PopulateStr("export_volume_s3_folder_path", "", "Result Type!"),
		EXPORT_VOLUME_S3_ACCESS_KEY: PopulateStr("export_volume_s3_access_key", "", "Result Type!"),
		EXPORT_VOLUME_S3_SECRET_KEY: PopulateStr("export_volume_s3_secret_key", "", "Result Type!"),
		EXPORT_VOLUME_S3_SECURE:     PopulateBool("export_volume_s3_secure", true, "To mount the volume"),
		EXPORT_VOLUME_S3_HOST:       PopulateStr("export_volume_s3_host", "", "Result Type!"),
		EXPORT_VOLUME_S3_REGION:     PopulateStr("export_volume_s3_region", "auto", "Region!"),

		// Scaner based variables
		// Host as variable
		HOST:          PopulateStr("host", "", "Host name"),
		IP_ADDRESS:    PopulateStr("ip", "", "Host name"),
		IP_ADDRESS_V4: PopulateStr("ip_v4", "", "Host name"),
		IP_ADDRESS_V6: PopulateStr("ip_v6", "", "Host name"),
		URL:           PopulateStr("url", "", "Host name"),
		FQDN:          PopulateStr("fqdn", "", "Host name"),
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
