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
	"sync"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/cache"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/matcher"
	"sigs.k8s.io/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Store struct {
	namespace string

	resCache   ctrlcache.Cache
	compCache  cache.Cache[*api.ComponentDefinition]
	validCache cache.Cache[api.Problems]
	adapters   Adapters

	depMatchers cache.Cache[*matcher.EventMatcher]
	relMatcher  *matcher.EventMatcher

	ctx    context.Context
	cancel context.CancelFunc

	mutex sync.RWMutex

	log *logkf.Logger
}

func NewStore(namespace string) *Store {
	ctx, cancel := context.WithCancel(context.Background())
	return &Store{
		namespace:   namespace,
		validCache:  cache.New[api.Problems](time.Minute * 15),
		depMatchers: cache.New[*matcher.EventMatcher](time.Minute * 15),
		ctx:         ctx,
		cancel:      cancel,
		log:         logkf.Global,
	}
}

func (str *Store) Open() error {
	ctx, cancel := context.WithTimeout(str.ctx, time.Minute*3)
	defer cancel()

	if err := str.init(ctx); err != nil {
		return err
	}

	if err := str.updateCaches(ctx, true); err != nil {
		return err
	}

	return nil
}

func (str *Store) init(ctx context.Context) error {
	str.mutex.Lock()
	defer str.mutex.Unlock()

	if err := v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return str.log.ErrorN("adding KubeFox CRs to scheme failed: %v", err)
	}

	cfg, err := ctrl.GetConfig()
	if err != nil {
		return str.log.ErrorN("loading K8s config failed: %v", err)
	}

	str.resCache, err = ctrlcache.New(cfg, ctrlcache.Options{
		Scheme:                      scheme.Scheme,
		DefaultNamespaces:           map[string]ctrlcache.Config{str.namespace: {}},
		ReaderFailOnMissingInformer: true,
	})
	if err != nil {
		return str.log.ErrorN("creating resource cache failed: %v", err)
	}

	err = str.initInformers(ctx,
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
	str.resCache.WaitForCacheSync(ctx)

	return nil
}

func (str *Store) initInformers(ctx context.Context, objs ...client.Object) error {
	for _, obj := range objs {
		// Getting the informer adds it to the cache.
		if inf, err := str.resCache.GetInformer(str.ctx, obj); err != nil {
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

func (str *Store) ComponentDef(comp *core.Component) (*api.ComponentDefinition, bool) {
	return str.compCache.Get(comp.GroupKey())
}

func (str *Store) IsGenesisAdapter(comp *core.Component) bool {
	r, found := str.ComponentDef(comp)
	if !found {
		return false
	}

	switch r.Type {
	case api.ComponentTypeHTTPAdapter:
		return true
	default:
		return false
	}
}

func (str *Store) Platform(ctx context.Context) (*v1alpha1.Platform, error) {
	obj := new(v1alpha1.Platform)
	return obj, str.get(config.Platform, obj, true)
}

func (str *Store) AppDeployment(ctx context.Context, name string) (*v1alpha1.AppDeployment, error) {
	obj := new(v1alpha1.AppDeployment)
	return obj, str.get(name, obj, true)
}

func (str *Store) Data(ctx context.Context, evtCtx *core.EventContext) (*api.Data, error) {
	if evtCtx.VirtualEnvironment == "" {
		return nil, core.ErrNotFound()
	}

	if evtCtx.ReleaseManifest != "" {
		manifest := &v1alpha1.ReleaseManifest{}
		if err := str.get(evtCtx.ReleaseManifest, manifest, true); err != nil {
			return nil, err
		}
		return manifest.Data, nil
	}

	ve := &v1alpha1.VirtualEnvironment{}
	if err := str.get(evtCtx.VirtualEnvironment, ve, true); err != nil {
		return nil, err
	}
	env := &v1alpha1.Environment{}
	if err := str.get(ve.Spec.Environment, env, false); err != nil {
		return nil, err
	}

	return ve.Data.MergeInto(&env.Data), nil
}

func (str *Store) ReleaseMatcher(ctx context.Context) (*matcher.EventMatcher, error) {
	str.mutex.RLock()
	if str.relMatcher != nil {
		str.mutex.RUnlock()
		return str.relMatcher, nil
	}
	str.mutex.RUnlock()

	// There are no matchers in cache, perform full reload.
	str.updateCaches(ctx, false)

	return str.relMatcher, nil
}

func (str *Store) DeploymentMatcher(ctx context.Context, evt *BrokerEvent) (*matcher.EventMatcher, error) {
	// Check cache.
	if depM, found := str.depMatchers.Get(evt.ContextKey); found {
		str.log.Debugf("found cached matcher for event context key '%s'", evt.ContextKey)
		return depM, nil
	}

	routes, err := str.buildRoutes(ctx, evt.AppDep, evt.Data, evt.Context)
	if err != nil {
		return nil, err
	}

	depM := matcher.New()
	if err := depM.AddRoutes(routes...); err != nil {
		return nil, err
	}
	str.depMatchers.Set(evt.ContextKey, depM)

	return depM, nil
}

// AttachEventContext gets the VirtualEnvironment, AppDeployment, and Adapters
// of the Event Context, attaches them to the BrokerEvent, and then validates
// the Event Context. Problems found during validation are returned.
func (str *Store) AttachEventContext(ctx context.Context, evt *BrokerEvent) error {
	var err error
	if evt.Data, err = str.Data(ctx, evt.Context); err != nil {
		return err
	}
	hash, _ := hashstructure.Hash(evt.Data, hashstructure.FormatV2, nil)
	evt.DataChecksum = fmt.Sprint(hash)

	if evt.AppDep, err = str.AppDeployment(ctx, evt.Context.AppDeployment); err != nil {
		return err
	}
	evt.ContextKey = fmt.Sprintf("%s-%d_%s-%s",
		evt.AppDep.Name, evt.AppDep.Generation, evt.Context.VirtualEnvironment, evt.DataChecksum)

	str.mutex.Lock()
	defer str.mutex.Unlock()

	allAdapters := str.adapters
	if allAdapters == nil {
		list := &v1alpha1.HTTPAdapterList{}
		if err := str.resCache.List(ctx, list); err != nil {
			return err
		}

		allAdapters = make(Adapters, len(list.Items))
		for _, a := range list.Items {
			allAdapters.Set(&a)
		}
		str.adapters = allAdapters
	}

	evt.Adapters = Adapters{}
	for _, c := range evt.AppDep.Spec.Components {
		for name, d := range c.Dependencies {
			if d.Type.IsAdapter() {
				a, _ := allAdapters.Get(name, d.Type)
				evt.Adapters.Set(a)
			}
		}
	}

	// Check validation cache.
	problems, found := str.validCache.Get(evt.ContextKey)
	if !found {
		problems, err = evt.AppDep.Validate(evt.Data,
			func(name string, typ api.ComponentType) (api.Adapter, error) {
				a, found := evt.Adapters[name]
				if !found {
					return nil, core.ErrNotFound()
				}
				return a, nil
			})
		if err != nil {
			return err
		}
		str.validCache.Set(evt.ContextKey, problems)
	}
	if len(problems) > 0 {
		b, _ := yaml.Marshal(problems)
		return core.ErrInvalid(fmt.Errorf("event context is invalid\n%s", b))
	}

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

	isAppDep := false
	switch obj.(type) {
	case *v1alpha1.AppDeployment:
		isAppDep = true

	case *v1alpha1.HTTPAdapter:
		// Clear adapters to force reload on next use.
		str.adapters = nil
	}

	go str.updateCaches(context.Background(), isAppDep)
}

func (str *Store) updateCaches(ctx context.Context, updateComps bool) error {
	str.mutex.Lock()
	defer str.mutex.Unlock()

	if updateComps {
		if compCache, err := str.buildComponentCache(ctx); err != nil {
			return err
		} else {
			str.compCache = compCache
		}
	}

	if relM, err := str.buildReleaseMatcher(ctx); err != nil {
		// Clear releaseMatcher so it is updated again on next use.
		str.relMatcher = nil
		str.log.Error(err)
		return err

	} else {
		str.relMatcher = relM
	}

	return nil
}

func (str *Store) buildComponentCache(ctx context.Context) (cache.Cache[*api.ComponentDefinition], error) {
	compCache := cache.New[*api.ComponentDefinition](time.Hour * 24)

	list := &v1alpha1.AppDeploymentList{}
	if err := str.resCache.List(ctx, list); err != nil {
		return nil, err
	}

	for _, appDep := range list.Items {
		for compName, compSpec := range appDep.Spec.Components {
			comp := &core.Component{
				Type:   string(compSpec.Type),
				App:    appDep.Spec.AppName,
				Name:   compName,
				Commit: compSpec.Commit,
			}
			compCache.Set(comp.GroupKey(), &compSpec.ComponentDefinition)
		}
	}

	p := &v1alpha1.Platform{}
	if err := str.resCache.Get(ctx, k8s.Key(config.Namespace, config.Platform), p); err != nil {
		return nil, err
	}

	for _, c := range p.Status.Components {
		if c.Name == api.PlatformComponentHTTPSrv {
			comp := &core.Component{
				Type:   string(c.Type),
				Name:   c.Name,
				Commit: c.Commit,
			}
			compCache.Set(comp.GroupKey(), &api.ComponentDefinition{
				Type: c.Type,
			})
		}
	}

	return compCache, nil
}

func (str *Store) buildReleaseMatcher(ctx context.Context) (*matcher.EventMatcher, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	envList := &v1alpha1.VirtualEnvironmentList{}
	if err := str.resCache.List(ctx, envList); err != nil {
		return nil, err
	}

	relM := matcher.New()
	for _, env := range envList.Items {
		release := env.Status.ActiveRelease
		if release == nil {
			str.log.Debugf("VirtualEnvironment '%s/%s' does not have an active Release", env.Namespace, env.Name)
			continue
		}

		avail := k8s.Condition(env.Status.Conditions, api.ConditionTypeActiveReleaseAvailable)
		if avail.Status == metav1.ConditionFalse {
			str.log.Debugf("VirtualEnvironment '%s/%s' Release is not available, reason: '%s'",
				env.Namespace, env.Name, avail.Reason)
			continue
		}

		evtCtx := &core.EventContext{
			Platform:           config.Platform,
			VirtualEnvironment: env.Name,
			ReleaseManifest:    release.ReleaseManifest,
		}

		for _, app := range release.Apps {
			appDep, err := str.AppDeployment(ctx, app.AppDeployment)
			if err != nil {
				str.log.Warn(err)
				continue
			}
			data, err := str.Data(ctx, evtCtx)
			if err != nil {
				str.log.Warn(err)
				continue
			}

			routes, err := str.buildRoutes(ctx, appDep, data, evtCtx)
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

	return relM, nil
}

func (str *Store) buildRoutes(
	ctx context.Context,
	appDep *v1alpha1.AppDeployment,
	data *api.Data,
	evtCtx *core.EventContext) ([]*core.Route, error) {

	var routes []*core.Route
	for compName, compSpec := range appDep.Spec.Components {
		comp := &core.Component{
			Type:   string(compSpec.Type),
			App:    appDep.Spec.AppName,
			Name:   compName,
			Commit: compSpec.Commit,
		}
		for _, r := range compSpec.Routes {
			route, err := core.NewRoute(r.Id, r.Rule)
			if err != nil {
				return nil, err
			}
			route.Component = comp
			route.EventContext = &core.EventContext{
				Platform:           evtCtx.Platform,
				AppDeployment:      appDep.Name,
				VirtualEnvironment: evtCtx.VirtualEnvironment,
				ReleaseManifest:    evtCtx.ReleaseManifest,
			}

			if err := route.Resolve(data); err != nil {
				return nil, err
			}
			routes = append(routes, route)
		}
	}

	return routes, nil
}

func (str *Store) get(name string, obj client.Object, namespaced bool) error {
	ctx, cancel := context.WithTimeout(str.ctx, time.Minute)
	defer cancel()

	nn := types.NamespacedName{Name: name}
	if namespaced {
		nn.Namespace = str.namespace
	}

	return str.resCache.Get(ctx, nn, obj)
}
