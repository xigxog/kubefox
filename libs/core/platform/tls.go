package platform

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/xigxog/kubefox/libs/core/utils"
	creds "google.golang.org/grpc/credentials"
	ktyps "k8s.io/apimachinery/pkg/types"
)

func NewGPRCSrvCreds(ctx context.Context, dir string) (creds.TransportCredentials, error) {
	// wait for cert file to exists
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if _, err := os.Stat(path.Join(dir, TLSCertFile)); err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	return creds.NewServerTLSFromFile(path.Join(dir, TLSCertFile), path.Join(dir, TLSKeyFile))
}

func NewGRPCClientCreds(caCertFile string, key ktyps.NamespacedName) (creds.TransportCredentials, error) {
	if c, err := creds.NewClientTLSFromFile(caCertFile, ""); err == nil {
		return c, nil
	}

	cp, err := utils.GetCAFromSecret(key)
	if err != nil {
		return nil, err
	}

	return creds.NewClientTLSFromCert(cp, ""), nil
}
