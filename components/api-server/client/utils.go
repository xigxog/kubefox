package client

import (
	"errors"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

func IsNotFound(err error) bool {
	return kerrors.IsNotFound(err) || errors.Is(err, ErrResourceNotFound)
}

func IsConflict(err error) bool {
	return kerrors.IsAlreadyExists(err) || errors.Is(err, ErrResourceConflict)
}
