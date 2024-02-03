# Quickstart

Welcome to the world of KubeFox! This technical guide will walk you through the
process of setting up a local Kubernetes Cluster using kind and deploying your
inaugural KubeFox App. From crafting Environments and deploying Apps to testing
and version control, we'll cover it all. Whether you're a seasoned developer or
just getting started, this guide will help you navigate the fundamentals of a
comprehensive software development lifecycle leveraging KubeFox. Let's dive in!

## Prerequisites

Ensure that the following tools are installed for this quickstart:

- [Docker](https://docs.docker.com/engine/install/) - A container toolset and
  runtime used to build KubeFox Components' OCI images and run a local
  Kubernetes Cluster via kind.
- [Fox](https://github.com/xigxog/kubefox-cli/releases/) - A CLI for
  communicating with the KubeFox Platform. Download the latest release and add
  the binary to your system's path. Or, if Go is installed, use `go install
  github.com/xigxog/fox@v0.8.0-alpha`.
- [Git](https://github.com/git-guides/install-git) - A distributed version
  control system.
- [Helm](https://helm.sh/docs/intro/install/) - Package manager for Kubernetes
  used to install the KubeFox Operator on Kubernetes.
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) - **K**uberentes
  **in** **D**ocker. A tool for running local Kubernetes Clusters using Docker
  container "nodes".
- [Kubectl](https://kubernetes.io/docs/tasks/tools/) - CLI for communicating
  with a Kubernetes Cluster's control plane, using the Kubernetes API.

Here are a few optional but recommended tools:

- [Go](https://go.dev/doc/install) - A programming language. The `hello-world`
  example App is written in Go, but Fox is able to compile it even without Go
  installed.
- [VS Code](https://code.visualstudio.com/download) - A lightweight but powerful
  source code editor. Helpful if you want to explore the `hello-world` App.

## Install KubeFox

Let's kick things off by setting up a local Kubernetes Cluster using kind. Use
the following command to create the Cluster.

```{ .shell .copy }
kind create cluster --wait 5m
```

??? example "Output"

    ```text
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

Now, install the KubeFox Helm Chart to initiate the KubeFox Operator on your
Kubernetes Cluster. The Operator manages KubeFox Platforms and Apps.

```{ .shell .copy }
helm upgrade kubefox kubefox \
  --repo https://xigxog.github.io/helm-charts \
  --create-namespace --namespace kubefox-system \
  --install --wait
```

??? example "Output"

    ```text
    Release "kubefox" does not exist. Installing it now.
    NAME: kubefox
    LAST DEPLOYED: Thu Jan  1 00:00:00 1970
    NAMESPACE: kubefox-system
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    ```

## Deploy

Awesome! You're all set to start the KubeFox Platform on the local kind Cluster
and deploy your first App. To begin, create a new directory and use Fox to
initialize the `hello-world` App. Run all subsequent commands from this
directory. The environment variable `FOX_INFO` tells Fox to to provide
additional output about what is going on. Employ the `--quickstart` flag to use
defaults and create a KubeFox Platform named `demo` in the `kubefox-demo`
Namespace.

```{ .shell .copy }
export FOX_INFO=true && \
  mkdir kubefox-quickstart && \
  cd kubefox-quickstart && \
  fox init --quickstart
```

??? example "Output"

    ```text
    info    Configuration successfully written to '/home/xadhatter/.config/kubefox/config.yaml'.

    info    Waiting for KubeFox Platform 'demo' to be ready...
    info    KubeFox initialized for the quickstart guide!
    ```

Notice the newly created directories and files. The `hello-world` App comprises
two Components, `frontend` and `backend`. There are also two example
Environments and VirtualEnvironments in the `hack/environments` directory.
Finally, Fox initialized a new Git repo for you. Take a look around!

Now, let's create some Environments and VirtualEnvironments. A
VirtualEnvironment inherits all specifications and data from its parent
Environment, but values can be overridden or added in the VirtualEnvironment. In
the provided examples the `subPath` variable is used to ensure unique routes
between contexts, as demonstrated in the `frontend` Component's `main.go` line
`17`.

```{ .go .no-copy linenums="17" }
k.Route("Path(`/{{.Vars.subPath}}/hello`)", sayHello)
```

Run the following command to examine the Environments and VirtualEnvironments
and apply them to Kubernetes using `kubectl`. Note the differences between the
two Environments' variables on lines numbered `11-12` and `30-31`.

```{ .shell .copy }
cat -b hack/environments/* && \
  kubectl apply --namespace kubefox-demo --filename hack/environments/
```

??? example "Output"

    ```text hl_lines="11 12 30 31"
        1 ---
        2 apiVersion: kubefox.xigxog.io/v1alpha1
        3 kind: Environment
        4 metadata:
        5   name: prod
        6 spec:
        7   releasePolicy:
        8     type: Stable
        9 data:
       10   vars:
       11     who: Universe
       12     subPath: prod
       13 ---
       14 apiVersion: kubefox.xigxog.io/v1alpha1
       15 kind: VirtualEnvironment
       16 metadata:
       17   name: prod
       18 spec:
       19   environment: prod
       20 ---
       21 apiVersion: kubefox.xigxog.io/v1alpha1
       22 kind: Environment
       23 metadata:
       24   name: qa
       25 spec:
       26   releasePolicy:
       27     type: Testing
       28 data:
       29   vars:
       30     who: World
       31     subPath: qa
       32 ---
       33 apiVersion: kubefox.xigxog.io/v1alpha1
       34 kind: VirtualEnvironment
       35 metadata:
       36   name: qa
       37 spec:
       38   environment: qa
    environment.kubefox.xigxog.io/prod created
    virtualenvironment.kubefox.xigxog.io/prod created
    environment.kubefox.xigxog.io/qa created
    virtualenvironment.kubefox.xigxog.io/qa created
    ```

Next deploy the `hello-world` App. Simply use the `publish` command, which not
only builds the OCI images for the Components but also loads them onto the kind
Cluster and deploys the App to the KubeFox Platform. You have the flexibility to
specify the name and version of the AppDeployment you're creating. We'll delve
into AppDeployment version later in this tutorial, so there's no need to worry
about it right now. If you don't provide a name, KubeFox defaults it to `<APP
NAME>-<GIT REF>`. In our case, it becomes `hello-world-main`. The initial run
might take a bit of time as it downloads dependencies, but subsequent runs will
be faster. If you want more detailed feedback, consider adding the `--verbose`
flag.

```{ .shell .copy }
fox publish --wait 5m
```

??? example "Output"

    ```text
    info    Building Component image 'localhost/kubefox/hello-world/backend:8bdd108ba636353020b95b75764b5edb18d5f914'.
    info    Loading Component image 'localhost/kubefox/hello-world/backend:8bdd108ba636353020b95b75764b5edb18d5f914' into kind cluster 'kind'.

    info    Building Component image 'localhost/kubefox/hello-world/frontend:8bdd108ba636353020b95b75764b5edb18d5f914'.
    info    Loading Component image 'localhost/kubefox/hello-world/frontend:8bdd108ba636353020b95b75764b5edb18d5f914' into kind cluster 'kind'.

    info    Waiting for KubeFox Platform 'demo' to be ready.
    info    Waiting for Component 'backend' to be ready.
    info    Waiting for Component 'frontend' to be ready.

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: AppDeployment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      finalizers:
      - kubefox.xigxog.io/release-protection
      generation: 1
      labels:
        app.kubernetes.io/name: hello-world
        kubefox.xigxog.io/app-branch: main
        kubefox.xigxog.io/app-commit: 8bdd108ba636353020b95b75764b5edb18d5f914
        kubefox.xigxog.io/app-commit-short: 8bdd108
      name: hello-world-main
      namespace: kubefox-demo
      resourceVersion: "13326"
      uid: 5ad9a257-01c0-43e0-b6be-92757a47ba7c
    details:
      description: A simple App demonstrating the use of KubeFox.
      title: Hello World
    spec:
      appName: hello-world
      branch: refs/heads/main
      commit: 8bdd108ba636353020b95b75764b5edb18d5f914
      commitTime: "1970-01-01T00:00:00Z"
      components:
        backend:
          commit: 8bdd108ba636353020b95b75764b5edb18d5f914
          defaultHandler: true
          envVarSchema:
            who:
              required: true
          type: KubeFox
        frontend:
          commit: 8bdd108ba636353020b95b75764b5edb18d5f914
          dependencies:
            backend:
              type: KubeFox
          routes:
          - envVarSchema:
              subPath:
                required: true
            id: 0
            rule: Path(`/{{.Vars.subPath}}/hello`)
          type: KubeFox
      containerRegistry: localhost/kubefox
    status:
      conditions:
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Component Deployments have minimum required Pods available.
        observedGeneration: 1
        reason: ComponentsAvailable
        status: "True"
        type: Available
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Component Deployments completed successfully.
        observedGeneration: 1
        reason: ComponentsDeployed
        status: "False"
        type: Progressing
    ```

Inspect what's running on Kubernetes.

```{ .shell .copy }
kubectl get pods --namespace kubefox-demo
```

??? example "Output"

    ```text
    NAME                                           READY   STATUS    RESTARTS   AGE
    demo-broker-grkcn                              1/1     Running   0          12s
    demo-httpsrv-7d8d946c57-rlt55                  1/1     Running   0          10s
    demo-nats-0                                    1/1     Running   0          18s
    hello-world-backend-8bdd108-577868c97b-29q2k   1/1     Running   0          2s
    hello-world-frontend-8bdd108-65fb98f59d-ll4sf  1/1     Running   0          2s
    ```

The Pods for two Components you deployed were created, `backend` and `frontend`.
The `broker`, `httpsrv`, and `nats` Pods are part of the KubeFox Platform
initiated by the Operator during Platform creation.

Typically, connections to KubeFox Apps are made through a public-facing load
balancer. For the simplicity of this guide use Fox to create a local proxy
instead. In a new terminal run the following command.

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
    HTTP proxy started on http://127.0.0.1:8080
    ```

Now, back in the original terminal, test the AppDeployment. KubeFox won't route
requests to the App until it's released, but you can still test AppDeployments
by manually providing context. KubeFox needs two pieces of information to route
an event, the AppDeployment to use and the VirtualEnvironment to inject. These
can be passed as headers or query parameters for HTTP requests.

```{ .shell .copy }
curl "http://localhost:8080/qa/hello?kf-dep=hello-world-main&kf-env=qa"
```

??? example "Output"

    ```text
    👋 Hello World!
    ```

Try switching to the `prod` VirtualEnvironment — this can be done seamlessly
with KubeFox without creating another AppDeployment. This is possible because
KubeFox injects context at request time instead of at deployment. Adding
VirtualEnvironments has nearly zero overhead! Be sure to change the URL path
from `/qa/hello` to `/prod/hello` to reflect the change of the `subPath`
variable.

```{ .shell .copy }
curl "http://localhost:8080/prod/hello?kf-dep=hello-world-main&kf-env=prod"
```

??? example "Output"

    ```text
    👋 Hello Universe!
    ```

## Release

To have KubeFox automatically route requests without manually specifying context
you need to create a Release. Once a AppDeployment is released, KubeFox will
match requests to Components' routes and automatically inject context. Before
creating a Release it is recommended to publish a versioned AppDeployment and
tag the Git repo. Unlike normal AppDeployments, which can be updated freely,
versioned AppDeployments are immutable. They provide a stable deployment that
can be promoted to higher VirtualEnvironments.

Tag the Git repo, publish the versioned AppDeployment, and release it to the
`qa` Environment with this command.

```{ .shell .copy }
fox publish --version v1 --create-tag && \
  fox release v1 --virtual-env qa
```

??? example "Output"

    ```text
    info    Component image 'localhost/kubefox/hello-world/backend:8bdd108ba636353020b95b75764b5edb18d5f914' exists, skipping build.
    info    Loading Component image 'localhost/kubefox/hello-world/backend:8bdd108ba636353020b95b75764b5edb18d5f914' into kind cluster 'kind'.

    info    Component image 'localhost/kubefox/hello-world/frontend:8bdd108ba636353020b95b75764b5edb18d5f914' exists, skipping build.
    info    Loading Component image 'localhost/kubefox/hello-world/frontend:8bdd108ba636353020b95b75764b5edb18d5f914' into kind cluster 'kind'.

    info    Creating tag 'v1'.

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: AppDeployment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      finalizers:
      - kubefox.xigxog.io/release-protection
      generation: 1
      labels:
        app.kubernetes.io/name: hello-world
        kubefox.xigxog.io/app-branch: main
        kubefox.xigxog.io/app-commit: 8bdd108ba636353020b95b75764b5edb18d5f914
        kubefox.xigxog.io/app-commit-short: 8bdd108
        kubefox.xigxog.io/app-tag: v1
        kubefox.xigxog.io/app-version: v1
      name: hello-world-v1
      namespace: kubefox-demo
      resourceVersion: "2257050"
      uid: 782a0938-7f9d-4bae-a6b5-900499fca6f7
    details:
      description: A simple App demonstrating the use of KubeFox.
      title: Hello World
    spec:
      appName: hello-world
      branch: refs/heads/main
      commit: 8bdd108ba636353020b95b75764b5edb18d5f914
      commitTime: "1970-01-01T00:00:00Z"
      components:
        backend:
          commit: 8bdd108ba636353020b95b75764b5edb18d5f914
          defaultHandler: true
          envVarSchema:
            who:
              required: true
          type: KubeFox
        frontend:
          commit: 8bdd108ba636353020b95b75764b5edb18d5f914
          dependencies:
            backend:
              type: KubeFox
          routes:
          - envVarSchema:
              subPath:
                required: true
            id: 0
            rule: Path(`/{{.Vars.subPath}}/hello`)
          type: KubeFox
      containerRegistry: localhost/kubefox
      tag: refs/tags/v1
      version: v1
    status:
      conditions:
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Component Deployments have minimum required Pods available.
        observedGeneration: 1
        reason: ComponentsAvailable
        status: "True"
        type: Available
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Component Deployments completed successfully.
        observedGeneration: 1
        reason: ComponentsDeployed
        status: "False"
        type: Progressing

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: VirtualEnvironment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      finalizers:
      - kubefox.xigxog.io/environment-protection
      generation: 2
      labels:
        kubefox.xigxog.io/environment: qa
      name: qa
      namespace: kubefox-demo
      resourceVersion: "5643"
      uid: f7d7a42f-bc3f-46f3-8ec4-c557e1a84fe3
    spec:
      environment: qa
      release:
        apps:
          hello-world:
            appDeployment: hello-world-v1
            version: v1
    status:
      activeRelease:
        activationTime: "1970-01-01T00:00:00Z"
        apps:
          hello-world:
            appDeployment: hello-world-v1
            version: v1
        id: 011aa5fd-0ed3-4920-8564-bb793435e97c
        requestTime: "1970-01-01T00:00:00Z"
      conditions:
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Release AppDeployments are available, Routes and Adapters are valid and
          compatible with the VirtualEnv.
        observedGeneration: 2
        reason: ContextAvailable
        status: "True"
        type: ActiveReleaseAvailable
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Release was activated.
        observedGeneration: 2
        reason: ReleaseActivated
        status: "False"
        type: ReleasePending
    ```

Test the same request as before, but this time without specifying context. Since
the App has been released, the request is matched by the Component's route, and
context information is automatically applied.

```{ .shell .copy }
curl "http://localhost:8080/qa/hello"
```

??? example "Output"

    ```text
    👋 Hello World!
    ```

Inspect the Pods running on Kubernetes now that you created another
AppDeployment and Release.

```{ .shell .copy }
kubectl get pods --namespace kubefox-demo
```

??? example "Output"

    ```text
    NAME                                           READY   STATUS    RESTARTS   AGE
    demo-broker-grkcn                              1/1     Running   0          6m11s
    demo-httpsrv-7d8d946c57-rlt55                  1/1     Running   0          6m9s
    demo-nats-0                                    1/1     Running   0          6m17s
    hello-world-backend-8bdd108-577868c97b-29q2k   1/1     Running   0          6m1s
    hello-world-frontend-8bdd108-65fb98f59d-ll4sf  1/1     Running   0          6m1s
    ```

Surprisingly, nothing has changed in the Pods running on Kubernetes. KubeFox
dynamically injects context per request, just like when you changed
VirtualEnvironments earlier with the query parameters.

## Update

Next, make a modification to the `frontend` Component, commit the changes, and
deploy. Open up `components/frontend/main.go` in your favorite editor and update
line `28` in the `sayHello` function to say something new.

```go linenums="22" hl_lines="7"
func sayHello(k kit.Kontext) error {
    r, err := k.Req(backend).Send()
    if err != nil {
        return err
    }

    msg := fmt.Sprintf("👋 Hello %s!", r.Str()) //(1)
    json := map[string]any{"msg": msg}
    html := fmt.Sprintf(htmlTmpl, msg)
    k.Log().Debug(msg)

    return k.Resp().SendAccepts(json, html, msg)
}
```

1. Update me to say `Hey` instead of `Hello`.

Fox operates against the current commit of the Git repo when deploying
Components. That means before deploying you need to commit the changes to record
them. You can then re-deploy the `main` AppDeployment and test. Check out how
the new commit hash is used to provide unique identifiers for Components. The
applicable lines are highlighted in the output below.

```{ .shell .copy }
git add . && \
  git commit -m "updated frontend to say Hey" && \
  fox publish --wait 5m
```

??? example "Output"

    ```text hl_lines="1 7 8 34 45"
    [main 6dcc993] updated frontend to say Hey
    1 file changed, 1 insertion(+)

    info    Component image 'localhost/kubefox/hello-world/backend:8bdd108ba636353020b95b75764b5edb18d5f914' exists, skipping build.
    info    Loading Component image 'localhost/kubefox/hello-world/backend:8bdd108ba636353020b95b75764b5edb18d5f914' into kind cluster 'kind'.

    info    Building Component image 'localhost/kubefox/hello-world/frontend:6dcc9937126fccaf119acae8785e3ab90808a998'.
    info    Loading Component image 'localhost/kubefox/hello-world/frontend:6dcc9937126fccaf119acae8785e3ab90808a998' into kind cluster 'kind'.

    info    Waiting for KubeFox Platform 'demo' to be ready...
    info    Waiting for Component 'backend' to be ready...
    info    Waiting for Component 'frontend' to be ready...

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: AppDeployment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      generation: 2
      labels:
        app.kubernetes.io/name: hello-world
        kubefox.xigxog.io/app-branch: main
        kubefox.xigxog.io/app-commit: 6dcc9937126fccaf119acae8785e3ab90808a998
        kubefox.xigxog.io/app-commit-short: 6dcc993
      name: hello-world-main
      namespace: kubefox-demo
      resourceVersion: "2258944"
      uid: 5ad9a257-01c0-43e0-b6be-92757a47ba7c
    details:
      description: A simple App demonstrating the use of KubeFox.
      title: Hello World
    spec:
      appName: hello-world
      branch: refs/heads/main
      commit: 6dcc9937126fccaf119acae8785e3ab90808a998
      commitTime: "1970-01-01T00:00:00Z"
      components:
        backend:
          commit: 8bdd108ba636353020b95b75764b5edb18d5f914
          defaultHandler: true
          envVarSchema:
            who:
              required: true
          type: KubeFox
        frontend:
          commit: 6dcc9937126fccaf119acae8785e3ab90808a998
          dependencies:
            backend:
              type: KubeFox
          routes:
          - envVarSchema:
              subPath:
                required: true
            id: 0
            rule: Path(`/{{.Vars.subPath}}/hello`)
          type: KubeFox
      containerRegistry: localhost/kubefox
    status:
      conditions:
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Component Deployments have minimum required Pods available.
        observedGeneration: 2
        reason: ComponentsAvailable
        status: "True"
        type: Available
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Component Deployments completed successfully.
        observedGeneration: 2
        reason: ComponentsDeployed
        status: "False"
        type: Progressing
    ```

Fox didn't rebuild the `backend` Component as no changes were made. Try testing
out the current Release and latest AppDeployment.

```{ .shell .copy }
echo -ne "VERSION\tVIRTENV\tOUTPUT" && \
  echo -ne "\nv1\tqa\t"; curl "http://localhost:8080/qa/hello" && \
  echo -ne "\nv1\tprod\t"; curl "http://localhost:8080/prod/hello?kf-dep=hello-world-v1&kf-env=prod" && \
  echo -ne "\nlatest\tqa\t"; curl "http://localhost:8080/qa/hello?kf-dep=hello-world-main&kf-env=qa" && \
  echo -ne "\nlatest\tprod\t"; curl "http://localhost:8080/prod/hello?kf-dep=hello-world-main&kf-env=prod"
```

??? example "Output"

    ```text
    VERSION VIRTENV OUTPUT
    v1      qa      👋 Hello World!
    v1      prod    👋 Hello Universe!
    latest  qa      👋 Hey World!
    latest  prod    👋 Hey Universe!
    ```

Now that you've created a new AppDeployment, take another look at the Pods
running on Kubernetes.

```{ .shell .copy }
kubectl get pods --namespace kubefox-demo
```

??? example "Output"

    ```text
    NAME                                           READY   STATUS    RESTARTS   AGE
    demo-broker-grkcn                              1/1     Running   0          6m11s
    demo-httpsrv-7d8d946c57-rlt55                  1/1     Running   0          6m9s
    demo-nats-0                                    1/1     Running   0          6m17s
    hello-world-backend-8bdd108-577868c97b-29q2k   1/1     Running   0          6m1s
    hello-world-frontend-6dcc993-7c9895669d-mj6vc  1/1     Running   0          18s
    hello-world-frontend-8bdd108-65fb98f59d-ll4sf  1/1     Running   0          6m1s
    ```

You might be surprised to find only three Component Pods running to support the
two AppDeployments and Release. Because the `backend` Component did not change
between AppDeployments, KubeFox shares a single Pod. Based on the context
applied at request time, routing to the correct Component version is dynamically
performed.

## Promote

Finally, publish the new version of the App, release it to the `qa`
VirtualEnvironment, and then promote version `v1` to the `prod`
VirtualEnvironment. Stricter policies of the `prod` VirtualEnvironment require
the Release to be stable. To achieve this, the Operator creates a
ReleaseManifest when it activates the Release. A ReleaseManifest is an immutable
snapshot of all constituents of the Release; the Environment,
VirtualEnvironment, Adapters, and AppDeployments. This ensures any changes made
to these resources after the Release is activated do not affect the Release.

Release the new version to `qa` and promote `v1` to `prod`.

```{ .shell .copy }
fox publish --version v2 --create-tag && \
  fox release v2 --virtual-env qa && \
  fox release v1 --virtual-env prod
```

??? example "Output"

    ```text
    info    Component image 'localhost/kubefox/hello-world/backend:8bdd108ba636353020b95b75764b5edb18d5f914' exists.
    info    Loading Component image 'localhost/kubefox/hello-world/backend:8bdd108ba636353020b95b75764b5edb18d5f914' into kind cluster 'kind'.

    info    Component image 'localhost/kubefox/hello-world/frontend:6dcc9937126fccaf119acae8785e3ab90808a998' exists.
    info    Loading Component image 'localhost/kubefox/hello-world/frontend:6dcc9937126fccaf119acae8785e3ab90808a998' into kind cluster 'kind'.

    info    Creating tag 'v2'.

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: AppDeployment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      finalizers:
      - kubefox.xigxog.io/release-protection
      generation: 1
      labels:
        app.kubernetes.io/name: hello-world
        kubefox.xigxog.io/app-branch: main
        kubefox.xigxog.io/app-commit: 6dcc9937126fccaf119acae8785e3ab90808a998
        kubefox.xigxog.io/app-commit-short: 6dcc993
        kubefox.xigxog.io/app-tag: v2
        kubefox.xigxog.io/app-version: v2
      name: hello-world-v2
      namespace: kubefox-demo
      resourceVersion: "2259758"
      uid: 7d3e6c48-71bd-428f-bd3a-245f73344538
    details:
      description: A simple App demonstrating the use of KubeFox.
      title: Hello World
    spec:
      appName: hello-world
      branch: refs/heads/main
      commit: 6dcc9937126fccaf119acae8785e3ab90808a998
      commitTime: "1970-01-01T00:00:00Z"
      components:
        backend:
          commit: 8bdd108ba636353020b95b75764b5edb18d5f914
          defaultHandler: true
          envVarSchema:
            who:
              required: true
          type: KubeFox
        frontend:
          commit: 6dcc9937126fccaf119acae8785e3ab90808a998
          dependencies:
            backend:
              type: KubeFox
          routes:
          - envVarSchema:
              subPath:
                required: true
            id: 0
            rule: Path(`/{{.Vars.subPath}}/hello`)
          type: KubeFox
      containerRegistry: localhost/kubefox
      tag: refs/tags/v2
      version: v2
    status:
      conditions:
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Component Deployments have minimum required Pods available.
        observedGeneration: 1
        reason: ComponentsAvailable
        status: "True"
        type: Available
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Component Deployments completed successfully.
        observedGeneration: 1
        reason: ComponentsDeployed
        status: "False"
        type: Progressing

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: VirtualEnvironment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      finalizers:
      - kubefox.xigxog.io/environment-protection
      generation: 3
      labels:
        kubefox.xigxog.io/environment: qa
      name: qa
      namespace: kubefox-demo
      resourceVersion: "5643"
      uid: f7d7a42f-bc3f-46f3-8ec4-c557e1a84fe3
    spec:
      environment: qa
      release:
        apps:
          hello-world:
            appDeployment: hello-world-v2
            version: v2
    status:
      activeRelease:
        activationTime: "1970-01-01T00:00:00Z"
        apps:
          hello-world:
            appDeployment: hello-world-v2
            version: v2
        id: 41d473dc-ccff-4878-9daa-abaf2db8d104
        requestTime: "1970-01-01T00:00:00Z"
      conditions:
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Release AppDeployments are available, Routes and Adapters are valid and
          compatible with the VirtualEnv.
        observedGeneration: 3
        reason: ContextAvailable
        status: "True"
        type: ActiveReleaseAvailable
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Release was activated.
        observedGeneration: 3
        reason: ReleaseActivated
        status: "False"
        type: ReleasePending
      releaseHistory:
      - activationTime: "1970-01-01T00:00:00Z"
        apps:
          hello-world:
            appDeployment: hello-world-v1
            version: v1
        archiveReason: Superseded
        archiveTime: "1970-01-01T00:00:00Z"
        id: 011aa5fd-0ed3-4920-8564-bb793435e97c
        requestTime: "1970-01-01T00:00:00Z"

    apiVersion: kubefox.xigxog.io/v1alpha1
    kind: VirtualEnvironment
    metadata:
      creationTimestamp: "1970-01-01T00:00:00Z"
      finalizers:
      - kubefox.xigxog.io/environment-protection
      generation: 2
      labels:
        kubefox.xigxog.io/environment: prod
      name: prod
      namespace: kubefox-demo
      resourceVersion: "5750"
      uid: 4980c5d9-208d-420b-9133-9ff3468eb7e6
    spec:
      environment: prod
      release:
        apps:
          hello-world:
            appDeployment: hello-world-v1
            version: v1
    status:
      activeRelease:
        activationTime: "1970-01-01T00:00:00Z"
        apps:
          hello-world:
            appDeployment: hello-world-v1
            version: v1
        id: 18875dc4-d347-45be-8221-c0dbc0c1b2e5
        releaseManifest: prod-5747-20240201-162421
        requestTime: "1970-01-01T00:00:00Z"
      conditions:
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Release AppDeployments are available, Routes and Adapters are valid and
          compatible with the VirtualEnv.
        observedGeneration: 2
        reason: ContextAvailable
        status: "True"
        type: ActiveReleaseAvailable
      - lastTransitionTime: "1970-01-01T00:00:00Z"
        message: Release was activated.
        observedGeneration: 2
        reason: ReleaseActivated
        status: "False"
        type: ReleasePending
    ```

Give the new Releases a spin! Notice the new output from the updated `frontend`
Component in `qa`, while `prod` still displays the original output from `v1`.
With Releases in both VirtualEnvironments there's no need to manually specify
the context.

```{ .shell .copy }
echo -ne "VERSION\tVIRTENV\tOUTPUT" && \
  echo -ne "\nv1\tprod\t"; curl "http://localhost:8080/prod/hello" && \
  echo -ne "\nv2\tqa\t"; curl "http://localhost:8080/qa/hello"
```

??? example "Output"

    ```text
    VERSION VIRTENV OUTPUT
    v1      prod    👋 Hello Universe!
    v2      qa      👋 Hey World!
    ```

Just for fun take one final look at all that you've created.

```{ .shell .copy}
kubectl api-resources --output name | \
  grep -e kubefox -e pods | \
  xargs -I % sh -c 'echo; kubectl get --show-kind --ignore-not-found --namespace kubefox-demo %'
```

??? example "Output"

    ```text
    NAME                                               READY   STATUS    RESTARTS   AGE
    pod/demo-broker-grkcn                              1/1     Running   0          6m21s
    pod/demo-httpsrv-7d8d946c57-rlt55                  1/1     Running   0          6m19s
    pod/demo-nats-0                                    1/1     Running   0          6m27s
    pod/hello-world-backend-8bdd108-577868c97b-29q2k   1/1     Running   0          6m11s
    pod/hello-world-frontend-6dcc993-7c9895669d-mj6vc  1/1     Running   0          28s
    pod/hello-world-frontend-8bdd108-65fb98f59d-ll4sf  1/1     Running   0          6m11s

    NAME                                                 APP           VERSION   AVAILABLE   REASON                PROGRESSING
    appdeployment.kubefox.xigxog.io/hello-world-main     hello-world             True        ComponentsAvailable   False
    appdeployment.kubefox.xigxog.io/hello-world-v1       hello-world   v1        True        ComponentsAvailable   False
    appdeployment.kubefox.xigxog.io/hello-world-v2       hello-world   v2        True        ComponentsAvailable   False

    NAME                                 AGE
    environment.kubefox.xigxog.io/prod   54m
    environment.kubefox.xigxog.io/qa     54m

    NAME                              AVAILABLE   EVENT TIMEOUT   EVENT MAX   LOG LEVEL
    platform.kubefox.xigxog.io/demo   True        30              5242880     info

    NAME                                                          ID                                     ENVIRONMENT   VIRTUALENVIRONMENT
    releasemanifest.kubefox.xigxog.io/prod-5747-20240201-162421   18875dc4-d347-45be-8221-c0dbc0c1b2e5   prod          prod

    NAME                                        ENVIRONMENT   MANIFEST                    AVAILABLE   REASON             PENDING   PENDING REASON
    virtualenvironment.kubefox.xigxog.io/prod   prod          prod-5747-20240201-162421   True        ContextAvailable   False     ReleaseActivated
    virtualenvironment.kubefox.xigxog.io/qa     qa                                        True        ContextAvailable   False     ReleaseActivated
    ```

Explore the rest of the documentation for more details. If you encounter any
problems please let us know on [GitHub
Issues](https://github.com/xigxog/kubefox/issues).
