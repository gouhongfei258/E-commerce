package biz

import (
	"github.com/go-kratos/kratos/v2/errors"
)

var (
	ErrUserNotFound      = errors.NotFound("USER_NOT_FOUND", "user not found")
	ErrUserDuplicate     = errors.Forbidden("USER_DUPLICATE", "username already exists")
	ErrUserPasswordWrong = errors.Forbidden("USER_PASSWORD_WRONG", "invalid password")
	ErrAddressNotFound   = errors.NotFound("ADDRESS_NOT_FOUND", "address not found")
	ErrAddressLimit      = errors.Forbidden("ADDRESS_LIMIT", "maximum 10 addresses per user")
)
