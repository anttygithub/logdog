package config

import (
	"log"
	"sync"

	"github.com/astaxie/beego/orm"
	"github.com/sdvdxl/falcon-logdog/models"
)

var netdevCache = make(map[string]map[string]interface{})
var netdevCacheLock = &sync.RWMutex{}

func reloadNetdevCache() {
	var err error
	o := orm.NewOrm()
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
		if err != nil {
			log.Fatalf("reloadNetdevCache:%s", err.Error())
		}
	}()
	netdevCacheLock.Lock()
	defer netdevCacheLock.Unlock()
	// load netdev
	var res []models.NetworkDevice
	if _, err = o.QueryTable(&models.NetworkDevice{}).Limit(-1).All(&res); err != nil {
		return
	}
	for _, v := range res {
		netdevCache[v.ManageIp] = OrmStructToMap(v)
	}
	log.Printf("reloadNetdevCache num:%d", len(netdevCache))
}

// FetchNetdevCache .
func FetchNetdevCache() map[string]map[string]interface{} {
	netdevCacheLock.RLock()
	if len(netdevCache) == 0 {
		netdevCacheLock.RUnlock()
		reloadNetdevCache()
		netdevCacheLock.RLock()
	}
	defer netdevCacheLock.RUnlock()
	rtn := make(map[string]map[string]interface{})
	for k, v := range netdevCache {
		rtn[k] = v
	}
	return rtn
}
