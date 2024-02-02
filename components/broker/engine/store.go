// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/cache"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/matcher"
	"github.com/xigxog/kubefox/vault"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type Store struct {
	resCache ctrlcache.Cache
	vaultCli *vault.Client

	validationCache cache.Cache[api.Problems]
	depMatcherCache cache.Cache[*matcher.EventMatcher]
	secretsCache    cache.Cache[*api.Data]

	compCache      map[string]*api.ComponentDefinition
	releaseMatcher *matcher.EventMatcher

	ctx    context.Context
	cancel context.CancelFunc

	log *logkf.Logger
}

func NewStore() *Store {
	ctx, cancel := context.WithCancel(context.Background())
	return &Store{
		validationCache: cache.New[api.Problems](time.Minute * 15),
		depMatcherCache: cache.New[*matcher.EventMatcher](time.Minute * 15),
		secretsCache:    cache.New[*api.Data](time.Minute * 15),
		ctx:             ctx,
		cancel:          cancel,
		log:             logkf.Global,
	}
}

func (str *Store) Open() error {
	ctx, cancel := context.WithTimeout(str.ctx, time.Minute*3)
	defer cancel()

	if err := str.init(ctx); err != nil {
		return err
	}

	if err := str.updateComponentCache(ctx); err != nil {
		return err
	}
	if err := str.updateReleaseMatcher(ctx); err != nil {
		return err
	}

	return nil
}

func (str *Store) init(initCtx context.Context) error {
	key := vault.Key{
		Instance:  config.Instance,
		Namespace: config.Namespace,
		Component: api.PlatformComponentBroker,
	}

	vaultCli, err := vault.New(vault.ClientOptions{
		Instance: config.Instance,
		Role:     vault.RoleName(key),
		URL:      config.VaultURL,
		CACert:   api.PathCACert,
	})
	if err != nil {
		return err
	}
	str.vaultCli = vaultCli

	if err := v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return str.log.ErrorN("adding KubeFox CRs to scheme failed: %v", err)
	}

	cfg, err := ctrl.GetConfig()
	if err != nil {
		return str.log.ErrorN("loading K8s config failed: %v", err)
	}

	str.resCache, err = ctrlcache.New(cfg, ctrlcache.Options{
		Scheme:                      scheme.Scheme,
		DefaultNamespaces:           map[string]ctrlcache.Config{config.Namespace: {}},
		ReaderFailOnMissingInformer: true,
	})
	if err != nil {
		return str.log.ErrorN("creating resource cache failed: %v", err)
	}

	err = str.initInformers(initCtx,
		&v1alpha1.Environment{},
		&v1alpha1.VirtualEnvironment{},
		&v1alpha1.ReleaseManifest{},
		&v1alpha1.HTTPAdapter{},
		&v1alpha1.AppDeployment{},
		&v1alpha1.Platform{},
	)
	if err != nil {
		return err
	}

	go func() {
		if err := str.resCache.Start(str.ctx); err != nil {
			str.log.Error(err)
		}
	}()
	str.resCache.WaitForCacheSync(initCtx)

	return nil
}

func (str *Store) initInformers(ctx context.Context, objs ...client.Object) error {
	for _, obj := range objs {
		// Getting the informer adds it to the cache.
		if inf, err := str.resCache.GetInformer(ctx, obj); err != nil {
			return err
		} else {
			inf.AddEventHandler(str)
		}
	}
	return nil
}

func (str *Store) Close() {
	str.cancel()
}

func (str *Store) ComponentDef(ctx context.Context, comp *core.Component) (*api.ComponentDefinition, error) {
	if str.compCache == nil {
		if err := str.updateComponentCache(ctx); err != nil {
			return nil, err
		}
	}

	def, found := str.compCache[comp.GroupKey()]
	if !found {
		return nil, core.ErrNotFound()
	}

	return def, nil
}

func (str *Store) IsGenesisAdapter(ctx context.Context, comp *core.Component) bool {
	r, err := str.ComponentDef(ctx, comp)
	if err != nil {
		return false
	}

	switch r.Type {
	case api.ComponentTypeHTTPAdapter:
		return true
	}

	return false
}

func (str *Store) Platform(ctx context.Context) (*v1alpha1.Platform, error) {
	obj := &v1alpha1.Platform{}
	return obj, str.resCache.Get(ctx, str.key(config.Platform), obj)
}

func (str *Store) AppDeployment(ctx context.Context, name string) (*v1alpha1.AppDeployment, error) {
	obj := &v1alpha1.AppDeployment{}
	return obj, str.resCache.Get(ctx, str.key(name), obj)
}

func (str *Store) ReleaseMatcher(ctx context.Context) (*matcher.EventMatcher, error) {
	if str.releaseMatcher != nil {
		return str.releaseMatcher, nil
	}

	// There are no matchers in cache, perform full reload.
	if err := str.updateReleaseMatcher(ctx); err != nil {
		return nil, err
	}

	return str.releaseMatcher, nil
}

func (str *Store) DeploymentMatcher(ctx *BrokerEventContext) (*matcher.EventMatcher, error) {
	// Check cache.
	if depM, found := str.depMatcherCache.Get(ctx.Key); found {
		str.log.Debugf("found cached matcher for event context key '%s'", ctx.Key)
		return depM, nil
	}

	routes, err := str.buildRoutes(ctx)
	if err != nil {
		return nil, err
	}

	depM := matcher.New()
	if err := depM.AddRoutes(routes...); err != nil {
		return nil, err
	}
	str.depMatcherCache.Set(ctx.Key, depM)

	return depM, nil
}

// AttachEventContext gets the VirtualEnvironment, AppDeployment, and if
// applicable, the ReleaseManifest of the Event Context and attaches them to the
// BrokerEvent. Secrets are retrieved from Vault and included in the Data
// member. The Event Context is then validated, problems found during validation
// are returned.
func (str *Store) AttachEventContext(ctx *BrokerEventContext) error {
	var (
		ve       = &v1alpha1.VirtualEnvironment{}
		appDep   = &v1alpha1.AppDeployment{}
		manifest *v1alpha1.ReleaseManifest
		data     *api.Data
		key      string
	)

	if err := str.resCache.Get(ctx, str.key(ctx.Event.Context.VirtualEnvironment), ve); err != nil {
		return err
	}
	if err := str.resCache.Get(ctx, str.key(ctx.Event.Context.AppDeployment), appDep); err != nil {
		return err
	}

	var dataProviderUID types.UID
	switch {
	case ctx.Event.Context.ReleaseManifest != "":
		manifest = &v1alpha1.ReleaseManifest{}
		if err := str.resCache.Get(ctx, str.key(ctx.Event.Context.ReleaseManifest), manifest); err != nil {
			return err
		}

		dataProviderUID = manifest.UID
		data = manifest.Data.DeepCopy()

		if err := str.mergeSecrets(ctx, manifest, data); err != nil {
			return err
		}

	default:
		env := &v1alpha1.Environment{}
		if err := str.resCache.Get(ctx, k8s.Key("", ve.Spec.Environment), env); err != nil {
			return err
		}

		dataProviderUID = ve.UID
		data = ve.Data.DeepCopy()
		data.Import(&env.Data)

		if err := str.mergeSecrets(ctx, env, data); err != nil {
			return err
		}
		if err := str.mergeSecrets(ctx, ve, data); err != nil {
			return err
		}

	}

	dataHash, _ := hashstructure.Hash(data, hashstructure.FormatV2, nil)
	key = fmt.Sprintf("%s-%d_%s-%s",
		appDep.UID, appDep.Generation, dataProviderUID, fmt.Sprint(dataHash))

	ctx.VirtualEnv = ve
	ctx.AppDeployment = appDep
	ctx.ReleaseManifest = manifest
	ctx.Data = data
	ctx.Key = key

	var problems, cached api.Problems
	// Only check cache if context has ReleaseManifest.
	if manifest != nil {
		str.log.Debugf("using cached problems for '%s'", key)
		cached, _ = str.validationCache.Get(key)
		problems = cached
	}
	if problems == nil {
		p, err := appDep.Validate(data,
			func(name string, typ api.ComponentType) (common.Adapter, error) {
				return str.Adapter(ctx, name, typ)
			})
		if err != nil {
			return err
		}
		problems = p
	}
	// Only update cache if context has ReleaseManifest.
	if manifest != nil && cached == nil {
		str.validationCache.Set(key, problems)
	}

	if len(problems) > 0 {
		b, _ := yaml.Marshal(problems)
		return core.ErrInvalid(fmt.Errorf("event context is invalid\n%s", b))
	}

	return nil
}

func (str *Store) Adapter(ctx *BrokerEventContext, name string, typ api.ComponentType) (common.Adapter, error) {
	if ctx.ReleaseManifest != nil {
		return ctx.ReleaseManifest.GetAdapter(name, typ)
	}

	var a common.Adapter
	switch typ {
	case api.ComponentTypeHTTPAdapter:
		a = &v1alpha1.HTTPAdapter{}

	default:
		return nil, core.ErrNotFound()
	}

	return a, str.resCache.Get(ctx, str.key(name), a)
}

func (str *Store) mergeSecrets(ctx context.Context, d api.DataProvider, data *api.Data) error {
	key := fmt.Sprintf("%s-%s", d.GetDataKey(), d.GetResourceVersion())
	if secs, _ := str.secretsCache.Get(key); secs != nil {
		str.log.Debugf("using cached secrets for '%s'", key)
		data.Merge(secs)
		return nil
	}

	secs := &api.Data{}
	if err := str.vaultCli.GetData(ctx, d.GetDataKey(), secs); k8s.IgnoreNotFound(err) != nil {
		return err
	}
	str.secretsCache.Set(key, secs)
	data.Merge(secs)

	return nil
}

func (str *Store) OnAdd(obj interface{}, isInInitialList bool) {
	if isInInitialList {
		str.log.Debugf("%T initialized", obj)
		return
	}
	str.onChange(obj, "added")
}

func (str *Store) OnUpdate(oldObj, obj interface{}) {
	str.onChange(obj, "updated")
}

func (str *Store) OnDelete(obj interface{}) {
	str.onChange(obj, "deleted")
}

func (str *Store) onChange(obj interface{}, op string) {
	str.log.Debugf("%T %s", obj, op)

	updateComps := false
	switch obj.(type) {
	case *v1alpha1.AppDeployment, *v1alpha1.Platform:
		updateComps = true
	}

	go func() {
		ctx, cancel := context.WithTimeout(str.ctx, time.Minute)
		defer cancel()

		if updateComps {
			str.updateComponentCache(ctx)
		}
		str.updateReleaseMatcher(ctx)
	}()
}

func (str *Store) updateComponentCache(ctx context.Context) error {
	compCache := map[string]*api.ComponentDefinition{}

	appDepList := &v1alpha1.AppDeploymentList{}
	if err := str.resCache.List(ctx, appDepList); err != nil {
		// Force rebuild on next used.
		str.compCache = nil
		return err
	}

	for _, appDep := range appDepList.Items {
		for compName, compSpec := range appDep.Spec.Components {
			comp := core.NewComponent(
				compSpec.Type,
				appDep.Spec.AppName,
				compName,
				compSpec.Commit,
			)
			compCache[comp.GroupKey()] = compSpec
		}
	}

	p := &v1alpha1.Platform{}
	if err := str.resCache.Get(ctx, str.key(config.Platform), p); err != nil {
		// Force rebuild on next used.
		str.compCache = nil
		return err
	}

	for _, c := range p.Status.Components {
		if c.Name == api.PlatformComponentHTTPSrv {
			comp := core.NewPlatformComponent(c.Type, c.Name, c.Commit)
			compCache[comp.GroupKey()] = &api.ComponentDefinition{
				Type: c.Type,
			}
		}
	}

	str.compCache = compCache
	return nil
}

func (str *Store) updateReleaseMatcher(ctx context.Context) error {
	veList := &v1alpha1.VirtualEnvironmentList{}
	if err := str.resCache.List(ctx, veList); err != nil {
		// Force rebuild on next use.
		str.releaseMatcher = nil
		return err
	}

	relM := matcher.New()
	for _, ve := range veList.Items {
		release := ve.Status.ActiveRelease
		if release == nil {
			str.log.Debugf("VirtualEnvironment '%s' does not have an active Release", ve.Name)
			continue
		}

		var data *api.Data
		switch {
		case release.ReleaseManifest != "":
			manifest := &v1alpha1.ReleaseManifest{}
			if err := str.resCache.Get(ctx, str.key(release.ReleaseManifest), manifest); err != nil {
				str.log.Warn(err)
				continue
			}
			data = &manifest.Data

		default:
			env := &v1alpha1.Environment{}
			if err := str.resCache.Get(ctx, k8s.Key("", ve.Spec.Environment), env); err != nil {
				str.log.Warn(err)
				continue
			}
			ve.Data.Import(&env.Data)
			data = &ve.Data
		}

		for appName, app := range release.Apps {
			appDep := &v1alpha1.AppDeployment{}
			if err := str.resCache.Get(ctx, str.key(app.AppDeployment), appDep); err != nil {
				str.log.Warn(err)
				continue
			}

			avail := k8s.Condition(appDep.Status.Conditions, api.ConditionTypeAvailable)
			if avail.Status == metav1.ConditionFalse {
				str.log.Debugf("AppDeployment '%s/%s' for App '%s' is not available, reason: '%s'",
					appDep.Namespace, appDep.Name, appName, avail.Reason)
				continue
			}

			brkCtx := &BrokerEventContext{
				Context:       ctx,
				AppDeployment: appDep,
				Data:          data,
				Event: &core.Event{
					Context: &core.EventContext{
						Platform:           config.Platform,
						VirtualEnvironment: ve.Name,
						AppDeployment:      appDep.Name,
						ReleaseManifest:    release.ReleaseManifest,
					},
				},
			}

			routes, err := str.buildRoutes(brkCtx)
			if err != nil {
				str.log.Warn(err)
				continue
			}
			if err := relM.AddRoutes(routes...); err != nil {
				str.log.Warn(err)
				continue
			}
		}
	}

	str.releaseMatcher = relM
	return nil
}

func (str *Store) buildRoutes(ctx *BrokerEventContext) ([]*core.Route, error) {
	var routes []*core.Route
	for compName, compSpec := range ctx.AppDeployment.Spec.Components {
		comp := core.NewComponent(
			compSpec.Type,
			ctx.AppDeployment.Spec.AppName,
			compName,
			compSpec.Commit,
		)
		for _, r := range compSpec.Routes {
			route, err := core.NewRoute(r.Id, r.Rule)
			if err != nil {
				return nil, err
			}
			route.Component = comp
			route.EventContext = &core.EventContext{
				Platform:           ctx.Event.Context.Platform,
				VirtualEnvironment: ctx.Event.Context.VirtualEnvironment,
				AppDeployment:      ctx.Event.Context.AppDeployment,
				ReleaseManifest:    ctx.Event.Context.ReleaseManifest,
			}

			if err := route.Resolve(ctx.Data); err != nil {
				return nil, err
			}
			routes = append(routes, route)
		}
	}

	return routes, nil
}

func (str *Store) key(name string) types.NamespacedName {
	return k8s.Key(config.Namespace, name)
}
