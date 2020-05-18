package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/martini-contrib/render"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"gopkg.in/mgo.v2/bson"
	"moul.io/http2curl"
)

func GetSystemStatus(req *http.Request, r render.Render) {
	// get system status of benchmark machine
	var result = map[string]string{}
	var l, _ = load.Avg()
	result["load:1"] = fmt.Sprintf("%v", l.Load1)
	result["load:5"] = fmt.Sprintf("%v", l.Load5)
	result["load:15"] = fmt.Sprintf("%v", l.Load15)
	var m, _ = mem.VirtualMemory()
	result["mem:total"] = fmt.Sprintf("total:%vM", m.Total>>20)
	result["mem:free"] = fmt.Sprintf("free:%vM", m.Free>>20)
	result["mem:buffers"] = fmt.Sprintf("buffers:%vM", m.Buffers>>20)
	result["mem:cached"] = fmt.Sprintf("cached:%vM", m.Cached>>20)
	result["mem:wired"] = fmt.Sprintf("wired:%vM", m.Wired>>20)
	result["mem:used"] = fmt.Sprintf("used:%.2f%%", m.UsedPercent)
	r.JSON(200, result)
}

func GetVegetaJobState(req *http.Request, r render.Render) {
	var jobId = req.FormValue("job_id")
	var job VegetaJob
	err := G_MongoDB.C("vegeta_jobs").FindId(bson.ObjectIdHex(jobId)).One(&job)
	var result = map[string]interface{}{}
	if err != nil {
		result["is_running"] = false
		result["current_rate"] = 0
	} else {
		result["is_running"] = job.IsRunning()
		result["current_rate"] = job.CurrentRate
	}
	r.JSON(200, result)
}

func GetBoomJobState(req *http.Request, r render.Render) {
	var jobId = req.FormValue("job_id")
	var job BoomJob
	err := G_MongoDB.C("boom_jobs").FindId(bson.ObjectIdHex(jobId)).One(&job)
	var result = map[string]interface{}{}
	if err != nil {
		result["is_running"] = false
		result["current_concurrency"] = 0
	} else {
		result["is_running"] = job.IsRunning()
		result["current_concurrency"] = job.CurrentConcurrency
	}
	r.JSON(200, result)
}

func TestParam(req *http.Request, r render.Render) {
	var host = req.FormValue("host")
	var envs = req.FormValue("envs")
	var url = req.FormValue("url")
	var header = req.FormValue("header")
	var params = req.FormValue("param")
	var data = req.FormValue("data")
	var method = req.FormValue("method")
	var jsonified = req.FormValue("jsonified") == "true"
	var headerMap map[string]interface{}
	var paramMap map[string]interface{}
	var dataMap map[string]interface{}
	var postBody []byte
	var envMap map[string]interface{}
	json.Unmarshal([]byte(header), &headerMap)
	json.Unmarshal([]byte(params), &paramMap)
	json.Unmarshal([]byte(envs), &envMap)
	ReplaceMapByEnvs(envMap, 1, headerMap, paramMap, dataMap)
	ReplaceStringByEnvs(envMap, 1, &data)
	if jsonified {
		postBody = []byte(data)
	} else {
		json.Unmarshal([]byte(data), &dataMap)
		postBody = BodyBytes(dataMap)
	}

	rq, _ := http.NewRequest(method, Urlcat(host, url, paramMap), bytes.NewReader(postBody))
	for k, vs := range headerMap {
		rq.Header.Add(k, vs.(string))
	}
	if host := req.Header.Get("Host"); host != "" {
		rq.Host = host
	}
	client := &http.Client{}
	var result = map[string]interface{}{}
	resp, err := client.Do(rq)
	if err == nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			result["err"] = err.Error()
		} else {
			command, err := http2curl.GetCurlCommand(rq)
			if err != nil {
				fmt.Println(err)
			}
			postBodyMap := map[string]interface{}{}
			json.Unmarshal(postBody, &postBodyMap)
			result["_debug_request"] = map[string]interface{}{
				"method":        method,
				"header":        headerMap,
				"url":           Urlcat(host, url, paramMap),
				"body":          postBodyMap,
				"_curl_command": command.String(),
			}

			// fmt.Sprintf("Method: [%s], Header: [%#v], URL: [%s], Body: [%s]",
			// 	method,
			// 	headerMap,
			// 	Urlcat(host, url, paramMap),
			// 	postBody,
			// )
			bodyMap := map[string]interface{}{}
			json.Unmarshal(body, &bodyMap)
			result["result"] = bodyMap
		}
	} else {
		result["err"] = err
	}
	r.JSON(200, result)
}
