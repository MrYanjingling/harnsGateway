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
)

// !!! IMPORTANT PLEASE READ FIRST !!!
// You SHOULD add new code at the end, and append comment of number
// Meanwhile, the corresponding error message SHOULD be appended in response.errors
// The order MUST be consistent between them
