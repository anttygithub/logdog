package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
)

type AlarmHistory struct {
	Id        int       `orm:"column(id);auto" description:"id"`
	Metric    string    `orm:"column(metric);size(255)" description:"device_type"`
	Endpoint  string    `orm:"column(endpoint);size(255)" description:"alarm_type"`
	Timestamp int64     `orm:"column(timestamp)" description:"path"`
	Value     string    `orm:"column(value)" description:"path"`
	Type      string    `orm:"column(type);size(255)" description:"path"`
	Tag       string    `orm:"column(tag);size(255)" description:"path"`
	Desc      string    `orm:"column(desc);size(255)" description:"json_filter"`
	Level     string    `orm:"column(level);size(255)" description:"json_filter"`
	Status    string    `orm:"column(status);size(255)" description:"record status"`
	Creator   uint      `orm:"column(creator);null" description:"updator"`
	Created   time.Time `orm:"column(created);type(timestamp);auto_now_add" description:"created time"`
	Updator   uint      `orm:"column(updator);null" description:"updator"`
	Updated   time.Time `orm:"column(updated);type(timestamp);auto_now" description:"last update time"`
	Remarks   string    `orm:"column(remarks);size(255);null" description:"remarks"`
}

func (t *AlarmHistory) TableName() string {
	return "alarm_history"
}

func init() {
	orm.RegisterModel(new(AlarmHistory))
}

// AddAlarmHistory insert a new AlarmHistory into database and returns
// last inserted Id on success.
func AddAlarmHistory(m *AlarmHistory) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetAlarmHistoryById retrieves AlarmHistory by Id. Returns error if
// Id doesn't exist
func GetAlarmHistoryById(id int) (v *AlarmHistory, err error) {
	o := orm.NewOrm()
	v = &AlarmHistory{Id: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllAlarmHistory retrieves all AlarmHistory matches certain condition. Returns empty list if
// no records exist
func GetAllAlarmHistory(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(AlarmHistory))
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

	var l []AlarmHistory
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

// UpdateAlarmHistory updates AlarmHistory by Id and returns error if
// the record to be updated doesn't exist
func UpdateAlarmHistoryById(m *AlarmHistory) (err error) {
	o := orm.NewOrm()
	v := AlarmHistory{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteAlarmHistory deletes AlarmHistory by Id and returns error if
// the record to be deleted doesn't exist
func DeleteAlarmHistory(id int) (err error) {
	o := orm.NewOrm()
	v := AlarmHistory{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&AlarmHistory{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
