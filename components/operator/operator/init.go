package operator

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/xigxog/kubefox/components/operator/templates"
	"github.com/xigxog/kubefox/libs/core/api/common"
	kfv1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/platform"
	"github.com/xigxog/kubefox/libs/core/vault"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ktyps "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	k8scli "sigs.k8s.io/controller-runtime/pkg/client"
	k8scfg "sigs.k8s.io/controller-runtime/pkg/client/config"
)

func New(ctx kubefox.KitContext) (*operator, error) {
	// init kubernetes client
	if err := kfv1a1.SchemeBuilder.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	k8sCfg, err := k8scfg.GetConfig()
	if err != nil {
		return nil, err
	}
	k8sCli, err := k8scli.New(k8sCfg, k8scli.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, err
	}

	op := &operator{
		k8sCli: k8sCli,
	}

	return op, op.init(ctx)
}

func (op *operator) init(ctx kubefox.KitContext) error {
	var err error
	pName := ctx.Platform()
	pNS := ctx.PlatformNamespace()

	if _, err := op.applyTemplate(ctx, "crds", nil); err != nil {
		return err
	}

	p := &kfv1a1.Platform{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "k8s.kubefox.io/v1alpha1",
			Kind:       "Platform",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pName,
			Namespace: pNS,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "kubefox",
				"app.kubernetes.io/instance":   pName,
				"app.kubernetes.io/version":    GitRef,
				"app.kubernetes.io/managed-by": pName + "-operator",
			},
		},
		Spec: common.PlatformSpec{
			Systems: map[uri.Key]*common.PlatformSystem{},
		},
	}
	if err := op.apply(ctx, p, true); err != nil {
		return err
	}

	data := &templates.Data{
		DevMode: true,
		Platform: templates.Platform{
			Name:    pName,
			Version: GitRef,
		},
		System: templates.System{
			Name:              platform.System,
			GitRef:            GitRef,
			GitHash:           GitHash,
			Namespace:         pNS,
			ContainerRegistry: ContainerRegistry,
		},
		Owner: &metav1.OwnerReference{
			APIVersion: p.APIVersion,
			Kind:       p.Kind,
			UID:        p.UID,
			Name:       p.Name,
		},
	}

	data.Component = templates.Component{
		Name:  "vault",
		Image: "ghcr.io/xigxog/vault:1.13.3-v0.0.1",
	}
	if _, err := op.applyTemplate(ctx, "vault", data); err != nil {
		return err
	}

	key := ktyps.NamespacedName{Namespace: pNS, Name: pName + "-vault"}
	vaultSS := &appsv1.StatefulSet{}
	// TODO turn this into a watch
	for {
		// ctx will eventually timeout ensuring this is not infinite
		if err := op.k8sCli.Get(ctx, key, vaultSS); err != nil {
			return err
		}
		if vaultSS.Status.ReadyReplicas > 0 {
			break
		}
		ctx.Log().Debug("Vault is not ready, waiting...")
		time.Sleep(10 * time.Second)
	}

	key = ktyps.NamespacedName{Namespace: pNS, Name: fmt.Sprintf("%s-%s", pName, platform.RootCASecret)}
	caCertSec := &corev1.Secret{}
	if err := op.k8sCli.Get(ctx, key, caCertSec); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(ctx.CACertPath()), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(ctx.CACertPath(), caCertSec.Data["ca.crt"], 0600); err != nil {
		return err
	}

	vaultURL := fmt.Sprintf("https://%s-vault:8200", pName)
	op.vaultCli, err = vault.NewClient(ctx, vault.OperatorRole, vaultURL, caCertSec.Data["ca.crt"])
	if err != nil {
		return err
	}

	opCertSec, err := op.vaultCli.Logical().WriteWithContext(ctx, "pki_int/issue/operator", map[string]interface{}{
		"common_name": fmt.Sprintf("%s-operator.%s", pName, pNS),
		"alt_names":   fmt.Sprintf("%s-operator,localhost", pName),
		"ip_sans":     "127.0.0.1",
		"ttl":         tenYears,
	})
	if err != nil {
		return err
	}
	if err := writeCert(opCertSec, platform.OperatorCertsDir); err != nil {
		return err
	}

	brkCertSec, err := op.vaultCli.Logical().WriteWithContext(ctx, "pki_int/issue/broker", map[string]interface{}{
		"common_name": "localhost",
		"ip_sans":     "127.0.0.1",
		"ttl":         tenYears,
	})
	if err != nil {
		return err
	}
	if err := writeCert(brkCertSec, platform.BrokerCertsDir); err != nil {
		return err
	}

	data.Component = templates.Component{
		Name:  "metacontroller",
		Image: "metacontrollerio/metacontroller:v4.10.3",
	}
	if _, err := op.applyTemplate(ctx, "metacontroller", data); err != nil {
		return err
	}

	return nil
}

func (op *operator) applyTemplate(ctx kubefox.KitContext, name string, data *templates.Data) ([]*unstructured.Unstructured, error) {
	objs, err := templates.Render(name, data)
	if err != nil {
		return nil, err
	}

	for _, obj := range objs {
		if err := op.apply(ctx, obj, true); err != nil {
			return objs, err
		}
	}

	return objs, nil
}

func (op *operator) apply(ctx kubefox.KitContext, obj k8scli.Object, retry bool) error {
	s := fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
	ctx.Log().Debugf("applying resource %s", s)

	if err := op.k8sCli.Patch(ctx, obj, k8scli.Apply, fieldOwner, k8scli.ForceOwnership); err != nil {
		if retry {
			ctx.Log().Warnf("error applying resource %s: %v, retrying...", s, err)
			time.Sleep(3 * time.Second)
			return op.apply(ctx, obj, false)
		} else {
			return err
		}
	}

	return nil
}

func writeCert(sec *vaultapi.Secret, dir string) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	tlsKey := fmt.Sprintf("%s", sec.Data["private_key"])
	if err := os.WriteFile(path.Join(dir, "tls.key"), []byte(tlsKey), 0600); err != nil {
		return err
	}
	tlsCrt := fmt.Sprintf("%s\n%s", sec.Data["certificate"], sec.Data["issuing_ca"])
	if err := os.WriteFile(path.Join(dir, "tls.crt"), []byte(tlsCrt), 0600); err != nil {
		return err
	}

	return nil
}
