package utils

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"time"

	creds "google.golang.org/grpc/credentials"
	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktyps "k8s.io/apimachinery/pkg/types"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
	k8scfg "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Certificate paths.
const (
	TLSCertFile     = "tls.crt"
	TLSKeyFile      = "tls.key"
	CACertFile      = "ca.crt"
	SvcAccTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

var (
	timeout = 15 * time.Second
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

	cp, err := GetCAFromSecret(key)
	if err != nil {
		return nil, err
	}

	return creds.NewClientTLSFromCert(cp, ""), nil
}

func GetSvcAccountToken(namespace, svcAccount string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if b, err := os.ReadFile(SvcAccTokenFile); err == nil {
		// Return token from file is it was successfully read.
		return string(b), nil
	}

	if namespace == "" {
		return "", fmt.Errorf("service account token not found at '%s'", SvcAccTokenFile)
	}

	client, err := k8sClient()
	if err != nil {
		return "", err
	}

	tr := &authv1.TokenRequest{}
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace, Name: svcAccount,
		},
	}
	if err := client.SubResource("token").Create(ctx, sa, tr); err != nil {
		return "", err
	}
	if tr.Status.Token == "" {
		return "", fmt.Errorf("no token was returned by kubernetes token request")
	}

	return tr.Status.Token, nil
}

func GetCAFromSecret(key ktyps.NamespacedName) (*x509.CertPool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := k8sClient()
	if err != nil {
		return nil, err
	}

	sec := &corev1.Secret{}
	if err = client.Get(ctx, key, sec); err != nil {
		return nil, err
	}

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(sec.Data["ca.crt"]) {
		err = fmt.Errorf("credentials: failed to append certificates")
		return nil, err
	}

	return cp, nil
}

func k8sClient() (k8s.Client, error) {
	cfg, err := k8scfg.GetConfig()
	if err != nil {
		return nil, err
	}

	return k8s.New(cfg, k8s.Options{})
}
