package response

import "strconv"

// messageIDs maps every error code to its i18n message ID. The actual message
// templates live in pkg/i18n/locales/*.yaml, keyed by these IDs.
var messageIDs = map[ErrCode]string{
	ErrCodeMalformedJSON:              "error.malformed_json",
	ErrCodeRequestBody:                "error.request_body",
	ErrCodeResourceExists:             "error.resource_exists",
	ErrCodeResourceNotFound:           "error.resource_not_found",
	ErrCodeLegalActionNotFound:        "error.legal_action_not_found",
	ErrCodeBooleanInvalid:             "error.boolean_invalid",
	ErrCodeInteger16Invalid:           "error.integer16_invalid",
	ErrCodeInteger32Invalid:           "error.integer32_invalid",
	ErrCodeInteger64Invalid:           "error.integer64_invalid",
	ErrCodeFloat32Invalid:             "error.float32_invalid",
	ErrCodeFloat64Invalid:             "error.float64_invalid",
	ErrCodeDeviceNotFound:             "error.device_not_found",
	ErrCodeVariableNotWritable:        "error.variable_not_writable",
	ErrCodeDeviceNotConnect:           "error.device_not_connect",
	ErrCodeDeviceOperatorUnSupported:  "error.device_operator_unsupported",
	ErrCodeTooManyJsonPatchOperations: "error.too_many_json_patch_operations",
	ErrCodeInternal:                   "error.internal",
	ErrCodeUserNotFound:               "error.user_not_found",
	ErrCodeUserNameExists:             "error.user_name_exists",
	ErrCodeUserNoExists:               "error.user_no_exists",
	ErrCodeVersionMismatch:            "error.version_mismatch",
	ErrCodeRoleNotFound:               "error.role_not_found",
	ErrCodeRoleNameExists:             "error.role_name_exists",
	ErrCodeRoleCodeExists:             "error.role_code_exists",
	ErrCodeOrgNotFound:                "error.org_not_found",
	ErrCodeOrgCodeExists:              "error.org_code_exists",
	ErrCodeOrgHasChildren:             "error.org_has_children",
	ErrCodeOrgParentNotFound:          "error.org_parent_not_found",
	ErrCodeOrgParentCycle:             "error.org_parent_cycle",
	ErrCodeLoginExpired:               "error.login_expired",
	ErrCodeUnauthorized:               "error.unauthorized",
	ErrCodeUsernamePasswordError:      "error.username_password_error",
	ErrCodeAccountDisabled:            "error.account_disabled",
	ErrCodeAccountLocked:              "error.account_locked",
	ErrCodeLoginFailed:                "error.login_failed",
	ErrCodeCaptchaEmpty:               "error.captcha_empty",
	ErrCodeCaptchaExpired:             "error.captcha_expired",
	ErrCodeCaptchaError:               "error.captcha_error",
	ErrCodeNotFound:                   "error.not_found",
	ErrCodeMethodNotAllowed:           "error.method_not_allowed",
}

func messageID(code ErrCode) string {
	if id, ok := messageIDs[code]; ok {
		return id
	}
	return strconv.Itoa(int(code))
}

// !!! IMPORTANT PLEASE READ FIRST !!!
// You SHOULD add new code at the end of enum firstly.

var (
	ErrMalformedJSON       = generateError(ErrCodeMalformedJSON)
	ErrRequestBody         = generateError(ErrCodeRequestBody)
	ErrLegalActionNotFound = generateError(ErrCodeLegalActionNotFound)
)
