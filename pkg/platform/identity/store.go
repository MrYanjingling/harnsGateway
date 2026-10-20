package identity

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"time"

	"harns-platform/pkg/apis"
)

// Store 是身份与组织模块的 SQL 访问层（用户/角色/组织/关联）。
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

const (
	userColumns = `id, version, user_name, password, user_no, nick_name, birth_date, gender,
 mail, phone, note, auth_key, auth_mode, last_login_time, login_failure_times, locked_until,
 expire_times, status, created_user, updated_user, created_time, updated_time`
	roleColumns = `id, version, name, code, description, status, created_user, updated_user, created_time, updated_time`
	orgColumns  = `id, version, code, name, parent_id, sort, description, remark, status, created_user, updated_user, created_time, updated_time`
)

// ---------- User ----------

// CreateUserWithRelations 在同一事务内写入用户及其角色/组织关联。
func (s *Store) CreateUserWithRelations(ctx context.Context, u *User, roleIDs, orgIDs []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO users (`+userColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		u.ID, u.Version, u.UserName, u.Password, u.UserNo, u.NickName, u.BirthDate, u.Gender,
		u.Mail, u.Phone, u.Note, u.AuthKey, u.AuthMode, u.LastLoginTime, u.LoginFailureTimes,
		u.LockedUntil, u.ExpireTimes, u.Status, u.CreatedUser, u.UpdatedUser, u.CreatedTime, u.UpdatedTime); err != nil {
		return err
	}
	if err := insertUserIDs(ctx, tx, "user_roles", "role_id", u.ID, roleIDs, u.CreatedTime); err != nil {
		return err
	}
	if err := insertUserIDs(ctx, tx, "user_orgs", "org_id", u.ID, orgIDs, u.CreatedTime); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) GetUser(ctx context.Context, id string) (*User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE id = ?`, id))
	if err != nil {
		return nil, err
	}
	if err := s.fillUserRelations(ctx, []*User{u}); err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByUserName 按用户名查询用户（含密码哈希与关联），用于登录。
func (s *Store) GetUserByUserName(ctx context.Context, userName string) (*User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE user_name = ?`, userName))
	if err != nil {
		return nil, err
	}
	if err := s.fillUserRelations(ctx, []*User{u}); err != nil {
		return nil, err
	}
	return u, nil
}

// RecordLoginFailure 失败次数+1；达到阈值时锁定账户并返回解锁时间。
func (s *Store) RecordLoginFailure(ctx context.Context, id string, maxFailed, lockMinutes int) (int, *time.Time, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET login_failure_times=login_failure_times+1, version=version+1 WHERE id=?`, id); err != nil {
		return 0, nil, err
	}

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT login_failure_times FROM users WHERE id=?`, id).Scan(&count); err != nil {
		return 0, nil, err
	}

	var lockedUntil *time.Time
	if maxFailed > 0 && count >= maxFailed {
		t := time.Now().Add(time.Duration(lockMinutes) * time.Minute)
		lockedUntil = &t
		if _, err := tx.ExecContext(ctx,
			`UPDATE users SET status=?, locked_until=?, version=version+1 WHERE id=?`,
			StatusDisabled, t, id); err != nil {
			return 0, nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, err
	}
	return count, lockedUntil, nil
}

// MarkLoginSuccess 登录成功后清零失败次数并刷新最后登录时间。
func (s *Store) MarkLoginSuccess(ctx context.Context, id string, now time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET login_failure_times=0, last_login_time=?, locked_until=NULL, version=version+1 WHERE id=?`,
		now, id)
	return err
}

// UnlockUser 解除锁定并清零失败次数。
func (s *Store) UnlockUser(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET status=?, login_failure_times=0, locked_until=NULL, version=version+1 WHERE id=?`,
		StatusEnabled, id)
	return err
}

// UpdateUser 更新用户并将 version+1（不使用乐观锁条件）。
func (s *Store) UpdateUser(ctx context.Context, u *User) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE users SET user_name=?, password=?, user_no=?, nick_name=?, birth_date=?, gender=?,
		 mail=?, phone=?, note=?, auth_key=?, auth_mode=?, expire_times=?, status=?,
		 updated_user=?, updated_time=?, version=version+1 WHERE id=?`,
		u.UserName, u.Password, u.UserNo, u.NickName, u.BirthDate, u.Gender, u.Mail, u.Phone,
		u.Note, u.AuthKey, u.AuthMode, u.ExpireTimes, u.Status, u.UpdatedUser, u.UpdatedTime, u.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// UpdateUserPassword 重置密码并将 version+1。
func (s *Store) UpdateUserPassword(ctx context.Context, id, hashed string, now time.Time, updatedUser string) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE users SET password=?, updated_user=?, updated_time=?, version=version+1 WHERE id=?`,
		hashed, updatedUser, now, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteUser 物理删除用户并清理关联（同一事务）。
func (s *Store) DeleteUser(ctx context.Context, id string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id=?`, id)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id=?`, id); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_orgs WHERE user_id=?`, id); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return n, nil
}

func (s *Store) ListUsers(ctx context.Context, q apis.PageQuery, keyword string) ([]*User, int64, error) {
	where, args := "1=1", []any{}
	if keyword != "" {
		where += " AND (user_name LIKE ? OR nick_name LIKE ? OR user_no LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}

	var total int64
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM users WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.db.QueryContext(ctx,
		"SELECT "+userColumns+" FROM users WHERE "+where+" ORDER BY created_time DESC LIMIT ? OFFSET ?",
		append(args, q.Limit(), q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]*User, 0, q.Limit())
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := s.fillUserRelations(ctx, list); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *Store) ExistsUserName(ctx context.Context, name, excludeID string) (bool, error) {
	return s.exists(ctx, "users", "user_name", name, excludeID)
}

func (s *Store) ExistsUserNo(ctx context.Context, userNo, excludeID string) (bool, error) {
	return s.exists(ctx, "users", "user_no", userNo, excludeID)
}

// MissingRoleIDs 返回 roles 中不存在的 id。
func (s *Store) MissingRoleIDs(ctx context.Context, ids []string) ([]string, error) {
	return s.missingIDs(ctx, "roles", ids)
}

// MissingOrgIDs 返回 orgs 中不存在的 id。
func (s *Store) MissingOrgIDs(ctx context.Context, ids []string) ([]string, error) {
	return s.missingIDs(ctx, "orgs", ids)
}

func (s *Store) fillUserRelations(ctx context.Context, users []*User) error {
	if len(users) == 0 {
		return nil
	}
	ids := make([]string, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	roleMap, err := s.relationMap(ctx, "user_roles", "role_id", ids)
	if err != nil {
		return err
	}
	orgMap, err := s.relationMap(ctx, "user_orgs", "org_id", ids)
	if err != nil {
		return err
	}

	roleIDs := collectRelationIDs(roleMap)
	roleObjs, err := s.rolesByIDs(ctx, roleIDs)
	if err != nil {
		return err
	}
	permMap, err := s.permissionsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return err
	}
	for id, r := range roleObjs {
		r.Permissions = permMap[id]
	}
	orgObjs, err := s.orgsByIDs(ctx, collectRelationIDs(orgMap))
	if err != nil {
		return err
	}

	for _, u := range users {
		u.RoleIDs = roleMap[u.ID]
		u.OrgIDs = orgMap[u.ID]
		u.Roles = pickRoles(u.RoleIDs, roleObjs)
		u.Orgs = pickOrgs(u.OrgIDs, orgObjs)

		codes := make([]string, 0, len(u.Roles))
		permissionSet := make(map[string]struct{})
		for _, r := range u.Roles {
			codes = append(codes, r.Code)
			for _, p := range r.Permissions {
				permissionSet[p] = struct{}{}
			}
		}
		u.RoleCodes = codes
		u.Permissions = make([]string, 0, len(permissionSet))
		for p := range permissionSet {
			u.Permissions = append(u.Permissions, p)
		}
		sort.Strings(u.Permissions)
	}
	return nil
}

func collectRelationIDs(m map[string][]string) []string {
	seen := make(map[string]struct{}, len(m))
	out := make([]string, 0, len(m))
	for _, ids := range m {
		for _, id := range ids {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

func (s *Store) rolesByIDs(ctx context.Context, ids []string) (map[string]*Role, error) {
	out := make(map[string]*Role, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+roleColumns+` FROM roles WHERE id IN (`+placeholders(len(ids))+`)`, toAny(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		r, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		out[r.ID] = r
	}
	return out, rows.Err()
}

func (s *Store) orgsByIDs(ctx context.Context, ids []string) (map[string]*Org, error) {
	out := make(map[string]*Org, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+orgColumns+` FROM orgs WHERE id IN (`+placeholders(len(ids))+`)`, toAny(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		o, err := scanOrg(rows)
		if err != nil {
			return nil, err
		}
		out[o.ID] = o
	}
	return out, rows.Err()
}

func pickRoles(ids []string, m map[string]*Role) []*Role {
	out := make([]*Role, 0, len(ids))
	for _, id := range ids {
		if r, ok := m[id]; ok {
			out = append(out, r)
		}
	}
	return out
}

func pickOrgs(ids []string, m map[string]*Org) []*Org {
	out := make([]*Org, 0, len(ids))
	for _, id := range ids {
		if o, ok := m[id]; ok {
			out = append(out, o)
		}
	}
	return out
}

func (s *Store) relationMap(ctx context.Context, table, column string, userIDs []string) (map[string][]string, error) {
	out := make(map[string][]string, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	q := `SELECT user_id, ` + column + ` FROM ` + table + ` WHERE user_id IN (` + placeholders(len(userIDs)) + `)`
	rows, err := s.db.QueryContext(ctx, q, toAny(userIDs)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var uid, rid string
		if err := rows.Scan(&uid, &rid); err != nil {
			return nil, err
		}
		out[uid] = append(out[uid], rid)
	}
	return out, rows.Err()
}

// ---------- Role ----------

func (s *Store) CreateRoleWithPermissions(ctx context.Context, r *Role, permissions []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO roles (`+roleColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		r.ID, r.Version, r.Name, r.Code, r.Description, r.Status,
		r.CreatedUser, r.UpdatedUser, r.CreatedTime, r.UpdatedTime); err != nil {
		return err
	}
	if err := insertRolePermissions(ctx, tx, r.ID, permissions); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) GetRole(ctx context.Context, id string) (*Role, error) {
	r, err := scanRole(s.db.QueryRowContext(ctx, `SELECT `+roleColumns+` FROM roles WHERE id = ?`, id))
	if err != nil {
		return nil, err
	}
	perms, err := s.permissionsByRoleIDs(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	r.Permissions = perms[id]
	return r, nil
}

func (s *Store) UpdateRoleWithPermissions(ctx context.Context, r *Role, permissions []string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE roles SET name=?, code=?, description=?, status=?,
		 updated_user=?, updated_time=?, version=version+1 WHERE id=?`,
		r.Name, r.Code, r.Description, r.Status, r.UpdatedUser, r.UpdatedTime, r.ID)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id=?`, r.ID); err != nil {
		return 0, err
	}
	if err := insertRolePermissions(ctx, tx, r.ID, permissions); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return n, nil
}

// DeleteRole 物理删除角色并清理用户角色/角色权限关联（同一事务）。
func (s *Store) DeleteRole(ctx context.Context, id string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM roles WHERE id=?`, id)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE role_id=?`, id); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id=?`, id); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return n, nil
}

func (s *Store) ListRoles(ctx context.Context, q apis.PageQuery, keyword string) ([]*Role, int64, error) {
	where, args := "1=1", []any{}
	if keyword != "" {
		where += " AND (name LIKE ? OR code LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}

	var total int64
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM roles WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx,
		"SELECT "+roleColumns+" FROM roles WHERE "+where+" ORDER BY created_time DESC LIMIT ? OFFSET ?",
		append(args, q.Limit(), q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]*Role, 0, q.Limit())
	for rows.Next() {
		r, err := scanRole(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, r)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := s.fillRolePermissions(ctx, list); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *Store) ExistsRoleName(ctx context.Context, name, excludeID string) (bool, error) {
	return s.exists(ctx, "roles", "name", name, excludeID)
}

func (s *Store) ExistsRoleCode(ctx context.Context, code, excludeID string) (bool, error) {
	return s.exists(ctx, "roles", "code", code, excludeID)
}

func (s *Store) roleIDsOfUsers(ctx context.Context, userIDs []string) (map[string][]string, error) {
	return s.relationMap(ctx, "user_roles", "role_id", userIDs)
}

func (s *Store) permissionsByRoleIDs(ctx context.Context, roleIDs []string) (map[string][]string, error) {
	out := make(map[string][]string, len(roleIDs))
	if len(roleIDs) == 0 {
		return out, nil
	}
	q := `SELECT role_id, permission FROM role_permissions WHERE role_id IN (` + placeholders(len(roleIDs)) + `)`
	rows, err := s.db.QueryContext(ctx, q, toAny(roleIDs)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var roleID, permission string
		if err := rows.Scan(&roleID, &permission); err != nil {
			return nil, err
		}
		out[roleID] = append(out[roleID], permission)
	}
	return out, rows.Err()
}

func (s *Store) fillRolePermissions(ctx context.Context, roles []*Role) error {
	if len(roles) == 0 {
		return nil
	}
	ids := make([]string, 0, len(roles))
	for _, r := range roles {
		ids = append(ids, r.ID)
	}
	perms, err := s.permissionsByRoleIDs(ctx, ids)
	if err != nil {
		return err
	}
	for _, r := range roles {
		r.Permissions = perms[r.ID]
	}
	return nil
}

func insertRolePermissions(ctx context.Context, tx *sql.Tx, roleID string, permissions []string) error {
	if len(permissions) == 0 {
		return nil
	}
	for _, p := range permissions {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO role_permissions (role_id, permission) VALUES (?, ?)`, roleID, p); err != nil {
			return err
		}
	}
	return nil
}

// ---------- Org ----------
// ---------- Org ----------

func (s *Store) CreateOrg(ctx context.Context, o *Org) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO orgs (`+orgColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		o.ID, o.Version, o.Code, o.Name, o.ParentID, o.Sort, o.Description, o.Remark, o.Status,
		o.CreatedUser, o.UpdatedUser, o.CreatedTime, o.UpdatedTime)
	return err
}

func (s *Store) GetOrg(ctx context.Context, id string) (*Org, error) {
	return scanOrg(s.db.QueryRowContext(ctx, `SELECT `+orgColumns+` FROM orgs WHERE id = ?`, id))
}

func (s *Store) UpdateOrg(ctx context.Context, o *Org) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE orgs SET code=?, name=?, parent_id=?, sort=?, description=?, remark=?, status=?,
		 updated_user=?, updated_time=?, version=version+1 WHERE id=?`,
		o.Code, o.Name, o.ParentID, o.Sort, o.Description, o.Remark, o.Status,
		o.UpdatedUser, o.UpdatedTime, o.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteOrg 物理删除组织并清理用户组织关联（同一事务）。
func (s *Store) DeleteOrg(ctx context.Context, id string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM orgs WHERE id=?`, id)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_orgs WHERE org_id=?`, id); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return n, nil
}

func (s *Store) ListOrgs(ctx context.Context) ([]*Org, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+orgColumns+` FROM orgs ORDER BY sort ASC, created_time ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*Org, 0)
	for rows.Next() {
		o, err := scanOrg(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

func (s *Store) ExistsOrgCode(ctx context.Context, code, excludeID string) (bool, error) {
	return s.exists(ctx, "orgs", "code", code, excludeID)
}

func (s *Store) CountChildren(ctx context.Context, parentID string) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM orgs WHERE parent_id = ?`, parentID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// ---------- helpers ----------

func (s *Store) exists(ctx context.Context, table, column, value, excludeID string) (bool, error) {
	q, args := "SELECT COUNT(1) FROM "+table+" WHERE "+column+" = ?", []any{value}
	if excludeID != "" {
		q += " AND id <> ?"
		args = append(args, excludeID)
	}
	var n int
	if err := s.db.QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *Store) missingIDs(ctx context.Context, table string, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	q := "SELECT id FROM " + table + " WHERE id IN (" + placeholders(len(ids)) + ")"
	rows, err := s.db.QueryContext(ctx, q, toAny(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existing := make(map[string]struct{}, len(ids))
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		existing[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	missing := make([]string, 0)
	for _, id := range ids {
		if _, ok := existing[id]; !ok {
			missing = append(missing, id)
		}
	}
	return missing, nil
}

func insertUserIDs(ctx context.Context, tx *sql.Tx, table, column, userID string, ids []string, now time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	q := "INSERT INTO " + table + " (user_id, " + column + ", created_time) VALUES (?,?,?)"
	for _, id := range ids {
		if _, err := tx.ExecContext(ctx, q, userID, id, now); err != nil {
			return err
		}
	}
	return nil
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("?,", n-1) + "?"
}

func toAny(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

type scanner interface{ Scan(dest ...any) error }

func scanUser(sc scanner) (*User, error) {
	u := &User{}
	err := sc.Scan(&u.ID, &u.Version, &u.UserName, &u.Password, &u.UserNo, &u.NickName,
		&u.BirthDate, &u.Gender, &u.Mail, &u.Phone, &u.Note, &u.AuthKey, &u.AuthMode,
		&u.LastLoginTime, &u.LoginFailureTimes, &u.LockedUntil, &u.ExpireTimes, &u.Status,
		&u.CreatedUser, &u.UpdatedUser, &u.CreatedTime, &u.UpdatedTime)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func scanRole(sc scanner) (*Role, error) {
	r := &Role{}
	err := sc.Scan(&r.ID, &r.Version, &r.Name, &r.Code, &r.Description, &r.Status,
		&r.CreatedUser, &r.UpdatedUser, &r.CreatedTime, &r.UpdatedTime)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func scanOrg(sc scanner) (*Org, error) {
	o := &Org{}
	err := sc.Scan(&o.ID, &o.Version, &o.Code, &o.Name, &o.ParentID, &o.Sort, &o.Description,
		&o.Remark, &o.Status, &o.CreatedUser, &o.UpdatedUser, &o.CreatedTime, &o.UpdatedTime)
	if err != nil {
		return nil, err
	}
	return o, nil
}
