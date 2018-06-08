package main

import (
	. "IPT/common"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"net/http"

	"strings"
)

var API_HOST string = "http://127.0.0.1:10334"

func HttpPostRequest(url, data string) (string, error) {
	httpClient := &http.Client{}

	request, err := http.NewRequest("POST", API_HOST+url, strings.NewReader(data))
	if nil != err {
		return "", err
	}
	request.Header.Add("Content-Type", "application/json")

	response, err := httpClient.Do(request)

	if nil != err {
		return "", err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if nil != err {
		return "", err
	}

	return string(body), nil
}

func sendTestDeployContract() {
	fileStr := "hello.avm"

	programHash := "ec47640feda118210e5f73d54353f3ab086fff72"

	bytes, err := ioutil.ReadFile(fileStr)
	if err != nil {
		fmt.Println(err.Error())
	}
	codeStr := BytesToHexString(bytes)

	mapParams := make(map[string]string)
	mapParams["Data"] = codeStr
	mapParams["ProgramHash"] = programHash

	jsonString, err := json.Marshal(mapParams)
	if err != nil {
		fmt.Println(err.Error())
	}

	result, err := HttpPostRequest("/api/v1/contract/deploy", string(jsonString))
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(result)
}

func sendTestInvokeContract() {
	codeHash := "3e183cc7144985828f191b8c25f57b443f8dc9de"
	programHash := "ec47640feda118210e5f73d54353f3ab086fff72"

	mapParams := make(map[string]interface{})
	mapParams["Data"] = codeHash
	mapParams["P1"] = "string"
	mapParams["P2"] = ""
	mapParams["P3"] = ""
	mapParams["ProgramHash"] = programHash

	jsonString, err := json.Marshal(mapParams)
	if err != nil {
		fmt.Println(err.Error())
	}

	result, err := HttpPostRequest("/api/v1/contract/invoke", string(jsonString))
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(result)
}

func main() {
	//sendTestDeployContract()
	sendTestInvokeContract()
}
