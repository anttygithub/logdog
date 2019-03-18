package config

import (
	"encoding/json"
	"log"
	"sort"
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
		// if e := recover(); e != nil {
		// 	err = e.(error)
		// }
		if err != nil {
			log.Fatalf("reloadFilterCache:%s", err.Error())
		}
	}()
	filterCacheLock.Lock()
	defer filterCacheLock.Unlock()
	// load filter
	var res []models.AlarmRule
	if _, err = o.QueryTable(&models.AlarmRule{}).Limit(-1).All(&res); err != nil {
		return
	}
	var mapres []map[string]interface{}
	for i := 0; i < len(res); i++ {
		mapres = append(mapres, OrmStructToMap(res[i]))
	}
	filterCache = make(map[string]Filters)
	for _, v := range mapres {
		if _, ok := v["alarm_type"]; !ok {
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
			filter[feildName] = vs
		}
		f := Filter{
			Filter:    filter,
			FilterLen: len(mapResult),
			Level:     v["level"].(string),
		}
		fs := Filters{}
		if _, ok := filterCache[v["alarm_type"].(string)]; ok {
			fs = filterCache[v["alarm_type"].(string)]
		}
		fs = append(fs, f)
		filterCache[v["alarm_type"].(string)] = fs
	}
	for k := range filterCache {
		sort.Sort(filterCache[k])
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

//sort filter
// 按照 Filter.FilterLen 从大到小排序
func (a Filters) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a Filters) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a Filters) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[j].FilterLen < a[i].FilterLen
}
