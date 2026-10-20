package identity

import (
	"time"

	"harns-platform/pkg/apis"
)

// 通用状态。
const (
	StatusEnabled  = "enabled"
	StatusDisabled = "disabled"
)

// 性别。
const (
	GenderMale   = "male"
	GenderFemale = "female"
)

// 认证模式。
const (
	AuthModeLocal  = "local"
	AuthModeLdap   = "ldap"
	AuthModeRadius = "radius"
)

// RootParentID 表示组织根节点的父 ID。
const RootParentID = "0"

// User 是用户实体，基础字段来自 apis.Meta。
// RoleIDs / OrgIDs 不是数据库列，仅供组装返回对象。
type User struct {
	apis.Meta
	UserName          string     `json:"userName"`
	Password          string     `json:"-"` // bcrypt 哈希，禁止外泄
	UserNo            string     `json:"userNo"`
	NickName          string     `json:"nickName"`
	BirthDate         *time.Time `json:"birthDate"`
	Gender            string     `json:"gender"`
	Mail              string     `json:"mail"`
	Phone             string     `json:"phone"`
	Note              string     `json:"note"`
	AuthKey           string     `json:"-"` // 认证密钥，禁止外泄
	AuthMode          string     `json:"authMode"`
	LastLoginTime     *time.Time `json:"lastLoginTime"`
	LoginFailureTimes int        `json:"loginFailureTimes"`
	LockedUntil       *time.Time `json:"-"`
	ExpireTimes       int        `json:"expireTimes"`
	Status            string     `json:"status"`

	RoleIDs []string `json:"-"`
	OrgIDs  []string `json:"-"`
	Roles   []*Role  `json:"-"`
	Orgs    []*Org   `json:"-"`

	RoleCodes   []string `json:"-"`
	Permissions []string `json:"-"`
}

// Role 是角色实体。
type Role struct {
	apis.Meta
	Name        string   `json:"name"`
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Permissions []string `json:"permissions"`
}

// Org 是组织(部门)实体。
type Org struct {
	apis.Meta
	Code        string `json:"code"`
	Name        string `json:"name"`
	ParentID    string `json:"parentId"`
	Sort        int    `json:"sort"`
	Description string `json:"description"`
	Remark      string `json:"remark"`
	Status      string `json:"status"`
}

// IsRoot 判断是否为根节点。
func (o *Org) IsRoot() bool {
	return o.ParentID == "" || o.ParentID == RootParentID
}
