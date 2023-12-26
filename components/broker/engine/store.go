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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Store struct {
	namespace string

	resCache  ctrlcache.Cache
	compCache cache.Cache[*api.ComponentDefinition]

	depMatchers cache.Cache[*core.EventMatcher]
	relMatcher  *core.EventMatcher

	ctx    context.Context
	cancel context.CancelFunc

	mutex sync.RWMutex

	log *logkf.Logger
}

func NewStore(namespace string) *Store {
	ctx, cancel := context.WithCancel(context.Background())
	return &Store{
		namespace:   namespace,
		depMatchers: cache.New[*core.EventMatcher](time.Minute * 15),
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
		&v1alpha1.ClusterVirtualEnv{},
		&v1alpha1.VirtualEnv{},
		&v1alpha1.VirtualEnvSnapshot{},
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
	return r.Type == api.ComponentTypeGenesis
}

func (str *Store) AppDeployment(ctx context.Context, evt *core.EventContext) (*v1alpha1.AppDeployment, error) {
	obj := new(v1alpha1.AppDeployment)
	return obj, str.get(evt.AppDeployment, obj, true)
}

func (str *Store) VirtualEnv(ctx context.Context, evt *core.EventContext) (*v1alpha1.VirtualEnvSnapshot, error) {
	if evt.VirtualEnv == "" {
		return nil, core.ErrNotFound()
	}

	if evt.VirtualEnvSnapshot != "" {
		snapshot := &v1alpha1.VirtualEnvSnapshot{}
		if err := str.get(evt.VirtualEnvSnapshot, snapshot, true); err != nil {
			return nil, err
		}
		return snapshot, nil
	}

	env := &v1alpha1.VirtualEnv{}
	if err := str.get(evt.VirtualEnv, env, true); err != nil {
		return nil, err
	}

	if env.Spec.Parent != "" {
		clusterEnv := &v1alpha1.ClusterVirtualEnv{}
		if err := str.get(env.Spec.Parent, clusterEnv, false); err != nil {
			return nil, err
		}
		env.MergeParent(clusterEnv)
	}

	hash, err := hashstructure.Hash(&env.Data, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, err
	}

	return &v1alpha1.VirtualEnvSnapshot{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.GroupVersion.Identifier(),
			Kind:       "VirtualEnvSnapshot",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: env.Namespace,
			Name: fmt.Sprintf("%s-%s-%s",
				env.Name, env.GetResourceVersion(), time.Now().UTC().Format("20060102-150405")),
		},
		Spec: v1alpha1.VirtualEnvSnapshotSpec{
			Source: v1alpha1.VirtualEnvSource{
				Name:            env.Name,
				ResourceVersion: env.ResourceVersion,
				DataChecksum:    fmt.Sprint(hash),
			},
		},
		Data:    &env.Data,
		Details: env.Details,
	}, nil
}

// TODO cache
func (str *Store) Adapters(ctx context.Context) (map[string]api.Adapter, error) {
	list := &v1alpha1.HTTPAdapterList{}
	if err := str.resCache.List(ctx, list); err != nil {
		return nil, err
	}

	adapters := map[string]api.Adapter{}
	for _, a := range list.Items {
		adapters[a.Name] = &a
	}

	return adapters, nil
}

func (str *Store) ReleaseMatcher(ctx context.Context) (*core.EventMatcher, error) {
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

func (str *Store) DeploymentMatcher(ctx context.Context, evtCtx *core.EventContext) (*core.EventMatcher, error) {
	dep, err := str.AppDeployment(ctx, evtCtx)
	if err != nil {
		return nil, err
	}
	env, err := str.VirtualEnv(ctx, evtCtx)
	if err != nil {
		return nil, err
	}
	id := fmt.Sprintf("%s-%s-%s-%s", dep.Name, dep.ResourceVersion, env.GetName(), env.GetResourceVersion())

	if depM, found := str.depMatchers.Get(id); found {
		return depM, nil
	}

	routes, err := str.buildRoutes(ctx, dep.Spec.Components, env.Data, evtCtx)
	if err != nil {
		return nil, err
	}

	depM := core.NewEventMatcher()
	if err := depM.AddRoutes(routes...); err != nil {
		return nil, err
	}
	str.depMatchers.Set(id, depM)

	return depM, nil
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
	_, isAppDep := obj.(*v1alpha1.AppDeployment)
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
		// Clear releaseMatcher so it is updated again on next access.
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

	for _, app := range list.Items {
		for compName, compSpec := range app.Spec.Components {
			comp := &core.Component{Name: compName, Commit: compSpec.Commit}
			compCache.Set(comp.GroupKey(), &compSpec.ComponentDefinition)
		}
	}

	p := &v1alpha1.Platform{}
	if err := str.resCache.Get(ctx, k8s.Key(config.Namespace, config.Platform), p); err != nil {
		return nil, err
	}

	for _, c := range p.Status.Components {
		if c.Name == api.PlatformComponentHTTPSrv {
			comp := &core.Component{Name: c.Name, Commit: c.Commit}
			compCache.Set(comp.GroupKey(), &api.ComponentDefinition{
				Type: api.ComponentTypeGenesis,
			})
		}
	}

	return compCache, nil
}

func (str *Store) buildReleaseMatcher(ctx context.Context) (*core.EventMatcher, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	envList := &v1alpha1.VirtualEnvList{}
	if err := str.resCache.List(ctx, envList); err != nil {
		return nil, err
	}

	relM := core.NewEventMatcher()
	for _, env := range envList.Items {
		release := env.Status.ActiveRelease
		if release == nil {
			str.log.Debugf("VirtualEnv '%s/%s' does not have an active Release", env.Namespace, env.Name)
			continue
		}

		evtCtx := &core.EventContext{
			AppDeployment:      release.AppDeployment.Name,
			VirtualEnv:         env.Name,
			VirtualEnvSnapshot: release.VirtualEnvSnapshot,
		}
		appDep, err := str.AppDeployment(ctx, evtCtx)
		if err != nil {
			str.log.Warn(err)
			continue
		}
		env, err := str.VirtualEnv(ctx, evtCtx)
		if err != nil {
			str.log.Warn(err)
			continue
		}

		routes, err := str.buildRoutes(ctx, appDep.Spec.Components, env.Data, evtCtx)
		if err != nil {
			str.log.Warn(err)
			continue
		}
		if err := relM.AddRoutes(routes...); err != nil {
			str.log.Warn(err)
			continue
		}
	}

	return relM, nil
}

func (str *Store) buildRoutes(
	ctx context.Context,
	comps map[string]*v1alpha1.Component,
	data *api.VirtualEnvData,
	evtCtx *core.EventContext) ([]*core.Route, error) {

	var routes []*core.Route
	for compName, compSpec := range comps {
		comp := &core.Component{Name: compName, Commit: compSpec.Commit}
		for _, r := range compSpec.Routes {
			// TODO? cache routes so template doesn't need to be parsed again
			rule, err := core.NewRule(r.Id, r.Rule)
			if err != nil {
				return nil, err
			}
			route := &core.Route{
				Rule:         rule,
				Component:    comp,
				EventContext: evtCtx,
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
