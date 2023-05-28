<!-- markdownlint-disable MD033 -->
# Quickstart

## Prerequisites

The following tools must be installed for this quickstart:

- [Docker](https://docs.docker.com/engine/install/) - A container toolset and
  runtime. Used to build KubeFox Components' OCI images and run a local
  Kubernetes cluster via Kind.
- [Fox CLI](https://github.com/xigxog/kubefox-cli/releases/tag/v0.1.0-alpha) -
  CLI for communicating with the KubeFox Platform. Download the latest release
  and put the binary on your path.
- [Helm](https://helm.sh/docs/intro/install/) - Package manager for Kubernetes.
  Used to install the KubeFox Platform on Kubernetes.
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) - **K**uberentes
  **in** **D**ocker. A tool for running local Kubernetes clusters using Docker
  container "nodes".
- [Kubectl](https://kubernetes.io/docs/tasks/tools/) - CLI for communicating
  with a Kubernetes cluster's control plane, using the Kubernetes API.

## Install KubeFox Platform

Let's start with setting up a local Kubernetes cluster using Kind. The following
command can be used to create the cluster. It exposes some extra ports to the
local host that allows communicating with the KubeFox Platform easier.

```shell
echo >/tmp/kind-cluster.yaml "\
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: kubefox
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 30080
        hostPort: 30080
      - containerPort: 30443
        hostPort: 30443
"
kind create cluster --config /tmp/kind-cluster.yaml --wait 5m
```

??? example "Output"

    ```text
    Creating cluster "kubefox" ...
    ✓ Ensuring node image (kindest/node:v1.26.3) 🖼
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
    ```

Next we need to add the XigXog Helm Repo and install the KubeFox Helm Chart.

```shell
helm repo add xigxog https://xigxog.github.io/helm-charts
helm repo update
helm install kubefox xigxog/kubefox --create-namespace --namespace kubefox-system --wait
```

??? example "Output"

    ```text
    "xigxog" has been added to your repositories
    Hang tight while we grab the latest from your chart repositories...
    ...Successfully got an update from the "xigxog" chart repository
    Update Complete. ⎈Happy Helming!⎈
    NAME: kubefox
    LAST DEPLOYED: Wed May  3 15:25:24 2023
    NAMESPACE: kubefox-system
    STATUS: deployed
    REVISION: 1
    ```

## Deploy KubeFox System

Great now we're ready to create your first KubeFox System and deploy it to the
Platform running on the local Kind Cluster. We'll be using a local Git repo for
the demo. To get things started let's create a new directory and use the Fox CLI
to initialize a "hello world" System. You'll want to run all the remaining
commands from this directory.

```shell
mkdir kubefox-demo
cd kubefox-demo
fox init
```

??? example "Output"

    ```text
    info    fox     Fox needs to create a secret to push and pull container images to GitHub Packages.
    info    fox     Copy this code 'XXXX-XXXX', then open 'https://github.com/login/device' in your browser.
    1. acme-corp
    2. wonka-industries
    Select the GitHub org to use (default 1):
    Enter the URL of the KubeFox API (default 'https://127.0.0.1:30443'):
    info    fox     Checking connectivity to Kubefox Platform at 'https://127.0.0.1:30443'
    info    fox     Writing files for a demo KubeFox system to './kubefox-demo'
    info    fox     KubeFox system initialization complete
    ```

You should see some newly created directories and files. The demo system
includes two components and an app. Take a look around!

Let's deploy the system and see what happens. The following command will
build the components OCI images. Push them up to the repo and then deploy them
onto the KubeFox Platform. The first run might take some time as it needs to
download dependencies, but future runs should be much faster.

```shell
fox publish --tag v1 --deploy
```

??? example "Output"

    ```text
    info    fox     Building component 'hello'
    info    fox     Building image 'ghcr.io/acme-corp/kubefox-demo/hello:4ccf3cb' for component 'hello'
    info    fox     Building component 'world'
    info    fox     Building image 'ghcr.io/acme-corp/kubefox-demo/world:4ccf3cb' for component 'world'
    info    fox     Creating system object 'system/kubefox-demo'
    info    fox     Creating system tag 'system/kubefox-demo/tag/v1'
    info    fox     Deploying system tag 'system/kubefox-demo/tag/v1'
    info    fox     System successfully published
    ```

Since the app uses environment vars we also need to create an environment for it
to use. A sample environment is provided in the `hack` dir.

```shell
fox apply --filename hack/env.yaml --tag v1
```

??? example "Output"

    ```text
    apiVersion: admin.kubefox.io/v1alpha1
    id: b3815e21-ae6c-4dda-ae6a-31eb4b8203bb
    kind: Environment
    metadata:
      description: A simple environment to use with KubeFox demo system.
      name: dev
      title: Development
    status: {}
    vars:
      who: John

    info    fox     Tag resource 'environment/dev/tag/v1' created
    ```

Awesome. Now we can try out our app. Since it has not been released none of the
routes are active yet. We'll need to manually specify the context of the genesis
event.

```shell
curl -k "https://localhost:30443/hello?kf-sys=kubefox-demo:tag:v1&kf-env=dev:tag:v1&kf-target=hello-world:hello"
```

??? example "Output"

    ```text
    👋 Hello John!
    ```

Now let's release the system so we don't have to specify all those details in
the request.

```shell
fox release --system kubefox-demo/tag/v1 --environment dev/tag/v1
```

??? example "Output"

    ```text
    apiVersion: admin.kubefox.io/v1alpha1
    environment: dev/tag/v1
    kind: Release
    status:
      ready: false
    system: kubefox-demo/tag/v1

    👋 Hello John!
    ```

Let's try the same request from above, but this time we won't specify the
context. Since the system has been released the request will get matched by the
app's route and the context information will be automatically injected by
KubeFox.

```shell
curl -k "https://localhost:30443/hello"
```

??? example "Output"

    ```text
    👋 Hello John!
    ```

Next we'll create a new version of our dev environment. Edit the `hack/env.yaml`
and set the `who` variable to be `world`. Then send the updated environment to
KubeFox.

```shell
echo >hack/env.yaml "\
apiVersion: admin.kubefox.io/v1alpha1
kind: Environment
metadata:
  name: dev
  title: Development
  description: A simple environment to use with KubeFox demo system.
vars:
  who: world
"
fox apply --filename hack/env.yaml --tag v2
```

??? example "Output"

    ```text
    apiVersion: admin.kubefox.io/v1alpha1
    id: ba8e34e0-3788-4651-8d4b-56a026a3da56
    kind: Environment
    metadata:
      description: A simple environment to use with KubeFox demo system.
      name: dev
      title: Development
    status: {}
    vars:
      who: world

    info    fox     Tag resource 'environment/dev/tag/v2' created
    ```

We probably want to test the new environment version before releasing so we'll
just manually specify the new environment.

```shell
curl -k "https://localhost:30443/hello?kf-env=dev:tag:v2"
```

??? example "Output"

    ```text
    👋 Hello world!
    ```

??? abstract "TODO"

    - Update and re-deploy components
    - Show manual switch to new system version
    - Create a QA environment and release to it
