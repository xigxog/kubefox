package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	klog "k8s.io/klog/v2"

	"k8s.io/client-go/tools/cache"
	"k8s.io/component-base/logs"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
)

func main() {
	flag.Parse()
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		klog.Fatal(err)
	}

	c, err := ctrlcache.New(ctrl.GetConfigOrDie(), ctrlcache.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		klog.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	klog.Info("b4 start")
	go func() {
		if err := c.Start(ctx); err != nil {
			klog.Fatal(err)
		}
	}()
	klog.Info("after start")

	inf, err := c.GetInformer(ctx, &v1alpha1.Platform{})
	if err != nil {
		klog.Fatal(err)
	}
	cache.WaitForCacheSync(ctx.Done(), inf.HasSynced)

	for {
		time.Sleep(time.Second)
	}

	buf := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	sentence, err := buf.ReadBytes('\n')
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(sentence))
	}

	p := &v1alpha1.Platform{}
	c.Get(ctx, types.NamespacedName{Namespace: "kfp-dev", Name: "dev"}, p)

	klog.Infof("%v", p)

	// stop := make(chan struct{})
	// defer close(stop)

	// cont.Run(stop)

	// select {}
}
