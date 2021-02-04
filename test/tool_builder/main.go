package main

import (
	"github.com/meddler-io/watchdog/producer"
)

func main() {

	// go sendData()
	// go sendData()
	// go sendData()
	gitCloner()
	// sendData()
	// go sendData2()
	// go sendData3()

}

func produce() {
	sendData()
}

func sendData() {

	producer.Produce(`
  { 

	"config": {

		"system": {
			"base_path": "/kaniko/fs",
			"input_dir":  "/inputs",
			"output_dir":  "/outputs",
			"results_json":  "/outputs/results_json/results.json"

		} ,

		"process": {
			"test": "variable"
		}

	},

	"substitute_var": true,
	"variables": {
		"input_dir" : "$input_dir",
		"output_dir" : "$output_dir"
	},

	"cmd": "/kaniko/executor",
	"args": [  "--context=git://github.com/MobSF/Mobile-Security-Framework-MobSF.git", "--destination=registry:5000/rouna311/mobsh:custom"  ,  "--dockerfile=Dockerfile" ,  "--insecure", "--insecure-pull", "--cleanup" , "--cache-dir=/tmp" , "--cache=true"	],

	"id": "outputbucket" ,

	"environ": 
	{
		"exec_timeout": "1000" ,   
		"TraceId":"5fde15c7ed17c3374c56990e" 
	},
		
	"dependencies": [

		{
			"id": "new-bucket",
			"alias": "Alias dependen 1",
			"type": "Type"
		},
		{
			"id": "inputbasket_2",
			"alias": "Alias dependen 1",
			"type": "Type"
		},
		{
			"id": "buckettest_3",
			"alias": "Alias dependen 1",
			"type": "Type"
		}

	]
  }`)

}

func gitCloner() {

	producer.Produce(`
  { 

	"config": {

		"system": {
			"base_path": "/kaniko/fs",
			"input_dir":  "/input",
			"output_dir":  "/output"

		} ,

		"process": {
			"test": "variable"
		}

	},

	"substitute_var": true,
	"variables": {
		"input_dir" : "$input_dir",
		"output_dir" : "$output_dir"
	},

	"cmd": "/root/Mobile-Security-Framework-MobSF/scripts/entrypoint.sh",
	"id": "outputbucket" ,

	"environ": 
	{
		"exec_timeout": "19" ,   
		"TraceId":"5fde15c7ed17c3374c56990e" 
	},
		
	"dependencies": [

		{
			"id": "new-bucket",
			"alias": "Alias dependen 1",
			"type": "Type"
		},
		{
			"id": "inputbasket_2",
			"alias": "Alias dependen 1",
			"type": "Type"
		},
		{
			"id": "buckettest_3",
			"alias": "Alias dependen 1",
			"type": "Type"
		}

	]
  }`)

}
