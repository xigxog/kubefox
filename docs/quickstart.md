# Quickstart

Welcome to the world of KubeFox! This technical guide will walk you through the
process of setting up a Kubernetes cluster using either kind or Azure and deploying your
inaugural KubeFox app. From crafting environments and deploying apps to testing
and version control, we'll cover it all. Whether you're a seasoned developer or
just getting started, this guide will help you navigate the fundamentals of a
comprehensive software development lifecycle leveraging KubeFox. Let's dive in!

## Prerequisites

Ensure that the following tools are installed for this quickstart:

- [Docker](https://docs.docker.com/engine/install/) - A container toolset and
  runtime used to build KubeFox components' OCI images and run a local
  Kubernetes cluster via kind.
- [Fox](https://github.com/xigxog/kubefox-cli/releases/) - A CLI for
  communicating with the KubeFox platform. Download the latest release and add
  the binary to your system's path.
- [Git](https://github.com/git-guides/install-git) - A distributed version
  control system.
- [Helm](https://helm.sh/docs/intro/install/) - Package manager for Kubernetes
  used to install the KubeFox platform on Kubernetes.
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) - **K**uberentes
  **in** **D**ocker. A tool for running local Kubernetes clusters using Docker
  container "nodes".
- [Kubectl](https://kubernetes.io/docs/tasks/tools/) - CLI for communicating
  with a Kubernetes cluster's control plane, using the Kubernetes API.

Here are a few optional but recommended tools:

- [Go](https://go.dev/doc/install) - A programming language. The `hello-world`
  sample app is written in Go, but Fox is able to compile it even without Go
  installed.
- [VS Code](https://code.visualstudio.com/download) - A lightweight but powerful
  source code editor. Helpful if you want to explore the `hello-world` app.
- [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) - CLI for communicating
  with the Azure control plane.  

## Setup Kubernetes

Let's kick things off by setting up our first Kubernetes cluster. Use
the below set of commands depending on which Kubernetes provider you would like to use.
If you already have a Kubernetes Cluster provisioned, you can skip this step.

=== "Local (kind)"
    Let's start with setting up a local Kubernetes cluster using the kind CLI. The following commands can be used to create and interact with the cluster.

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

=== "Azure (AKS)"
    Let's start with setting up a remote Kubernetes cluster using the Azure CLI. The following commands can be used to create and interact with the cluster.
    First you need to login to Azure.

    ```{ .shell .copy }
    az login
    ```
    Lets set the other required variables for this quickstart on Azure. Make sure to modify the below accordingly before pasting into your terminal.
    ```{ .shell .copy }
    export AZ_LOCATION=eastus2 && \
    export AZ_RESOURCE_GROUP=kf-quickstart-infra-eus2-rg && \
    export AZ_AKS_NAME=kf-quickstart-eus2-aks-01
    ```
    Next you need to create a Resource Group for the AKS cluster, and then deploy AKS within this Resource Group.

    ```{ .shell .copy }
    az group create --location ${AZ_LOCATION} --name ${AZ_RESOURCE_GROUP}
    ```
    ??? example "Output"

        ```text
        $ az group create --location ${AZ_LOCATION} --name ${AZ_RESOURCE_GROUP}
        {
          "id": "/subscriptions/00000000-0000-0000-0000-00000000/resourceGroups/kf-quickstart-infra-eus2-rg",
          "location": "eastus2",
          "managedBy": null,
          "name": "kf-quickstart-infra-eus2-rg",
          "properties": {
            "provisioningState": "Succeeded"
          },
          "tags": null,
          "type": "Microsoft.Resources/resourceGroups"
        }
        ```
    *Note: This command to create your AKS cluster will take > 5 minutes to complete.*
    ```{ .shell .copy }
    az aks create \
        --resource-group ${AZ_RESOURCE_GROUP} \
        --tier free \
        --name ${AZ_AKS_NAME} \
        --location ${AZ_LOCATION} \
        --node-count 1 \
        --node-vm-size "Standard_B2s"
    ```
    Once your AKS cluster is ready, you can use the below command to add the cluster to your kubectl configuration.

    ```{ .shell .copy }
    az aks get-credentials \
      --resource-group ${AZ_RESOURCE_GROUP}  \
      --name ${AZ_AKS_NAME}
    ```

    Please run the below command if you need to setup a container registry on Azure using ACR. If you already have a registry which is accessible from Azure, you can skip this step. Any OCI compliant container registry can be used here (ex DockerHub, GHCR, Harbor).

    ??? "Create Container Registry on Azure (optional)"
        Create your Azure Container Registry in the same resource group as your AKS cluster.
        Please modify the below to set the name of your ACR in the AZ_ACR_NAME environment variable, otherwise it will auto-generate a random numeric string.
    
        ```{ .shell .copy }
        export AZ_ACR_NAME="acr${RANDOM}" && \
          az acr create --name ${AZ_ACR_NAME} --sku Basic --admin-enabled true --resource-group ${AZ_RESOURCE_GROUP}
        ```
        Run the below commands to set the endpoints and token required to access your registry. Save this output somewhere safe.
        These environment variables will be used by the fox CLI later to configure your environment.
        
        ```{ .shell .copy }
        FOX_REGISTRY_ADDRESS=$(az acr show-endpoints -n ${AZ_ACR_NAME} --resource-group ${AZ_RESOURCE_GROUP} --output tsv --query loginServer) && \
          FOX_REGISTRY_TOKEN=$(az acr login --name ${AZ_ACR_NAME} --expose-token --output tsv --query accessToken) && \
          FOX_REGISTRY_USERNAME="00000000-0000-0000-0000-000000000000"  
        ```

## Install KubeFox Platform

Now, install the KubeFox Helm Chart to initiate the KubeFox operator on your
Kubernetes cluster. The operator manages KubeFox platforms and apps.

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

## Setup Fox

Here we will setup the Fox CLI.

=== "Install using Go"

    ```{ .shell .copy }
    go install github.com/xigxog/fox@alpha-v0.8.0
    ```

=== "Install using Bash"

    ```{ .shell .copy }
    curl -sL "https://github.com/xigxog/fox/releases/latest/download/fox-$(uname -s | tr 'A-Z' 'a-z')-amd64.tar.gz" | tar xvz - -C /tmp && sudo mv /tmp/fox /usr/local/bin/fox 
    ```

  To begin, create a new directory and use Fox to initialize the `hello-world` app. Run all subsequent
commands from this directory. The environment variable `FOX_INFO` tells Fox to
to provide additional output about what is going on. Employ the `--quickstart`
flag to use defaults and create a KubeFox platform named `demo` in the
`kubefox-demo` namespace.

??? "Remote Registry Note"
    If you are using a remote registry, please ensure the following environment variables are set before running the quickstart:  FOX_REGISTRY_ADDRESS, FOX_REGISTRY_TOKEN, and FOX_REGISTRY_USERNAME. If you setup the Azure Container Registry as a part of this quickstart, these variables will already be set.

```{ .shell .copy }
export FOX_INFO=true && \
  mkdir kubefox-quickstart && \
  cd kubefox-quickstart && \
  fox init --quickstart
```

??? example "Output"

    ```text
    $ export FOX_INFO=true && \
        mkdir kubefox-quickstart && \
        cd kubefox-quickstart && \
        fox init --quickstart
    info    Waiting for KubeFox Platform 'demo' to be ready...
    info    KubeFox initialized for the quickstart guide!
    ```


Notice the newly created directories and files. The `hello-world` app comprises
two components and example environments. Fox also initialized a new Git repo for
you. Take a look around!

## Deploy

Awesome! You're all set to create your first KubeFox app and deploy it to a
KubeFox platform running on your kubernetes cluster. 

Now, let's create some environments. Two example environments are available in
the `hack` directory. The `subPath` variable ensures unique routes between
environments, as demonstrated in the frontend component's `main.go` line 21.

```{ .go .no-copy linenums="21" }
k.Route("Path(`/{{.Env.subPath}}/hello`)", sayHello)
```

Run the following command to examine the environments and apply them to
Kubernetes using `kubectl`. Note the differences between the two environments'
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

Next, deploy the `hello-world` app. The `publish` command builds the component's
OCI images, loads them onto the kind cluster, and deploys the app to the KubeFox
platform. The initial run might take some time as it downloads dependencies, but
subsequent runs will be faster. Optionally add the the `--verbose` flag for
extra output.

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
              required: true
              type: string
              unique: true
          routes:
          - id: 0
            rule: Path(`/{{.Env.subpath}}/hello`)
    status:
      available: false
    ```

Inspect what's running on Kubernetes.

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

The pods for two components you deployed were created, `hello-world-backend` and
`hello-world-frontend`. The `broker`, `httpsrv`, and `nats` pods are part of the
KubeFox platform initiated by the operator during platform creation.

Typically, connections to KubeFox apps are made through a public-facing load
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
    $ fox proxy 8080
    HTTP proxy started on http://127.0.0.1:8080
    ```

Now, back in the original terminal, test the deployment. KubeFox won't route
requests to the app until it's released, but you can still test deployments by
manually providing context. KubeFox needs two pieces of information to route an
event, the deployment to use and the environment to inject. These can be passed
as headers or query parameters.

```{ .shell .copy }
curl "http://localhost:8080/qa/hello?kf-dep=main&kf-env=qa"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/qa/hello?kf-dep=main&kf-env=qa"
    👋 Hello World!
    ```

Try switching to the `prod` environment — this can be done seamlessly with
KubeFox without creating another deployment. This is possible because KubeFox
injects context at request time instead of at deployment. Adding environments
has nearly zero overhead! Be sure to change the `subPath` from `qa` to `prod` to
reflect the change of environment.

```{ .shell .copy }
curl "http://localhost:8080/prod/hello?kf-dep=main&kf-env=prod"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/prod/hello?kf-dep=main&kf-env=prod"
    👋 Hello Universe!
    ```

## Release

To have KubeFox automatically route requests without specifying context you need
to create a release. Once a deployment is released, KubeFox will match requests
to components' routes and automatically inject context. Before creating a
release it is recommended to publish a versioned deployment and tag the Git
repo. Unlike normal deployments, which can be updated freely, versioned
deployments are immutable. They provide a stable deployment that can be promoted
to higher environments.

Tag the Git repo, publish the versioned deployment, and release it to the `qa`
environment with this command.

```{ .shell .copy }
fox publish --version v0.1.0 --create-tag && \
  fox release v0.1.0 --virtual-env qa --wait 5m
```

??? example "Output"

    ```text
    $ fox publish --version v0.1.0 --create-tag && \
        fox release v0.1.0 --virtual-env qa --wait 5m
    info    Component image 'localhost/kubefox/hello-world/backend:bb702a1' exists, skipping build.
    info    Loading component image 'localhost/kubefox/hello-world/backend:bb702a1' into kind cluster 'kind'.

    info    Component image 'localhost/kubefox/hello-world/frontend:bb702a1' exists, skipping build.
    info    Loading component image 'localhost/kubefox/hello-world/frontend:bb702a1' into kind cluster 'kind'.

    info    Creating tag 'v0.1.0'.

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
              required: true
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

Test the same request as before, but this time without specifying context. Since
the app has been released, the request is matched by the component's route, and
context information is automatically applied.

```{ .shell .copy }
curl "http://localhost:8080/qa/hello"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/qa/hello"
    👋 Hello World!
    ```

Inspect the pods running on Kubernetes now that you performed another deployment
and release.

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

Surprisingly, nothing has changed in the pods running on Kubernetes. KubeFox
dynamically injects context per request, just like when you changed environments
earlier with the query parameters.

## Update

Next, make a modification to the `frontend` component, commit the changes, and
deploy. Open up `components/frontend/main.go` in your favorite editor and update
line 32 in the `sayHello` function to say something new.

```go linenums="25" hl_lines="7"
func sayHello(k kit.Kontext) error {
    r, err := k.Req(backend).Send()
    if err != nil {
        return err
    }

    msg := fmt.Sprintf("👋 Hello %s!", r.Str()) //(1)
    k.Log().Debug(msg)

    json := map[string]any{"msg": msg}
    html := fmt.Sprintf(htmlTmpl, msg)
    return k.Resp().SendAccepts(json, html, msg)
}
```

1. Update me to say `Hey` instead of `Hello`.

Fox operates against the current commit of the Git repo when deploying
components. That means before deploying you need to commit the changes to record
them. You can then re-deploy the `main` deployment and test. Take note of the
commit hash, in the example output below it is `780e2db`. Commit hashes are used
to version components.

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
              required: true
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

??? example "Output"

    ```text
    $ curl "http://localhost:8080/qa/hello?kf-dep=main&kf-env=qa"
    👋 Hey World!
    ```

```{ .shell .copy }
curl "http://localhost:8080/prod/hello?kf-dep=main&kf-env=prod"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/prod/hello?kf-dep=main&kf-env=prod"
    👋 Hey Universe!
    ```

```{ .shell .copy }
curl "http://localhost:8080/qa/hello"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/qa/hello"
    👋 Hello World!
    ```

Inspect the pods running on Kubernetes with the new deployment.

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
between deployments, KubeFox shares a single pod. Based on the context applied
at request time, routing to the correct component versions is dynamically
performed.

## Promote

Finally, publish the new version of the app, release it to the `qa`
environment, and then promote version `v0.1.0` to the `prod` environment.
Stricter policies of the `prod` environment require an environment snapshot to
exist for a release. Similar to a versioned deployment, an environment snapshot
is immutable ensuring a stable release even if the source environment is
changed.

Release the new version to `qa`.

```{ .shell .copy }
fox publish --version v0.1.1 --create-tag && \
  fox release v0.1.1 --virtual-env qa --wait 5m
```

??? example "Output"

    ```text
    $ fox publish --version v0.1.1 --create-tag && \
        fox release v0.1.1 --virtual-env qa --wait 5m
    info    Component image 'localhost/kubefox/hello-world/backend:bb702a1' exists.
    info    Loading component image 'localhost/kubefox/hello-world/backend:bb702a1' into kind cluster 'kind'.

    info    Component image 'localhost/kubefox/hello-world/frontend:780e2db' exists.
    info    Loading component image 'localhost/kubefox/hello-world/frontend:780e2db' into kind cluster 'kind'.

    info    Creating tag 'v0.1.1'.

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
              required: true
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
    ```

Test the release. You should see the output from the updated `frontend`
component.

```{ .shell .copy }
curl "http://localhost:8080/qa/hello"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/qa/hello"
    👋 Hey World!
    ```

Now, release the original version `v0.1.0` to the `prod` environment and create
the required environment snapshot.

```{ .shell .copy }
fox release v0.1.0 --virtual-env prod --create-snapshot
```

??? example "Output"

    ```text
    $ fox release v0.1.0 --virtual-env prod --create-snapshot

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
      virtualEnvSnapshot: prod-1520-19700101-000000
    status:
      current: null
    ```

Send a request to see it working.

```{ .shell .copy }
curl "http://localhost:8080/prod/hello"
```

??? example "Output"

    ```text
    $ curl "http://localhost:8080/prod/hello"
    👋 Hello Universe!
    ```

## Cleanup

Once you are done with the quickstart, you can optionally use the below commands to clean up the Kubernetes environment.

=== "Local (kind)"

    ```{ .shell .copy }
    kind delete cluster
    ```

=== "Azure (AKS)"

    ```{ .shell .copy }
    az aks delete -g ${AZ_RESOURCE_GROUP} --name ${AZ_AKS_NAME} && \
      az acr delete -g ${AZ_RESOURCE_GROUP} --name ${AZ_ACR_NAME} && \
      az group delete -g ${AZ_RESOURCE_GROUP} 
    ```

## Feedback

Explore the rest of the documentation for more details. If you encounter any
problems please let us know on [GitHub
Issues](https://github.com/xigxog/kubefox/issues).
