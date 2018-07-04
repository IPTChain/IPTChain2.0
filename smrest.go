package main

import (
	. "IPT/common"
	"encoding/json"
	"flag"
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
	fileStr := "ico.avm"

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

func sendTestInvokeContract(p1 string, p2, p3 interface{}) {
	codeHash := "2a007930cfd2e72413bbd93e4e183dc23075e1c8"
	programHash := "ec47640feda118210e5f73d54353f3ab086fff72"

	mapParams := make(map[string]interface{})
	mapParams["Data"] = codeHash
	mapParams["P1"] = p1
	mapParams["P2"] = p2
	mapParams["P3"] = p3
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
	action := flag.String("contract", "deploy", "config file name")
	param1 := flag.String("p1", "deploy", "config file name")
	param2 := flag.String("p2", "0", "config file name")
	param3 := flag.String("p3", "0", "config file name")
	flag.Parse()
	if *action == "deploy" {
		sendTestDeployContract()
	} else {
		sendTestInvokeContract(*param1, *param2, *param3)
	}
}
