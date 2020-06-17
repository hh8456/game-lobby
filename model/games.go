package model

import (
	"time"
)

/******sql******
CREATE TABLE `agent` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uid` int(10) DEFAULT NULL,
  `invite_code` int(10) DEFAULT NULL COMMENT '邀请码',
  `is_agent` int(10) DEFAULT NULL COMMENT '是否代理',
  `is_senior_agent` int(10) DEFAULT NULL COMMENT '是否总代',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_uid` (`uid`) USING BTREE,
  KEY `idx_is_agent` (`is_agent`) USING BTREE,
  KEY `idx_is_senior_agent` (`is_senior_agent`) USING BTREE,
  KEY `idx_invite_code` (`invite_code`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
******sql******/
// Agent [...]
type Agent struct {
	ID            int64 `gorm:"primary_key;column:id;type:bigint(20);not null" json:"-"`
	UId           int   `gorm:"unique;column:uid;type:int(10)" json:"uid"`
	InviteCode    int   `gorm:"index;column:invite_code;type:int(10)" json:"invite_code"`         // 邀请码
	IsAgent       int   `gorm:"index;column:is_agent;type:int(10)" json:"is_agent"`               // 是否代理
	IsSeniorAgent int   `gorm:"index;column:is_senior_agent;type:int(10)" json:"is_senior_agent"` // 是否总代
}

/******sql******
CREATE TABLE `invite_code_in_use` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `invite_code` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `invite_code` (`invite_code`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4
******sql******/
// InviteCodeInUse [...]
type InviteCodeInUse struct {
	ID         int64 `gorm:"primary_key;column:id;type:bigint(20);not null" json:"-"`
	InviteCode int64 `gorm:"unique;column:invite_code;type:bigint(20)" json:"invite_code"`
}

/******sql******
CREATE TABLE `player_base_info` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `wxid` varchar(512) DEFAULT NULL,
  `wxid_crc32` bigint(20) unsigned DEFAULT NULL,
  `uid` int(10) DEFAULT NULL,
  `head_pic` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_estonian_ci DEFAULT NULL,
  `invite_code` int(10) DEFAULT NULL,
  `diamond` int(10) DEFAULT NULL,
  `gold` int(10) DEFAULT NULL,
  `sex` tinyint(2) DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL,
  `reg_date` datetime DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `idx_wxid` (`wxid`) USING BTREE,
  KEY `idx_uid` (`uid`) USING BTREE,
  KEY `idx_wx_crc32_id` (`wxid_crc32`,`wxid`) USING BTREE,
  KEY `idx_reg_date` (`reg_date`) USING BTREE,
  KEY `idx_name` (`name`) USING BTREE,
  KEY `idx_invite_code` (`invite_code`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=584 DEFAULT CHARSET=utf8mb4
******sql******/
// PlayerBaseInfo [...]
type PlayerBaseInfo struct {
	ID         int64     `gorm:"primary_key;column:id;type:bigint(20);not null" json:"-"`
	Wxid       string    `gorm:"unique;column:wxid;type:varchar(512)" json:"wxid"`
	WxidCrc32  int64     `gorm:"index:idx_wx_crc32_id;column:wxid_crc32;type:bigint(20) unsigned" json:"wxid_crc32"`
	UId        int       `gorm:"index;column:uid;type:int(10)" json:"uid"`
	HeadPic    string    `gorm:"column:head_pic;type:varchar(512)" json:"head_pic"`
	InviteCode int       `gorm:"index;column:invite_code;type:int(10)" json:"invite_code"`
	Diamond    int       `gorm:"column:diamond;type:int(10)" json:"diamond"`
	Gold       int       `gorm:"column:gold;type:int(10)" json:"gold"`
	Sex        int8      `gorm:"column:sex;type:tinyint(2)" json:"sex"`
	Name       string    `gorm:"index;column:name;type:varchar(64)" json:"name"`
	RegDate    time.Time `gorm:"index;column:reg_date;type:datetime" json:"reg_date"`
}

/******sql******
CREATE TABLE `subordinate` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uid` int(10) DEFAULT NULL,
  `subordinate_uid` int(10) DEFAULT NULL COMMENT 'uid 的直接下级',
  `establish_contact_date` datetime DEFAULT NULL COMMENT 'subordinate 成为 uid 下级的时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_subordinate` (`subordinate_uid`) USING BTREE,
  KEY `idx_uid_sub_date` (`uid`,`subordinate_uid`,`establish_contact_date`) USING BTREE,
  KEY `idx_date` (`establish_contact_date`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
******sql******/
// Subordinate [...]
type Subordinate struct {
	ID                   int64     `gorm:"primary_key;column:id;type:bigint(20);not null" json:"-"`
	UId                  int       `gorm:"index:idx_uid_sub_date;column:uid;type:int(10)" json:"uid"`
	SubordinateUId       int       `gorm:"unique;column:subordinate_uid;type:int(10)" json:"subordinate_uid"`                                      // uid 的直接下级
	EstablishContactDate time.Time `gorm:"index:idx_uid_sub_date;index;column:establish_contact_date;type:datetime" json:"establish_contact_date"` // subordinate 成为 uid 下级的时间
}

/******sql******
CREATE TABLE `uid_in_use` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uid` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_uid` (`uid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4
******sql******/
// UIdInUse [...]
type UIdInUse struct {
	ID  int64 `gorm:"primary_key;column:id;type:bigint(20);not null" json:"-"`
	UId int   `gorm:"unique;column:uid;type:int(11)" json:"uid"`
}
