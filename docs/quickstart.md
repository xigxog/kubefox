# Quickstart

## Prerequisites

The following tools must be installed for this quickstart:

- [Docker](https://docs.docker.com/engine/install/) - A container toolset and
  runtime. Used to build KubeFox components' OCI images and run a local
  Kubernetes cluster via kind.
- [Fox](https://github.com/xigxog/kubefox-cli/releases/) -
  CLI for communicating with the KubeFox platform. Download the latest release
  and put the binary on your path.
- [Git](https://github.com/git-guides/install-git) - A distributed version
  control system.
- [Helm](https://helm.sh/docs/intro/install/) - Package manager for Kubernetes.
  Used to install the KubeFox platform on Kubernetes.
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) - **K**uberentes
  **in** **D**ocker. A tool for running local Kubernetes clusters using Docker
  container "nodes".
- [Kubectl](https://kubernetes.io/docs/tasks/tools/) - CLI for communicating
  with a Kubernetes cluster's control plane, using the Kubernetes API.

Here are a few optional but recommended tools:

- [Go](https://go.dev/doc/install) - Programming language. The
  `hello-world` sample app is written in Go but Fox is able to compile it even
  without Go installed.
- [VS Code](https://code.visualstudio.com/download) - A lightweight but powerful
  source code editor. Helpful if you want to explore the `hello-world` app.

## Install KubeFox Platform

You will start with setting up a local Kubernetes cluster using kind. The
following command can be used to create the cluster.

```{ .shell .copy }
kind create cluster --wait 5m
```

??? example "Output"

    ```text
    $ kind create cluster --wait 5m
    Creating cluster "kind" ...
    ✓ Ensuring node image (kindest/node:v1.27.3) 🖼
    ✓ Preparing nodes 📦
    ✓ Writing configuration 📜
    ✓ Starting control-plane 🕹️
    ✓ Installing CNI 🔌
    ✓ Installing StorageClass 💾
    ✓ Waiting ≤ 5m0s for control-plane = Ready ⏳
    • Ready after 15s 💚
    Set kubectl context to "kind-kind"
    You can now use your cluster with:

    kubectl cluster-info --context kind-kind

    Have a nice day! 👋
    ```

Next you need to install the KubeFox Helm Chart. This will start the KubeFox
operator on your Kubernetes cluster. It is responsible for managing KubeFox
platforms and apps.

```{ .shell .copy }
helm upgrade kubefox kubefox \
  --repo https://xigxog.github.io/helm-charts \
  --create-namespace --namespace kubefox-system \
  --install --wait
```

??? example "Output"

    ```text
    $ helm upgrade kubefox kubefox \
        --repo https://xigxog.github.io/helm-charts \
        --create-namespace --namespace kubefox-system \
        --install --wait

    Release "kubefox" does not exist. Installing it now.
    NAME: kubefox
    LAST DEPLOYED: Thu Jan  1 00:00:00 1970
    NAMESPACE: kubefox-system
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    ```

## Deploy App

Great now you're ready to create your first KubeFox app and deploy it to a
KubeFox platform running on the local kind cluster. You'll be using a local Git
repo for the demo. To get things started create a new directory and use Fox to
initialize the `hello-world` app. You'll want to run all the remaining commands
from this directory. The export command tells Fox to print some extra info about
what is going on. You can also pass the `--verbose` flag to print debug
statements. If this is your first time using Fox it will run you through
one-time setup. You can answer the prompts as indicated below.

```{ .shell .copy }
export FOX_INFO=true && \
  mkdir kubefox-hello-world && \
  cd kubefox-hello-world
```

```{ .shell .copy }
fox init
```

Answer the prompts:

```text
Are you only using KubeFox with local kind cluster? [y/N] y
Enter the kind cluster's name (default 'kind'): kind
Would you like to initialize the 'hello-world' KubeFox app? [y/N] y
Enter URL for remote Git repo (optional): # Press ENTER to skip
Would you like to create a KubeFox platform? [Y/n] y
Enter the KubeFox platform's name (required): demo
Enter the Kubernetes namespace of the KubeFox platform (default 'kubefox-demo'): kubefox-demo
```

??? example "Output"

    ```text
    $ export FOX_INFO=true && \
        mkdir kubefox-hello-world && \
        cd kubefox-hello-world

    $ fox init
    info    It looks like this is the first time you are using 🦊 Fox. Welcome!

    info    🦊 Fox needs some information from you to configure itself. The setup process only
    info    needs to be run once, but if you ever want change things you can use the
    info    command 'fox config setup'.

    info    Please make sure your workstation has Docker installed (https://docs.docker.com/engine/install)
    info    and that KubeFox is installed (https://docs.kubefox.io/install) on your Kubernetes cluster.

    info    If you don't have a Kubernetes cluster you can run one locally with kind (https://kind.sigs.k8s.io)
    info    to experiment with KubeFox.

    info    🦊 Fox needs a place to store the component images it will build, normally this is
    info    a remote container registry. However, if you only want to use KubeFox locally
    info    with kind you can skip this step.
    Are you only using KubeFox with local kind cluster? [y/N] y
    Enter the kind cluster's name (default 'kind'): kind

    info    Configuration successfully written to '/home/xadhatter/.config/kubefox/config.yaml'.

    info    Congrats, you are ready to use KubeFox!
    info    Check out the quickstart for next steps (https://docs.kubefox.io/quickstart/).
    info    If you run into any problems please let us know on GitHub (https://github.com/xigxog/kubefox/issues).

    info    Let's initialize a KubeFox app!

    info    To get things started quickly 🦊 Fox can create a 'hello-world' KubeFox app which
    info    includes two components and example environments for testing.
    Would you like to initialize the 'hello-world' KubeFox app? [y/N] y
    Enter URL for remote Git repo (optional):

    info    You need to have a KubeFox platform instance running to deploy your components.
    info    Don't worry, 🦊 Fox can create one for you.
    Would you like to create a KubeFox platform? [Y/n] y
    Enter the KubeFox platform's name (required): demo
    Enter the Kubernetes namespace of the KubeFox platform (default 'kubefox-demo'): kubefox-demo

    info    KubeFox app initialization complete!
    ```

You should see some newly created directories and files. The `hello-world` app
includes two components and example environments. You might have noticed Fox
initialized a new Git repo for you. Take a look around!

When you're ready it's time to create some environments. Two example
environments are provided in the `hack` dir. Take a quick look at them, then use
`kubectl` to apply them to Kubernetes. The `subPath` variable is used to ensure
unique routes between environments. You can see how it used in the `frontend`
component's `main.go` line 21.

```{ .go .no-copy linenums="21" }
k.Route("Path(`/{{.Env.subPath}}/hello`)", sayHello)
```

Run the following commands to continue. Note the changes to the two environment
variables on lines numbered 11,12 and 23,24.

```{ .shell .copy }
cat -b hack/environments/* && \
  kubectl apply --namespace kubefox-demo --filename hack/environments/
```

??? example "Output"

    ```text hl_lines="13 14 26 27"
    $ cat -b hack/environments/* && \
        kubectl apply --namespace kubefox-demo --filename hack/environments/
     1  apiVersion: kubefox.xigxog.io/v1alpha1
     2  kind: VirtualEnv
     3  metadata:
     4    name: prod
     5  spec:
     6    releasePolicy:
     7      appDeploymentPolicy: VersionRequired
     8      virtualEnvPolicy: SnapshotRequired
     9  data:
    10    vars:
    11      subPath: prod
    12      who: Universe

    13  apiVersion: kubefox.xigxog.io/v1alpha1
    14  kind: VirtualEnv
    15  metadata:
    16    name: qa
    17  spec:
    18    releasePolicy:
    19      appDeploymentPolicy: VersionOptional
    20      virtualEnvPolicy: SnapshotOptional
    21  data:
    22    vars:
    23      subPath: qa
    24      who: World

    virtualenv.kubefox.xigxog.io/prod created
    virtualenv.kubefox.xigxog.io/qa created
    ```

Now you'll deploy the app. The following command will build the component's OCI
images, load them onto the kind cluster, and deploy them to a KubeFox platform.
The first run might take some time as it needs to download dependencies, but
future runs should be much faster. Remember, you can add the `--verbose` flag
for extra output if you want to see what's going on behind the scenes. Note that
the platform created earlier might still be initializing, it could take a few
minutes for all the pods to be ready. But don't worry, future deployments will
be much faster.

```{ .shell .copy }
fox publish main --wait 5m
```

??? example "Output"

    ```text
    $ fox publish main --wait 5m
    info    Building component image 'localhost/kubefox/hello-world/backend:bb702a1'.
    info    Loading component image 'localhost/kubefox/hello-world/backend:bb702a1' into kind cluster 'kind'.

    info    Building component image 'localhost/kubefox/hello-world/frontend:bb702a1'.
    info    Loading component image 'localhost/kubefox/hello-world/frontend:bb702a1' into kind cluster 'kind'.

    info    Waiting for KubeFox platform 'demo' to be ready.
    info    Waiting for component 'backend' to be ready.
    info    Waiting for component 'frontend' to be ready.

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: AppDeployment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 1
      name: main
      namespace: kubefox-demo
      resourceVersion: "13326"
      uid: 5ad9a257-01c0-43e0-b6be-92757a47ba7c
    details:
      app:
        description: A simple app demonstrating the use of KubeFox.
        title: Hello World
    spec:
      app:
        branch: refs/heads/main
        commit: bb702a1
        commitTime: "1970-01-01T00:00:00Z"
        containerRegistry: localhost/kubefox/hello-world
        name: hello-world
      components:
        backend:
          type: kubefox
          commit: bb702a1
          defaultHandler: true
          envSchema:
            who:
              required: true
              type: string
              unique: false
        frontend:
          type: kubefox
          commit: bb702a1
          dependencies:
            backend:
              type: kubefox
          envSchema:
            subPath:
              required: false
              type: string
              unique: true
          routes:
          - id: 0
            rule: Path(`/{{.Env.subpath}}/hello`)
    status:
      available: false
    ```

Now take a quick look at what is running on Kubernetes.

```{ .shell .copy }
kubectl get pods --namespace kubefox-demo
```

??? example "Output"

    ```text
    $ kubectl get pods --namespace kubefox-demo
    NAME                                            READY   STATUS    RESTARTS   AGE
    demo-broker-grkcn                               1/1     Running   0          12s
    demo-httpsrv-7d8d946c57-rlt55                   1/1     Running   0          10s
    demo-nats-0                                     1/1     Running   0          18s
    hello-world-backend-bb702a1-8577fc876-bpf4j     1/1     Running   0          2s
    hello-world-frontend-bb702a1-5d998f5cb-t9qp6    1/1     Running   0          2s
    ```

You can see the two components running that you just deployed,
`hello-world-backend` and `hello-world-frontend`. The `broker`, `httpsrv`, and
`nats` pods are part of the KubeFox platform and were started by the operator
when you created the platform.

Normally connections to the KubeFox platform would be made through a public
facing load balancer, but setting that up is outside the scope of this
quickstart. Instead you can use Fox to create a local proxy. In a new terminal
start this command.

```{ .shell .copy }
fox proxy 8080
```

??? info "macOS Network Warning"

    <figure markdown>
      ![macosx-warning](images/fox-mac-net-warn.png)
    </figure>

    If you are using macOS you might notice this dialog popup when you start the
    proxy. This is expected as Fox starts a local HTTP server. The server is
    bound to the `localhost` interface and is only accessible from your
    workstation. Please press `Allow` to continue.

??? example "Output"

    ```text
    $ fox proxy 8080
    HTTP proxy started on http://127.0.0.1:8080
    ```

Now back in the original terminal you can test the deployment. When KubeFox
deploys an app it starts the components but will not route requests to it until
it is released. But you can still test deployments by providing some context.
KubeFox needs two pieces of information, the deployment to use and the
environment to inject. These can be passed as headers or query parameters.

```{ .shell .copy }
curl "http://localhost:8080/qa/hello?kf-dep=main&kf-env=qa"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/qa/hello?kf-dep=main&kf-env=qa"
    👋 Hello World!
    ```

Next try switching to the `prod` environment created earlier. With KubeFox
there is no need to create another deployment to switch environments, simply
change the query parameter. This is possible because KubeFox injects context at
request time instead of at deployment. Adding environments has nearly zero
overhead! Be sure to change the `subPath` from `qa` to `prod` to reflect the
change of environment.

```{ .shell .copy }
curl "http://localhost:8080/prod/hello?kf-dep=main&kf-env=prod"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/prod/hello?kf-dep=main&kf-env=prod"
    👋 Hello Universe!
    ```

## Release App

Now you will release the app to the `qa` environment. Once released you don't
have to specify all those details in the request, routes are matched
automatically. First you'll publish a versioned deployment. Unlike normal
deployments which can be updated freely, versioned deployments are immutable.
They provide a stable deployment that can be promoted to higher environments
when needed. Whenever you create a versioned deployment is it recommended to tag
the Git repo to make keeping track of versions easier.

```{ .shell .copy }
git tag v0.1.0 && \
  fox publish --version v0.1.0 && \
  fox release v0.1.0 --virtual-env qa --wait 5m
```

??? example "Output"

    ```text
    $ git tag v0.1.0 && \
        fox publish --version v0.1.0 && \
        fox release v0.1.0 --virtual-env qa --wait 5m
    info    Component image 'localhost/kubefox/hello-world/backend:bb702a1' exists, skipping build.
    info    Loading component image 'localhost/kubefox/hello-world/backend:bb702a1' into kind cluster 'kind'.

    info    Component image 'localhost/kubefox/hello-world/frontend:bb702a1' exists, skipping build.
    info    Loading component image 'localhost/kubefox/hello-world/frontend:bb702a1' into kind cluster 'kind'.

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: AppDeployment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 1
      name: v0-1-0
      namespace: kubefox-demo
      resourceVersion: "2257050"
      uid: 782a0938-7f9d-4bae-a6b5-900499fca6f7
    details:
      app:
        description: A simple app demonstrating the use of KubeFox.
        title: Hello World
    spec:
      app:
        branch: refs/heads/main
        commit: bb702a1
        commitTime: "1970-01-01T00:00:00Z"
        containerRegistry: localhost/kubefox/hello-world
        name: hello-world
        tag: refs/tags/v0.1.0
      components:
        backend:
          type: kubefox
          commit: bb702a1
          defaultHandler: true
          envSchema:
            who:
              required: true
              type: string
              unique: false
        frontend:
          type: kubefox
          commit: bb702a1
          dependencies:
            backend:
              type: kubefox
          envSchema:
            subPath:
              required: false
              type: string
              unique: true
          routes:
          - id: 0
            rule: Path(`/{{.Env.subpath}}/hello`)
    status:
      available: false

    info    Waiting for KubeFox Platform 'demo' to be ready...
    info    Waiting for component 'backend' to be ready...
    info    Waiting for component 'frontend' to be ready...

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: Release
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 1
      name: qa
      namespace: kubefox-demo
      resourceVersion: "4369"
      uid: 43b96900-72fc-4499-af10-fc87103d99da
    spec:
      appDeployment:
        name: v0-1-0
        version: v0.1.0
    status:
      current: null
    ```

Try the same request from above, but this time don't specify the context. Since
the app has been released the request will get matched by the component's route
and the context information will be automatically injected by KubeFox.

```{ .shell .copy }
curl "http://localhost:8080/qa/hello"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/qa/hello"
    👋 Hello World!
    ```

Take another look at the pods running on Kubernetes now that you performed
another deployment and release.

```{ .shell .copy }
kubectl get pods --namespace kubefox-demo
```

??? example "Output"

    ```text
    $ kubectl get pods --namespace kubefox-demo
    NAME                                            READY   STATUS    RESTARTS   AGE
    demo-broker-grkcn                               1/1     Running   0          6m11s
    demo-httpsrv-7d8d946c57-rlt55                   1/1     Running   0          6m9s
    demo-nats-0                                     1/1     Running   0          6m17s
    hello-world-backend-bb702a1-8577fc876-bpf4j     1/1     Running   0          6m1s
    hello-world-frontend-bb702a1-5d998f5cb-t9qp6    1/1     Running   0          6m1s
    ```

Surprise, nothing has changed! KubeFox is dynamically injecting the context per
request just like when you changed environments above.

## Update and Promote App

Next you'll make a modification to the `frontend` component and deploy it. Open
up `components/frontend/main.go` in your favorite editor and update line 32 in
the `sayHello` function to say something new.

```go linenums="26" hl_lines="7"
func sayHello(k kit.Kontext) error {
  r, err := k.Req(backend).Send()
  if err != nil {
    return err
  }

  msg := fmt.Sprintf("👋 Hello %s!", r.Str()) //(1)
  k.Log().Debug(msg)

  a := strings.ToLower(k.Header("accept"))
  switch {
  case strings.Contains(a, "application/json"):
    return k.Resp().SendJSON(map[string]any{"msg": msg})

  case strings.Contains(a, "text/html"):
    return k.Resp().SendHTML(fmt.Sprintf(html, msg))

  default:
    return k.Resp().SendStr(msg)
  }
}
```

1. Update me to say `Hey` instead of `Hello`.

Fox operates against the current commit of the Git repo when deploying
components. That means before deploying you need to commit the changes to record
them. Then you can update the `main` deployment and test. Take note of the hash
generated by the commit, in the example output below it is `780e2db`. The commit
hashes are used to version the components as can be seen in the container OCI
image tags.

```{ .shell .copy }
git add . && \
  git commit -m "updated frontend to say Hey" && \
  fox publish main --wait 5m
```

??? example "Output"

    ```text hl_lines="4 10 49"
    $ git add . && \
        git commit -m "updated frontend to say Hey" && \
        fox publish main --wait 5m
    [main 780e2db] updated frontend to say Hey
    1 file changed, 1 insertion(+)

    info    Component image 'localhost/kubefox/hello-world/backend:bb702a1' exists, skipping build.
    info    Loading component image 'localhost/kubefox/hello-world/backend:bb702a1' into kind cluster 'kind'.

    info    Building component image 'localhost/kubefox/hello-world/frontend:780e2db'.
    info    Loading component image 'localhost/kubefox/hello-world/frontend:780e2db' into kind cluster 'kind'.

    info    Waiting for KubeFox platform 'demo' to be ready...
    info    Waiting for component 'backend' to be ready...
    info    Waiting for component 'frontend' to be ready...

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: AppDeployment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 1
      name: main
      namespace: kubefox-demo
      resourceVersion: "2258944"
      uid: 5ad9a257-01c0-43e0-b6be-92757a47ba7c
    details:
      app:
        description: A simple app demonstrating the use of KubeFox.
        title: Hello World
    spec:
      app:
        branch: refs/heads/main
        commit: 780e2db
        commitTime: "1970-01-01T00:00:00Z"
        containerRegistry: localhost/kubefox/hello-world
        name: hello-world
      components:
        backend:
          type: kubefox
          commit: bb702a1
          defaultHandler: true
          envSchema:
            who:
              required: true
              type: string
              unique: false
        frontend:
          type: kubefox
          commit: 780e2db
          dependencies:
            backend:
              type: kubefox
          envSchema:
            subPath:
              required: false
              type: string
              unique: true
          routes:
          - id: 0
            rule: Path(`/{{.Env.subpath}}/hello`)
    status:
      available: true
    ```

Fox didn't rebuild the `backend` component as no changes were made. Try testing
out the updated deployment and current release.

```{ .shell .copy }
curl "http://localhost:8080/qa/hello?kf-dep=main&kf-env=qa"
```

```{ .shell .copy }
curl "http://localhost:8080/prod/hello?kf-dep=main&kf-env=prod"
```

```{ .shell .copy }
curl "http://localhost:8080/qa/hello"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/qa/hello?kf-dep=main&kf-env=qa"
    👋 Hey World!

    $ curl "http://localhost:8080/prod/hello?kf-dep=main&kf-env=prod"
    👋 Hey Universe!

    $ curl "http://localhost:8080/qa/hello"
    👋 Hello World!
    ```

Again, take look at the pods running on Kubernetes with the new deployment.

```{ .shell .copy }
kubectl get pods --namespace kubefox-demo
```

??? example "Output"

    ```text
    $ kubectl get pods --namespace kubefox-demo
    NAME                                             READY   STATUS    RESTARTS   AGE
    demo-broker-pkw8s                                1/1     Running   0          13m
    demo-httpsrv-7d8d946c57-rlt55                    1/1     Running   0          13m
    demo-nats-0                                      1/1     Running   0          14m
    hello-world-backend-bb702a1-54bcbf6648-5hb9r     1/1     Running   0          12m
    hello-world-frontend-780e2db-59ffcbc668-h7sk9    1/1     Running   0          18s
    hello-world-frontend-bb702a1-584cd8dbdd-lm6ww    1/1     Running   0          12m
    ```

You might be surprised to find only three component pods running to support the
two deployments and release. Because the `backend` component did not change
between deployments KubeFox is able to share a single pod. Not only are
environments injected per request, routing is performed dynamically.

For fun publish the new version of the app, release it to the `qa` environment,
then promote version `v0.1.0` to the `prod` environment. Because of the stricter
policies set in the `prod` environment a snapshot is required to release. Much
like a versioned deployment an environment snapshot is immutable ensuring a
stable release even if the `prod` environment is changed.

Check out those blazing fast the releases.

```{ .shell .copy }
git tag v0.1.1 && \
  fox publish --version v0.1.1 --wait 5m && \
  fox release v0.1.1 --virtual-env qa && \
  fox release v0.1.0 --virtual-env prod --create-snapshot
```

??? example "Output"

    ```text
    $ git tag v0.1.1 && \
        fox publish --version v0.1.1 --wait 5m && \
        fox release v0.1.1 --virtual-env qa && \
        fox release v0.1.0 --virtual-env prod --create-snapshot
    info    Component image 'localhost/kubefox/hello-world/backend:bb702a1' exists.
    info    Loading component image 'localhost/kubefox/hello-world/backend:bb702a1' into kind cluster 'kind'.

    info    Component image 'localhost/kubefox/hello-world/frontend:780e2db' exists.
    info    Loading component image 'localhost/kubefox/hello-world/frontend:780e2db' into kind cluster 'kind'.

    info    Waiting for KubeFox Platform 'demo' to be ready...
    info    Waiting for component 'backend' to be ready...
    info    Waiting for component 'frontend' to be ready...

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: AppDeployment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 1
      name: v0-1-1
      namespace: kubefox-demo
      resourceVersion: "2259758"
      uid: 7d3e6c48-71bd-428f-bd3a-245f73344538
    details:
      app:
        description: A simple app demonstrating the use of KubeFox.
        title: Hello World
    spec:
      app:
        branch: refs/heads/main
        commit: 780e2db
        commitTime: "1970-01-01T00:00:00Z"
        containerRegistry: localhost/kubefox/hello-world
        name: hello-world
        tag: refs/tags/v0.1.1
      components:
        backend:
          type: kubefox
          commit: bb702a1
          defaultHandler: true
          envSchema:
            who:
              required: true
              type: string
              unique: false
        frontend:
          type: kubefox
          commit: 780e2db
          dependencies:
            backend:
              type: kubefox
          envSchema:
            subPath:
              required: false
              type: string
              unique: true
          routes:
          - id: 0
            rule: Path(`/{{.Env.subpath}}/hello`)
    status:
      available: false

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: Release
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 1
      name: qa
      namespace: kubefox-demo
      resourceVersion: "2259782"
      uid: 43b96900-72fc-4499-af10-fc87103d99da
    spec:
      appDeployment:
        name: v0-1-1
        version: v0.1.1
    status:
      current:
        appDeployment:
          name: v0-1-0
          version: v0.1.0
        requestTime: "2023-11-29T18:10:16Z"

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: Release
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 1
      name: prod
      namespace: kubefox-demo
      resourceVersion: "2259824"
      uid: c047e0af-9621-4b88-b659-53e4b4e02cf0
    spec:
      appDeployment:
        name: v0-1-0
        version: v0.1.0
    status:
      current: null
    ```

Test it out when everything is done.

```{ .shell .copy }
curl "http://localhost:8080/qa/hello"
```

```{ .shell .copy }
curl "http://localhost:8080/prod/hello"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/qa/hello"
    👋 Hey World!

    $ curl "http://localhost:8080/prod/hello"
    👋 Hello Universe!
    ```

Check out the rest of the docs for more. If you run into any problems please let
us know on [GitHub Issues](https://github.com/xigxog/kubefox/issues).
