package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/cache"
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/matcher"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Store struct {
	namespace string

	resCache      ctrlcache.Cache
	compSpecCache cache.Cache[*api.ComponentDefinition]
	compSpecKV    jetstream.KeyValue

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
		namespace:     namespace,
		depMatchers:   cache.New[*matcher.EventMatcher](time.Minute * 15),
		relMatcher:    new(matcher.EventMatcher),
		compSpecCache: cache.New[*api.ComponentDefinition](time.Hour * 24),
		ctx:           ctx,
		cancel:        cancel,
		log:           logkf.Global,
	}
}

func (str *Store) Open(compSpecKV jetstream.KeyValue) error {
	str.compSpecKV = compSpecKV
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

	ctx, cancel := context.WithTimeout(str.ctx, time.Minute)
	defer cancel()

	err = str.initInformers(ctx,
		&v1alpha1.VirtualEnv{},
		&v1alpha1.ResolvedEnvironment{},
		&v1alpha1.AppDeployment{},
		&v1alpha1.Release{},
		&v1alpha1.Platform{},
		&appsv1.DaemonSet{},
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

func (str *Store) RegisterComponent(ctx context.Context, comp *kubefox.Component, reg *api.ComponentDefinition) error {
	str.log.Debugf("registering component '%s' of type '%s'", comp.GroupKey(), reg.Type)
	b, err := json.Marshal(reg)
	if err != nil {
		return err
	}

	if _, err := str.compSpecKV.Put(ctx, comp.GroupKey(), b); err != nil {
		return err
	}
	str.compSpecCache.Set(comp.GroupKey(), reg)

	return nil
}

func (str *Store) Component(ctx context.Context, comp *kubefox.Component) (*api.ComponentDefinition, error) {
	compSpec, found := str.compSpecCache.Get(comp.GroupKey())
	if !found {
		entry, err := str.compSpecKV.Get(ctx, comp.GroupKey())
		if errors.Is(err, nats.ErrKeyNotFound) {
			return nil, kubefox.ErrRouteNotFound(fmt.Errorf("component is not registered"))
		} else if err != nil {
			return nil, err
		}

		compSpec = &api.ComponentDefinition{}
		err = json.Unmarshal(entry.Value(), compSpec)
		if err != nil {
			return nil, err
		}

		str.compSpecCache.Set(comp.GroupKey(), compSpec)
	}

	return compSpec, nil
}

func (str *Store) IsGenesisAdapter(ctx context.Context, comp *kubefox.Component) bool {
	r, err := str.Component(ctx, comp)
	if err != nil || r == nil {
		return false
	}
	return r.Type == api.ComponentTypeGenesis
}

func (str *Store) AppDeployment(name string) (*v1alpha1.AppDeployment, error) {
	obj := new(v1alpha1.AppDeployment)
	return obj, str.get(name, obj, true)
}

func (str *Store) Environment(name string) (*v1alpha1.VirtualEnv, error) {
	obj := new(v1alpha1.VirtualEnv)
	return obj, str.get(name, obj, false)
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

func (str *Store) ReleaseEnv(name string) (*v1alpha1.ResolvedEnvironment, error) {
	rel := new(v1alpha1.Release)
	if err := str.get(name, rel, true); err != nil {
		return nil, err
	}

	envName := fmt.Sprintf("%s-%s", rel.Spec.Environment.Name, rel.Spec.Environment.ResourceVersion)
	env := new(v1alpha1.ResolvedEnvironment)

	return env, str.get(envName, env, true)
}

// TODO return a map of node names to broker pod id. This will allow running
// broker without host network. Broker just sends back correct ip during
// subscribe.... is this really needed? use headless service!
func (str *Store) BrokerMap() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(str.ctx, time.Minute)
	defer cancel()

	ls := labels.NewSelector()
	// r, err := labels.NewRequirement()
	// ls.Add()

	dsList := new(appsv1.DaemonSetList)
	if err := str.resCache.List(ctx, dsList, &client.ListOptions{
		Namespace:     str.namespace,
		LabelSelector: ls,
	}); err != nil {
		return nil, err
	}

	return map[string]string{}, nil
}

func (str *Store) ReleaseMatcher(ctx context.Context) (*matcher.EventMatcher, error) {
	str.mutex.RLock()
	if !str.relMatcher.IsEmpty() {
		str.mutex.RUnlock()
		return str.relMatcher, nil
	}
	str.mutex.RUnlock()

	// There are no matchers in cache, perform full reload.
	return str.updateReleaseMatcher(ctx)
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
	id := fmt.Sprintf("%s-%s-%s-%s", dep.Name, dep.ResourceVersion, env.Name, env.ResourceVersion)

	if depM, found := str.depMatchers.Get(id); found {
		return depM, nil
	}

	routes, err := str.buildRoutes(ctx, dep.Spec.Components, env.Data.Vars, evtCtx)
	if err != nil {
		return nil, err
	}

	str.mutex.Lock()
	defer str.mutex.Unlock()

	depM := matcher.New()
	depM.AddRoutes(routes)
	str.depMatchers.Set(id, depM)

	return depM, nil
}

func (str *Store) OnAdd(obj interface{}, isInInitialList bool) {
	switch obj.(type) {
	case *v1alpha1.Platform:
		str.log.Debug("platform added")
		str.updateReleaseMatcher(context.Background())

	default:
		return
	}
}

func (str *Store) OnUpdate(oldObj, obj interface{}) {
	switch obj.(type) {
	case *v1alpha1.Platform:
		str.log.Debug("platform updated")
		str.updateReleaseMatcher(context.Background())

	default:
		return
	}
}

func (str *Store) OnDelete(obj interface{}) {
	switch obj.(type) {
	case *v1alpha1.Platform:
		str.log.Debug("platform deleted")
		str.updateReleaseMatcher(context.Background())

	default:
		return
	}
}

func (str *Store) updateReleaseMatcher(ctx context.Context) (*matcher.EventMatcher, error) {
	relM, err := str.buildReleaseMatcher(ctx)
	if err != nil {
		// Clear releaseMatcher so it is updated again on next access.
		relM = matcher.New()
		str.log.Error(err)
	}

	str.mutex.Lock()
	str.relMatcher = relM
	str.mutex.Unlock()

	return relM, err
}

func (str *Store) buildReleaseMatcher(ctx context.Context) (*matcher.EventMatcher, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	platList := &v1alpha1.PlatformList{}
	if err := str.resCache.List(ctx, platList); err != nil {
		return nil, err
	}

	relList := []v1alpha1.Release{}
	for _, p := range platList.Items {
		for _, env := range p.Spec.Environments {
			if env.Release == nil || env.Release.Name == "" {
				continue
			}
			rel := v1alpha1.Release{}
			if err := str.get(env.Release.Name, &rel, true); client.IgnoreNotFound(err) != nil {
				return nil, err
			} else if apierrors.IsNotFound(err) {
				str.log.Warnf("active release '%s' not found", env.Release.Name)
				continue
			}
			relList = append(relList, rel)
		}
	}

	relM := matcher.New()
	for _, rel := range relList {
		var (
			comps map[string]*v1alpha1.Component
			vars  map[string]*api.Val
		)

		if appDep, err := str.AppDeployment(rel.Spec.AppDeployment.Name); err != nil {
			return nil, err
		} else {
			comps = appDep.Spec.Components
		}

		if env, err := str.ReleaseEnv(rel.Name); err != nil {
			return nil, err
		} else {
			vars = env.Data.Vars
		}

		evtCtx := &kubefox.EventContext{Release: rel.Name}

		routes, err := str.buildRoutes(ctx, comps, vars, evtCtx)
		if err != nil {
			return nil, err
		}
		if err := relM.AddRoutes(routes); err != nil {
			return nil, err
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
	for compName, resComp := range comps {
		comp := &kubefox.Component{Name: compName, Commit: resComp.Commit}
		compSpec, err := str.Component(ctx, comp)
		if err != nil {
			return nil, err
		}

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
