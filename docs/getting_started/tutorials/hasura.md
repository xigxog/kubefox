# Hasura Tutorial

Welcome to the world of KubeFox! This technical guide will walk you through the
process of setting up a Kubernetes cluster using kind ("kind" is **K**ubernetes
**in** **D**ocker).

A couple of notes:

1. We'll add instructions for Azure soon - it's really not that different but we
want to vet it first.
2. This tutorial currently gets the cluster ready for the LiveStream we're doing
   on the 13th of March.  We'll turn it into a standalone tutorial soon!

## Prerequisites

Ensure that the following tools are installed for this tutorial:

- [Docker](https://docs.docker.com/engine/install/) - A container toolset and
  runtime used to build KubeFox Components' OCI images and run a local
  Kubernetes Cluster via kind.
- [Fox](https://github.com/xigxog/kubefox-cli/releases/) - A CLI for
  communicating with the KubeFox Platform. Installation instructions are below.
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
  source code editor. Helpful if you want to explore the `hello-world` app.
- [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) -
  CLI for communicating with the Azure control plane.

## Setup Kubernetes

*Note:  If you went through the Quickstart using kind, we recommend that you
first delete the cluster you created to ensure that you're starting with a clean slate.*

Let's kick things off by setting up a Kubernetes cluster. Use the following
commands depending on which Kubernetes provider you would like to use. If you
already have a Kubernetes Cluster provisioned, you can skip this step.

=== "Local (kind)"

    Setup a Kubernetes cluster on your workstation using kind and Docker. Kind
    is an excellent tool specifically designed for quickly establishing a
    cluster for testing purposes.

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

## Setup KubeFox

In this step you will install the KubeFox Helm Chart to initiate the KubeFox
Operator on your Kubernetes cluster. The operator manages KubeFox Platforms and
Apps.

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

Now install Fox, a CLI tool used to interact with KubeFox and prepare your Apps
for deployment and release.

=== "Install using Go"

    ```{ .shell .copy }
    go install github.com/xigxog/fox@latest
    ```

=== "Install using Bash"

    ```{ .shell .copy }
    curl -sL "https://github.com/xigxog/fox/releases/latest/download/fox-$(uname -s | tr 'A-Z' 'a-z')-amd64.tar.gz" | \
      tar xvz - -C /tmp && \
      sudo mv /tmp/fox /usr/local/bin/fox
    ```

=== "Install Manually"

    Download the [latest Fox release](https://github.com/xigxog/fox/releases/latest){:target="_blank"} for your OS and extract the `fox` binary to a directory on your path.

## Deploy

Awesome! You're all set to start the KubeFox Platform on the your newly created
cluster and deploy your first KubeFox Hasura App. To begin, create a new directory and
use Fox to initialize the `hello-world` App. Run all subsequent commands from
this directory. The environment variable `FOX_INFO` tells Fox to to provide
additional output about what is going on. Employ the `--hasura` flag to simplify
the Hasura `demo` in the `kubefox-demo` Namespace.

```{ .shell .copy }
mkdir kubefox-hasura && \
  cd kubefox-hasura && \
  export FOX_INFO=true && \
  fox init --hasura 
```

??? example "Output"

    ```text
    info    Configuration successfully written to '/home/xadhatter/.config/kubefox/config.yaml'.

    info    Waiting for KubeFox Platform 'demo' to be ready...
    info    KubeFox initialized for the quickstart guide!
    ```

From here, please refer to the CNCF LiveStream from March 13th.  Rest Assured,
we'll update this shortly to turn it into a standalone tutorial!