package models

import "github.com/astaxie/beego/orm"

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
	RorName          string `orm:"column(ror_name);size(255);null" description:"ror_name"`
}

func (t *NetworkDevice) TableName() string {
	return "network_device"
}

func init() {
	orm.RegisterModel(new(NetworkDevice))
}
