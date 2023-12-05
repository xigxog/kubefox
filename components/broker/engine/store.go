package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/cache"
	"github.com/xigxog/kubefox/components/broker/config"
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/matcher"
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
		&v1alpha1.ClusterVirtualEnv{},
		&v1alpha1.VirtualEnv{},
		&v1alpha1.VirtualEnvSnapshot{},
		&v1alpha1.AppDeployment{},
		&v1alpha1.Release{},
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

func (str *Store) Component(ctx context.Context, comp *kubefox.Component) (*api.ComponentDefinition, bool) {
	return str.compCache.Get(comp.GroupKey())
}

func (str *Store) IsGenesisAdapter(ctx context.Context, comp *kubefox.Component) bool {
	r, found := str.Component(ctx, comp)
	if !found {
		return false
	}
	return r.Type == api.ComponentTypeGenesis
}

func (str *Store) AppDeployment(name string) (*v1alpha1.AppDeployment, error) {
	obj := new(v1alpha1.AppDeployment)
	return obj, str.get(name, obj, true)
}

func (str *Store) Environment(name string) (v1alpha1.VirtualEnvObject, error) {
	env := &v1alpha1.VirtualEnv{}
	err := str.get(name, env, true)
	switch {
	case err == nil:
		envSnap := &v1alpha1.VirtualEnvSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Name:      env.Name,
				Namespace: env.Namespace,
			},
			Data: v1alpha1.EnvSnapshotData{
				Source: v1alpha1.EnvSource{
					Kind:            env.Kind,
					Name:            env.Name,
					ResourceVersion: env.ResourceVersion,
				},
				SnapshotTime: metav1.Now(),
			},
		}

		if env.Spec.Parent != "" {
			clusterEnv := &v1alpha1.ClusterVirtualEnv{}
			if err := str.get(env.Spec.Parent, clusterEnv, false); err != nil {
				return nil, err
			}
			v1alpha1.MergeVirtualEnvironment(envSnap, clusterEnv)
		}
		v1alpha1.MergeVirtualEnvironment(envSnap, env)

		return envSnap, nil

	case k8s.IsNotFound(err):
		clusterEnv := &v1alpha1.ClusterVirtualEnv{}
		return clusterEnv, str.get(name, clusterEnv, false)

	default:
		return nil, err
	}
}

func (str *Store) Release(name string) (*v1alpha1.Release, error) {
	obj := new(v1alpha1.Release)
	return obj, str.get(name, obj, true)
}

func (str *Store) ReleaseAppDeployment(name string) (*v1alpha1.AppDeployment, error) {
	rel := new(v1alpha1.Release)
	if err := str.get(name, rel, true); err != nil {
		return nil, err
	}
	return str.AppDeployment(rel.Spec.AppDeployment.Name)
}

func (str *Store) ReleaseEnv(name string) (v1alpha1.VirtualEnvObject, error) {
	rel := new(v1alpha1.Release)
	if err := str.get(name, rel, true); err != nil {
		return nil, err
	}

	if rel.Spec.VirtualEnvSnapshot != "" {
		env := &v1alpha1.VirtualEnvSnapshot{}
		return env, str.get(rel.Spec.VirtualEnvSnapshot, env, true)
	}

	return str.Environment(rel.Name)
}

func (str *Store) ReleaseMatcher(ctx context.Context) (*matcher.EventMatcher, error) {
	str.mutex.RLock()
	if str.relMatcher != nil && !str.relMatcher.IsEmpty() {
		str.mutex.RUnlock()
		return str.relMatcher, nil
	}
	str.mutex.RUnlock()

	// There are no matchers in cache, perform full reload.
	str.updateCaches(ctx, false)

	return str.relMatcher, nil
}

func (str *Store) DeploymentMatcher(ctx context.Context, evtCtx *kubefox.EventContext) (*matcher.EventMatcher, error) {
	dep, err := str.AppDeployment(evtCtx.Deployment)
	if err != nil {
		return nil, err
	}
	env, err := str.Environment(evtCtx.Environment)
	if err != nil {
		return nil, err
	}
	id := fmt.Sprintf("%s-%s-%s-%s", dep.Name, dep.ResourceVersion, env.GetName(), env.GetResourceVersion())

	if depM, found := str.depMatchers.Get(id); found {
		return depM, nil
	}

	routes, err := str.buildRoutes(ctx, dep.Spec.App.Components, env.GetData().Vars, evtCtx)
	if err != nil {
		return nil, err
	}

	depM := matcher.New()
	depM.AddRoutes(routes)
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
		str.relMatcher = matcher.New()
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
		for compName, compSpec := range app.Spec.App.Components {
			comp := &kubefox.Component{Name: compName, Commit: compSpec.Commit}
			compCache.Set(comp.GroupKey(), &compSpec.ComponentDefinition)
		}
	}

	p := &v1alpha1.Platform{}
	if err := str.resCache.Get(ctx, k8s.Key(config.Namespace, config.Platform), p); err != nil {
		return nil, err
	}

	for _, c := range p.Status.Components {
		if c.Name == api.PlatformComponentHTTPSrv {
			comp := &kubefox.Component{Name: c.Name, Commit: c.Commit}
			compCache.Set(comp.GroupKey(), &api.ComponentDefinition{
				Type: api.ComponentTypeGenesis,
			})
		}
	}

	return compCache, nil
}

func (str *Store) buildReleaseMatcher(ctx context.Context) (*matcher.EventMatcher, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	relList := &v1alpha1.ReleaseList{}
	if err := str.resCache.List(ctx, relList); err != nil {
		return nil, err
	}

	relM := matcher.New()
	for _, rel := range relList.Items {
		var (
			comps map[string]*v1alpha1.Component
			vars  map[string]*api.Val
		)

		var appDepName string
		if rel.Status.Current != nil {
			appDepName = rel.Status.Current.AppDeployment.Name
		}
		if appDepName == "" {
			str.log.Debugf("Release '%s/%s' does not have a current AppDeployment", rel.Namespace, rel.Name)
			continue
		}

		if appDep, err := str.AppDeployment(appDepName); err != nil {
			str.log.Warn(err)
			continue
		} else {
			comps = appDep.Spec.App.Components
		}
		if env, err := str.ReleaseEnv(rel.Name); err != nil {
			str.log.Warn(err)
			continue
		} else {
			vars = env.GetData().Vars
		}

		evtCtx := &kubefox.EventContext{Release: rel.Name}

		routes, err := str.buildRoutes(ctx, comps, vars, evtCtx)
		if err != nil {
			str.log.Warn(err)
			continue
		}
		if err := relM.AddRoutes(routes); err != nil {
			str.log.Warn(err)
			continue
		}
	}

	return relM, nil
}

func (str *Store) buildRoutes(
	ctx context.Context,
	comps map[string]*v1alpha1.Component,
	vars map[string]*api.Val,
	evtCtx *kubefox.EventContext) ([]*kubefox.Route, error) {

	routes := make([]*kubefox.Route, 0)
	for compName, compSpec := range comps {
		comp := &kubefox.Component{Name: compName, Commit: compSpec.Commit}
		for _, r := range compSpec.Routes {
			route := &kubefox.Route{
				RouteSpec:    r,
				Component:    comp,
				EventContext: evtCtx,
			}
			if err := route.Resolve(vars, sprig.FuncMap()); err != nil {
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
