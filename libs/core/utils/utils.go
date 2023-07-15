package utils

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktyps "k8s.io/apimachinery/pkg/types"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
	k8scfg "sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	svcAccTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

var (
	mutex     sync.Mutex
	k8sClient k8s.Client
	timeout   = 10 * time.Second
)

type Comparable[T any] interface {
	Equals(T) bool
}

func Contains[T Comparable[T]](s []T, e T) bool {
	for _, v := range s {
		if v.Equals(e) {
			return true
		}
	}
	return false
}

func ResolveFlag(curr, envVar, def string) string {
	if curr != "" {
		return curr
	}

	if e := os.Getenv(envVar); e != "" {
		return e
	} else {
		return def
	}
}

func ResolveFlagBool(curr bool, envVar string, def bool) bool {
	if curr != def {
		return curr
	}

	if e, err := strconv.ParseBool(os.Getenv(envVar)); err == nil {
		return e
	} else {
		return def
	}
}

// GetParamOrHeader looks for query parameters and headers for the provided
// keys. Keys are checked in order. Query parameters take precedence over
// headers.
func GetParamOrHeader(httpReq *http.Request, keys ...string) string {
	for _, key := range keys {
		val := httpReq.URL.Query().Get(strings.ToLower(key))
		if val == "" {
			val = httpReq.Header.Get(key)
		}
		if val != "" {
			return val
		}
	}

	return ""
}

// SystemNamespace returns the name of the Kubernetes Namespace that contains
// all System objects. The format is 'kfs-{Instance}-{System}'. The 'kfs' prefix
// stands for 'KubeFox System'.
//
// If any arg is empty an empty string is returned.
func SystemNamespace(platform, system string) string {
	if platform == "" || system == "" {
		return ""
	}

	return fmt.Sprintf("kfs-%s-%s", platform, system)
}

func GetSvcAccountToken(namespace, svcAccount string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if b, err := os.ReadFile(svcAccTokenFile); err == nil {
		return string(b), nil
	}

	client, err := getK8sClient()
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

	client, err := getK8sClient()
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

func getK8sClient() (k8s.Client, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if k8sClient != nil {
		return k8sClient, nil
	}

	cfg, err := k8scfg.GetConfig()
	if err != nil {
		return nil, err
	}

	k8sClient, err = k8s.New(cfg, k8s.Options{})
	if err != nil {
		return nil, err
	}

	return k8sClient, nil
}
