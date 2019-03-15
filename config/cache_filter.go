package config

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/astaxie/beego/orm"
	"github.com/sdvdxl/falcon-logdog/models"
)

// Filter .
type Filter struct {
	Filter    map[string][]string
	FilterLen int
	Level     string
}

// Filters .
type Filters []Filter

var filterCache = make(map[string]Filters)
var filterCacheLock = &sync.RWMutex{}

func reloadFilterCache() {
	var err error
	o := orm.NewOrm()
	defer func() {
		if err != nil {
			log.Fatalf("reloadFilterCache:%s", err.Error())
		}
	}()
	var res []orm.Params
	filterCacheLock.Lock()
	defer filterCacheLock.Unlock()
	// load filter
	if _, err = o.QueryTable(models.AlarmRule{}).Limit(-1).Values(&res); err != nil {
		return
	}
	for _, v := range res {
		_, ok1 := v["device_type"]
		_, ok2 := v["alarm_type"]
		if !ok1 || !ok2 {
			continue
		}
		jsonStr := v["json_filter"].(string)
		if jsonStr == "" {
			continue
		}
		var mapResult map[string]interface{}
		err := json.Unmarshal([]byte(jsonStr), &mapResult)
		if err != nil {
			log.Fatalf("JsonToMapDemo err: %s", err.Error())

		}
		filter := make(map[string][]string)
		for feildName, value := range mapResult {
			vs := strings.Split(value.(string), ",")
			if feildName == "device_type" {
				filter["device_type"] = []string{v["device_type"].(string)}
				continue
			}
			filter[feildName] = vs
		}
		f := Filter{
			Filter:    filter,
			FilterLen: len(mapResult),
			Level:     v["level"].(string),
		}
		fs := Filters{}
		if _, ok3 := filterCache[v["device_type"].(string)+"??"+v["alarm_type"].(string)]; ok3 {
			fs = filterCache[v["device_type"].(string)+"??"+v["alarm_type"].(string)]
		}
		fs = append(fs, f)
		filterCache[v["device_type"].(string)+"??"+v["alarm_type"].(string)] = fs
	}
	log.Printf("reloadFilterCache:%v", filterCache)
}

// FetchFilterCache .
func FetchFilterCache() map[string]Filters {
	filterCacheLock.RLock()
	if len(filterCache) == 0 {
		filterCacheLock.RUnlock()
		reloadFilterCache()
		filterCacheLock.RLock()
	}
	defer filterCacheLock.RUnlock()
	rtn := make(map[string]Filters)
	for k, v := range filterCache {
		rtn[k] = v
	}
	return rtn
}

//TODO sort filter
