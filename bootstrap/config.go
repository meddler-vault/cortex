package bootstrap

import (
	"encoding/json"
	"fmt"
)

// DependencySchema
type DependencySchema struct {
	Identifier string `json:"id"`
	Alias      string `json:"alias"`
	Type       string `json:"asset"`
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

type TaskResultBase struct {
	Status          string `json:"exec_status" validate:"required"`
	Message         string `json:"message" validate:"required"`
	WatchdogVersion string `json:"watchdog_version" validate:"required"`
	Identifier      string `json:"identifier" validate:"required"`
}

type TaskResult struct {
	TaskResultBase
	Response string `json:"response" ` // success_endpoint
}
