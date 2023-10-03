package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/cache"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"github.com/xigxog/kubefox/libs/core/matcher"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	k8scache "k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Matcher interface {
	Match(evt *kubefox.Event) *kubefox.MatchedEvent
}

type Store struct {
	namespace string

	resCache ctrlcache.Cache
	routes   cache.Cache[[]*kubefox.Route]
	routesKV nats.KeyValue

	depMatchersByDep cache.Cache[map[string]*DeploymentMatcher]
	depMatchersByEnv cache.Cache[map[string]*DeploymentMatcher]
	relMatchers      ReleaseMatchers

	envInf ctrlcache.Informer
	depInf ctrlcache.Informer
	relInf ctrlcache.Informer

	ctx    context.Context
	cancel context.CancelFunc

	mutex sync.Mutex

	log *logkf.Logger
}

type DeploymentMatcher struct {
	Deployment  string
	Environment string
	Matchers    []*matcher.EventMatcher
	Error       error
}

type ReleaseMatchers map[string]*ReleaseMatcher

type ReleaseMatcher struct {
	Release  string
	Matchers []*matcher.EventMatcher
	Error    error
}

func NewStore(namespace string) *Store {
	ctx, cancel := context.WithCancel(context.Background())
	return &Store{
		namespace:        namespace,
		depMatchersByDep: cache.New[map[string]*DeploymentMatcher](time.Hour),
		depMatchersByEnv: cache.New[map[string]*DeploymentMatcher](time.Hour),
		relMatchers:      make(map[string]*ReleaseMatcher),
		routes:           cache.New[[]*kubefox.Route](time.Hour * 24),
		ctx:              ctx,
		cancel:           cancel,
		log:              logkf.Global,
	}
}

func (str *Store) Open(routesKV nats.KeyValue) error {
	str.routesKV = routesKV
	if err := v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return str.log.ErrorN("adding KubeFox CRs to scheme failed: %v", err)
	}

	cfg, err := ctrl.GetConfig()
	if err != nil {
		return str.log.ErrorN("loading K8s config failed: %v", err)
	}
	c, err := ctrlcache.New(cfg, ctrlcache.Options{
		Scheme:     scheme.Scheme,
		Namespaces: []string{str.namespace},
	})
	if err != nil {
		return str.log.ErrorN("creating resource cache failed: %v", err)
	}
	str.resCache = c

	go func() {
		if err := c.Start(str.ctx); err != nil {
			str.log.Error(err)
		}
	}()

	// Getting the informer starts the sync process for the resource kind.
	if inf, err := c.GetInformer(str.ctx, &v1alpha1.Deployment{}); err != nil {
		return str.log.ErrorN("deployment informer failed: %v", err)
	} else {
		str.depInf = inf
		k8scache.WaitForCacheSync(str.ctx.Done(), inf.HasSynced)
		inf.AddEventHandler(str)
	}
	if inf, err := c.GetInformer(str.ctx, &v1alpha1.Environment{}); err != nil {
		return str.log.ErrorN("environment informer failed: %v", err)
	} else {
		str.envInf = inf
		k8scache.WaitForCacheSync(str.ctx.Done(), inf.HasSynced)
		inf.AddEventHandler(str)
	}
	if inf, err := c.GetInformer(str.ctx, &v1alpha1.Release{}); err != nil {
		return str.log.ErrorN("release informer failed: %v", err)
	} else {
		str.relInf = inf
		k8scache.WaitForCacheSync(str.ctx.Done(), inf.HasSynced)
		inf.AddEventHandler(str)
	}

	return nil
}

func (str *Store) Close() {
	str.cancel()
}

func (str *Store) GetDeployment(name string) (*v1alpha1.Deployment, error) {
	obj := new(v1alpha1.Deployment)
	return obj, str.get(name, obj, true)
}

func (str *Store) GetEnvironment(name string) (*v1alpha1.Environment, error) {
	obj := new(v1alpha1.Environment)
	return obj, str.get(name, obj, false)
}

func (str *Store) GetRelease(name string) (*v1alpha1.Release, error) {
	obj := new(v1alpha1.Release)
	return obj, str.get(name, obj, false)
}

func (str *Store) GetReleaseMatchers() (ReleaseMatchers, error) {
	if len(str.relMatchers) > 0 {
		return str.relMatchers, nil
	}

	// There are no matchers in cache, perform full reload.
	str.mutex.Lock()
	defer str.mutex.Unlock()

	ctx, cancel := context.WithTimeout(str.ctx, time.Minute)
	defer cancel()

	relList := new(v1alpha1.ReleaseList)
	if err := str.resCache.List(ctx, relList, &client.ListOptions{Namespace: str.namespace}); err != nil {
		return nil, err
	}

	relMatchers := make(ReleaseMatchers)
	for _, rel := range relList.Items {
		matchers, err := str.buildMatchers(rel.Spec.Deployment.Components, rel.Spec.Environment.Vars)
		if err != nil {
			return nil, err
		}
		relMatchers[rel.Name] = &ReleaseMatcher{
			Release:  rel.Name,
			Matchers: matchers,
			Error:    err,
		}
	}
	str.relMatchers = relMatchers

	return str.relMatchers, nil
}

func (str *Store) GetDeploymentMatcher(depName, envName string) (*DeploymentMatcher, error) {
	if byEnv, found := str.depMatchersByDep.Get(depName); found {
		if depM, found := byEnv[envName]; found {
			str.depMatchersByEnv.Get(envName) // touch env to reset TTL
			return depM, nil
		}
	}

	str.mutex.Lock()
	defer str.mutex.Unlock()

	dep, err := str.GetDeployment(depName)
	if err != nil {
		return nil, err
	}
	env, err := str.GetEnvironment(envName)
	if err != nil {
		return nil, err
	}
	matchers, err := str.buildMatchers(dep.Spec.Components, env.Spec.Vars)
	if err != nil {
		return nil, err
	}

	depMatcher := &DeploymentMatcher{
		Deployment:  depName,
		Environment: envName,
		Matchers:    matchers,
	}

	if m, found := str.depMatchersByDep.Get(depName); found {
		m[envName] = depMatcher
	} else {
		str.depMatchersByDep.Set(depName, map[string]*DeploymentMatcher{envName: depMatcher})
	}
	if m, found := str.depMatchersByEnv.Get(envName); found {
		m[depName] = depMatcher
	} else {
		str.depMatchersByEnv.Set(envName, map[string]*DeploymentMatcher{depName: depMatcher})
	}

	return depMatcher, nil
}

func (rms ReleaseMatchers) Match(evt *kubefox.Event) *kubefox.MatchedEvent {
	for _, rm := range rms {
		if rm.Error != nil {
			continue
		}
		for _, m := range rm.Matchers {
			if comp, rt, match := m.Match(evt); match {
				evt.Release = rm.Release
				evt.Target = comp
				return &kubefox.MatchedEvent{
					Event:   evt,
					RouteId: int64(rt.Id),
				}
			}
		}
	}

	return nil
}

func (dm *DeploymentMatcher) Match(evt *kubefox.Event) *kubefox.MatchedEvent {
	if dm.Error != nil {
		return nil
	}

	for _, m := range dm.Matchers {
		if comp, rt, match := m.Match(evt); match {
			evt.Target = comp
			return &kubefox.MatchedEvent{
				Event:   evt,
				RouteId: int64(rt.Id),
			}
		}
	}

	return nil
}

func (str *Store) OnAdd(obj interface{}, isInInitialList bool) {
	str.mutex.Lock()
	defer str.mutex.Unlock()

	switch rel := obj.(type) {
	case *v1alpha1.Release:
		matchers, err := str.buildMatchers(rel.Spec.Deployment.Components, rel.Spec.Environment.Vars)
		str.relMatchers[rel.Name] = &ReleaseMatcher{
			Release:  rel.Name,
			Matchers: matchers,
			Error:    err,
		}

	default:
		return
	}
}

func (str *Store) OnUpdate(oldObj, obj interface{}) {
	str.mutex.Lock()
	defer str.mutex.Unlock()

	switch v := obj.(type) {
	case *v1alpha1.Deployment:
		if byEnv, found := str.depMatchersByDep.Get(v.Name); found {
			for envName, depMatcher := range byEnv {
				str.depMatchersByEnv.Get(envName) // touch to reset TTL
				if env, err := str.GetEnvironment(envName); err != nil {
					depMatcher.Matchers = nil
					depMatcher.Error = err
				} else {
					depMatcher.Matchers, depMatcher.Error = str.buildMatchers(v.Spec.Components, env.Spec.Vars)
				}
			}
		}

	case *v1alpha1.Environment:
		if byDep, found := str.depMatchersByEnv.Get(v.Name); found {
			for depName, depMatcher := range byDep {
				str.depMatchersByDep.Get(depName) // touch to reset TTL
				if dep, err := str.GetDeployment(depName); err != nil {
					depMatcher.Matchers = nil
					depMatcher.Error = err
				} else {
					depMatcher.Matchers, depMatcher.Error = str.buildMatchers(dep.Spec.Components, v.Spec.Vars)
				}
			}
		}

	case *v1alpha1.Release:
		relMatcher, found := str.relMatchers[v.Name]
		if !found {
			relMatcher = &ReleaseMatcher{Release: v.Name}
			str.relMatchers[v.Name] = relMatcher
		}
		relMatcher.Matchers, relMatcher.Error = str.buildMatchers(v.Spec.Deployment.Components, v.Spec.Environment.Vars)

	default:
		return
	}
}

func (str *Store) OnDelete(obj interface{}) {
	str.mutex.Lock()
	defer str.mutex.Unlock()

	switch v := obj.(type) {
	case *v1alpha1.Deployment:
		if byEnv, found := str.depMatchersByDep.Get(v.Name); found {
			for envName := range byEnv {
				if m, found := str.depMatchersByEnv.Get(envName); found {
					delete(m, v.Name)
					if len(m) == 0 {
						str.depMatchersByEnv.Delete(envName)
					}
				}
			}
			str.depMatchersByDep.Delete(v.Name)
		}

	case *v1alpha1.Environment:
		if byDep, found := str.depMatchersByEnv.Get(v.Name); found {
			for depName := range byDep {
				if m, found := str.depMatchersByDep.Get(depName); found {
					delete(m, v.Name)
					if len(m) == 0 {
						str.depMatchersByDep.Delete(depName)
					}
				}
			}
			str.depMatchersByEnv.Delete(v.Name)
		}

	case *v1alpha1.Release:
		delete(str.relMatchers, v.Name)

	default:
		return
	}
}

func (str *Store) buildMatchers(comps map[string]*v1alpha1.Component, vars map[string]*kubefox.Var) ([]*matcher.EventMatcher, error) {
	matchers := make([]*matcher.EventMatcher, 0)
	for compName, resComp := range comps {
		comp := &kubefox.Component{Name: compName, Commit: resComp.Commit}
		compRts, found := str.routes.Get(comp.GroupKey())
		if !found {
			entry, err := str.routesKV.Get(comp.GroupKey())
			if err != nil {
				return nil, err
			}
			compReg := new(kubefox.ComponentRegistration)
			err = json.Unmarshal(entry.Value(), compReg)
			if err != nil {
				return nil, err
			}
			str.routes.Set(comp.GroupKey(), compReg.Routes)
			compRts = compReg.Routes
		}

		evtMatcher, err := matcher.New(comp)
		if err != nil {
			return nil, err
		}
		for _, r := range compRts {
			if err := r.Resolve(vars); err != nil {
				return nil, err
			}
		}
		if err := evtMatcher.AddRoutes(compRts); err != nil {
			return nil, fmt.Errorf("route issue with component '%s.%s': %w", comp.Name, comp.Commit, err)
		}
		matchers = append(matchers, evtMatcher)
	}
	sort.SliceStable(matchers, func(i, j int) bool {
		return matchers[i].Id() < matchers[j].Id()
	})

	return matchers, nil
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
