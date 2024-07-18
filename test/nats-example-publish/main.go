package main

import (
	"log"

	producernats "github.com/meddler-vault/cortex/producer-nats"
)

const message = `
{ 

"config": {

    "system": {
        "base_path": "/tmp/watchdog-test",
        "input_dir":  "/input",
        "output_dir":  "/output",
        "git_path": "/output/git-path",
        "results_json":  "/outputs/results_json/results.json",
        "git_remote": "https://github.com/studiogangster/sensibull-realtime-options-api-ingestor.git",
        "git_mode": false,

        "mount_volume": true,
        "mount_volume_path": "minio-volume",
        "mount_volume_s3_folder_path": "purger/nested 1/",
        "mount_volume_s3_object_path": "./nested-2/",

        "mount_volume_s3_access_key": "uaaGAF0jnXVHa7KV5eOa",
        "mount_volume_s3_secret_key": "kiwty0-Xigruc-zyfnyj",
        "mount_volume_s3_bucket": "minio-vapt",
        "mount_volume_s3_host": "s3.meddler.io",
        "mount_volume_s3_secure": true

    } ,

    "process": {
        "test": "variable"
    }

},

"substitute_var": true,
"variables": {
    "input_dir" : "$input_dir",
    "output_dir" : "$output_dir",
    "git_path" : "$git_path",
    "mount_volume_s3_folder_path": "$mount_volume_s3_folder_path",
    "mount_volume_s3_object_path": "$mount_volume_s3_object_path",


    
},

"cmd": ["/opt/test.sh" ],
"args": [  ],
"entrypoint": [ "bin/sh"  ],

"id": "outputbucket" ,

"environ": 
{
    "exec_timeout": "1000" ,   
    "TraceId":"5fde15c7ed17c3374c56990e" 
},
    
"dependencies": [

  

]
}

`

// Kaniko image build message
const kaniko_build_image = `
{ 

"config": {

    "system": {
        "base_path": "/tmp/watchdog-test",
        "input_dir":  "/input",
        "output_dir":  "/output",
        "git_path": "/output/git-path",
        "results_json":  "/outputs/results_json/results.json",
        "git_remote": "https://github.com/studiogangster/sensibull-realtime-options-api-ingestor.git",
        "git_mode": false,

        "mount_volume": true,
        "mount_volume_path": "minio-volume",
        "mount_volume_s3_folder_path": "purger/nested 1/",
        "mount_volume_s3_object_path": "./nested-2/",

        "mount_volume_s3_access_key": "uaaGAF0jnXVHa7KV5eOa",
        "mount_volume_s3_secret_key": "kiwty0-Xigruc-zyfnyj",
        "mount_volume_s3_bucket": "minio-vapt",
        "mount_volume_s3_host": "s3.meddler.io",
        "mount_volume_s3_secure": true

    } ,

    "process": {
        "test": "variable"
    }

},

"substitute_var": true,
"variables": {
    "input_dir" : "$input_dir",
    "output_dir" : "$output_dir",
    "git_path" : "$git_path",
    "mount_volume_s3_folder_path": "$mount_volume_s3_folder_path",
    "mount_volume_s3_object_path": "$mount_volume_s3_object_path"


    
},

"cmd": [
    "--context" , "$mount_volume_s3_folder_path" ,"--cache=false",            "--cache-dir=$input_dir/cache",            "--cleanup",            "--insecure=false",
    "--no-push"

],
"args": [  ],
"entrypoint": [ 
    "/kaniko/executor" ],

"id": "outputbucket" ,

"environ": 
{
    "exec_timeout": "1000" ,   
    "TraceId":"5fde15c7ed17c3374c56990e" 
},
    
"dependencies": [

  

]
}

`
const kaniko_build_image_topic = "tool_build"

// Kaniko image build message
const sample_message_nomad_tool = `
{ 

"config": {

    "system": {
        "base_path": "/tmp/watchdog-test",
        "input_dir":  "/input",
        "output_dir":  "/output",
        "git_path": "/output/git-path",
        "results_json":  "/outputs/results_json/results.json",
        "git_remote": "https://github.com/studiogangster/sensibull-realtime-options-api-ingestor.git",
        "git_mode": false,

        "mount_volume": true,
        "mount_volume_path": "minio-volume",
        "mount_volume_s3_folder_path": "purger/nested 1/",
        "mount_volume_s3_object_path": "./nested-2/",

        "mount_volume_s3_access_key": "uaaGAF0jnXVHa7KV5eOa",
        "mount_volume_s3_secret_key": "kiwty0-Xigruc-zyfnyj",
        "mount_volume_s3_bucket": "minio-vapt",
        "mount_volume_s3_host": "s3.meddler.io",
        "mount_volume_s3_secure": true

    } ,

    "process": {
        "test": "variable"
    }

},

"substitute_var": true,
"variables": {
    "input_dir" : "$input_dir",
    "output_dir" : "$output_dir",
    "git_path" : "$git_path",
    "mount_volume_s3_folder_path": "$mount_volume_s3_folder_path",
    "mount_volume_s3_object_path": "$mount_volume_s3_object_path"


    
},

"cmd": [
    "--context" , "$mount_volume_s3_folder_path" ,"--cache=false",            "--cache-dir=$input_dir/cache",            "--cleanup",            "--insecure=false",
    "--no-push"

],
"args": [  ],
"entrypoint": [ 
    "/kaniko/executor" ],

"id": "outputbucket" ,

"environ": 
{
    "exec_timeout": "1000" ,   
    "TraceId":"5fde15c7ed17c3374c56990e" 
},
    
"dependencies": [

  

]
}

`
const sample_message_nomad_topic = "6693fe232b4fdd0a70e41e32:6694310e67412eddafb7ed60"

func main() {

	er := producernats.Produce("whitehat", "4Jy6P)$Ep@c^SenL", "rmq.meddler.io:443", sample_message_nomad_topic, sample_message_nomad_tool)
	log.Println("Error", er)
}
