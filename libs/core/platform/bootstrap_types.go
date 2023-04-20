package platform

import (
	"crypto/tls"
)

type BootstrapResponse struct {
	Certificate tls.Certificate `json:"certificate"`
}
