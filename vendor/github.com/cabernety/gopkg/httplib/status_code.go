package httplib

const (
	STATUS_OK                  = 0
	STATUS_FAILED              = 1
	STATUS_FORM_VALIDATE_ERR   = 2
	STATUS_NOT_FOUND           = 4
	STATUS_INTERNAL_SERVER_ERR = 5
	STATUS_UNAUTHORIZED        = 6

	// 参数错误
	STATUS_PARAM_ERR = 7

	// user
	STATUS_USER_EXISTS  = 100
	STATUS_EMAIL_EXISTS = 101
	// 修改密码
	STATUS_PASSWORD_CONFIRM_FAILED  = 102 // 新密码与确认密码不一致
	STATUS_NEW_OLD_PASSWORD_SAME    = 103 // 新密码与原密码一致
	STATUS_OLD_PASSWORD_AUTH_FAILED = 104 // 原始密码验证失败

	// form validation
	STATUS_FIELD_REQUIRED     = 2000 // 字段必填
	STATUS_FIELD_REGEX_FAILED = 2001 // 字段格式不正确
)
