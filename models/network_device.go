package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/astaxie/beego/orm"
)

type NetworkDevice struct {
	Id               int    `orm:"column(id);auto" description:"id"`
	Name             string `orm:"column(name);size(255)" description:"name"`
	NetDeviceAssetid string `orm:"column(net_device_assetid);size(255);null" description:"net_device_assetid"`
	Idc              string `orm:"column(idc);size(255);null" description:"idc"`
	StandardModel    string `orm:"column(standard_model);size(255);null" description:"standard_model"`
	Status           string `orm:"column(status);size(255);null" description:"status"`
	Sn               string `orm:"column(sn);size(255);null" description:"sn"`
	DeviceType       string `orm:"column(device_type);size(255);null" description:"device_type"`
	Supplier         string `orm:"column(supplier);size(255);null" description:"supplier"`
	DeviceModel      string `orm:"column(device_model);size(255);null" description:"device_model"`
	DeviceManager    string `orm:"column(device_manager);size(255);null" description:"device_manager"`
	NetdevFunc       string `orm:"column(netdev_func);size(255);null" description:"netdev_func"`
	ManageIp         string `orm:"column(manage_ip);size(255);null" description:"manage_ip"`
	Shelf            string `orm:"column(shelf);size(255);null" description:"shelf"`
	NetworkArea      string `orm:"column(network_area);size(255);null" description:"network_area"`
	SchemaVersion    string `orm:"column(schema_version);size(255);null" description:"schema_version"`
	TorName          string `orm:"column(tor_name);size(255);null" description:"tor_name"`
}

func (t *NetworkDevice) TableName() string {
	return "network_device"
}

func init() {
	orm.RegisterModel(new(NetworkDevice))
}

// AddNetworkDevice insert a new NetworkDevice into database and returns
// last inserted Id on success.
func AddNetworkDevice(m *NetworkDevice) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetNetworkDeviceById retrieves NetworkDevice by Id. Returns error if
// Id doesn't exist
func GetNetworkDeviceById(id int) (v *NetworkDevice, err error) {
	o := orm.NewOrm()
	v = &NetworkDevice{Id: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllNetworkDevice retrieves all NetworkDevice matches certain condition. Returns empty list if
// no records exist
func GetAllNetworkDevice(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(NetworkDevice))
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

	var l []NetworkDevice
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

// UpdateNetworkDevice updates NetworkDevice by Id and returns error if
// the record to be updated doesn't exist
func UpdateNetworkDeviceById(m *NetworkDevice) (err error) {
	o := orm.NewOrm()
	v := NetworkDevice{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteNetworkDevice deletes NetworkDevice by Id and returns error if
// the record to be deleted doesn't exist
func DeleteNetworkDevice(id int) (err error) {
	o := orm.NewOrm()
	v := NetworkDevice{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&NetworkDevice{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
