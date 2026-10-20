package response

type ErrCode int

const (
	_                                 ErrCode = 10000 + iota
	ErrCodeMalformedJSON                      // 10001
	ErrCodeRequestBody                        // 10002
	ErrCodeResourceExists                     // 10003
	ErrCodeResourceNotFound                   // 10004
	ErrCodeLegalActionNotFound                // 10005
	ErrCodeBooleanInvalid                     // 10006
	ErrCodeInteger16Invalid                   // 10007
	ErrCodeInteger32Invalid                   // 10008
	ErrCodeInteger64Invalid                   // 10009
	ErrCodeFloat32Invalid                     // 10010
	ErrCodeFloat64Invalid                     // 10011
	ErrCodeDeviceNotFound                     // 10012
	ErrCodeVariableNotWritable                // 10013
	ErrCodeDeviceNotConnect                   // 10014
	ErrCodeDeviceOperatorUnSupported          // 10015
	ErrCodeTooManyJsonPatchOperations         // 10016
	ErrCodeInternal                           // 10017
	ErrCodeUserNotFound                       // 10018
	ErrCodeUserNameExists                     // 10019
	ErrCodeUserNoExists                       // 10020
	ErrCodeVersionMismatch                    // 10021
	ErrCodeRoleNotFound                       // 10022
	ErrCodeRoleNameExists                     // 10023
	ErrCodeRoleCodeExists                     // 10024
	ErrCodeOrgNotFound                        // 10025
	ErrCodeOrgCodeExists                      // 10026
	ErrCodeOrgHasChildren                     // 10027
	ErrCodeOrgParentNotFound                  // 10028
	ErrCodeOrgParentCycle                     // 10029
	ErrCodeLoginExpired                       // 10030
	ErrCodeUnauthorized                       // 10031
	ErrCodeUsernamePasswordError              // 10032
	ErrCodeAccountDisabled                    // 10033
	ErrCodeAccountLocked                      // 10034
	ErrCodeLoginFailed                        // 10035
	ErrCodeCaptchaEmpty                       // 10036
	ErrCodeCaptchaExpired                     // 10037
	ErrCodeCaptchaError                       // 10038
	ErrCodeNotFound                           // 10039
	ErrCodeMethodNotAllowed                   // 10040
)

// !!! IMPORTANT PLEASE READ FIRST !!!
// You SHOULD add new code at the end, and append comment of number
// Meanwhile, the corresponding error message SHOULD be appended in response.errors
// The order MUST be consistent between them
