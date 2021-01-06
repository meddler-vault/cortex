package bootstrap

//MessageSpec...
type MessageSpec struct {
	Identifier    string            `json:"id"`
	SystemEnviron map[string]string `json:"system_environ"`
	Environ       map[string]string `json:"environ"`
	Entrypoint    []string          `json:"entrypoint"`
	Cmd           []string          `json:"cmd"`
	Config        map[string]string `json:"config"`
}
