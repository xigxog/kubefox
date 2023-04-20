package platform

import (
	"github.com/xigxog/kubefox/libs/core/utils"
	creds "google.golang.org/grpc/credentials"
)

func NewGPRCSrvCreds(namespace string) (creds.TransportCredentials, error) {
	if c, err := creds.NewServerTLSFromFile(TLSCertFile, TLSKeyFile); err == nil {
		return c, nil
	} else if namespace == "" {
		return nil, err
	}

	cert, _, err := utils.GetCertFromSecret(namespace, CertSecret)
	if err != nil {
		return nil, err
	}

	return creds.NewServerTLSFromCert(&cert), nil
}

func NewGRPCClientCreds(namespace string) (creds.TransportCredentials, error) {
	if c, err := creds.NewClientTLSFromFile(CACertFile, ""); err == nil {
		return c, nil
	} else if namespace == "" {
		return nil, err
	}

	_, cp, err := utils.GetCertFromSecret(namespace, CertSecret)
	if err != nil {
		return nil, err
	}

	return creds.NewClientTLSFromCert(cp, ""), nil
}
