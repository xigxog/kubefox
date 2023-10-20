<!-- markdownlint-disable MD033 -->

# Quickstart

## Prerequisites

The following tools must be installed for this quickstart:

- [Docker](https://docs.docker.com/engine/install/) - A container toolset and
  runtime. Used to build KubeFox Components' OCI images and run a local
  Kubernetes cluster via Kind.
- [Fox CLI](https://github.com/xigxog/kubefox-cli/releases/) -
  CLI for communicating with the KubeFox Platform. Download the latest release
  and put the binary on your path.
- [Git](https://github.com/git-guides/install-git) - A distributed version
  control system.
- [Helm](https://helm.sh/docs/intro/install/) - Package manager for Kubernetes.
  Used to install the KubeFox Platform on Kubernetes.
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

Let's start with setting up a local Kubernetes cluster using Kind. The following
command can be used to create the cluster. It exposes some extra ports to the
local host that allows communicating with the KubeFox Platform easier.

```shell
echo >/tmp/kind-cluster.yaml "\
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: kubefox-demo
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 8080
        hostPort: 30080
      - containerPort: 8443
        hostPort: 30443
"
kind create cluster --config /tmp/kind-cluster.yaml --wait 5m
kubectl config use-context kind-kubefox-demo
```

??? example "Output"

    ```text
    $ kind create cluster --name kubefox --wait 5m
    Creating cluster "kubefox" ...
    ✓ Ensuring node image (kindest/node:v1.27.3) 🖼
    ✓ Preparing nodes 📦
    ✓ Writing configuration 📜
    ✓ Starting control-plane 🕹️
    ✓ Installing CNI 🔌
    ✓ Installing StorageClass 💾
    ✓ Waiting ≤ 5m0s for control-plane = Ready ⏳
    • Ready after 15s 💚
    Set kubectl context to "kind-kubefox"
    You can now use your cluster with:

    kubectl cluster-info --context kind-kubefox

    Have a nice day! 👋

    $ kubectl config use-context kind-kubefox-demo
    Switched to context "kind-kubefox-demo".
    ```

Next we need to add the XigXog Helm Repo and install the KubeFox Helm Chart.

```shell
helm install kubefox-demo kubefox \
  --repo https://xigxog.github.io/helm-charts \
  --create-namespace --namespace kubefox-system
```

??? example "Output"

    ```text
    $ helm install kubefox-demo kubefox \
        --repo https://xigxog.github.io/helm-charts \
        --create-namespace --namespace kubefox-system

    NAME: kubefox-demo
    LAST DEPLOYED: Thu Jan  1 00:00:00 1970
    NAMESPACE: kubefox-system
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    ```

## Deploy KubeFox App

Great now we're ready to create your first KubeFox app and deploy it to the a
Platform running on the local Kind Cluster. We'll be using a local Git repo for
the demo. To get things started let's create a new directory and use the Fox to
initialize the `hello-world` app. You'll want to run all the remaining commands
from this directory. The export command tells Fox to print some extra info about
what is going on. If you run into problems you can also pass the `--verbose`
flag to print debug statements.

```shell
export FOX_INFO=true
mkdir kubefox-hello-world && cd kubefox-hello-world
fox init
```

Answer the prompts:

```text
Are you only using KubeFox with local Kind cluster? [y/N] y
Enter the Kind cluster's name (default 'kind'): kubefox-demo
Would you like to initialize the 'hello-world' KubeFox app? [y/N] y
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

    info    If you don't have a Kubernetes cluster you can run one locally with Kind (https://kind.sigs.k8s.io)
    info    to experiment with KubeFox.

    info    Fox needs a place to store the KubeFox Component images it will build, normally
    info    this is a remote container registry. However, if you only want to use KubeFox
    info    locally with Kind you can skip this step.
    Are you only using KubeFox with local Kind cluster? [y/N] y
    Enter the Kind cluster's name (default 'kind'): kubefox-demo

    info    Configuration successfully written to '/home/xadhatter/.config/kubefox/config.yaml'.

    info    Congrats, you are ready to use KubeFox!
    info    Check out the quickstart for next steps (https://docs.kubefox.io/quickstart/).
    info    If you run into any problems please let us know on GitHub (https://github.com/xigxog/kubefox/issues).

    info    Let's initialize a KubeFox app!

    info    To get things started quickly 🦊 Fox can create a 'hello-world' KubeFox app which
    info    includes two components and example environments for testing.
    Would you like to initialize the 'hello-world' KubeFox app? [y/N] y

    info    KubeFox app initialization complete!
    ```

You should see some newly created directories and files. The `hello-world` app
includes two components and example environments. You might also notice Fox
initialized a new Git repo for you. Take a look around!

Let's get some environments created. Two sample environments are provided in the
`hack` dir. Let's take a quick look at them, then use `kubectl` to send them to
Kubernetes. Take note of the environment variable `who`.

```shell
cat hack/env-universe.yaml
cat hack/env-world.yaml
kubectl apply --filename hack/
```

??? example "Output"

    ```text
    $ cat hack/env-universe.yaml
    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: Environment
    metadata:
      name: universe
    spec:
      vars:
        who: Universe

    $ cat hack/env-world.yaml
    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: Environment
    metadata:
      name: world
    spec:
      vars:
        who: World

    $ kubectl apply --filename hack/
    environment.kubefox.xigxog.io/universe created
    environment.kubefox.xigxog.io/world created
    ```

Now let's deploy the app and see what happens. The following command will build
the component's OCI images, load them onto the kind cluster, and deploy them to
the KubeFox platform. The first run might take some time as it needs to download
dependencies, but future runs should be much faster. Remember, you can add the
`--verbose` flag for extra output if you want to see what's going on behind the
scenes. Note that the first deployment will initialize the platform so it might
take a couple minutes for all the pods to be ready. But don't worry, future
deployments will be very fast.

```shell
fox publish my-deployment --wait 5m
```

Answer the prompts:

```text
Would you like to create a KubeFox platform? [Y/n] y
Enter the KubeFox platform's name (required): dev
Enter the Kubernetes namespace of the KubeFox platform (default 'kubefox-dev'): kubefox-dev
```

??? example "Output"

    ```text
    $ fox publish my-deployment --wait 5m
    info    Building component image 'localhost/kubefox/hello-world/backend:68beae1'.
    info    Loading component image 'localhost/kubefox/hello-world/backend:68beae1' into Kind cluster 'kubefox-demo'.
    
    info    Building component image 'localhost/kubefox/hello-world/frontend:68beae1'.
    info    Loading component image 'localhost/kubefox/hello-world/frontend:68beae1' into Kind cluster 'kubefox-demo'.

    info    You need to have a KubeFox platform instance running to deploy your components.
    info    Don't worry, 🦊 Fox can create one for you.
    Would you like to create a KubeFox platform? [Y/n] y
    Enter the KubeFox platform's name (required): dev
    Enter the Kubernetes namespace of the KubeFox platform (default 'kubefox-dev'): kubefox-dev

    info    Component image 'localhost/kubefox/hello-world/backend:68beae1' exists.
    info    Component image 'localhost/kubefox/hello-world/frontend:68beae1' exists.

    info    Waiting for KubeFox platform 'dev' to be ready.
    info    Waiting for component 'backend' to be ready.
    info    Waiting for component 'frontend' to be ready.

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: Deployment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 1
      name: my-deployment
      namespace: kubefox-dev
      resourceVersion: "13326"
      uid: 5ad9a257-01c0-43e0-b6be-92757a47ba7c
    spec:
      app:
        commit: 68beae1
        containerRegistry: localhost/kubefox/hello-world
        description: A simple app demonstrating the use of KubeFox.
        gitRef: refs/heads/main
        name: hello-world
        title: Hello World
      components:
        backend:
          commit: 68beae1
          env: {}
        frontend:
          commit: 68beae1
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
    hello-world-backend-6c42fbb-8577fc876-bpf4j    1/1     Running   0          2s
    hello-world-frontend-6c42fbb-5d998f5cb-t9qp6   1/1     Running   0          2s
    ```

Awesome. When KubeFox deploys an app it starts the components but will not
automatically send requests to it until it is released. But you can still test
deployments by providing some context. KubeFox needs two pieces of information,
the deployment to use and the environment to inject. These can be passed as
headers or query parameters. Give it a shot.

```shell
curl "http://localhost:30080/hello?kf-dep=my-deployment&kf-env=world"
```

??? example "Output"

    ```text
    $ curl "http://localhost:30080/hello?kf-dep=my-deployment&kf-env=world"
    👋 Hello World!
    ```

Next try switching to the `universe` environment created earlier. With KubeFox
there is no need to create another deployment to switch environments, simply
change the query parameter!

```shell
curl "http://localhost:30080/hello?kf-dep=my-deployment&kf-env=universe"
```

??? example "Output"

    ```text
    $ curl "http://localhost:30080/hello?kf-dep=my-deployment&kf-env=universe"
    👋 Hello Universe!
    ```

## Release KubeFox App

Now let's release the app so we don't have to specify all those details in the
request. It is recommended to tag the repo for releases so we'll do that first,
and then switch to the tag for the release. It is important to understand that
Fox works against the active state of the Git repo. To deploy or release a
different version of your app simply checkout the tag, branch, or commit you
want and let Fox do the rest.

```shell
git tag v0.1.0
git checkout v0.1.0
fox release dev --env world --wait 5m
git switch -
```

??? example "Output"

    ```text
    $ git tag v0.1.0
    $ git checkout v0.1.0
    Note: switching to 'v0.1.0'.

    You are in 'detached HEAD' state. You can look around, make experimental
    changes and commit them, and you can discard any commits you make in this
    state without impacting any branches by switching back to a branch.

    If you want to create a new branch to retain commits you create, you may
    do so (now or later) by using -c with the switch command. Example:

      git switch -c <new-branch-name>

    Or undo this operation with:

      git switch -

    Turn off this advice by setting config variable advice.detachedHead to false

    HEAD is now at 6c42fbb And so it begins...

    $ fox release dev --env world --wait 5m
    info    Component image 'localhost/kubefox/hello-world/backend:68beae1' exists.
    info    Component image 'localhost/kubefox/hello-world/frontend:68beae1' exists.

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
          commit: 6c42fbb
          containerRegistry: localhost/kubefox/hello-world
          description: A simple app demonstrating the use of KubeFox.
          gitRef: refs/tags/v0.1.0
          name: hello-world
          title: Hello World
        components:
          backend:
            commit: 6c42fbb
            env: {}
          frontend:
            commit: 6c42fbb
            env: {}
      environment:
        name: world
    status: {}
    ```

Try the same request from above, but this time don't specify the context. Since
the app has been released the request will get matched by the component's route
and the context information will be automatically injected by KubeFox.

```shell
curl "http://localhost:30080/hello"
```

??? example "Output"

    ```text
    $ curl "http://localhost:30080/hello"
    👋 Hello World!
    ```

Let's take another look at the pods running on Kubernetes now that we performed
a release.

```shell
kubectl get pods --namespace kubefox-dev
```

??? example "Output"

    ```text
    $ kubectl get pods --namespace kubefox-dev
    NAME                                           READY   STATUS    RESTARTS   AGE
    dev-broker-grkcn                               1/1     Running   0          6m11s
    dev-nats-0                                     1/1     Running   0          6m17s
    hello-world-backend-6c42fbb-8577fc876-bpf4j    1/1     Running   0          6m1s
    hello-world-frontend-6c42fbb-5d998f5cb-t9qp6   1/1     Running   0          6m1s
    ```

Surprise, nothing has changed! KubeFox dynamically injects context at request
time, not at deployment time. That means it can continue to use the same pods.

Check out the rest of the docs for more. If you run into any problems please let
us know on [GitHub Issues](https://github.com/xigxog/kubefox/issues).
