package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"consumernats"
)

func main() {

	// go sendData()
	consumernats.Start()
}

func sendData() {

	url := "http://localhost:8000/api/v2/toolbuilder/test"
	method := "POST"

	payload := strings.NewReader(`
  { 

	"config": {

		"system": {
			"base_path": "/tmp/test"

		}

	},

      "id": "outputbucket" ,
      "environ": 
        {

		  "exec_timeout": "10" ,  
		  "INPUTDIR": "inpudir",
          "TraceId":"5fde15c7ed17c3374c56990e" ,
          "fprocess":"echo \\$fprocess $fprocess $BASEPATH "  

		} ,
		
		"dependencies": [

			{
				"id": "inputbasket_1",
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

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("sec-ch-ua", "\"Google Chrome\";v=\"87\", \" Not;A Brand\";v=\"99\", \"Chromium\";v=\"87\"")
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Origin", "http://localhost:4200")
	req.Header.Add("Sec-Fetch-Site", "same-site")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Referer", "http://localhost:4200/")
	req.Header.Add("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
func sendData2() {

	url := "http://localhost:8000/api/v2/toolbuilder/test"
	method := "POST"

	payload := strings.NewReader(`
  { 

	"config": {

		"system": {
			"bafse_path": "/tmp/test"

		}

	},

      "id": "outputbucket" ,
      "environ": 
        {


		  "exec_timeout": "10" ,  
		  "INPUTDIR": "inpudir",
          "TraceId":"5fde15c7ed17c3374c56990e" ,
          "fprocess":"echo \\$fprocess $fprocess $BASEPATH "  

		} ,
		
		"dependencies": [

			{
				"id": "inputbasket_1",
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

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("sec-ch-ua", "\"Google Chrome\";v=\"87\", \" Not;A Brand\";v=\"99\", \"Chromium\";v=\"87\"")
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Origin", "http://localhost:4200")
	req.Header.Add("Sec-Fetch-Site", "same-site")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Referer", "http://localhost:4200/")
	req.Header.Add("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
func sendData3() {

	url := "http://localhost:8000/api/v2/toolbuilder/test"
	method := "POST"

	payload := strings.NewReader(`
  { 

	"config": {

		"system": {
			"base_path": "/tmp/test/3"

		}

	},

      "id": "outputbucket" ,
      "environ": 
        {

		  "exec_timeout": "10" ,  
		  "INPUTDIR": "inpudir",
          "TraceId":"5fde15c7ed17c3374c56990e" ,
          "fprocess":"echo \\$fprocess $fprocess $BASEPATH "  

		} ,
		
		"dependencies": [

			{
				"id": "inputbasket_1",
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

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("sec-ch-ua", "\"Google Chrome\";v=\"87\", \" Not;A Brand\";v=\"99\", \"Chromium\";v=\"87\"")
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Origin", "http://localhost:4200")
	req.Header.Add("Sec-Fetch-Site", "same-site")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Referer", "http://localhost:4200/")
	req.Header.Add("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
