package bootstrap

//DependencySchema
type DependencySchema struct {
	Identifier string `json:"id"`
	Alias      string `json:"alias"`
	Type       string `json:"asset"`
}

// MessageDataSpec
type MessageDataSpec struct {
	Dependencies []DependencySchema `json:"dependencies"`
}

//MessageSpec...
type MessageSpec struct {
	MessageDataSpec
	Identifier    string            `json:"id"`             // For changing status, ingesting data, persisting FS on Storage (Bucket Name)
	SystemEnviron map[string]string `json:"system_environ"` // SystemEnviron. Variables to inject to Watchdog & override for a particular task
	Environ       map[string]string `json:"environ"`        // Environ. Variables to inject / override inside the actial process
	Entrypoint    []string          `json:"entrypoint"`     // Override entrypoint
	Cmd           []string          `json:"cmd"`            // Override CMD
	Config        map[string]string `json:"config"`         // TODO: Config:
	Variables     map[string]string `json:"variables"`      // TODO: Variables: Replace placeholders with actual value

}
