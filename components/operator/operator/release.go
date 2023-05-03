package operator

import (
	"fmt"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	tv1a1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	kubev1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/maker"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/component"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/platform"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (op *operator) ProcessRelease(kit kubefox.Kit) error {
	k := kit.Request().Kube()

	req := &Request[kubev1a1.Release]{}
	if err := k.Unmarshal(req); err != nil {
		return ErrEvent(kit, err)
	}
	kit.Log().Infof("processing %s hook for %s", k.GetHook(), req.GetObject())

	switch k.GetHook() {
	case kubefox.Customize:
		return CustomizeEvent(kit)

	case kubefox.Sync:
		attachments := []runtime.Object{}
		spec := req.GetObject().Spec
		sys, _ := uri.New(uri.Authority, uri.System, spec.System)
		env, _ := uri.New(uri.Authority, uri.Environment, spec.Environment)

		for _, c := range spec.Components {
			comp := component.New(component.Fields{
				App:     c.App,
				Name:    c.Name,
				GitHash: c.GitHash,
			})
			for _, r := range c.Routes {
				if r.Type == "http" {
					name := fmt.Sprintf("%s-%s-%s", sys.Name(), env.Name(), c.Key())
					mw := maker.New[tv1a1.Middleware](maker.Props{
						Group: "traefik.containo.us",
						Name:  name,
						// Organization: org,
						System: sys.Name(),
						// SystemRef:      spec.System,
						SystemId:    spec.SystemId,
						Environment: env.Name(),
						// EnvironmentRef: spec.Environment,
						EnvironmentId: spec.EnvironmentId,
						Component:     c.Name,
						CompHash:      c.ShortHash(),
					})
					mw.Spec = tv1a1.MiddlewareSpec{
						Headers: &dynamic.Headers{
							CustomRequestHeaders: map[string]string{
								kubefox.RelEnvHeader:    env.HTTPKey(),
								kubefox.RelSysHeader:    sys.HTTPKey(),
								kubefox.RelTargetHeader: comp.GetHTTPKey(),
							},
						},
					}

					ig := maker.New[tv1a1.IngressRoute](maker.Props{
						Group: "traefik.containo.us",
						Name:  name,
						// Organization: org,
						System: sys.Name(),
						// SystemRef:      spec.System,
						SystemId:    spec.SystemId,
						Environment: env.Name(),
						// EnvironmentRef: spec.Environment,
						EnvironmentId: spec.EnvironmentId,
						Component:     c.Name,
						CompHash:      c.ShortHash(),
					})
					ig.Spec = tv1a1.IngressRouteSpec{
						EntryPoints: []string{"websecure"},
						Routes: []tv1a1.Route{
							{
								Kind:        "Rule",
								Match:       r.Match,
								Middlewares: []tv1a1.MiddlewareRef{{Name: name}},
								Services: []tv1a1.Service{
									{
										LoadBalancerSpec: tv1a1.LoadBalancerSpec{
											Name: platform.BrkService,
											Port: intstr.FromInt(8080),
										},
									},
								},
							},
						},
					}

					attachments = append(attachments, mw, ig)
				}
			}
		}

		return SyncEvent(kit, nil, attachments...)

	default:
		return ErrEvent(kit, fmt.Errorf("unknown hook type: %s", k.GetHook()))
	}
}
