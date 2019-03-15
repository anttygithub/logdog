package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
)

type SyslogKeywork struct {
	Id             int       `orm:"column(id);auto" description:"id"`
	DeviceType     string    `orm:"column(device_type);size(255)" description:"device_type"`
	AlarmType      string    `orm:"column(alarm_type);size(255)" description:"alarm_type"`
	Path           string    `orm:"column(path);size(255)" description:"path"`
	Prefix         string    `orm:"column(prefix);size(255);null" description:"path"`
	Suffix         string    `orm:"column(suffix);size(255)" description:"path"`
	Tag            string    `orm:"column(tag);size(255)" description:"path"`
	SysylogKeywrod string    `orm:"column(sysylog_keywrod)" description:"json_filter"`
	Status         string    `orm:"column(status)" description:"record status"`
	Creator        uint      `orm:"column(creator);null" description:"creator"`
	Created        time.Time `orm:"column(created);type(timestamp);auto_now_add" description:"created time"`
	Updator        uint      `orm:"column(updator);null" description:"updator"`
	Updated        time.Time `orm:"column(updated);type(timestamp);auto_now" description:"last update time"`
}

func (t *SyslogKeywork) TableName() string {
	return "syslog_keywork"
}

func init() {
	orm.RegisterModel(new(SyslogKeywork))
}

// AddSyslogKeywork insert a new SyslogKeywork into database and returns
// last inserted Id on success.
func AddSyslogKeywork(m *SyslogKeywork) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetSyslogKeyworkById retrieves SyslogKeywork by Id. Returns error if
// Id doesn't exist
func GetSyslogKeyworkById(id int) (v *SyslogKeywork, err error) {
	o := orm.NewOrm()
	v = &SyslogKeywork{Id: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllSyslogKeywork retrieves all SyslogKeywork matches certain condition. Returns empty list if
// no records exist
func GetAllSyslogKeywork(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(SyslogKeywork))
	// query k=v
	for k, v := range query {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		if strings.Contains(k, "isnull") {
			qs = qs.Filter(k, (v == "true" || v == "1"))
		} else {
			qs = qs.Filter(k, v)
		}
	}
	// order by:
	var sortFields []string
	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			// 1) for each sort field, there is an associated order
			for i, v := range sortby {
				orderby := ""
				if order[i] == "desc" {
					orderby = "-" + v
				} else if order[i] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
			qs = qs.OrderBy(sortFields...)
		} else if len(sortby) != len(order) && len(order) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range sortby {
				orderby := ""
				if order[0] == "desc" {
					orderby = "-" + v
				} else if order[0] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return nil, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 {
			return nil, errors.New("Error: unused 'order' fields")
		}
	}

	var l []SyslogKeywork
	qs = qs.OrderBy(sortFields...)
	if _, err = qs.Limit(limit, offset).All(&l, fields...); err == nil {
		if len(fields) == 0 {
			for _, v := range l {
				ml = append(ml, v)
			}
		} else {
			// trim unused fields
			for _, v := range l {
				m := make(map[string]interface{})
				val := reflect.ValueOf(v)
				for _, fname := range fields {
					m[fname] = val.FieldByName(fname).Interface()
				}
				ml = append(ml, m)
			}
		}
		return ml, nil
	}
	return nil, err
}

// UpdateSyslogKeywork updates SyslogKeywork by Id and returns error if
// the record to be updated doesn't exist
func UpdateSyslogKeyworkById(m *SyslogKeywork) (err error) {
	o := orm.NewOrm()
	v := SyslogKeywork{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteSyslogKeywork deletes SyslogKeywork by Id and returns error if
// the record to be deleted doesn't exist
func DeleteSyslogKeywork(id int) (err error) {
	o := orm.NewOrm()
	v := SyslogKeywork{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&SyslogKeywork{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
