package platform

import "github.com/xigxog/kubefox/libs/core/component"

// Default TLS cert files
const (
	CertDir     = "/kubefox/runtime/tls"
	TLSCertFile = CertDir + "/tls.crt"
	TLSKeyFile  = CertDir + "/tls.key"
	CACertFile  = CertDir + "/ca.crt"
)

// K8s resource names
const (
	APISrvSvcAccount     = "kfp-api-server"
	BrokerSvcAccount     = "kfp-broker"
	RuntimeSrvSvcAccount = "kfp-runtime-server"
	OprtrSvcAccount      = "kfp-operator"

	NATSCertSecret  = "kfp-tls-cert-nats"
	CertSecret      = "kfp-tls-cert-runtime"
	EnvConfigMap    = "kfp-env"
	BrkService      = "kfp-broker"
	ImagePullSecret = "kfp-image-pull-secret"
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
	RuntimeSrvComp = component.New(component.Fields{
		App:     App,
		Name:    "runtime-server",
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
