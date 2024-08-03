package bootstrap

import (
	"encoding/json"
	"fmt"
)

// DependencySchema
type DependencySchemaDeprecated struct {
	// Identifier string `json:"id"`
	// Alias      string `json:"alias"`
	// Type       string `json:"asset"`
}

// DependencySchema
type DependencySchema struct {
	MOUNT_VOLUME_VARIABLE      *string `json:"mount_volume_variable" `       // To mount the volume path or not. It is mandatoru to successfuly mount if true else the process fails
	MOUNT_VOLUME_PATH          *string `json:"mount_volume_path" `           // Relative volume mount point on base_path
	MOUNT_VOLUME_FOLDER_PATH   *string `json:"mount_volume_s3_folder_path" ` // if empty..go to object path to sunc the file
	MOUNT_VOLUME_OBJECT_PATH   *string `json:"mount_volume_s3_object_path" ` // if empty..the folder is synced else the object is synced
	MOUNT_VOLUME_S3_ACCESS_KEY *string `json:"mount_volume_s3_access_key" `
	MOUNT_VOLUME_BUCKET        *string `json:"mount_volume_s3_bucket" `
	MOUNT_VOLUME_S3_SECRET_KEY *string `json:"mount_volume_s3_secret_key" `
	MOUNT_VOLUME_S3_SECURE     *bool   `json:"mount_volume_s3_secure" ` // To mount the volume path or not. It is mandatoru to successfuly mount if true else the process fails
	MOUNT_VOLUME_S3_HOST       *string `json:"mount_volume_s3_host" `
	MOUNT_VOLUME_S3_REGION     *string `json:"mount_volume_s3_region" `
}

// MessageDataSpec
type MessageDataSpec struct {
	Dependencies []DependencySchema `json:"dependencies"`
}

// type MessageConfig struct {
// 	System   SystemConstants   `json:"system"`
// 	Process  ProcessConstants  `json:"process"`
// 	Reserved ReservedConstants `json:"reserved"`
// }

// EnvironMap is a custom type for handling environment variables
type EnvironMap map[string]string

// UnmarshalJSON custom unmarshaler for EnvironMap
func (e *EnvironMap) UnmarshalJSON(data []byte) error {
	var tempMap map[string]interface{}
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return err
	}

	result := make(map[string]string)
	for key, value := range tempMap {
		switch v := value.(type) {
		case bool:
			result[key] = fmt.Sprintf("%v", v)
		case string:
			result[key] = v
		case float64:
			result[key] = fmt.Sprintf("%f", v)
		case int:
			result[key] = fmt.Sprintf("%d", v)
		default:
			result[key] = fmt.Sprintf("%v", v)
		}
	}

	*e = result
	return nil
}

// MessageSpec...
type MessageSpec struct {
	MessageDataSpec
	Identifier string `json:"id"` // For changing status, ingesting data, persisting FS on Storage (Bucket Name)
	// SystemEnviron map[string]string `json:"system_environ"` // SystemEnviron. Variables to inject to Watchdog & override for a particular task
	// Environ             map[string]string `json:"environ"`                                 // Environ. Variables to inject / override inside the actial process
	Environ             EnvironMap        `json:"environ"`
	Entrypoint          []string          `json:"entrypoint"`                              // Override entrypoint
	Cmd                 []string          `json:"cmd" validate:"required,cmd"`             // Override CMD
	Args                []string          `json:"args" validate:"required,args"`           // Override ARGS
	SubstituteVariables bool              `json:"substitute_var" validate:"required,args"` // Parse ARGS for variables
	Variables           map[string]string `json:"variables"`                               // TODO: Variables: Replace placeholders with actual value

	Config Constants `json:"config"` // TODO: Config:

	SuccessEndpoint string `json:"success_endpoint" validate:"required,success_endpoint"` // success_endpoint
	FailureEndpoint string `json:"failure_endpoint" validate:"required,failure_endpoint"` // success_endpoint

}

// Define the custom type for Status
type TaskStatus string

const (
	ENQUEUED  TaskStatus = "ENQUEUED"
	INITIATED TaskStatus = "INITIATED"
	COMPLETED TaskStatus = "COMPLETED"
	TIMEOUT   TaskStatus = "TIMEOUT"
	SUCCESS   TaskStatus = "SUCCESS"
	FAILURE   TaskStatus = "FAILURE"
	UNKNOWN   TaskStatus = "UNKNOWN"
)

type TaskResult struct {
	TaskStatus      TaskStatus `json:"exec_status" validate:"required"`
	Message         string     `json:"message" validate:"required"`
	WatchdogVersion string     `json:"watchdog_version" validate:"required"`
	Identifier      string     `json:"identifier" validate:"required"`

	Response string `json:"response" ` // success_endpoint
}
