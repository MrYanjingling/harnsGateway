package audit

// Operation types recorded by the audit log.
const (
	OpLogin  = "LOGIN"
	OpLogout = "LOGOUT"
	OpCreate = "CREATE"
	OpUpdate = "UPDATE"
	OpDelete = "DELETE"
	OpQuery  = "QUERY"
)

// Operation describes what a request did.
type Operation struct {
	Type string
	Desc string
}
