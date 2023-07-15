package client

import (
	"errors"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

func IsNotFound(err error) bool {
	return kerrors.IsNotFound(err) || errors.Is(err, kubefox.ErrResourceNotFound)
}

func IsConflict(err error) bool {
	return kerrors.IsAlreadyExists(err) || errors.Is(err, kubefox.ErrResourceConflict)
}
