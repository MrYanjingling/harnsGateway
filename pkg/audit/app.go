package audit

import "strings"

// Application identifiers stored in audit_logs.app.
const (
	AppIOT = "IOT"
	AppEMS = "EMS"
	AppEAM = "EAM"
)

func normalizeApp(app string) string {
	return strings.ToUpper(strings.TrimSpace(app))
}
