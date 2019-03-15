package config

import (
	"log"

	"github.com/astaxie/beego/orm"
)

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
