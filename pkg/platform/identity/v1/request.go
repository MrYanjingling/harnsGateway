package v1

import "harns-platform/pkg/apis"

// ---------- User ----------

// CreateUserReq 创建用户请求。
type CreateUserReq struct {
	UserName    string   `json:"userName" binding:"required,max=64"`
	Password    string   `json:"password" binding:"omitempty,min=6,max=64"`
	UserNo      string   `json:"userNo" binding:"required,max=64"`
	NickName    string   `json:"nickName" binding:"required,max=64"`
	BirthDate   string   `json:"birthDate" binding:"omitempty,datetime=2006-01-02"`
	Gender      string   `json:"gender" binding:"omitempty,oneof=male female"`
	Mail        string   `json:"mail" binding:"omitempty,email"`
	Phone       string   `json:"phone" binding:"omitempty,max=32"`
	Note        string   `json:"note" binding:"omitempty,max=255"`
	AuthKey     string   `json:"authKey" binding:"omitempty,max=128"`
	AuthMode    string   `json:"authMode" binding:"omitempty,oneof=local ldap radius"`
	ExpireTimes int      `json:"expireTimes" binding:"omitempty,min=0"`
	Status      string   `json:"status" binding:"omitempty,oneof=enabled disabled"`
	RoleIDs     []string `json:"roleIds" binding:"omitempty,dive,max=32"`
	OrgIDs      []string `json:"orgIds" binding:"omitempty,dive,max=32"`
}

// UpdateUserReq 更新用户请求，字段可选，nil 表示不修改。
type UpdateUserReq struct {
	UserName    *string `json:"userName" binding:"omitempty,max=64"`
	Password    *string `json:"password" binding:"omitempty,min=6,max=64"`
	UserNo      *string `json:"userNo" binding:"omitempty,max=64"`
	NickName    *string `json:"nickName" binding:"omitempty,max=64"`
	BirthDate   *string `json:"birthDate" binding:"omitempty,datetime=2006-01-02"`
	Gender      *string `json:"gender" binding:"omitempty,oneof=male female"`
	Mail        *string `json:"mail" binding:"omitempty,email"`
	Phone       *string `json:"phone" binding:"omitempty,max=32"`
	Note        *string `json:"note" binding:"omitempty,max=255"`
	AuthKey     *string `json:"authKey" binding:"omitempty,max=128"`
	AuthMode    *string `json:"authMode" binding:"omitempty,oneof=local ldap radius"`
	ExpireTimes *int    `json:"expireTimes" binding:"omitempty,min=0"`
	Status      *string `json:"status" binding:"omitempty,oneof=enabled disabled"`
}

// UserQuery 用户分页查询参数。
type UserQuery struct {
	apis.PageQuery
	Keyword string `json:"keyword" form:"keyword" binding:"omitempty,max=64"`
}

// ---------- Role ----------

// CreateRoleReq 创建角色请求。
type CreateRoleReq struct {
	Name        string `json:"name" binding:"required,max=64"`
	Code        string `json:"code" binding:"required,max=64"`
	Description string `json:"description" binding:"omitempty,max=255"`
	Status      string `json:"status" binding:"omitempty,oneof=enabled disabled"`
}

// UpdateRoleReq 更新角色请求，字段可选。
type UpdateRoleReq struct {
	Name        *string `json:"name" binding:"omitempty,max=64"`
	Code        *string `json:"code" binding:"omitempty,max=64"`
	Description *string `json:"description" binding:"omitempty,max=255"`
	Status      *string `json:"status" binding:"omitempty,oneof=enabled disabled"`
}

// RoleQuery 角色分页查询参数。
type RoleQuery struct {
	apis.PageQuery
	Keyword string `json:"keyword" form:"keyword" binding:"omitempty,max=64"`
}

// ---------- Org ----------

// CreateOrgReq 创建组织请求。
type CreateOrgReq struct {
	Code        string `json:"code" binding:"required,max=64"`
	Name        string `json:"name" binding:"required,max=64"`
	ParentID    string `json:"parentId" binding:"omitempty,max=32"`
	Sort        int    `json:"sort" binding:"omitempty,min=0"`
	Description string `json:"description" binding:"omitempty,max=255"`
	Remark      string `json:"remark" binding:"omitempty,max=255"`
	Status      string `json:"status" binding:"omitempty,oneof=enabled disabled"`
}

// UpdateOrgReq 更新组织请求，字段可选。
type UpdateOrgReq struct {
	Code        *string `json:"code" binding:"omitempty,max=64"`
	Name        *string `json:"name" binding:"omitempty,max=64"`
	ParentID    *string `json:"parentId" binding:"omitempty,max=32"`
	Sort        *int    `json:"sort" binding:"omitempty,min=0"`
	Description *string `json:"description" binding:"omitempty,max=255"`
	Remark      *string `json:"remark" binding:"omitempty,max=255"`
	Status      *string `json:"status" binding:"omitempty,oneof=enabled disabled"`
}
