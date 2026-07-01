package biz

import (
	"github.com/go-kratos/kratos/v2/errors"
)

var (
	ErrCategoryNotFound    = errors.NotFound("CATEGORY_NOT_FOUND", "category not found")
	ErrCategoryHasChildren = errors.Forbidden("CATEGORY_HAS_CHILDREN", "category has children, cannot delete")
	ErrBrandNotFound       = errors.NotFound("BRAND_NOT_FOUND", "brand not found")
	ErrSPUNotFound         = errors.NotFound("SPU_NOT_FOUND", "spu not found")
	ErrSKUNotFound         = errors.NotFound("SKU_NOT_FOUND", "sku not found")
	ErrStockInsufficient   = errors.Forbidden("STOCK_INSUFFICIENT", "stock insufficient")
	ErrStockLockConflict   = errors.Forbidden("STOCK_LOCK_CONFLICT", "stock lock conflict, retry later")
	ErrStockOpDuplicated   = errors.Forbidden("STOCK_OPERATION_DUPLICATED", "stock operation already performed for this order")
	ErrSKUAttrInvalid      = errors.BadRequest("SKU_ATTR_INVALID", "sku attribute is invalid")
	ErrSPUStatusInvalid    = errors.Forbidden("STATUS_TRANSITION_INVALID", "spu status transition not allowed")
)
