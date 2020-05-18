package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/martini-contrib/render"
)

var Pipelines = template.FuncMap{
	"strftime": Strftime,
	"json":     Json,
}

func Strftime(ts int64) string {
	if ts == 0 {
		return ""
	}
	t := time.Unix(ts, 0)
	return t.Format("2006-01-02 15:04:05")
}

func Json(value interface{}) string {
	if value == nil {
		return "{}"
	}
	result, _ := json.Marshal(value)
	return string(result)
}

func BodyBytes(data map[string]interface{}) []byte {
	var buffer bytes.Buffer
	i := 0
	for k, v := range data {
		var item = fmt.Sprintf("%s=%v", k, v)
		buffer.WriteString(item)
		if i < len(data)-1 {
			buffer.WriteString("&")
		}
		i++
	}
	return buffer.Bytes()
}

// hostWithSchema should starts with https:// or http://
func Urlcat(hostWithSchema string, urls string, params map[string]interface{}) string {
	var u, _ = url.Parse(fmt.Sprintf("%s%s", hostWithSchema, urls))
	var values, _ = url.ParseQuery(u.RawQuery)
	for k, v := range params {
		values.Add(k, fmt.Sprintf("%v", v))
	}
	u.RawQuery = values.Encode()
	return u.String()
}

func GenMethodSelectors(method string) []MethodSelector {
	methods := make([]MethodSelector, 5)
	methods[0] = MethodSelector{"GET", false}
	methods[1] = MethodSelector{"POST", false}
	methods[2] = MethodSelector{"PUT", false}
	methods[3] = MethodSelector{"DELETE", false}
	methods[4] = MethodSelector{"HEAD", false}
	if method == "GET" {
		methods[0].Selected = true
	} else if method == "POST" {
		methods[1].Selected = true
	} else if method == "PUT" {
		methods[2].Selected = true
	} else if method == "DELETE" {
		methods[3].Selected = true
	} else if method == "HEADER" {
		methods[4].Selected = true
	} else {
		methods[0].Selected = true
	}
	return methods
}

func GenTeamSelectors(team string) []TeamSelector {
	var teams = make([]TeamSelector, len(G_AlexTeams)+1)
	teams[0] = TeamSelector{"", false}
	for i := 1; i <= len(G_AlexTeams); i++ {
		teams[i] = TeamSelector{G_AlexTeams[i-1], false}
	}
	for i := 0; i <= len(G_AlexTeams); i++ {
		if teams[i].Team == team {
			teams[i].Selected = true
		}
	}
	return teams
}

func MaxInt(nums ...int) int {
	max := nums[0]
	for _, num := range nums {
		if num > max {
			max = num
		}
	}
	return max
}

func RenderTemplate(r render.Render, tmpl string, context map[string]interface{}) {
	context["ShowLayout"] = G_ShowLayout
	r.HTML(200, tmpl, context)
}

type ConcurrentSet struct {
	// thread safe string set
	d     map[string]bool
	mutex sync.Mutex
}

func NewConcurrentSet() *ConcurrentSet {
	return &ConcurrentSet{map[string]bool{}, sync.Mutex{}}
}

func (this *ConcurrentSet) Empty() bool {
	return len(this.d) == 0
}

func (this *ConcurrentSet) Size() int {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return len(this.d)
}

func (this *ConcurrentSet) Put(key string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.d[key] = true
}

func (this *ConcurrentSet) Exists(key string) bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	_, ok := this.d[key]
	return ok
}

func (this *ConcurrentSet) Delete(key string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	delete(this.d, key)
}

// func ReplaceMagicStrings(ori string) string {
// 	fmt.Printf("Original data := %s\n", ori)
// 	ori = strings.Replace(ori, "!RANDOM", randomString(5), -1)
// 	ori = strings.Replace(ori, "!B64RANDOM", base64.StdEncoding.EncodeToString([]byte(randomString(5))), -1)
// 	ori = strings.Replace(ori, "!UUID", genUUID(), -1)
// 	fmt.Printf("New data := %s\n", ori)

// 	return ori
// }

// func ReplaceEnvs(job *VegetaJob) error {
// 	var envMap map[string]interface{}
// 	err := json.Unmarshal([]byte(job.Envs), &envMap)
// 	if err != nil {
// 		fmt.Printf("invalid envs: %s\n", err)
// 		return err
// 	}
// 	for k, v := range envMap {
// 		for _, seed := range job.Seeds {
// 			headerValue := seed.Header
// 		}
// 	}

// 	return nil
// }

func randomString(length int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func genUUID() string {
	uuid := uuid.Must(uuid.NewV4())

	return uuid.String()
}

// ReplaceByEnvs replace the value of source which contains envs key
// source should be a pointer
// for example
// envs = {"host":"a.b.com","port":8080}
// source = {"myHost":"$host","myPort":"$port","myFoo":"some comments"}
// newVal would be {"myHost": "a.b.com", "myPort": 8080, "myFoo": "some comments"}
func ReplaceMapByEnvs(envs map[string]interface{}, autonum int, sources ...map[string]interface{}) {
	for _, source := range sources {
		for k, v := range source {
			if strV, ok := v.(string); ok {
				for envK, envV := range envs {
					if strings.Contains(strV, fmt.Sprintf("$%s", envK)) {
						tmp := strings.Replace(strV, fmt.Sprintf("$%s", envK), envV.(string), -1)
						source[k] = tmp
					}
				}
				// magic strings
				source[k] = strings.Replace(source[k].(string), "!RANDOM", randomString(5), -1)
				source[k] = strings.Replace(source[k].(string), "!B64RANDOM", base64.StdEncoding.EncodeToString([]byte(randomString(5))), -1)
				source[k] = strings.Replace(source[k].(string), "!UUID", genUUID(), -1)
				source[k] = strings.Replace(source[k].(string), "!AUTONUM", fmt.Sprint(autonum), -1)
				if val, err := strconv.Atoi(source[k].(string)); err == nil {
					source[k] = val
				}
			}
		}
	}
}

// ReplaceStringByEnvs replace the string by envs, write the result back to source
func ReplaceStringByEnvs(envs map[string]interface{}, autonum int, source *string) error {
	// replace by envs
	for envK, envV := range envs {
		if strings.Contains(*source, fmt.Sprintf("$%s", envK)) {
			*source = strings.Replace(*source, fmt.Sprintf("$%s", envK), envV.(string), -1)
		}
	}

	// magic functions
	*source = strings.Replace(*source, "!RANDOM", randomString(5), -1)
	*source = strings.Replace(*source, "!B64RANDOM", base64.StdEncoding.EncodeToString([]byte(randomString(5))), -1)
	*source = strings.Replace(*source, "!UUID", genUUID(), -1)
	*source = strings.Replace(*source, "!AUTONUM", fmt.Sprint(autonum), -1)
	return nil
}
