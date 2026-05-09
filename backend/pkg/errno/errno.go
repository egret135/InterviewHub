package errno

import "errors"

var (
	ErrNotFound    = errors.New("NOT_FOUND")
	ErrInternal    = errors.New("INTERNAL")
	ErrBadRequest  = errors.New("BAD_REQUEST")
)
