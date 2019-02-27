package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/fsnotify/fsnotify"
	"github.com/go-errors/errors"
	"github.com/hpcloud/tail"
	"github.com/sdvdxl/falcon-logdog/models"
)

type Config struct {
	Metric      string      //度量名称,比如log.console 或者log
	Timer       int         // 每隔多长时间（秒）上报
	Host        string      //主机名称
	Agent       string      //agent api url
	WatchFiles  []WatchFile `json:"files"`
	LogLevel    string
	AlarmRuleDB DBConfig
}
type DBConfig struct {
	Enabled  bool
	Address  string
	DbName   string
	User     string
	Password string
}

//mod by dennis
type DresultFile struct {
	FileName string
	ModTime  time.Time
	LogTail  *tail.Tail
}

//mod by dennis
type WatchFile struct {
	Path        string //路径
	Prefix      string //log前缀
	Suffix      string //log后缀
	Keywords    []keyWord
	PathIsFile  bool          //path 是否是文件
	ResultFiles []DresultFile `json:"-"` //add by dennis,目录下的文件可能有多个
}

type keyWord struct {
	Exp      string
	Tag      string
	FixedExp string         `json:"-"` //替换
	Regex    *regexp.Regexp `json:"-"`
	Level    string         `json:"level"`
	Idc      string         `json:"idc"`
	Use      string         `json:"use"`
}

//说明：这7个字段都是必须指定
type PushData struct {
	Metric    string   `json:"metric"`    //统计纬度
	Endpoint  string   `json:"endpoint"`  //主机
	Timestamp int64    `json:"timestamp"` //unix时间戳,秒
	Value     []string `json:"value"`     // 代表该metric在当前时间点的值。modify by dennis,直接代表匹配关键字的字符串
	//Step      int     `json:"step"`      //  表示该数据采集项的汇报周期，这对于后续的配置监控策略很重要，必须明确指定。  modify by nic
	//COUNTER：指标在存储和展现的时候，会被计算为speed，即（当前值 - 上次值）/ 时间间隔
	//COUNTER：指标在存储和展现的时候，会被计算为speed，即（当前值 - 上次值）/ 时间间隔

	Type string `json:"type"` //只能是COUNTER或者GAUGE二选一，前者表示该数据采集项为计时器类型，后者表示其为原值 (注意大小写)   modify by nic
	//GAUGE：即用户上传什么样的值，就原封不动的存储
	//COUNTER：指标在存储和展现的时候，会被计算为speed，即（当前值 - 上次值）/ 时间间隔
	Tag    string `json:"tag"`    //一组逗号分割的键值对, 对metric进一步描述和细化, 可以是空字符串. 比如idc=lg，比如service=xbox等，多个tag之间用逗号分割   modify by nic
	Status string `json:"status"` //add by nic
	Desc   string `json:"desc"`   //add by nic
	Level  string `json:"level"`
}

const configFile = "cfg.json"

var (
	Cfg         *Config
	fixExpRegex = regexp.MustCompile(`[\W]+`)
)

func init() {

	var err error
	Cfg, err = ReadConfig(configFile)
	if err != nil {
		log.Fatal("ReadConfig ERROR: ", err)
	}
	if err = checkConfig(Cfg); err != nil {
		log.Fatal(err)
	}

	go func() {
		ConfigFileWatcher()
	}()
	err = InitDatabase()
	if err != nil {
		log.Fatal("InitDatabase ERROR: ", err)
	}
	go func() {
		for {
			time.Sleep(time.Second * 300)
			reloadNetdevCache()
			fetchAlarmCache()
		}
	}()

	fmt.Println("INFO: config:", Cfg)
}

func ReadConfig(configFile string) (*Config, error) {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config *Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	fmt.Println(config.LogLevel)

	// // 检查配置项目
	// if err := checkConfig(config); err != nil {
	// 	return nil, err
	// }

	log.Println("config init success, start to work ...")
	return config, nil
}

// 检查配置项目是否正确
func checkConfig(config *Config) error {
	var err error

	//检查 host
	if config.Host == "" {
		if config.Host, err = os.Hostname(); err != nil {
			return err
		}

		log.Println("host not set will use system's name:", config.Host)

	}
	if config.AlarmRuleDB.Enabled {
		if config.LogLevel == "DEBUG" {
			orm.Debug = true
		} else {
			orm.Debug = false
		}
		fetchAlarmCache()
	} else {
		log.Println("INFO:the config.AlarmRuleDB.Enabled is not true")
	}
	for i, v := range config.WatchFiles {
		//检查路径
		fInfo, err := os.Stat(v.Path)
		if err != nil {
			return err
		}

		if !fInfo.IsDir() {
			config.WatchFiles[i].PathIsFile = true
		}

		//检查后缀,如果没有,则默认为.log
		config.WatchFiles[i].Prefix = strings.TrimSpace(v.Prefix)
		config.WatchFiles[i].Suffix = strings.TrimSpace(v.Suffix)
		if config.WatchFiles[i].Suffix == "" {
			log.Println("file pre ", config.WatchFiles[i].Path, "suffix is no set, will use .log")
			config.WatchFiles[i].Suffix = ".log"
		}

		//agent不检查,可能后启动agent

		//检查keywords
		if len(v.Keywords) == 0 {
			return errors.New("ERROR: keyword list not set")
		}

		for _, keyword := range v.Keywords {
			if keyword.Exp == "" || keyword.Tag == "" {
				return errors.New("ERROR: keyword's exp and tag are requierd")
			}
		}

		// 设置正则表达式
		for j, keyword := range v.Keywords {

			if config.WatchFiles[i].Keywords[j].Regex, err = regexp.Compile(keyword.Exp); err != nil {
				return err
			}

			log.Println("INFO: tag:", keyword.Tag, "regex", config.WatchFiles[i].Keywords[j].Regex.String())

			config.WatchFiles[i].Keywords[j].FixedExp = string(fixExpRegex.ReplaceAll([]byte(keyword.Exp), []byte(".")))
		}
	}

	return nil
}

//ConfigFileWatcher 配置文件监控,可以实现热更新
func ConfigFileWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Name == configFile && (event.Op == fsnotify.Chmod || event.Op == fsnotify.Rename || event.Op == fsnotify.Write || event.Op == fsnotify.Create) {
					log.Println("modified config file", event.Name, "will reaload config")
					if cfg, err := ReadConfig(configFile); err != nil {
						log.Println("ERROR: config has error, will not use old config", err)
					} else if checkConfig(Cfg) != nil {
						log.Println("ERROR: config has error, will not use old config", err)
					} else {
						log.Println("config reload success")
						Cfg = cfg
					}
				}
			case err := <-watcher.Errors:
				log.Fatal(err)
			}
		}
	}()

	err = watcher.Add(".")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

var alarmCache = make(map[string][]keyWord)
var alarmCacheLock = &sync.RWMutex{}

func reloadAlarmCache() {
	var err error
	o := orm.NewOrm()
	defer func() {
		if err != nil {
			log.Fatalf("reloadAlarmCache:%s", err.Error())
		}
	}()
	var res []models.AlarmRule
	alarmCacheLock.Lock()
	defer alarmCacheLock.Unlock()
	if _, err = o.QueryTable(models.AlarmRule{}).Filter("status", "active").Limit(-1).All(&res); err != nil {
		return
	}

	alarmCache = make(map[string][]keyWord)
	for _, v := range res {
		tmpkwarray := []keyWord{}
		if _, ok := alarmCache[v.Path+"??"+v.Prefix+"??"+v.Suffix]; ok {
			tmpkwarray = alarmCache[v.Path+"??"+v.Prefix+"??"+v.Suffix]
		}
		// 裂解idc和用途
		idcs := []string{}
		if v.Idc == "" {
			idcs = []string{"*"}
		} else {
			idcs = strings.Split(v.Idc, ",")
		}
		uses := []string{}
		if v.Use == "" {
			uses = []string{"*"}
		} else {
			uses = strings.Split(v.Use, ",")
		}
		log.Println("idcs:", idcs)
		log.Println("uses:", uses)
		for _, idc := range idcs {
			for _, use := range uses {
				kw := keyWord{
					Exp:   v.Rule,
					Tag:   v.Tag,
					Level: v.Level,
					Idc:   idc,
					Use:   use,
				}
				tmpkwarray = append(tmpkwarray, kw)
			}
		}
		alarmCache[v.Path+"??"+v.Prefix+"??"+v.Suffix] = tmpkwarray
	}
}

// fetchAlarmCache .
func fetchAlarmCache() {
	alarmCacheLock.RLock()
	if len(alarmCache) == 0 {
		alarmCacheLock.RUnlock()
		reloadAlarmCache()
		alarmCacheLock.RLock()
	}
	defer alarmCacheLock.RUnlock()
	WFS := []WatchFile{}
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
		WFS = append(WFS, tmpWF)
	}
	Cfg.WatchFiles = WFS
	log.Printf("load alarm db cache:%#v", Cfg)
	return
}

var netdevCache = make(map[string]models.NetworkDevice)
var netdevCacheLock = &sync.RWMutex{}

func reloadNetdevCache() {
	var err error
	o := orm.NewOrm()
	defer func() {
		if err != nil {
			log.Fatalf("reloadNetdevCache:%s", err.Error())
		}
	}()
	var res []models.NetworkDevice
	netdevCacheLock.Lock()
	defer netdevCacheLock.Unlock()
	// load netdev
	if _, err = o.QueryTable(models.NetworkDevice{}).Limit(-1).All(&res); err != nil {
		return
	}

	netdevCache = make(map[string]models.NetworkDevice)
	for _, v := range res {
		netdevCache[v.ManageIp] = v
	}
	log.Printf("reloadNetdevCache:%v", netdevCache)
}

// FetchNetdevCache .
func FetchNetdevCache() map[string]models.NetworkDevice {
	netdevCacheLock.RLock()
	if len(netdevCache) == 0 {
		netdevCacheLock.RUnlock()
		reloadNetdevCache()
		netdevCacheLock.RLock()
	}
	defer netdevCacheLock.RUnlock()
	rtn := make(map[string]models.NetworkDevice)
	for k, v := range netdevCache {
		rtn[k] = v
	}
	return rtn
}

// InitDatabase .
func InitDatabase() error {
	key := Cfg.AlarmRuleDB.User
	sectet := Cfg.AlarmRuleDB.Password
	IPPort := Cfg.AlarmRuleDB.Address
	database := Cfg.AlarmRuleDB.DbName
	connectString := key + ":" + sectet + "@tcp(" + IPPort + ")/" + database
	err := orm.RegisterDriver("mysql", orm.DRMySQL)
	if err != nil {
		return err
	}
	err = orm.RegisterDataBase("default", "mysql", connectString)
	if err != nil {
		return err
	}
	if Cfg.LogLevel == "DEBUG" {
		orm.Debug = true
	} else {
		orm.Debug = false
	}
	log.Println("init database successed")
	return nil
}
