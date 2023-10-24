<!-- markdownlint-disable MD033 -->

# Quickstart

## Prerequisites

The following tools must be installed for this quickstart:

- [Docker](https://docs.docker.com/engine/install/) - A container toolset and
  runtime. Used to build KubeFox Components' OCI images and run a local
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

```shell
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

```shell
helm install kubefox-demo kubefox \
  --repo https://xigxog.github.io/helm-charts \
  --create-namespace --namespace kubefox-system \
  --wait
```

??? example "Output"

    ```text
    $ helm install kubefox-demo kubefox \
        --repo https://xigxog.github.io/helm-charts \
        --create-namespace --namespace kubefox-system \
        --wait

    NAME: kubefox-demo
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

```shell
export FOX_INFO=true
```

```shell
mkdir kubefox-hello-world && cd kubefox-hello-world
```

```shell
fox init
```

Answer the prompts:

```text
Are you only using KubeFox with local kind cluster? [y/N] y
Enter the kind cluster's name (default 'kind'): kind
Would you like to initialize the 'hello-world' KubeFox app? [y/N] y
Enter URL for remote Git repo (optional): # Press ENTER to skip
```

??? example "Output"

    ```text
    $ mkdir kubefox-hello-world && cd kubefox-hello-world

    $ fox init
    info    It looks like this is the first time you are using 🦊 Fox. Welcome!

    info    Fox needs some information from you to configure itself. The setup process only
    info    needs to be run once, but if you ever want change things you can use the
    info    command 'fox config setup'.

    info    Please make sure your workstation has Docker installed (https://docs.docker.com/engine/install)
    info    and that KubeFox is installed (https://docs.kubefox.io/install) on your Kubernetes cluster.

    info    If you don't have a Kubernetes cluster you can run one locally with kind (https://kind.sigs.k8s.io)
    info    to experiment with KubeFox.

    info    Fox needs a place to store the KubeFox Component images it will build, normally
    info    this is a remote container registry. However, if you only want to use KubeFox
    info    locally with kind you can skip this step.
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

    info    KubeFox app initialization complete!
    ```

You should see some newly created directories and files. The `hello-world` app
includes two components and example environments. You might also notice Fox
initialized a new Git repo for you. Take a look around!

When you're ready it's time to create some environments. Two example
environments are provided in the `hack` dir. Take a quick look at them, then use
`kubectl` to apply them to Kubernetes. The `subPath` variable is used to ensure
unique routes between environments. You can see how it used in the `frontend`
component's `main.go` line 12.

```go linenums="12"
k.Route("Path(`/{{.Env.subPath}}/hello`)", sayHello)
```

Run the following commands to continue.

```shell
cat hack/environments/universe.yaml
```

```shell
cat hack/environments/world.yaml
```

```shell
kubectl apply --filename hack/environments/
```

??? example "Output"

    ```text
    $ cat hack/environments/universe.yaml
    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: Environment
    metadata:
      name: universe
    spec:
      vars:
        subPath: big
        who: Universe

    $ cat hack/environments/world.yaml
    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: Environment
    metadata:
      name: world
    spec:
      vars:
        subPath: small
        who: World

    $ kubectl apply --filename hack/environments/
    environment.kubefox.xigxog.io/universe created
    environment.kubefox.xigxog.io/world created
    ```

Now you'll deploy the app. The following command will build the component's OCI
images, load them onto the kind cluster, and deploy them to a KubeFox platform.
The first run might take some time as it needs to download dependencies, but
future runs should be much faster. Remember, you can add the `--verbose` flag
for extra output if you want to see what's going on behind the scenes. Note that
the first deployment will initialize a platform so it might take a couple
minutes for all the pods to be ready. But don't worry, future deployments will
be very fast.

```shell
fox publish alpha --wait 5m
```

Answer the prompts:

```text
Would you like to create a KubeFox platform? [Y/n] y
Enter the KubeFox platform's name (required): dev
Enter the Kubernetes namespace of the KubeFox platform (default 'kubefox-dev'): kubefox-dev
```

??? example "Output"

    ```text
    $ fox publish alpha --wait 5m
    info    Building component image 'localhost/kubefox/hello-world/backend:bb702a1'.
    info    Loading component image 'localhost/kubefox/hello-world/backend:bb702a1' into kind cluster 'kind'.

    info    Building component image 'localhost/kubefox/hello-world/frontend:bb702a1'.
    info    Loading component image 'localhost/kubefox/hello-world/frontend:bb702a1' into kind cluster 'kind'.

    info    You need to have a KubeFox platform instance running to deploy your components.
    info    Don't worry, 🦊 Fox can create one for you.
    Would you like to create a KubeFox platform? [Y/n] y
    Enter the KubeFox platform's name (required): dev
    Enter the Kubernetes namespace of the KubeFox platform (default 'kubefox-dev'): kubefox-dev

    info    Waiting for KubeFox platform 'dev' to be ready.
    info    Waiting for component 'backend' to be ready.
    info    Waiting for component 'frontend' to be ready.

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: Deployment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 1
      name: alpha
      namespace: kubefox-dev
      resourceVersion: "13326"
      uid: 5ad9a257-01c0-43e0-b6be-92757a47ba7c
    spec:
      app:
        commit: bb702a1
        containerRegistry: localhost/kubefox/hello-world
        description: A simple app demonstrating the use of KubeFox.
        gitRef: refs/heads/main
        name: hello-world
        title: Hello World
      components:
        backend:
          commit: bb702a1
          env: {}
        frontend:
          commit: bb702a1
          env: {}
    status: {}
    ```

Now take a quick look at what is running on Kubernetes.

```shell
kubectl get pods --namespace kubefox-dev
```

??? example "Output"

    ```text
    $ kubectl get pods --namespace kubefox-dev
    NAME                                           READY   STATUS    RESTARTS   AGE
    dev-broker-grkcn                               1/1     Running   0          12s
    dev-nats-0                                     1/1     Running   0          18s
    hello-world-backend-bb702a1-8577fc876-bpf4j    1/1     Running   0          2s
    hello-world-frontend-bb702a1-5d998f5cb-t9qp6   1/1     Running   0          2s
    ```

You can see the two components running that you just deployed,
`hello-world-backend` and `hello-world-frontend`. The `broker` and `nats` pods
are part of the KubeFox platform and were started by the operator when you
created the platform above.

Normally connections to the KubeFox platform would be made through a public
facing load balancer, but setting that up is outside the scope of this
quickstart. Instead you can use Fox to create a local proxy. In a new terminal
start this command.

```shell
fox proxy 8080
```

??? example "Output"

    ```text
    $ fox proxy 8080
    HTTP proxy started on http://127.0.0.1:8080
    ```

Now back in the original terminal you can test the deployment. When KubeFox
deploys an app it starts the components but will not automatically send requests
to it until it is released. But you can still test deployments by providing some
context. KubeFox needs two pieces of information, the deployment to use and the
environment to inject. These can be passed as headers or query parameters.

```shell
curl "http://localhost:8080/small/hello?kf-dep=alpha&kf-env=world"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/small/hello?kf-dep=alpha&kf-env=world"
    👋 Hello World!
    ```

Next try switching to the `universe` environment created earlier. With KubeFox
there is no need to create another deployment to switch environments, simply
change the query parameter. This is possible because KubeFox injects context at
request time instead of at deployment. Adding environments has nearly zero
overhead! Be sure to change the `subPath` from `small` to `big` to reflect the
change of environment.

```shell
curl "http://localhost:8080/big/hello?kf-dep=alpha&kf-env=universe"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/big/hello?kf-dep=alpha&kf-env=universe"
    👋 Hello Universe!
    ```

## Release App

Now you will release the app so you don't have to specify all those details in
the request. It is recommended to tag the repo for releases to help keep track
of versions. Fox works against the active state of the Git repo. To deploy or
release a different version of your app simply checkout the tag, branch, or
commit you want and let Fox do the rest.

```shell
git tag v0.1.0 && git checkout v0.1.0
```

```shell
fox release dev --env world --wait 5m
```

```shell
git switch -
```

??? example "Output"

    ```text
    $ git tag v0.1.0 && git checkout v0.1.0
    HEAD is now at bb702a1 And so it begins...

    # You might see a note from Git about being in a 'detached HEAD' state. It
    # can be disabled in the future by running `git config --global advice.detachedHead false`,
    # if you prefer.

    $ fox release dev --env world --wait 5m
    info    Component image 'localhost/kubefox/hello-world/backend:bb702a1' exists.
    info    Loading component image 'localhost/kubefox/hello-world/backend:bb702a1' into kind cluster 'kind'.

    info    Component image 'localhost/kubefox/hello-world/frontend:bb702a1' exists.
    info    Loading component image 'localhost/kubefox/hello-world/frontend:bb702a1' into kind cluster 'kind'.

    info    Waiting for KubeFox platform 'dev' to be ready.
    info    Waiting for component 'frontend' to be ready.
    info    Waiting for component 'backend' to be ready.

    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 1
      name: dev
      namespace: kubefox-dev
      resourceVersion: "4369"
      uid: 43b96900-72fc-4499-af10-fc87103d99da
    spec:
      deployment:
        app:
          commit: bb702a1
          containerRegistry: localhost/kubefox/hello-world
          description: A simple app demonstrating the use of KubeFox.
          gitRef: refs/tags/v0.1.0
          name: hello-world
          title: Hello World
        components:
          backend:
            commit: bb702a1
            env: {}
          frontend:
            commit: bb702a1
            env: {}
      environment:
        name: world
    status: {}
    ```

Try the same request from above, but this time don't specify the context. Since
the app has been released the request will get matched by the component's route
and the context information will be automatically injected by KubeFox.

```shell
curl "http://localhost:8080/small/hello"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/small/hello"
    👋 Hello World!
    ```

Take another look at the pods running on Kubernetes now that you performed a
release.

```shell
kubectl get pods --namespace kubefox-dev
```

??? example "Output"

    ```text
    $ kubectl get pods --namespace kubefox-dev
    NAME                                           READY   STATUS    RESTARTS   AGE
    dev-broker-grkcn                               1/1     Running   0          6m11s
    dev-nats-0                                     1/1     Running   0          6m17s
    hello-world-backend-bb702a1-8577fc876-bpf4j    1/1     Running   0          6m1s
    hello-world-frontend-bb702a1-5d998f5cb-t9qp6   1/1     Running   0          6m1s
    ```

Surprise, nothing has changed! KubeFox is dynamically injecting the context per
request just like when you changed environments above.

## Version App

Next you'll make a modification to the `frontend` component and deploy it. Open
up `components/frontend/main.go` in your favorite editor and update line 22 in
the `sayHello` function to say something new.

```go linenums="16" hl_lines="7"
func sayHello(k kit.Kontext) error {
    r, err := k.Component("backend").Send()
    if err != nil {
        return err
    }

    msg := fmt.Sprintf("👋 Hey %s!", r.Str()) //(1)
    k.Log().Info(msg)

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

As noted earlier Fox operates against the current commit of the Git repo. That
means before deploying you need to commit the changes to record them. Then you
can publish a new deployment, `beta`, and test. Take note of the hash
generated by the commit, in the example output below it is `780e2db`.

```shell
git add .
```

```shell
git commit -m "updated frontend to say Hey"
```

```shell
fox publish beta --wait 5m
```

??? example "Output"

    ```text
    $ git add .

    $ git commit -m "updated frontend to say Hey"
    [main 780e2db] updated frontend to say Hey
    1 file changed, 1 insertion(+)

    $ fox publish beta --wait 5m
    info    Component image 'localhost/kubefox/hello-world/backend:bb702a1' exists, skipping build.
    info    Loading component image 'localhost/kubefox/hello-world/backend:bb702a1' into kind cluster 'kind'.

    info    Building component image 'localhost/kubefox/hello-world/frontend:780e2db'.
    info    Loading component image 'localhost/kubefox/hello-world/frontend:780e2db' into kind cluster 'kind'.

    info    Waiting for KubeFox platform 'dev' to be ready...
    info    Waiting for component 'backend' to be ready...
    info    Waiting for component 'frontend' to be ready...

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: Deployment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 2
      name: beta
      namespace: kubefox-dev
      resourceVersion: "6707"
      uid: b6be4df9-fe9d-4bc8-9544-120afd0fbfd9
    spec:
      app:
        commit: 780e2db
        containerRegistry: localhost/kubefox/hello-world
        description: A simple app demonstrating the use of KubeFox.
        gitRef: refs/heads/main
        name: hello-world
        title: Hello World
      components:
        backend:
          commit: bb702a1
          env: {}
        frontend:
          commit: 780e2db
          env: {}
    status: {}
    ```

Notice that Fox didn't rebuild the `backend` component as no changes were made.
Try testing out the new deployment. You can even switch back to `alpha`
to verify the changes.

```shell
curl "http://localhost:8080/small/hello?kf-dep=beta&kf-env=world"
```

```shell
curl "http://localhost:8080/big/hello?kf-dep=beta&kf-env=universe"
```

```shell
curl "http://localhost:8080/big/hello?kf-dep=alpha&kf-env=universe"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/small/hello?kf-dep=beta&kf-env=world"
    👋 Hey World!

    $ curl "http://localhost:8080/big/hello?kf-dep=beta&kf-env=universe"
    👋 Hey Universe!

    $ curl "http://localhost:8080/big/hello?kf-dep=alpha&kf-env=universe"
    👋 Hello Universe!
    ```

Again, take look at the pods running on Kubernetes with the new deployment.

```shell
kubectl get pods --namespace kubefox-dev
```

??? example "Output"

    ```text
    $ kubectl get pods --namespace kubefox-dev
    NAME                                            READY   STATUS    RESTARTS   AGE
    dev-broker-pkw8s                                1/1     Running   0          13m
    dev-nats-0                                      1/1     Running   0          14m
    hello-world-backend-bb702a1-54bcbf6648-5hb9r    1/1     Running   0          12m
    hello-world-frontend-780e2db-59ffcbc668-h7sk9   1/1     Running   0          18s
    hello-world-frontend-bb702a1-584cd8dbdd-lm6ww   1/1     Running   0          12m
    ```

You might be surprised to find only three component pods running to support the
two deployments and release. Because the `backend` component did not change
between deployments KubeFox is able to share a single pod. Not only are
environments injected per request, routing is performed dynamically.

For fun tag the new version and update the existing `dev` release using the
`world` environment. Then release the original version `v0.1.0` using the
`universe` environment. Notice how fast the releases are.

```shell
git tag v0.1.1 && git checkout v0.1.1 && fox release dev --env world --wait 5m
```

```shell
git checkout v0.1.0 && fox release qa --env universe --wait 5m
```

```shell
git checkout main
```

??? example "Output"

    ```text
    $ git tag v0.1.1 && git checkout v0.1.1 && fox release dev --env world --wait 5m
    HEAD is now at 780e2db updated frontend to say Hey
    info    Component image 'localhost/kubefox/hello-world/backend:bb702a1' exists.
    info    Loading component image 'localhost/kubefox/hello-world/backend:bb702a1' into kind cluster 'kind'.

    info    Component image 'localhost/kubefox/hello-world/frontend:780e2db' exists.
    info    Loading component image 'localhost/kubefox/hello-world/frontend:780e2db' into kind cluster 'kind'.

    info    Waiting for KubeFox platform 'dev' to be ready...
    info    Waiting for component 'backend' to be ready...
    info    Waiting for component 'frontend' to be ready...

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: Release
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 3
      name: dev
      namespace: kubefox-dev
      resourceVersion: "2511"
      uid: 9fc19bbd-db75-4a36-a601-955078563d5c
    spec:
      deployment:
        app:
          commit: 780e2db
          containerRegistry: localhost/kubefox/hello-world
          description: A simple app demonstrating the use of KubeFox.
          gitRef: refs/tags/v0.1.1
          name: hello-world
          title: Hello World
        components:
          backend:
            commit: bb702a1
            env: {}
          frontend:
            commit: 780e2db
            env: {}
      environment:
        name: world
    status: {}

    $ git checkout v0.1.0 && fox release qa --env universe --wait 5m
    Previous HEAD position was 780e2db updated frontend to say Hey
    HEAD is now at bb702a1 And so it begins...
    info    Component image 'localhost/kubefox/hello-world/backend:bb702a1' exists.
    info    Loading component image 'localhost/kubefox/hello-world/backend:bb702a1' into kind cluster 'kind'.

    info    Component image 'localhost/kubefox/hello-world/frontend:bb702a1' exists.
    info    Loading component image 'localhost/kubefox/hello-world/frontend:bb702a1' into kind cluster 'kind'.

    info    Waiting for KubeFox platform 'dev' to be ready...
    info    Waiting for component 'backend' to be ready...
    info    Waiting for component 'frontend' to be ready...

    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 1
      name: qa
      namespace: kubefox-dev
      resourceVersion: "2328"
      uid: 64294db4-79a5-45b8-873a-d093e5aa2851
    spec:
      deployment:
        app:
          commit: bb702a1
          containerRegistry: localhost/kubefox/hello-world
          description: A simple app demonstrating the use of KubeFox.
          gitRef: refs/tags/v0.1.0
          name: hello-world
          title: Hello World
        components:
          backend:
            commit: bb702a1
            env: {}
          frontend:
            commit: bb702a1
            env: {}
      environment:
        name: universe
    status: {}

    $ git checkout main
    Previous HEAD position was bb702a1 And so it begins...
    Switched to branch 'main'
    ```

Test it out when everything is done.

```shell
curl "http://localhost:8080/small/hello"
```

```shell
curl "http://localhost:8080/big/hello"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/small/hello"
    👋 Hey World!

    $ curl "http://localhost:8080/big/hello"
    👋 Hello Universe!
    ```

Check out the rest of the docs for more. If you run into any problems please let
us know on [GitHub Issues](https://github.com/xigxog/kubefox/issues).
