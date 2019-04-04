package config

import (
	"log"
	"reflect"
	"strings"

	"github.com/astaxie/beego/orm"
)

// InitDatabase .
func InitDatabase() error {
	key := Cfg.AlarmRuleDB.User
	sectet := Cfg.AlarmRuleDB.Password
	IPPort := Cfg.AlarmRuleDB.Address
	database := Cfg.AlarmRuleDB.DbName
	connectString := key + ":" + sectet + "@tcp(" + IPPort + ")/" + database + "?charset=utf8&loc=Local"
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

// OrmStructToMap .
func OrmStructToMap(obj interface{}) map[string]interface{} {
	obj1 := reflect.TypeOf(obj)
	obj2 := reflect.ValueOf(obj)
	var data = make(map[string]interface{})
	for i := 0; i < obj1.NumField(); i++ {
		tag0 := obj1.Field(i).Tag.Get("orm")
		tag1 := strings.Split(tag0, ")")
		if len(tag1) == 1 || len(tag1[0]) <= 7 { //7 is the len of "column("
			continue
		}
		tag2 := tag1[0][7:]
		data[tag2] = obj2.Field(i).Interface()
	}
	return data
}
