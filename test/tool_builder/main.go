package main

import (
	"github.com/meddler-io/watchdog/producer"
)

func main() {

	// go sendData()
	// go sendData()
	// go sendData()
	sendData()
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
			"base_path": "/workspace",
			"input_dir":  "input_dirs",
			"output_dir":  "output_dirs"

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
	"args": [  "--context=dir:///$input_dir/new-bucket" ,  "--no-push"  ,  "--destination=image" , "--tarPath=$output_dir/test.tar" ,  "--dockerfile=Dockerfile"  ],

	"id": "outputbucket" ,

	"environ": 
	{

		"exec_timeout": "1000" ,  
		"$INPUTDIR": "inpudir",
		"TraceId":"5fde15c7ed17c3374c56990e" ,
		"fprocess":"echo \\$fprocess $fprocess $BASEPATH "  

	} ,
		
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
