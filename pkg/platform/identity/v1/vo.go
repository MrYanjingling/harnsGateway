package v1

import "time"

// UserVo 是对外返回的用户对象，不含 password / authKey。
type UserVo struct {
	ID                string     `json:"id"`
	Version           int64      `json:"version"`
	UserName          string     `json:"userName"`
	UserNo            string     `json:"userNo"`
	NickName          string     `json:"nickName"`
	BirthDate         string     `json:"birthDate"`
	Gender            string     `json:"gender"`
	Mail              string     `json:"mail"`
	Phone             string     `json:"phone"`
	Note              string     `json:"note"`
	AuthMode          string     `json:"authMode"`
	LastLoginTime     *time.Time `json:"lastLoginTime"`
	LoginFailureTimes int        `json:"loginFailureTimes"`
	ExpireTimes       int        `json:"expireTimes"`
	Status            string     `json:"status"`
	CreatedUser       string     `json:"createdUser"`
	UpdatedUser       string     `json:"updatedUser"`
	CreatedTime       time.Time  `json:"createdTime"`
	UpdatedTime       time.Time  `json:"updatedTime"`
	RoleIDs           []string   `json:"roleIds"`
	OrgIDs            []string   `json:"orgIds"`
	Roles             []*RoleVo  `json:"roles"`
	Orgs              []*OrgVo   `json:"orgs"`
}

// RoleVo 是对外返回的角色对象。
type RoleVo struct {
	ID          string    `json:"id"`
	Version     int64     `json:"version"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedUser string    `json:"createdUser"`
	UpdatedUser string    `json:"updatedUser"`
	CreatedTime time.Time `json:"createdTime"`
	UpdatedTime time.Time `json:"updatedTime"`
}

// OrgVo 是对外返回的组织对象，树状查询时携带 Children。
type OrgVo struct {
	ID          string    `json:"id"`
	Version     int64     `json:"version"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	ParentID    string    `json:"parentId"`
	Sort        int       `json:"sort"`
	Description string    `json:"description"`
	Remark      string    `json:"remark"`
	Status      string    `json:"status"`
	CreatedUser string    `json:"createdUser"`
	UpdatedUser string    `json:"updatedUser"`
	CreatedTime time.Time `json:"createdTime"`
	UpdatedTime time.Time `json:"updatedTime"`
	Children    []*OrgVo  `json:"children,omitempty"`
}
