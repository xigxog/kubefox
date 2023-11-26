package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	vapi "github.com/hashicorp/vault/api"
	vauth "github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"
)

const (
	TenYears = "87600h"
)

type Flags struct {
	platformVaultName string
	compVaultName     string
	compSvcName       string
	compIP            string
	vaultURL          string
	logFormat         string
	logLevel          string
}

var (
	f Flags
)

func main() {
	flag.StringVar(&f.platformVaultName, "platform-vault-name", "", "Vault name of Platform. (required)")
	flag.StringVar(&f.compVaultName, "component-vault-name", "", "Vault name of Component to bootstrap. (required)")
	flag.StringVar(&f.compSvcName, "component-service-name", "", "Service name of Component to bootstrap. (required)")
	flag.StringVar(&f.compIP, "component-ip", "", "IP address of Component to bootstrap. (required)")
	flag.StringVar(&f.vaultURL, "vault-url", "", "URL of Vault server. (required)")
	flag.StringVar(&f.logFormat, "log-format", "console", `Log format; one of ["json", "console"].`)
	flag.StringVar(&f.logLevel, "log-level", "debug", `Log level; one of ["debug", "info", "warn", "error"].`)
	flag.Parse()

	utils.CheckRequiredFlag("platform-vault-name", f.platformVaultName)
	utils.CheckRequiredFlag("component-vault-name", f.compVaultName)
	utils.CheckRequiredFlag("component-service-name", f.compSvcName)
	utils.CheckRequiredFlag("component-ip", f.compIP)
	utils.CheckRequiredFlag("vault-url", f.vaultURL)

	logkf.Global = logkf.
		BuildLoggerOrDie(f.logFormat, f.logLevel).
		WithPlatformComponent(api.PlatformComponentBootstrap)
	defer logkf.Global.Sync()
	log := logkf.Global

	log.DebugInterface("flags:", f)

	for retry := 0; retry < 3; retry++ {
		log.Infof("generating certificates for component %s, attempt %d of 3", f.compVaultName, retry+1)
		if err := generateCerts(log); err != nil {
			log.Errorf("error generating certificates: %v", err)
			time.Sleep(time.Second * time.Duration(rand.Intn(2)+1))

		} else {
			return
		}
	}

	log.Fatal("unable to bootstrap component")
}

func generateCerts(log *logkf.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, err := vaultClient(ctx)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("pki/int/platform/%s/issue/%s", f.platformVaultName, f.compVaultName)

	s, err := c.Logical().WriteWithContext(ctx, path, map[string]interface{}{
		"common_name": fmt.Sprintf("%s.svc", f.compSvcName),
		"alt_names":   fmt.Sprintf("%s@%s,%s,localhost", f.compVaultName, f.compSvcName, f.compSvcName),
		"ip_sans":     fmt.Sprintf("%s,127.0.0.1", f.compIP),
		"ttl":         TenYears,
	})
	if err != nil {
		return err
	}

	cert := fmt.Sprintf("%s\n%s", s.Data["certificate"], s.Data["issuing_ca"])
	key := fmt.Sprintf("%s", s.Data["private_key"])

	if err := os.WriteFile(api.PathTLSCert, []byte(cert), 0600); err != nil {
		return err
	}
	log.Infof("wrote tls certificate to %s", api.PathTLSCert)

	if err := os.WriteFile(api.PathTLSKey, []byte(key), 0600); err != nil {
		return err
	}
	log.Infof("wrote tls private key to %s", api.PathTLSKey)

	return nil
}

func vaultClient(ctx context.Context) (*vapi.Client, error) {
	cfg := vapi.DefaultConfig()
	cfg.Address = f.vaultURL
	cfg.MaxRetries = 3
	cfg.HttpClient.Timeout = time.Second * 15
	cfg.ConfigureTLS(&vapi.TLSConfig{
		CACert: api.PathCACert,
	})

	vault, err := vapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(api.PathSvcAccToken)
	if err != nil {
		return nil, err
	}
	token := vauth.WithServiceAccountToken(string(b))
	auth, err := vauth.NewKubernetesAuth(f.compVaultName, token)
	if err != nil {
		return nil, err
	}
	authInfo, err := vault.Auth().Login(ctx, auth)
	if err != nil {
		return nil, err
	}
	if authInfo == nil {
		return nil, fmt.Errorf("error logging in with kubernetes auth: no auth info was returned")
	}

	return vault, nil
}
