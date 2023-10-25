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
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"
)

const (
	TenYears = "87600h"
)

type Flags struct {
	instance    string
	platform    string
	vaultName   string
	component   string
	componentIP string
	namespace   string
	vaultAddr   string
	vaultRole   string
	logFormat   string
	logLevel    string
}

var (
	f Flags
)

func main() {
	flag.StringVar(&f.instance, "instance", "", "Name of KubeFox instance. (required)")
	flag.StringVar(&f.platform, "platform", "", "Name of Platform. (required)")
	flag.StringVar(&f.component, "component", "", "Name of Component to bootstrap. (required)")
	flag.StringVar(&f.componentIP, "component-ip", "", "IP address of Component to bootstrap. (required)")
	flag.StringVar(&f.namespace, "namespace", "", "Namespace of Platform. (required)")
	flag.StringVar(&f.vaultName, "vault-name", "", "Name of Platform in Vault. (required)")
	flag.StringVar(&f.vaultRole, "vault-role", "", "Name of Vault role to authenticate with. (required)")
	flag.StringVar(&f.vaultAddr, "vault-addr", "127.0.0.1:8200", "Address and port of Vault server.")
	flag.StringVar(&f.logFormat, "log-format", "console", `Log format; one of ["json", "console"].`)
	flag.StringVar(&f.logLevel, "log-level", "debug", `Log level; one of ["debug", "info", "warn", "error"].`)
	flag.Parse()

	utils.CheckRequiredFlag("instance", f.instance)
	utils.CheckRequiredFlag("platform", f.platform)
	utils.CheckRequiredFlag("component", f.component)
	utils.CheckRequiredFlag("component-ip", f.componentIP)
	utils.CheckRequiredFlag("namespace", f.namespace)
	utils.CheckRequiredFlag("vault-name", f.vaultName)
	utils.CheckRequiredFlag("vault-role", f.vaultRole)

	logkf.Global = logkf.
		BuildLoggerOrDie(f.logFormat, f.logLevel).
		WithInstance(f.instance).
		WithPlatform(f.platform).
		WithService("bootstrap")
	defer logkf.Global.Sync()
	log := logkf.Global

	log.Infof("gitCommit: %s, gitRef: %s", kubefox.GitCommit, kubefox.GitRef)
	log.DebugInterface("flags:", f)

	for retry := 0; retry < 3; retry++ {
		log.Infof("generating certificates using vault role %s, attempt %d of 3", f.vaultRole, retry+1)
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

	cname := fmt.Sprintf("%s-%s.%s", f.platform, f.component, f.namespace)
	path := fmt.Sprintf("pki/int/platform/%s/issue/%s", f.vaultName, f.component)

	s, err := c.Logical().WriteWithContext(ctx, path, map[string]interface{}{
		"common_name": cname,
		"alt_names":   fmt.Sprintf("%s@%s,localhost", f.component, cname),
		"ip_sans":     fmt.Sprintf("%s,127.0.0.1", f.componentIP),
		"ttl":         TenYears,
	})
	if err != nil {
		return err
	}

	cert := fmt.Sprintf("%s\n%s", s.Data["certificate"], s.Data["issuing_ca"])
	key := fmt.Sprintf("%s", s.Data["private_key"])

	if err := os.WriteFile(kubefox.PathTLSCert, []byte(cert), 0600); err != nil {
		return err
	}
	log.Infof("wrote tls certificate to %s", kubefox.PathTLSCert)

	if err := os.WriteFile(kubefox.PathTLSKey, []byte(key), 0600); err != nil {
		return err
	}
	log.Infof("wrote tls private key to %s", kubefox.PathTLSKey)

	return nil
}

func vaultClient(ctx context.Context) (*vapi.Client, error) {
	cfg := vapi.DefaultConfig()
	cfg.Address = fmt.Sprintf("https://%s", f.vaultAddr)
	cfg.MaxRetries = 3
	cfg.HttpClient.Timeout = time.Second * 15
	cfg.ConfigureTLS(&vapi.TLSConfig{
		CACert: kubefox.PathCACert,
	})

	vault, err := vapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(kubefox.PathSvcAccToken)
	if err != nil {
		return nil, err
	}
	token := vauth.WithServiceAccountToken(string(b))
	auth, err := vauth.NewKubernetesAuth(f.vaultRole, token)
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
