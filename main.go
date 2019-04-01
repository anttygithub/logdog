package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/fsnotify/fsnotify"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hpcloud/tail"
	"github.com/sdvdxl/falcon-logdog/config"
	"github.com/sdvdxl/falcon-logdog/log"
	"github.com/sdvdxl/falcon-logdog/models"
	cmap "github.com/streamrail/concurrent-map"
)

var (
	workers  chan bool
	keywords cmap.ConcurrentMap
)

func main() {

	workers = make(chan bool, runtime.NumCPU()*2)
	keywords = cmap.New()
	runtime.GOMAXPROCS(runtime.NumCPU())

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(int64(config.Cfg.Timer)))
		for range ticker.C {
			//del by dennis,无关键字匹配不上报
			//fillData()

			postData()
		}
	}()

	go func() {
		setLogFile()

		log.Info("watch file", config.Cfg.WatchFiles)

		for i := 0; i < len(config.Cfg.WatchFiles); i++ {
			readFileAndSetTail(&(config.Cfg.WatchFiles[i]))
			go logFileWatcher(&(config.Cfg.WatchFiles[i]))

		}

	}()

	select {}
}

func logFileWatcher(file *config.WatchFile) {
	// modify by nic: 创建、重命名、删除文件时，更新[]ResultFiles
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	// log.Info("watch file 2222222222222222222", file)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// log.Info("-----------------EVENT------------------%v file %v, basePath:%v", event.Op, event.Name, path.Base(event.Name))
				var resf config.DresultFile
				if event.Op == fsnotify.Create {
					// log.Debug("--------------ResultFiles CREATE 1----------------", file.ResultFiles)
					filepath.Walk(event.Name, func(path string, info os.FileInfo, err error) error {
						resf.FileName = event.Name
						resf.ModTime = info.ModTime()
						reg := regexp.MustCompile(`.log$`)
						if len(reg.FindAllString((resf.FileName), -1)) != 0 {
							for i := 0; i < len(file.ResultFiles); i++ {
								logTail := file.ResultFiles[i].LogTail
								if event.Name == file.ResultFiles[i].FileName {
									if logTail != nil {
										// log.Debug("-------------- 1 STOP tail ---------------", i, file.ResultFiles)
										logTail.Stop()
									}
									file.ResultFiles = append(file.ResultFiles[:i], file.ResultFiles[i+1:]...)
									file.ResultFiles = append(file.ResultFiles, resf)
								} else if i == len(file.ResultFiles)-1 {
									file.ResultFiles = append(file.ResultFiles, resf)
								}
								readFileAndSetTail(file)
							}
							//首次创建日志文件
							if len(file.ResultFiles) == 0 {
								file.ResultFiles = append(file.ResultFiles, resf)
								readFileAndSetTail(file)
							}
						}
						return err
					})
					// log.Debug("--------------ResultFiles CREATE 2----------------", file.ResultFiles, len(file.ResultFiles))
				} else if event.Op == fsnotify.Rename || event.Op == fsnotify.Remove {
					// log.Debug("--------------ResultFiles REMOVE/RENAME 2----------------", file.ResultFiles)
					filepath.Walk(event.Name, func(path string, info os.FileInfo, err error) error {
						for i := 0; i < len(file.ResultFiles); i++ {
							if event.Name == file.ResultFiles[i].FileName {
								logTail := file.ResultFiles[i].LogTail
								if logTail != nil {
									// log.Debug("-------------- 2 STOP tail ---------------", i, file.ResultFiles)
									logTail.Stop()
								}
								file.ResultFiles = append(file.ResultFiles[:i], file.ResultFiles[i+1:]...)
							}
						}
						return err
					})
					// log.Debug("--------------ResultFiles REMOVE/RENAME 2----------------", file.ResultFiles)
				}
			case err := <-watcher.Errors:
				log.Error(err)
			}
		}
	}()

	watchPath := file.Path
	if file.PathIsFile {
		watchPath = filepath.Dir(file.Path)
	}
	err = watcher.Add(watchPath)
	if err != nil {
		log.Fatal(err)

	}
	<-done
}

func readFileAndSetTail(file *config.WatchFile) {
	//add by dennis,处理日志文件数组
	if len(file.ResultFiles) < 1 {
		return
	}
	/*删除
	if file.ResultFile.FileName == "" {
		return
	}
	*/

	//mod by dennis,处理该监控目录下的多个日志文件
	for i := 0; i < len(file.ResultFiles); i++ {
		tailb := file.ResultFiles[i].LogTail
		if tailb == nil {
			_, err := os.Stat(file.ResultFiles[i].FileName)
			if err != nil {
				log.Error(file.ResultFiles[i].FileName, err)
				return
			}

			log.Info("read file", file.ResultFiles[i].FileName)
			//mod by dennis,从文件末尾开始读
			tail, err := tail.TailFile(file.ResultFiles[i].FileName, tail.Config{Follow: true, Location: &tail.SeekInfo{Offset: 0, Whence: 2}})
			if err != nil {
				log.Fatal(err)
			}

			file.ResultFiles[i].LogTail = tail
			filename := file.ResultFiles[i].FileName

			go func() {
				for line := range tail.Lines {
					// log.Debug("log line: ", line.Text)
					//mod by dennis,传入更多参数
					handleKeywords(*file, filename, line.Text)
				}
			}()
		}
	}
}

func setLogFile() {
	c := config.Cfg
	for i, v := range c.WatchFiles {
		if v.PathIsFile {
			//add by dennis,多个日志放入数组
			var resf config.DresultFile
			resf.FileName = v.Path
			c.WatchFiles[i].ResultFiles = append(c.WatchFiles[i].ResultFiles, resf)
			/* 删除
			   c.WatchFiles[i].ResultFile.FileName = v.Path
			*/
			continue
		}

		filepath.Walk(v.Path, func(path string, info os.FileInfo, err error) error {
			cfgPath := v.Path
			if strings.HasSuffix(cfgPath, "/") {
				cfgPath = string([]rune(cfgPath)[:len(cfgPath)-1])
			}
			log.Debug(path)

			//只读取root目录的log
			if filepath.Dir(path) != cfgPath && info.IsDir() {
				log.Debug(path, "not in root path, ignoring , Dir:", path, "cofig path:", cfgPath)
				return err
			}

			log.Debug("path", path, "prefix:", v.Prefix, "suffix:", v.Suffix, "base:", filepath.Base(path), "isFile", !info.IsDir())
			if strings.HasPrefix(filepath.Base(path), v.Prefix) && strings.HasSuffix(path, v.Suffix) && !info.IsDir() {
				//add by dennis,多个日志文件放入数组
				var resf config.DresultFile
				resf.FileName = path
				resf.ModTime = info.ModTime()
				c.WatchFiles[i].ResultFiles = append(c.WatchFiles[i].ResultFiles, resf)
				/* 删除
				if c.WatchFiles[i].ResultFile.FileName == "" || info.ModTime().After(c.WatchFiles[i].ResultFile.ModTime) {
					c.WatchFiles[i].ResultFile.FileName = path
					c.WatchFiles[i].ResultFile.ModTime = info.ModTime()
				} */
				return err
			}

			return err
		})

	}
}

// 查找关键词
func handleKeywords(file config.WatchFile, filename string, line string) {
	// log.Debug(filename)
	// log.Debugf("WatchFile:%+v", file)
	//get network device cache
	ip := getIPFromLog(line)
	if ip == "" {
		return
	}
	allNetdevs := config.FetchNetdevCache()
	if _, ok := allNetdevs[ip]; !ok {
		return
	}
	Netdev := allNetdevs[ip]

	for _, p := range file.Keywords {
		//modify by dennis
		value := ""
		if p.Regex.MatchString(line) { //match the keyword
			log.Debugf("exp:%v match ===> line: %v ", p.Regex.String(), line)
			if !matchNetdevType(Netdev, p.DeviceType) { //match network device type and keyword cache
				continue
			}
			title, level := matchFilter(Netdev, p.AlarmType) //match network device and filter cache
			if title == "" && level == "" {
				continue
			}
			//modify by dennis
			value = line

			var data config.PushData
			//mod by dennis
			UUID := GenerateID("T")
			hashkey := filename + "|" + p.Tag + "=" + p.Exp + "|" + p.DeviceType + "|" + UUID
			log.Debugf("hashkey:%s", hashkey)
			if v, ok := keywords.Get(hashkey); ok {
				d := v.(config.PushData)
				d.Value = append(d.Value, value)
				data = d
			} else {
				stringValue := []string{value}
				data = config.PushData{Metric: config.Cfg.Metric,
					Endpoint:  ip,
					Timestamp: time.Now().Unix(),
					Value:     stringValue,
					//Step:        config.Cfg.Timer,  //modify by nic
					Type:   "network",                                                                                                     //modify by nic
					Tag:    "filename=" + filename + ",prefix=" + file.Prefix + ",suffix=" + file.Suffix + "," + p.Tag + "=" + p.FixedExp, //modify by nic
					Status: "PROBLEM",                                                                                                     //add by nic
					Desc:   title,                                                                                                         //add by nic
					Level:  level,                                                                                                         //add by nic
				}
			}

			keywords.Set(hashkey, data)
		}

	}
}

func postData() {
	// modify by nic: 每次只post一条数据
	c := config.Cfg
	workers <- true

	go func() {
		if len(keywords.Items()) != 0 {
			for k, v := range keywords.Items() {
				keywords.Remove(k)
				bytes, err := json.Marshal(v.(config.PushData))

				if err != nil {
					log.Error("marshal push v.(config.PushData", v.(config.PushData), err)
					return
				}

				log.Debug("pushing data:", string(bytes))
				insertAlarmHistory(v.(config.PushData))
				// resp, err := http.Post(c.Agent, "plain/text", strings.NewReader(string(bytes)))
				resp, err := http.Post(c.Agent, "application/json", strings.NewReader(string(bytes)))
				if err != nil {
					log.Error(" post data ", string(bytes), " to agent ", err)
				} else {
					defer resp.Body.Close()
					bytes, _ = ioutil.ReadAll(resp.Body)
					log.Debug(string(bytes))
				}
			}
		}

		<-workers
	}()

}

func getIPFromLog(line string) (ip string) {
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)\.(\d+)`)
	if len(re.FindAllString(line, -1)) == 0 {
		log.Error("getIPFromLog error:", re.FindAllString(line, -1))
		return ""
	}
	return re.FindAllString(line, -1)[0]

}

func matchNetdevType(Netdev map[string]interface{}, DeviceType string) bool {
	if _, ok := Netdev["device_type"]; !ok {
		return false
	}
	if Netdev["device_type"].(string) != DeviceType {
		return false
	}
	return true
}

func matchFilter(Netdev map[string]interface{}, AlarmType string) (title, alarmLevel string) {
	allFilters := config.FetchFilterCache()
	if _, ok := allFilters[AlarmType]; !ok {
		return
	}
	Filters := allFilters[AlarmType]

	for i := 0; i < len(Filters); i++ { //handle one filter
		filter := Filters[i]
		flag := true
		for k, v := range filter.Filter { //handle one field
			if _, ok := Netdev[k]; !ok {
				flag = false
				log.Debugf("can not match the filter :%s from netdev cache", k)
				break
			}
			if !ArrayIn(v, Netdev[k].(string)) {
				flag = false
				log.Debugf("can not match the filter [%s:%#v] netdev[%s] from cache", k, v, Netdev[k].(string))
				break
			}
		}
		if flag == true {
			title = AlarmType
			alarmLevel = filter.Level
			return
		}
	}
	return
}

// ArrayIn .
func ArrayIn(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}

//insert alarm into the table alarm_history
func insertAlarmHistory(a config.PushData) {
	o := orm.NewOrm()
	m := models.AlarmHistory{
		Metric:    a.Metric,
		Endpoint:  a.Endpoint,
		Timestamp: a.Timestamp,
		Value:     fmt.Sprint(a.Value),
		Type:      a.Type,
		Tag:       a.Tag,
		Status:    a.Status,
		Desc:      a.Desc,
		Level:     a.Level,
	}
	var err error
	o.Begin()
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
		if err != nil {
			log.Fatalf("insert alarm history fail:%s", err.Error())
			o.Rollback()
		} else {
			o.Commit()
		}
	}()
	_, err = o.Insert(&m)
	return
}

// GenerateID generate id from nano time with prefix
func GenerateID(prefix string) (id string) {
	id = strconv.FormatInt(time.Now().UnixNano()/10000, 10)
	if len(id) > 15 {
		id = id[:15]
	}
	if len(id) == 14 {
		id = "0" + id
	}
	id = prefix + id
	if len(id) > 16 {
		id = id[:16]
	}
	return
}
