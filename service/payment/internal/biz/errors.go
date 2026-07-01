package biz

import (
	"github.com/go-kratos/kratos/v2/errors"
)

var (
	ErrPaymentNotFound        = errors.NotFound("PAYMENT_NOT_FOUND", "payment not found")
	ErrPaymentAlreadyProcessed = errors.Forbidden("PAYMENT_ALREADY_PROCESSED", "payment has already been processed")
	ErrPaymentStatusInvalid   = errors.Forbidden("PAYMENT_STATUS_INVALID", "payment status transition not allowed")
	ErrPaymentAmountMismatch  = errors.BadRequest("PAYMENT_AMOUNT_MISMATCH", "payment amount does not match order amount")
	ErrPaymentProviderFail    = errors.InternalServer("PAYMENT_PROVIDER_FAIL", "payment provider processing failed")
	ErrPaymentInvalidMethod   = errors.BadRequest("PAYMENT_INVALID_METHOD", "unsupported payment method")
)
