package identity

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"

	"harns-platform/pkg/apis"
	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/logging"
	"harns-platform/pkg/platform/identity/v1"
	"harns-platform/pkg/platform/principal"
)

// SessionInvalidator clears the login sessions of a user.
type SessionInvalidator interface {
	InvalidateUserSessions(ctx context.Context, userID string) error
}

// Service 是身份与组织模块的业务层。
type Service struct {
	store              *Store
	sessionInvalidator SessionInvalidator
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

// SetSessionInvalidator sets the hook used to invalidate user sessions.
func (s *Service) SetSessionInvalidator(inv SessionInvalidator) {
	s.sessionInvalidator = inv
}

func (s *Service) invalidateSessions(ctx context.Context, userID string) {
	if s.sessionInvalidator != nil && userID != "" {
		_ = s.sessionInvalidator.InvalidateUserSessions(ctx, userID)
	}
}

// ---------- User ----------

func (s *Service) CreateUser(ctx context.Context, req *v1.CreateUserReq) (*User, error) {
	defer logging.Trace("identity.CreateUser", "userName", req.UserName)()
	if dup, err := s.store.ExistsUserName(ctx, req.UserName, ""); err != nil {
		return nil, response.ErrInternal(err)
	} else if dup {
		return nil, response.ErrUserNameExists(req.UserName)
	}
	if dup, err := s.store.ExistsUserNo(ctx, req.UserNo, ""); err != nil {
		return nil, response.ErrInternal(err)
	} else if dup {
		return nil, response.ErrUserNoExists(req.UserNo)
	}

	roleIDs := dedupe(req.RoleIDs)
	orgIDs := dedupe(req.OrgIDs)
	if missing, err := s.store.MissingRoleIDs(ctx, roleIDs); err != nil {
		return nil, response.ErrInternal(err)
	} else if len(missing) > 0 {
		return nil, response.ErrRoleNotFound(missing[0])
	}
	if missing, err := s.store.MissingOrgIDs(ctx, orgIDs); err != nil {
		return nil, response.ErrInternal(err)
	} else if len(missing) > 0 {
		return nil, response.ErrOrgNotFound(missing[0])
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(passwordOrDefault(req.Password, req.UserName)), bcrypt.DefaultCost)
	if err != nil {
		return nil, response.ErrInternal(err)
	}

	now := time.Now()
	operator := principal.UserID(ctx)
	u := &User{
		Meta: apis.Meta{
			ID:          ulid.Make().String(),
			Version:     0,
			CreatedUser: operator,
			UpdatedUser: operator,
			CreatedTime: now,
			UpdatedTime: now,
		},
		UserName:    req.UserName,
		Password:    string(hashed),
		UserNo:      req.UserNo,
		NickName:    req.NickName,
		BirthDate:   parseDate(req.BirthDate),
		Gender:      req.Gender,
		Mail:        req.Mail,
		Phone:       req.Phone,
		Note:        req.Note,
		AuthKey:     req.AuthKey,
		AuthMode:    orDefault(req.AuthMode, AuthModeLocal),
		ExpireTimes: req.ExpireTimes,
		Status:      orDefault(req.Status, StatusEnabled),
		RoleIDs:     roleIDs,
		OrgIDs:      orgIDs,
	}
	if err := s.store.CreateUserWithRelations(ctx, u, roleIDs, orgIDs); err != nil {
		return nil, response.ErrInternal(err)
	}
	return s.GetUser(ctx, u.ID)
}

func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
	defer logging.Trace("identity.GetUser", "id", id)()
	u, err := s.store.GetUser(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.ErrUserNotFound(id)
	}
	if err != nil {
		return nil, response.ErrInternal(err)
	}
	return u, nil
}

// GetUserByUserName 按用户名查询用户（含密码哈希与关联），用于登录。
func (s *Service) GetUserByUserName(ctx context.Context, userName string) (*User, error) {
	u, err := s.store.GetUserByUserName(ctx, userName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.ErrUserNotFound(userName)
	}
	if err != nil {
		return nil, response.ErrInternal(err)
	}
	return u, nil
}

// RecordLoginFailure 记录一次登录失败；达到阈值时锁定账户。
func (s *Service) RecordLoginFailure(ctx context.Context, id string, maxFailed, lockMinutes int) (int, *time.Time, error) {
	count, lockedUntil, err := s.store.RecordLoginFailure(ctx, id, maxFailed, lockMinutes)
	if err != nil {
		return 0, nil, response.ErrInternal(err)
	}
	return count, lockedUntil, nil
}

// MarkLoginSuccess 登录成功：清零失败次数、刷新最后登录时间。
func (s *Service) MarkLoginSuccess(ctx context.Context, id string, now time.Time) error {
	if err := s.store.MarkLoginSuccess(ctx, id, now); err != nil {
		return response.ErrInternal(err)
	}
	return nil
}

// UnlockUser 解除账户锁定。
func (s *Service) UnlockUser(ctx context.Context, id string) error {
	if err := s.store.UnlockUser(ctx, id); err != nil {
		return response.ErrInternal(err)
	}
	return nil
}

func (s *Service) UpdateUser(ctx context.Context, id string, req *v1.UpdateUserReq) (*User, error) {
	defer logging.Trace("identity.UpdateUser", "id", id)()
	u, err := s.store.GetUser(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.ErrUserNotFound(id)
	} else if err != nil {
		return nil, response.ErrInternal(err)
	}

	if req.UserName != nil && *req.UserName != u.UserName {
		if dup, err := s.store.ExistsUserName(ctx, *req.UserName, id); err != nil {
			return nil, response.ErrInternal(err)
		} else if dup {
			return nil, response.ErrUserNameExists(*req.UserName)
		}
		u.UserName = *req.UserName
	}
	if req.UserNo != nil && *req.UserNo != u.UserNo {
		if dup, err := s.store.ExistsUserNo(ctx, *req.UserNo, id); err != nil {
			return nil, response.ErrInternal(err)
		} else if dup {
			return nil, response.ErrUserNoExists(*req.UserNo)
		}
		u.UserNo = *req.UserNo
	}
	passwordChanged := false
	if req.Password != nil && *req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, response.ErrInternal(err)
		}
		u.Password = string(hashed)
		passwordChanged = true
	}
	setIf(&u.NickName, req.NickName)
	setIf(&u.Gender, req.Gender)
	setIf(&u.Mail, req.Mail)
	setIf(&u.Phone, req.Phone)
	setIf(&u.Note, req.Note)
	setIf(&u.AuthKey, req.AuthKey)
	setIf(&u.AuthMode, req.AuthMode)
	setIf(&u.Status, req.Status)
	if req.ExpireTimes != nil {
		u.ExpireTimes = *req.ExpireTimes
	}
	if req.BirthDate != nil {
		u.BirthDate = parseDate(*req.BirthDate)
	}
	u.UpdatedUser = principal.UserID(ctx)
	u.UpdatedTime = time.Now()

	rows, err := s.store.UpdateUser(ctx, u)
	if err != nil {
		return nil, response.ErrInternal(err)
	}
	if rows == 0 {
		return nil, response.ErrUserNotFound(id)
	}
	if passwordChanged || u.Status == StatusDisabled {
		s.invalidateSessions(ctx, id)
	}
	return s.GetUser(ctx, id)
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	defer logging.Trace("identity.DeleteUser", "id", id)()
	rows, err := s.store.DeleteUser(ctx, id)
	if err != nil {
		return response.ErrInternal(err)
	}
	if rows == 0 {
		return response.ErrUserNotFound(id)
	}
	s.invalidateSessions(ctx, id)
	return nil
}

// ResetPassword 将用户密码重置为「用户名+@123」。
func (s *Service) ResetPassword(ctx context.Context, id string) (*User, error) {
	defer logging.Trace("identity.ResetPassword", "id", id)()
	u, err := s.store.GetUser(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.ErrUserNotFound(id)
	} else if err != nil {
		return nil, response.ErrInternal(err)
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(defaultPassword(u.UserName)), bcrypt.DefaultCost)
	if err != nil {
		return nil, response.ErrInternal(err)
	}
	rows, err := s.store.UpdateUserPassword(ctx, id, string(hashed), time.Now(), principal.UserID(ctx))
	if err != nil {
		return nil, response.ErrInternal(err)
	}
	if rows == 0 {
		return nil, response.ErrUserNotFound(id)
	}
	s.invalidateSessions(ctx, id)
	return s.GetUser(ctx, id)
}

func (s *Service) ListUsers(ctx context.Context, q apis.PageQuery, keyword string) ([]*User, int64, error) {
	defer logging.Trace("identity.ListUsers", "page", q.Page, "pageSize", q.PageSize, "keyword", keyword)()
	list, total, err := s.store.ListUsers(ctx, q, keyword)
	if err != nil {
		return nil, 0, response.ErrInternal(err)
	}
	return list, total, nil
}

// ---------- Role ----------

func (s *Service) CreateRole(ctx context.Context, req *v1.CreateRoleReq) (*Role, error) {
	defer logging.Trace("identity.CreateRole", "name", req.Name)()
	if dup, err := s.store.ExistsRoleName(ctx, req.Name, ""); err != nil {
		return nil, response.ErrInternal(err)
	} else if dup {
		return nil, response.ErrRoleNameExists(req.Name)
	}
	if dup, err := s.store.ExistsRoleCode(ctx, req.Code, ""); err != nil {
		return nil, response.ErrInternal(err)
	} else if dup {
		return nil, response.ErrRoleCodeExists(req.Code)
	}

	now := time.Now()
	operator := principal.UserID(ctx)
	r := &Role{
		Meta: apis.Meta{
			ID:          ulid.Make().String(),
			Version:     0,
			CreatedUser: operator,
			UpdatedUser: operator,
			CreatedTime: now,
			UpdatedTime: now,
		},
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      orDefault(req.Status, StatusEnabled),
	}
	if err := s.store.CreateRole(ctx, r); err != nil {
		return nil, response.ErrInternal(err)
	}
	return r, nil
}

func (s *Service) GetRole(ctx context.Context, id string) (*Role, error) {
	defer logging.Trace("identity.GetRole", "id", id)()
	r, err := s.store.GetRole(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.ErrRoleNotFound(id)
	}
	if err != nil {
		return nil, response.ErrInternal(err)
	}
	return r, nil
}

func (s *Service) UpdateRole(ctx context.Context, id string, req *v1.UpdateRoleReq) (*Role, error) {
	defer logging.Trace("identity.UpdateRole", "id", id)()
	r, err := s.store.GetRole(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.ErrRoleNotFound(id)
	} else if err != nil {
		return nil, response.ErrInternal(err)
	}

	if req.Name != nil && *req.Name != r.Name {
		if dup, err := s.store.ExistsRoleName(ctx, *req.Name, id); err != nil {
			return nil, response.ErrInternal(err)
		} else if dup {
			return nil, response.ErrRoleNameExists(*req.Name)
		}
		r.Name = *req.Name
	}
	if req.Code != nil && *req.Code != r.Code {
		if dup, err := s.store.ExistsRoleCode(ctx, *req.Code, id); err != nil {
			return nil, response.ErrInternal(err)
		} else if dup {
			return nil, response.ErrRoleCodeExists(*req.Code)
		}
		r.Code = *req.Code
	}
	setIf(&r.Description, req.Description)
	setIf(&r.Status, req.Status)
	r.UpdatedUser = principal.UserID(ctx)
	r.UpdatedTime = time.Now()

	rows, err := s.store.UpdateRole(ctx, r)
	if err != nil {
		return nil, response.ErrInternal(err)
	}
	if rows == 0 {
		return nil, response.ErrRoleNotFound(id)
	}
	return s.GetRole(ctx, id)
}

func (s *Service) DeleteRole(ctx context.Context, id string) error {
	defer logging.Trace("identity.DeleteRole", "id", id)()
	rows, err := s.store.DeleteRole(ctx, id)
	if err != nil {
		return response.ErrInternal(err)
	}
	if rows == 0 {
		return response.ErrRoleNotFound(id)
	}
	return nil
}

func (s *Service) ListRoles(ctx context.Context, q apis.PageQuery, keyword string) ([]*Role, int64, error) {
	defer logging.Trace("identity.ListRoles", "page", q.Page, "pageSize", q.PageSize, "keyword", keyword)()
	list, total, err := s.store.ListRoles(ctx, q, keyword)
	if err != nil {
		return nil, 0, response.ErrInternal(err)
	}
	return list, total, nil
}

// ---------- Org ----------

func (s *Service) CreateOrg(ctx context.Context, req *v1.CreateOrgReq) (*Org, error) {
	defer logging.Trace("identity.CreateOrg", "code", req.Code)()
	if dup, err := s.store.ExistsOrgCode(ctx, req.Code, ""); err != nil {
		return nil, response.ErrInternal(err)
	} else if dup {
		return nil, response.ErrOrgCodeExists(req.Code)
	}

	parentID := normalizeParentID(req.ParentID)
	if parentID != "" {
		if err := s.ensureParentExists(ctx, parentID); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	operator := principal.UserID(ctx)
	o := &Org{
		Meta: apis.Meta{
			ID:          ulid.Make().String(),
			Version:     0,
			CreatedUser: operator,
			UpdatedUser: operator,
			CreatedTime: now,
			UpdatedTime: now,
		},
		Code:        req.Code,
		Name:        req.Name,
		ParentID:    parentID,
		Sort:        req.Sort,
		Description: req.Description,
		Remark:      req.Remark,
		Status:      orDefault(req.Status, StatusEnabled),
	}
	if err := s.store.CreateOrg(ctx, o); err != nil {
		return nil, response.ErrInternal(err)
	}
	return o, nil
}

func (s *Service) GetOrg(ctx context.Context, id string) (*Org, error) {
	defer logging.Trace("identity.GetOrg", "id", id)()
	o, err := s.store.GetOrg(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.ErrOrgNotFound(id)
	}
	if err != nil {
		return nil, response.ErrInternal(err)
	}
	return o, nil
}

func (s *Service) UpdateOrg(ctx context.Context, id string, req *v1.UpdateOrgReq) (*Org, error) {
	defer logging.Trace("identity.UpdateOrg", "id", id)()
	o, err := s.store.GetOrg(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.ErrOrgNotFound(id)
	} else if err != nil {
		return nil, response.ErrInternal(err)
	}

	if req.Code != nil && *req.Code != o.Code {
		if dup, err := s.store.ExistsOrgCode(ctx, *req.Code, id); err != nil {
			return nil, response.ErrInternal(err)
		} else if dup {
			return nil, response.ErrOrgCodeExists(*req.Code)
		}
		o.Code = *req.Code
	}
	if req.ParentID != nil {
		parentID := normalizeParentID(*req.ParentID)
		if parentID != "" {
			if parentID == id {
				return nil, response.ErrOrgParentCycle(id)
			}
			desc, err := s.isDescendant(ctx, id, parentID)
			if err != nil {
				return nil, response.ErrInternal(err)
			}
			if desc {
				return nil, response.ErrOrgParentCycle(parentID)
			}
			if err := s.ensureParentExists(ctx, parentID); err != nil {
				return nil, err
			}
		}
		o.ParentID = parentID
	}
	setIf(&o.Name, req.Name)
	if req.Sort != nil {
		o.Sort = *req.Sort
	}
	setIf(&o.Description, req.Description)
	setIf(&o.Remark, req.Remark)
	setIf(&o.Status, req.Status)
	o.UpdatedUser = principal.UserID(ctx)
	o.UpdatedTime = time.Now()

	rows, err := s.store.UpdateOrg(ctx, o)
	if err != nil {
		return nil, response.ErrInternal(err)
	}
	if rows == 0 {
		return nil, response.ErrOrgNotFound(id)
	}
	return s.GetOrg(ctx, id)
}

func (s *Service) DeleteOrg(ctx context.Context, id string) error {
	defer logging.Trace("identity.DeleteOrg", "id", id)()
	cnt, err := s.store.CountChildren(ctx, id)
	if err != nil {
		return response.ErrInternal(err)
	}
	if cnt > 0 {
		return response.ErrOrgHasChildren(id)
	}
	rows, err := s.store.DeleteOrg(ctx, id)
	if err != nil {
		return response.ErrInternal(err)
	}
	if rows == 0 {
		return response.ErrOrgNotFound(id)
	}
	return nil
}

// TreeOrgs 返回组织树（根节点列表，子节点挂到 Children）。
func (s *Service) TreeOrgs(ctx context.Context) ([]*Org, error) {
	defer logging.Trace("identity.TreeOrgs")()
	orgs, err := s.store.ListOrgs(ctx)
	if err != nil {
		return nil, response.ErrInternal(err)
	}
	return orgs, nil
}

func (s *Service) ensureParentExists(ctx context.Context, parentID string) error {
	if _, err := s.store.GetOrg(ctx, parentID); errors.Is(err, sql.ErrNoRows) {
		return response.ErrOrgParentNotFound(parentID)
	} else if err != nil {
		return response.ErrInternal(err)
	}
	return nil
}

// isDescendant 沿父链上行判断 candidate 是否为 nodeID 的后代，用于防止成环。
func (s *Service) isDescendant(ctx context.Context, nodeID, candidate string) (bool, error) {
	cur := candidate
	for cur != "" && cur != RootParentID {
		if cur == nodeID {
			return true, nil
		}
		p, err := s.store.GetOrg(ctx, cur)
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		cur = p.ParentID
	}
	return false, nil
}

// ---------- helpers ----------

func normalizeParentID(v string) string {
	if v == RootParentID {
		return ""
	}
	return v
}

// defaultPassword 返回默认密码：用户名 + @123。
func defaultPassword(userName string) string {
	return userName + "@123"
}

// passwordOrDefault 未显式传入密码时使用默认密码。
func passwordOrDefault(password, userName string) string {
	if password != "" {
		return password
	}
	return defaultPassword(userName)
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func setIf(dst *string, v *string) {
	if v != nil {
		*dst = *v
	}
}

func parseDate(v string) *time.Time {
	if v == "" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", v, time.Local)
	if err != nil {
		return nil
	}
	return &t
}

// dedupe 去重并丢弃空串，保持原顺序。
func dedupe(ids []string) []string {
	out := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
