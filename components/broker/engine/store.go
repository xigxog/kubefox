package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/cache"
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/matcher"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	k8scache "k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TplComponentKey = "component/%s"
	TplAdapterKey   = "adapter/%s"
)

type Store struct {
	namespace string

	resCache     ctrlcache.Cache
	compRegCache cache.Cache[*kubefox.ComponentReg]
	compRegKV    jetstream.KeyValue

	depMatchers cache.Cache[*matcher.EventMatcher]
	relMatcher  *matcher.EventMatcher

	envInf ctrlcache.Informer
	depInf ctrlcache.Informer
	relInf ctrlcache.Informer
	dsInf  ctrlcache.Informer

	ctx    context.Context
	cancel context.CancelFunc

	mutex sync.RWMutex

	log *logkf.Logger
}

func NewStore(namespace string) *Store {
	ctx, cancel := context.WithCancel(context.Background())
	return &Store{
		namespace:    namespace,
		depMatchers:  cache.New[*matcher.EventMatcher](time.Minute * 15),
		relMatcher:   new(matcher.EventMatcher),
		compRegCache: cache.New[*kubefox.ComponentReg](time.Hour * 24),
		ctx:          ctx,
		cancel:       cancel,
		log:          logkf.Global,
	}
}

func (str *Store) Open(compRegKV jetstream.KeyValue) error {
	str.compRegKV = compRegKV
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

func (str *Store) RegisterComponent(ctx context.Context, comp *kubefox.Component, reg *kubefox.ComponentReg) error {
	id := fmt.Sprintf(TplComponentKey, comp.GroupKey())
	return str.setReg(ctx, id, reg)
}

func (str *Store) RegisterAdapter(ctx context.Context, comp *kubefox.Component) error {
	id := fmt.Sprintf(TplAdapterKey, comp.Key())
	return str.setReg(ctx, id, &kubefox.ComponentReg{})
}

func (str *Store) setReg(ctx context.Context, id string, reg *kubefox.ComponentReg) error {
	str.log.Debugf("registering component '%s'", id)
	b, err := json.Marshal(reg)
	if err != nil {
		return err
	}

	if _, err := str.compRegKV.Put(ctx, id, b); err != nil {
		return err
	}
	str.compRegCache.Set(id, reg)

	return nil
}

func (str *Store) Component(ctx context.Context, comp *kubefox.Component) (*kubefox.ComponentReg, error) {
	return str.getReg(ctx, fmt.Sprintf(TplComponentKey, comp.GroupKey()))
}

func (str *Store) Adapter(ctx context.Context, comp *kubefox.Component) (*kubefox.ComponentReg, error) {
	return str.getReg(ctx, fmt.Sprintf(TplAdapterKey, comp.Key()))
}

func (str *Store) IsAdapter(ctx context.Context, comp *kubefox.Component) bool {
	r, err := str.Adapter(ctx, comp)
	return r != nil && err == nil
}

func (str *Store) getReg(ctx context.Context, id string) (*kubefox.ComponentReg, error) {
	compReg, found := str.compRegCache.Get(id)
	if !found {
		entry, err := str.compRegKV.Get(ctx, id)
		if errors.Is(err, nats.ErrKeyNotFound) {
			return nil, fmt.Errorf("%w: component is not registered", ErrRouteNotFound)
		} else if err != nil {
			return nil, err
		}

		compReg = &kubefox.ComponentReg{}
		err = json.Unmarshal(entry.Value(), compReg)
		if err != nil {
			return nil, err
		}

		str.compRegCache.Set(id, compReg)
	}

	return compReg, nil
}

func (str *Store) Deployment(name string) (*v1alpha1.Deployment, error) {
	obj := new(v1alpha1.Deployment)
	return obj, str.get(name, obj, true)
}

func (str *Store) Environment(name string) (*v1alpha1.Environment, error) {
	obj := new(v1alpha1.Environment)
	return obj, str.get(name, obj, false)
}

func (str *Store) Release(name string) (*v1alpha1.Release, error) {
	obj := new(v1alpha1.Release)
	return obj, str.get(name, obj, true)
}

// TODO return a map of node names to broker pod id. This will allow running
// broker without host network. Broker just sends back correct ip during
// subscribe.
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

	// There are no matchers in cache, perform full reload.
	str.mutex.RUnlock()
	return str.buildReleaseMatcher(ctx)
}

func (str *Store) DeploymentMatcher(ctx context.Context, evtCtx *kubefox.EventContext) (*matcher.EventMatcher, error) {
	dep, err := str.Deployment(evtCtx.Deployment)
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

	routes, err := str.buildRoutes(ctx, dep.Spec.Components, env.Spec.Vars, evtCtx)
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
	case *v1alpha1.Deployment:
		str.log.Debug("deployment added")

	case *v1alpha1.Environment:
		str.log.Debug("environment added")

	case *v1alpha1.Release:
		str.log.Debug("release added")
		str.buildReleaseMatcher(context.Background())

	default:
		return
	}
}

func (str *Store) OnUpdate(oldObj, obj interface{}) {
	switch obj.(type) {
	case *v1alpha1.Deployment:
		str.log.Debug("deployment updated")

	case *v1alpha1.Environment:
		str.log.Debug("environment updated")

	case *v1alpha1.Release:
		str.log.Debug("release updated")
		str.buildReleaseMatcher(context.Background())

	default:
		return
	}
}

func (str *Store) OnDelete(obj interface{}) {
	switch obj.(type) {
	case *v1alpha1.Deployment:
		str.log.Debug("deployment deleted")

	case *v1alpha1.Environment:
		str.log.Debug("environment deleted")

	case *v1alpha1.Release:
		str.log.Debug("release deleted")
		str.buildReleaseMatcher(context.Background())

	default:
		return
	}
}

func (str *Store) buildReleaseMatcher(ctx context.Context) (*matcher.EventMatcher, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	relList := new(v1alpha1.ReleaseList)
	if err := str.resCache.List(ctx, relList, &client.ListOptions{Namespace: str.namespace}); err != nil {
		return nil, err
	}

	relM := matcher.New()
	for _, rel := range relList.Items {
		comps := rel.Spec.Deployment.Components
		vars := rel.Spec.Environment.Vars
		evtCtx := &kubefox.EventContext{Release: rel.Name}

		routes, err := str.buildRoutes(ctx, comps, vars, evtCtx)
		if err != nil {
			return nil, err
		}
		if err := relM.AddRoutes(routes); err != nil {
			return nil, err
		}
	}

	str.mutex.Lock()
	defer str.mutex.Unlock()
	str.relMatcher = relM

	return relM, nil
}

func (str *Store) buildRoutes(
	ctx context.Context,
	comps map[string]*v1alpha1.Component,
	vars map[string]*kubefox.Val,
	evtCtx *kubefox.EventContext) ([]*kubefox.Route, error) {

	routes := make([]*kubefox.Route, 0)
	for compName, resComp := range comps {
		comp := &kubefox.Component{Name: compName, Commit: resComp.Commit}
		compReg, err := str.Component(ctx, comp)
		if err != nil {
			return nil, err
		}

		for _, r := range compReg.Routes {
			rCopy := *r
			rCopy.Component = comp
			rCopy.EventContext = evtCtx
			if err := rCopy.Resolve(vars); err != nil {
				return nil, err
			}
			routes = append(routes, &rCopy)
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
