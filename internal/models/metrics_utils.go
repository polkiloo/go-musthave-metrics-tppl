package models

import "errors"

var (
	ErrInvalidMetricType = errors.New("invalid metric type")
	ErrMissingValue      = errors.New("missing value")
	ErrBothValueAndDelta = errors.New("both value and delta")
	ErrTypeMismatch      = errors.New("type mismatch")
)
