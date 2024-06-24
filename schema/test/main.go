package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/meddler-vault/cortex/logger"

	"github.com/xeipuuv/gojsonschema"
)

func generateJSONSchema(baseSchema string, schema string) (schemaDefinition []byte, err error) {

	var dataset map[string]interface{}

	var schemas []byte
	var properties []byte

	// Child schema (Properties) loaded first, so as to override default definitions
	if properties, err = ioutil.ReadFile(schema); err != nil {
		return
	}

	if err = json.Unmarshal(properties, &dataset); err != nil {
		return
	}

	// Base Schema overrides the given schema
	if schemas, err = ioutil.ReadFile(baseSchema); err != nil {
		return
	}

	if err = json.Unmarshal(schemas, &dataset); err != nil {
		return
	}

	if schemaDefinition, err = json.Marshal(dataset); err != nil {
		return
	}

	// fmt.Println(string(schemaDefinition))
	return

}

func main() {

	schemaBytes, err := generateJSONSchema("/Users/meddler/Office/Workspaces/Secoflex/secoflex/modules/watchdog/schema/definition/v1.json", "/Users/meddler/Office/Workspaces/Secoflex/secoflex/modules/watchdog/schema/properties/test.json")
	if err != nil {
		panic(err)
	}

	schemaLoader := gojsonschema.NewSchemaLoader()
	schemaJSONLoader := gojsonschema.NewStringLoader(string(schemaBytes))

	// err = schemaLoader.AddSchemas(schemaJSONLoader)

	schema, err := schemaLoader.Compile(schemaJSONLoader)

	if err != nil {
		panic(err)
	}

	documentLoader := gojsonschema.NewStringLoader(`
	[
	{

		"severity" : "HIGH",
		"uri": "filestore//:/s/dasd d dasd"
	}
	]
	`)

	result, err := schema.Validate(documentLoader)

	for errs := range result.Errors() {
		logger.Println(errs)
	}

	logger.Println(result.Valid())
	logger.Println(err)
	return

}
func _main() {

	schemaLoader := gojsonschema.NewSchemaLoader()
	// schemaLoader := gojsonschema.NewSchemaLoader("file:///Users/meddler/Office/Workspaces/Secoflex/secoflex/modules/watchdog/schema/definition/v1.json")
	propertiesLoader := gojsonschema.NewReferenceLoader("file:///Users/meddler/Office/Workspaces/Secoflex/secoflex/modules/watchdog/schema/propertes/test.json")

	schemaLoader.AddSchemas(propertiesLoader)

	_schemaLoader, err := schemaLoader.Compile(propertiesLoader)

	documentLoader := gojsonschema.NewStringLoader(`
	{
		"severity": "HIGH",

		"test": "/foo/bar",
		"ip": "dasdsa:ddas"
	}
	`)

	result, err := _schemaLoader.Validate(documentLoader)
	if err != nil {
		panic(err.Error())
	}

	if result.Valid() {
		fmt.Printf("The document is valid\n")
	} else {
		fmt.Printf("The document is not valid. see errors :\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
	}
}
