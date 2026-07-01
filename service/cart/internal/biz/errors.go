package biz

import (
	"github.com/go-kratos/kratos/v2/errors"
)

var (
	ErrCartItemNotFound    = errors.NotFound("CART_ITEM_NOT_FOUND", "cart item not found")
	ErrCartQuantityInvalid = errors.BadRequest("CART_QUANTITY_INVALID", "cart quantity must be greater than 0")
)
