# KubeFox

[![build](https://github.com/xigxog/kubefox/actions/workflows/build.yaml/badge.svg)](https://github.com/xigxog/kubefox/actions/workflows/build.yaml)
[![release](https://github.com/xigxog/kubefox/actions/workflows/release.yaml/badge.svg)](https://github.com/xigxog/kubefox/actions/workflows/release.yaml)
[![Go Report
Card](https://goreportcard.com/badge/github.com/xigxog/kubefox)](https://goreportcard.com/report/github.com/xigxog/kubefox)

- Website: <https://www.kubefox.io>
- Documentation: <https://docs.kubefox.io>
- Quickstart: <https://docs.kubefox.io/quickstart.html>

<figure markdown>
<img src="./docs/images/KubeFox Horizontal Layout (400 x 100) revised.svg" width=100% height=100%>
</figure>

KubeFox is an SDK, platform and infrastructure to enable rapid construction and
deployment of secure and robust applications for Kubernetes, and which
drastically reduces bureaucracy and burdensome DevOps activities.

Teams and even individual developers can create and rapidly prototype code on
the same cluster in what appear to be individual sandboxes. Behind the scenes,
KubeFox provides the following capabilities:

- **Deployment Distillation:**

  With KubeFox, you deploy at an application level. You don't need to worry
  about which components you have added or modified. KubeFox tracks the
  repository and builds, containerizes and distills deployments to only those
  components that have changed. You can read more about Deployment Distillation
  [here](https://docs.kubefox.io/concepts/deployment_distillation.html).

- **Versioned Deployments**

  When you deploy an application with KubeFox, that application is automatically
  versioned. KubeFox will ensure that traffic is restricted to the components
  that composed the application when it was deployed. Note that individual
  deployments may share one or more components. This enables KubeFox to prevent
  over-provisioning; the deployments can run on the same cluster but it appears
  that each deployment is running in its own invidual sandbox. That extends to
  deployment telemetry, which will reflect data from each version. You can read
  more about Versioned Deployments
  [here](https://docs.kubefox.io/concepts/versioned_deployments.html).

- **Virtual Environments**

  KubeFox Virtual Environments (VEs for short) are lightweight, malleable
  constructs that enable engineering teams and developers to create virtual
  sandboxes in which they can run prototypes or POCs. Custom configurations
  (consisting of environment variables and secrets) can be created and deployed
  easily - either in concert with a deployment or independently. And they're
  controllable by the developers themselves, eliminating the overhead associated
  with traditional environments. You can read more about Virtual Environments
  [here](https://docs.kubefox.io/concepts/virtual_environments.html).

- **Federated metrics, tracing and logging**

  Multiple versions of components and applications can be deployed to a single
  cluster. That would seem to complicate telemetry but KubeFox provides for
  telemetry at global, application and component levels. You can read more about
  Telemetry [here](https://docs.kubefox.io/concepts/telemetry.html).

- **Dynamic Routing and Guaranteed Delivery**

  KubeFox is capable of managing multiple versions of applications and
  components because it is aware of the composition of those applications, and
  traffic is shaped and validated to a specific version of an application at
  runtime. You can read more about Dynamic Routing
  [here](https://docs.kubefox.io/concepts/dynamic_routing.html).

- **Zero Trust**

  Zero Trust is an intrinsic property of a KubeFox application. All requests are
  validated at origin (Genesis Events in KubeFox) and at all
  component-to-component boundaries.

## Getting Started and Documentation

The best way to get started with KubeFox is to go through our
[Quickstart](https://docs.kubefox.io/quickstart.html) and read through our
[docs](https://docs.kubefox.io). The Quickstart currently supports
[kind](https://kind.sigs.k8s.io/) and [Microsoft
Azure](https://learn.microsoft.com/en-us/azure/?product=popular). The Quickstart
can be navigated in less than a half hour and gives you a good overview of the
power of KubeFox.

## Local Development

This will help get your workstation setup for local development, allowing you to
run a local instance of the `broker` and `httpsrv` with a debugger. First,
create a new kind cluster.

```shell
kind create cluster --wait 5m
```

Then, build all the components and load their container images into the kind
cluster.

```shell
make all KIND_LOAD=true
```

Now, install the KubeFox Operator using the current branch name as the image
tag.

```shell
helm install kubefox kubefox \
    --repo https://xigxog.github.io/helm-charts \
    --create-namespace --namespace kubefox-system \
    --set log.format=console --set log.level=debug --set image.tag=$(git rev-parse --abbrev-ref HEAD) \
    --wait
```

Next, create a new directory and initialize a KubeFox App with Fox. When
prompted to initialize the `hello-world` App respond with `y` and for the
Platform name enter `debug` and press `Enter` to accept the default namespace of
`kubefox-debug`. The name and namespace must match exactly for the local
development environment to work.

```shell
export FOX_INFO=true && \
    mkdir -p /tmp/kubefox/hello-world && \
    cd /tmp/kubefox/hello-world && \
    fox init && \
    fox publish --wait 5m && \
    kubectl apply -f ./hack/environments/ -n kubefox-debug ; \
    cd -
```

The output should look something like this.

```text
info    Let's initialize a KubeFox App!

info    To get things started quickly 🦊 Fox can create a 'hello-world' KubeFox App which
info    includes two components and example environments for testing.
Would you like to initialize the 'hello-world' KubeFox App? [y/N] y
Enter URL for remote Git repo (default 'https://github.com/xigxog/hello-world.git'):

info    You need to have a KubeFox Platform instance running to deploy your components.
info    Don't worry, 🦊 Fox can create one for you.
Would you like to create a KubeFox Platform? [Y/n] y
Enter the KubeFox Platform's name (required): debug
Enter the Kubernetes namespace of the KubeFox Platform (default 'kubefox-debug'):
```

Great! Now that there is a local instance you just need to enable debug mode and
start the `broker` locally. There is a script that takes care of updating the
Platform and setting up port forwards to your workstation for Vault and NATS.

```shell
make debug
```

If everything went smoothly you can now start local instances of the `broker`
and `httpsrv` using the commands provided by `make debug`. Or if using VSCode
you can start the launch configurations named `broker` and `httpsrv`. Once both
are started use `curl` to perform a request and you should see corresponding log
output in the debug console.

```shell
curl -v "http://localhost:8080/qa/hello?kf-dep=hello-world-main&kf-ve=qa"
```

## License

XigXog is committed to open source and our software is licensed under the
[Mozilla Public License Version
2.0](https://github.com/xigxog/kubefox/blob/main/LICENSE).
