package config

import (
	"log"
	"strings"
	"sync"

	"github.com/astaxie/beego/orm"
	"github.com/sdvdxl/falcon-logdog/models"
)

var alarmCache = make(map[string][]keyWord)
var alarmCacheLock = &sync.RWMutex{}

func reloadKeywordCache() {
	var err error
	if !Cfg.AlarmRuleDB.Enabled {
		log.Println("INFO:the config.AlarmRuleDB.Enabled is not true")
		if err = checkConfig(Cfg); err != nil {
			log.Fatal(err)
		}
		return
	}
	o := orm.NewOrm()
	defer func() {
		if err != nil {
			log.Fatalf("reloadKeywordCache:%s", err.Error())
		}
	}()
	var res []models.SyslogKeyword
	alarmCacheLock.Lock()
	defer alarmCacheLock.Unlock()
	if _, err = o.QueryTable(models.SyslogKeyword{}).Filter("status", "active").Limit(-1).All(&res); err != nil {
		return
	}

	alarmCache = make(map[string][]keyWord)
	for _, v := range res {
		key := v.Path + "??" + v.Prefix + "??" + v.Suffix
		kw := keyWord{
			DeviceType: v.DeviceType,
			AlarmType:  v.AlarmType,
			Exp:        v.SyslogKeyword,
			Tag:        v.Tag,
		}
		tmpkwarray := []keyWord{}
		if _, ok := alarmCache[key]; ok {
			tmpkwarray = alarmCache[key]
		}
		tmpkwarray = append(tmpkwarray, kw)
		alarmCache[key] = tmpkwarray

	}
	fetchKeywordCache()
	if err = checkConfig(Cfg); err != nil {
		log.Fatal(err)
	}
	log.Printf("reloadKeywordCache num:%+v", alarmCache)
}

// fetchKeywordCache .
func fetchKeywordCache() {
	// alarmCacheLock.RLock()
	// if len(alarmCache) == 0 {
	// 	alarmCacheLock.RUnlock()
	// 	reloadKeywordCache()
	// 	alarmCacheLock.RLock()
	// }
	// defer alarmCacheLock.RUnlock()
	WFS := []*WatchFile{}
	for k, v := range alarmCache {
		str := strings.Split(k, "??")
		if len(str) != 3 {
			log.Fatalf("alarm rule config error,please check :%s", k)
			return
		}
		tmpWF := WatchFile{
			Path:     str[0],
			Prefix:   str[1],
			Suffix:   str[2],
			Keywords: v,
		}
		WFS = append(WFS, &tmpWF)
	}
	if len(WFS) != 1 || len(Cfg.WatchFiles) != 1 { //暂时只考虑1个watchfiles的config
		log.Fatal("alarm rule config error,please check len(WFS) != len(Cfg.WatchFiles) != 1")
		return
	}
	for i := 0; i < len(Cfg.WatchFiles); i++ {
		Cfg.WatchFiles[i].Path = WFS[i].Path //暂时只考虑keyword和Path更新
		Cfg.WatchFiles[i].Keywords = WFS[i].Keywords
	}
	// log.Printf("load alarm db cache:%#v", Cfg)
	return
}
