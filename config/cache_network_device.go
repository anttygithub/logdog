package config

import (
	"log"
	"sync"

	"github.com/astaxie/beego/orm"
	"github.com/sdvdxl/falcon-logdog/models"
)

var netdevCache = make(map[string]orm.Params)
var netdevCacheLock = &sync.RWMutex{}

func reloadNetdevCache() {
	var err error
	o := orm.NewOrm()
	defer func() {
		if err != nil {
			log.Fatalf("reloadNetdevCache:%s", err.Error())
		}
	}()
	var res []orm.Params
	netdevCacheLock.Lock()
	defer netdevCacheLock.Unlock()
	// load netdev
	if _, err = o.QueryTable(models.NetworkDevice{}).Limit(-1).Values(&res); err != nil {
		return
	}
	for _, v := range res {
		netdevCache[v["manage_ip"].(string)] = v
	}
	log.Printf("reloadNetdevCache:%v", netdevCache)
}

// FetchNetdevCache .
func FetchNetdevCache() map[string]orm.Params {
	netdevCacheLock.RLock()
	if len(netdevCache) == 0 {
		netdevCacheLock.RUnlock()
		reloadNetdevCache()
		netdevCacheLock.RLock()
	}
	defer netdevCacheLock.RUnlock()
	rtn := make(map[string]orm.Params)
	for k, v := range netdevCache {
		rtn[k] = v
	}
	return rtn
}
