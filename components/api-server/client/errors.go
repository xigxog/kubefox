package client

import "errors"

var (
	ErrResourceConflict  = errors.New("resource already exists")
	ErrResourceNotFound  = errors.New("resource not found")
	ErrResourceForbidden = errors.New("resource forbidden")
	ErrBadRequest        = errors.New("bad request")
)
