package identity

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"harns-platform/pkg/apis"
	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/logging"
	"harns-platform/pkg/platform/httpx"
	"harns-platform/pkg/platform/identity/v1"
)

// InstallHandler 在给定路由组上注册用户、角色、组织接口。
func InstallHandler(rg *gin.RouterGroup, svc *Service) {
	users := rg.Group("/users")
	users.POST("", httpx.RequirePermission(httpx.PermUserWrite), httpx.WithOperation(httpx.OpCreate, "创建用户"), createUser(svc))
	users.GET("", httpx.RequirePermission(httpx.PermUserRead), httpx.WithOperation(httpx.OpQuery, "分页查询用户"), listUsers(svc))
	users.GET("/:id", httpx.RequirePermission(httpx.PermUserRead), httpx.WithOperation(httpx.OpQuery, "查询用户详情"), getUserById(svc))
	users.PUT("/:id", httpx.RequirePermission(httpx.PermUserWrite), httpx.WithOperation(httpx.OpUpdate, "更新用户"), updateUserById(svc))
	users.PUT("/:id/password", httpx.RequirePermission(httpx.PermUserWrite), httpx.WithOperation(httpx.OpUpdate, "重置用户密码"), resetUserPassword(svc))
	users.DELETE("/:id", httpx.RequirePermission(httpx.PermUserWrite), httpx.WithOperation(httpx.OpDelete, "删除用户"), deleteUserById(svc))

	roles := rg.Group("/roles")
	roles.POST("", httpx.RequirePermission(httpx.PermRoleWrite), httpx.WithOperation(httpx.OpCreate, "创建角色"), createRole(svc))
	roles.GET("", httpx.RequirePermission(httpx.PermRoleRead), httpx.WithOperation(httpx.OpQuery, "分页查询角色"), listRoles(svc))
	roles.GET("/:id", httpx.RequirePermission(httpx.PermRoleRead), httpx.WithOperation(httpx.OpQuery, "查询角色详情"), getRoleById(svc))
	roles.PUT("/:id", httpx.RequirePermission(httpx.PermRoleWrite), httpx.WithOperation(httpx.OpUpdate, "更新角色"), updateRoleById(svc))
	roles.DELETE("/:id", httpx.RequirePermission(httpx.PermRoleWrite), httpx.WithOperation(httpx.OpDelete, "删除角色"), deleteRoleById(svc))

	orgs := rg.Group("/orgs")
	orgs.GET("/tree", httpx.RequirePermission(httpx.PermOrgRead), httpx.WithOperation(httpx.OpQuery, "查询组织树"), treeOrgs(svc))
	orgs.POST("", httpx.RequirePermission(httpx.PermOrgWrite), httpx.WithOperation(httpx.OpCreate, "创建组织"), createOrg(svc))
	orgs.GET("/:id", httpx.RequirePermission(httpx.PermOrgRead), httpx.WithOperation(httpx.OpQuery, "查询组织详情"), getOrgById(svc))
	orgs.PUT("/:id", httpx.RequirePermission(httpx.PermOrgWrite), httpx.WithOperation(httpx.OpUpdate, "更新组织"), updateOrgById(svc))
	orgs.DELETE("/:id", httpx.RequirePermission(httpx.PermOrgWrite), httpx.WithOperation(httpx.OpDelete, "删除组织"), deleteOrgById(svc))
}

// ---------- User handlers ----------

func createUser(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.CreateUserReq
		if !httpx.BindJSON(c, &req) {
			return
		}
		u, err := svc.CreateUser(c.Request.Context(), &req)
		if err != nil {
			logging.V(2).InfoS("Failed to create user", "err", err)
			writeErr(c, err)
			return
		}
		c.Header(apis.ETag, strconv.FormatInt(u.Version, 10))
		response.WriteOK(c, toUserVo(u))
	}
}

func getUserById(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, err := svc.GetUser(c.Request.Context(), c.Param("id"))
		if err != nil {
			writeErr(c, err)
			return
		}
		c.Header(apis.ETag, strconv.FormatInt(u.Version, 10))
		response.WriteOK(c, toUserVo(u))
	}
}

func updateUserById(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.UpdateUserReq
		if !httpx.BindJSON(c, &req) {
			return
		}
		u, err := svc.UpdateUser(c.Request.Context(), c.Param("id"), &req)
		if err != nil {
			logging.V(2).InfoS("Failed to update user", "err", err)
			writeErr(c, err)
			return
		}
		c.Header(apis.ETag, strconv.FormatInt(u.Version, 10))
		response.WriteOK(c, toUserVo(u))
	}
}

func deleteUserById(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := svc.DeleteUser(c.Request.Context(), c.Param("id")); err != nil {
			logging.V(2).InfoS("Failed to delete user", "err", err)
			writeErr(c, err)
			return
		}
		response.WriteOK(c, nil)
	}
}

// resetUserPassword 将用户密码重置为「用户名+@123」。
func resetUserPassword(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, err := svc.ResetPassword(c.Request.Context(), c.Param("id"))
		if err != nil {
			logging.V(2).InfoS("Failed to reset user password", "err", err)
			writeErr(c, err)
			return
		}
		c.Header(apis.ETag, strconv.FormatInt(u.Version, 10))
		response.WriteOK(c, toUserVo(u))
	}
}

func listUsers(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var q v1.UserQuery
		if !httpx.BindQuery(c, &q) {
			return
		}
		q.Normalize()

		list, total, err := svc.ListUsers(c.Request.Context(), q.PageQuery, q.Keyword)
		if err != nil {
			logging.V(2).InfoS("Failed to list users", "err", err)
			writeErr(c, err)
			return
		}
		response.WriteOK(c, apis.NewPageResp(toUserVos(list), total, q.PageQuery))
	}
}

// ---------- Role handlers ----------

func createRole(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.CreateRoleReq
		if !httpx.BindJSON(c, &req) {
			return
		}
		r, err := svc.CreateRole(c.Request.Context(), &req)
		if err != nil {
			logging.V(2).InfoS("Failed to create role", "err", err)
			writeErr(c, err)
			return
		}
		c.Header(apis.ETag, strconv.FormatInt(r.Version, 10))
		response.WriteOK(c, toRoleVo(r))
	}
}

func getRoleById(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		r, err := svc.GetRole(c.Request.Context(), c.Param("id"))
		if err != nil {
			writeErr(c, err)
			return
		}
		c.Header(apis.ETag, strconv.FormatInt(r.Version, 10))
		response.WriteOK(c, toRoleVo(r))
	}
}

func updateRoleById(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.UpdateRoleReq
		if !httpx.BindJSON(c, &req) {
			return
		}
		r, err := svc.UpdateRole(c.Request.Context(), c.Param("id"), &req)
		if err != nil {
			logging.V(2).InfoS("Failed to update role", "err", err)
			writeErr(c, err)
			return
		}
		c.Header(apis.ETag, strconv.FormatInt(r.Version, 10))
		response.WriteOK(c, toRoleVo(r))
	}
}

func deleteRoleById(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := svc.DeleteRole(c.Request.Context(), c.Param("id")); err != nil {
			logging.V(2).InfoS("Failed to delete role", "err", err)
			writeErr(c, err)
			return
		}
		response.WriteOK(c, nil)
	}
}

func listRoles(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var q v1.RoleQuery
		if !httpx.BindQuery(c, &q) {
			return
		}
		q.Normalize()

		list, total, err := svc.ListRoles(c.Request.Context(), q.PageQuery, q.Keyword)
		if err != nil {
			logging.V(2).InfoS("Failed to list roles", "err", err)
			writeErr(c, err)
			return
		}
		response.WriteOK(c, apis.NewPageResp(toRoleVos(list), total, q.PageQuery))
	}
}

// ---------- Org handlers ----------

func createOrg(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.CreateOrgReq
		if !httpx.BindJSON(c, &req) {
			return
		}
		o, err := svc.CreateOrg(c.Request.Context(), &req)
		if err != nil {
			logging.V(2).InfoS("Failed to create org", "err", err)
			writeErr(c, err)
			return
		}
		c.Header(apis.ETag, strconv.FormatInt(o.Version, 10))
		response.WriteOK(c, toOrgVo(o))
	}
}

func getOrgById(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, err := svc.GetOrg(c.Request.Context(), c.Param("id"))
		if err != nil {
			writeErr(c, err)
			return
		}
		c.Header(apis.ETag, strconv.FormatInt(o.Version, 10))
		response.WriteOK(c, toOrgVo(o))
	}
}

func updateOrgById(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.UpdateOrgReq
		if !httpx.BindJSON(c, &req) {
			return
		}
		o, err := svc.UpdateOrg(c.Request.Context(), c.Param("id"), &req)
		if err != nil {
			logging.V(2).InfoS("Failed to update org", "err", err)
			writeErr(c, err)
			return
		}
		c.Header(apis.ETag, strconv.FormatInt(o.Version, 10))
		response.WriteOK(c, toOrgVo(o))
	}
}

func deleteOrgById(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := svc.DeleteOrg(c.Request.Context(), c.Param("id")); err != nil {
			logging.V(2).InfoS("Failed to delete org", "err", err)
			writeErr(c, err)
			return
		}
		response.WriteOK(c, nil)
	}
}

func treeOrgs(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgs, err := svc.TreeOrgs(c.Request.Context())
		if err != nil {
			logging.V(2).InfoS("Failed to query org tree", "err", err)
			writeErr(c, err)
			return
		}
		response.WriteOK(c, buildOrgTree(orgs))
	}
}

// ---------- shared ----------

// writeErr 业务错误 400，内部错误 500。
func writeErr(c *gin.Context, err error) {
	status := http.StatusBadRequest
	if response.ErrorCode(err) == int(response.ErrCodeInternal) {
		status = http.StatusInternalServerError
	}
	response.WriteError(c, status, err)
}

// ToUserVo converts a User entity to its external VO.
func ToUserVo(u *User) *v1.UserVo {
	return toUserVo(u)
}

func toUserVo(u *User) *v1.UserVo {
	vo := &v1.UserVo{
		ID:                u.ID,
		Version:           u.Version,
		UserName:          u.UserName,
		UserNo:            u.UserNo,
		NickName:          u.NickName,
		Gender:            u.Gender,
		Mail:              u.Mail,
		Phone:             u.Phone,
		Note:              u.Note,
		AuthMode:          u.AuthMode,
		LastLoginTime:     u.LastLoginTime,
		LoginFailureTimes: u.LoginFailureTimes,
		ExpireTimes:       u.ExpireTimes,
		Status:            u.Status,
		CreatedUser:       u.CreatedUser,
		UpdatedUser:       u.UpdatedUser,
		CreatedTime:       u.CreatedTime,
		UpdatedTime:       u.UpdatedTime,
		RoleIDs:           u.RoleIDs,
		OrgIDs:            u.OrgIDs,
		Roles:             toRoleVos(u.Roles),
		Orgs:              toOrgVos(u.Orgs),
		Permissions:       u.Permissions,
	}
	if u.BirthDate != nil {
		vo.BirthDate = u.BirthDate.Format("2006-01-02")
	}
	return vo
}

func toUserVos(users []*User) []*v1.UserVo {
	out := make([]*v1.UserVo, 0, len(users))
	for _, u := range users {
		out = append(out, toUserVo(u))
	}
	return out
}

func toRoleVo(r *Role) *v1.RoleVo {
	return &v1.RoleVo{
		ID:          r.ID,
		Version:     r.Version,
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description,
		Status:      r.Status,
		Permissions: r.Permissions,
		CreatedUser: r.CreatedUser,
		UpdatedUser: r.UpdatedUser,
		CreatedTime: r.CreatedTime,
		UpdatedTime: r.UpdatedTime,
	}
}

func toRoleVos(roles []*Role) []*v1.RoleVo {
	out := make([]*v1.RoleVo, 0, len(roles))
	for _, r := range roles {
		out = append(out, toRoleVo(r))
	}
	return out
}

func toOrgVo(o *Org) *v1.OrgVo {
	return &v1.OrgVo{
		ID:          o.ID,
		Version:     o.Version,
		Code:        o.Code,
		Name:        o.Name,
		ParentID:    o.ParentID,
		Sort:        o.Sort,
		Description: o.Description,
		Remark:      o.Remark,
		Status:      o.Status,
		CreatedUser: o.CreatedUser,
		UpdatedUser: o.UpdatedUser,
		CreatedTime: o.CreatedTime,
		UpdatedTime: o.UpdatedTime,
	}
}

func toOrgVos(orgs []*Org) []*v1.OrgVo {
	out := make([]*v1.OrgVo, 0, len(orgs))
	for _, o := range orgs {
		out = append(out, toOrgVo(o))
	}
	return out
}

// buildOrgTree 把扁平组织列表构建成树（根节点数组）。
func buildOrgTree(orgs []*Org) []*v1.OrgVo {
	nodes := make(map[string]*v1.OrgVo, len(orgs))
	for _, o := range orgs {
		nodes[o.ID] = toOrgVo(o)
	}
	roots := make([]*v1.OrgVo, 0)
	for _, o := range orgs {
		node := nodes[o.ID]
		if o.IsRoot() {
			roots = append(roots, node)
			continue
		}
		parent, ok := nodes[o.ParentID]
		if !ok {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}
	return roots
}
