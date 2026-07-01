package biz

import (
	"github.com/go-kratos/kratos/v2/errors"
)

var (
	ErrOrderNotFound       = errors.NotFound("ORDER_NOT_FOUND", "order not found")
	ErrOrderStatusInvalid  = errors.Forbidden("ORDER_STATUS_INVALID", "order status transition not allowed")
	ErrOrderCannotCancel   = errors.Forbidden("ORDER_CANNOT_CANCEL", "order cannot be cancelled in current status")
	ErrOrderItemEmpty      = errors.BadRequest("ORDER_ITEM_EMPTY", "order must contain at least one item")
	ErrOrderAddressRequired = errors.BadRequest("ORDER_ADDRESS_REQUIRED", "shipping address is required")
	ErrOrderUnauthorized   = errors.Forbidden("ORDER_UNAUTHORIZED", "order does not belong to the current user")
	ErrStockInsufficient   = errors.Forbidden("STOCK_INSUFFICIENT", "stock insufficient")
)
