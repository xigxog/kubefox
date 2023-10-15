package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/xigxog/kubefox/libs/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/cache"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"github.com/xigxog/kubefox/libs/core/matcher"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	k8scache "k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
	dsInf  ctrlcache.Informer

	ctx    context.Context
	cancel context.CancelFunc

	mutex sync.Mutex

	log *logkf.Logger
}

func NewStore(namespace string) *Store {
	ctx, cancel := context.WithCancel(context.Background())
	return &Store{
		namespace:        namespace,
		depMatchersByDep: cache.New[map[string]*DeploymentMatcher](time.Minute * 15),
		depMatchersByEnv: cache.New[map[string]*DeploymentMatcher](time.Minute * 15),
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
		Scheme:            scheme.Scheme,
		DefaultNamespaces: map[string]ctrlcache.Config{str.namespace: {}},
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
	if inf, err := c.GetInformer(str.ctx, &appsv1.DaemonSet{}); err != nil {
		return str.log.ErrorN("daemonset informer failed: %v", err)
	} else {
		str.dsInf = inf
		k8scache.WaitForCacheSync(str.ctx.Done(), inf.HasSynced)
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
	return obj, str.get(name, obj, true)
}

// TODO return a map of node names to broker pod id. This will allow running
// broker without host network. Broker just sends back correct ip during
// subscribe.
func (str *Store) GetBrokerMap() (map[string]string, error) {
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

func (str *Store) GetRelMatchers() (ReleaseMatchers, error) {
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

func (str *Store) GetDepMatcher(depName, envName string) (*DeploymentMatcher, error) {
	dep, err := str.GetDeployment(depName)
	if err != nil {
		return nil, err
	}
	env, err := str.GetEnvironment(envName)
	if err != nil {
		return nil, err
	}
	depId := fmt.Sprintf("%s-%s", dep.Name, dep.ResourceVersion)
	envId := fmt.Sprintf("%s-%s", dep.Name, dep.ResourceVersion)

	if byEnv, found := str.depMatchersByDep.Get(depId); found {
		if depM, found := byEnv[envId]; found {
			str.depMatchersByEnv.Get(envId) // touch env to reset TTL
			return depM, nil
		}
	}

	str.mutex.Lock()
	defer str.mutex.Unlock()

	matchers, err := str.buildMatchers(dep.Spec.Components, env.Spec.Vars)
	if err != nil {
		return nil, err
	}

	depMatcher := &DeploymentMatcher{
		Deployment:  depName,
		Environment: envName,
		Matchers:    matchers,
	}

	if m, found := str.depMatchersByDep.Get(depId); found {
		m[envName] = depMatcher
	} else {
		str.depMatchersByDep.Set(depId, map[string]*DeploymentMatcher{envId: depMatcher})
	}
	if m, found := str.depMatchersByEnv.Get(envId); found {
		m[depId] = depMatcher
	} else {
		str.depMatchersByEnv.Set(envId, map[string]*DeploymentMatcher{depId: depMatcher})
	}

	return depMatcher, nil
}

func (str *Store) OnAdd(obj interface{}, isInInitialList bool) {
	str.mutex.Lock()
	defer str.mutex.Unlock()

	switch rel := obj.(type) {
	case *v1alpha1.Deployment:
		str.log.Debug("deployment added")

	case *v1alpha1.Environment:
		str.log.Debug("environment added")

	case *v1alpha1.Release:
		str.log.Debug("release added")
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
		str.log.Debug("deployment updated")

	case *v1alpha1.Environment:
		str.log.Debug("environment updated")

	case *v1alpha1.Release:
		str.log.Debug("release updated")
		matchers, err := str.buildMatchers(v.Spec.Deployment.Components, v.Spec.Environment.Vars)
		str.relMatchers[v.Name] = &ReleaseMatcher{
			Release:  v.Name,
			Matchers: matchers,
			Error:    err,
		}

	default:
		return
	}
}

func (str *Store) OnDelete(obj interface{}) {
	str.mutex.Lock()
	defer str.mutex.Unlock()

	switch v := obj.(type) {
	case *v1alpha1.Deployment:
		str.log.Debug("deployment deleted")

	case *v1alpha1.Environment:
		str.log.Debug("environment deleted")

	case *v1alpha1.Release:
		str.log.Debug("release deleted")
		delete(str.relMatchers, v.Name)

	default:
		return
	}
}

func (str *Store) buildMatchers(comps map[string]*v1alpha1.Component, vars map[string]*kubefox.Val) ([]*matcher.EventMatcher, error) {
	matchers := make([]*matcher.EventMatcher, 0)
	for compName, resComp := range comps {
		comp := &kubefox.Component{Name: compName, Commit: resComp.Commit}
		compRts, found := str.routes.Get(comp.GroupKey())
		if !found {
			entry, err := str.routesKV.Get(comp.GroupKey())
			if err != nil {
				return nil, err
			}
			compReg := new(kubefox.ComponentReg)
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
