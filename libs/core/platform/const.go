package platform

import (
	"path"
	"time"

	"github.com/xigxog/kubefox/libs/core/component"
)

// Certificate paths
const (
	RootDir  = "/kubefox"
	CertsDir = "certs"

	TLSCertFile = "tls.crt"
	TLSKeyFile  = "tls.key"
	CACertFile  = "ca.crt"
)

var (
	CACertPath       = path.Join(RootDir, CertsDir, CACertFile)
	BrokerCertsDir   = path.Join(RootDir, CertsDir, "broker")
	OperatorCertsDir = path.Join(RootDir, CertsDir, "operator")
)

var (
	// long timeout as we need to wait for Vault
	StartupTimeout = 5 * time.Minute
	GRPCServiceCfg = `{
		"methodConfig": [{
		  "name": [{"service": "", "method": ""}],
		  "waitForReady": true,
		  "retryPolicy": {
			  "MaxAttempts": 4,
			  "InitialBackoff": "1s",
			  "MaxBackoff": "60s",
			  "BackoffMultiplier": 15.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE" ]
		  }
		}]}`
)

// K8s resource names
const (
	APISrvSvcAccount   = "api-server"
	BrokerSvcAccount   = "broker"
	OperatorSvcAccount = "operator"

	RootCASecret    = "root-ca"
	UnsealKeySecret = "unseal-key"

	BrokerService   = "broker"
	ImagePullSecret = "image-pull-secret"
)

// Arg names
const (
	SvcAccountTokenArg = "svc-account-token"
	TargetArg          = "target"
)

const (
	System  = "kubefox"
	Env     = "kubefox"
	App     = "platform"
	GitHash = "0000000"
)

var (
	APISrvComp = component.New(component.Fields{
		App:     App,
		Name:    "api-server",
		GitHash: GitHash,
	})
	OperatorComp = component.New(component.Fields{
		App:     App,
		Name:    "operator",
		GitHash: GitHash,
	})
	HTTPIngressAdapt = component.New(component.Fields{
		App:     App,
		Name:    "http-ingress-adapter",
		GitHash: GitHash,
	})
	K8sAdapt = component.New(component.Fields{
		App:     App,
		Name:    "kubernetes-adapter",
		GitHash: GitHash,
	})
)
