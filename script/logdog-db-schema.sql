create database logdog
  DEFAULT CHARACTER SET utf8
  DEFAULT COLLATE utf8_general_ci;
USE logdog;
SET NAMES utf8;

DROP TABLE if exists `alarm_rule`;
CREATE TABLE `alarm_rule` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT 'rule name',
  `idc` varchar(255) NOT NULL DEFAULT '' COMMENT 'machine room',
  `use` varchar(255) NOT NULL DEFAULT '' COMMENT 'the device use for ',
  `path` varchar(255) NOT NULL DEFAULT '/log' COMMENT 'path',
  `prefix` varchar(255) DEFAULT NULL COMMENT 'prefix',
  `suffix` varchar(255) NOT NULL DEFAULT '.log' COMMENT 'suffix',
  `tag` varchar(255) NOT NULL DEFAULT 'error-tag' COMMENT 'tag',
  `rule` text NOT NULL COMMENT 'the rule in regular expression for match the keywork ',
  `level` enum('critical','major','warning','minor','info') NOT NULL DEFAULT 'info' COMMENT 'alarm level(''critical'',''major'',''warning'',''minor'',''info'')',
  `status` enum('active','inactive') NOT NULL DEFAULT 'active' COMMENT 'record status',
  `creator` int(10) unsigned DEFAULT '0' COMMENT 'creator',
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updator` int(10) unsigned DEFAULT '0' COMMENT 'updator',
  `updated` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'last update time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;

DROP TABLE if exists `network_device`;
CREATE TABLE `network_device` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT 'name',
  `net_device_assetid` varchar(255) DEFAULT '' COMMENT 'net_device_assetid',
  `idc` varchar(255) DEFAULT NULL COMMENT 'idc',
  `standard_model` varchar(255) DEFAULT '' COMMENT 'standard_model',
  `status` varchar(255) DEFAULT '' COMMENT 'status',
  `sn` varchar(255) DEFAULT '' COMMENT 'sn',
  `device_type` varchar(255) DEFAULT '' COMMENT 'device_type',
  `supplier` varchar(255) DEFAULT '' COMMENT 'supplier',
  `device_model` varchar(255) DEFAULT '' COMMENT 'device_model',
  `device_manager` varchar(255) DEFAULT '' COMMENT 'device_manager',
  `netdev_func` varchar(255) DEFAULT '' COMMENT 'netdev_func',
  `manage_ip` varchar(255) DEFAULT '' COMMENT 'manage_ip',
  `shelf` varchar(255) DEFAULT '' COMMENT 'shelf',
  `network_area` varchar(255) DEFAULT '' COMMENT 'network_area',
  `schema_version` varchar(255) DEFAULT '' COMMENT 'schema_version',
  `tor_name` varchar(255) DEFAULT '' COMMENT 'tor_name',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;
