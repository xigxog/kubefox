Ensure that the following tools are installed for this exercise (each of the
keywords below contains a link to the installation web page for the referenced
product):

- **[Docker Desktop](https://docs.docker.com/get-docker/)** - The easiest way to get
  going with Docker is to install Docker Desktop.  Docker is a container toolset and
  runtime used to build KubeFox Component OCI images and run a local
  Kubernetes Cluster via kind.  Click the link corresponding to the OS you wish
  you use.
    - **[MacOS](https://docs.docker.com/desktop/install/mac-install/)** - Install
      Docker Desktop on Mac.
    - **[Windows](https://docs.docker.com/desktop/install/windows-install/)** -
      Install Docker Desktop on Windows.
    - **[Linux](https://docs.docker.com/desktop/install/linux-install/)** - Install
      Docker Desktop on Linux.
- **[Fox](https://docs.kubefox.io/getting_started/fox-cli.html)** - A CLI for
  communicating with the KubeFox Platform.
- **[Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)** - A distributed version
  control system.  If this is the first time you're installing git, go through
  the steps to ensure that your identity is established (see "Your Identity" on the next
  **[page](https://git-scm.com/book/en/v2/Getting-Started-First-Time-Git-Setup)**).
- **[Helm](https://helm.sh/docs/intro/install/)** - Package manager for Kubernetes
  used to install the KubeFox Operator on Kubernetes.
- **[kind](https://kind.sigs.k8s.io/docs/user/quick-start/)** - kind is **K**uberentes
  **in** **D**ocker. kind is a tool for running local Kubernetes Clusters using Docker
  container "nodes", hence Docker must be running before you use kind to create
  your Kubernetes cluster. 
- **[Kubectl](https://kubernetes.io/docs/tasks/tools/)** - CLI for communicating
  with a Kubernetes Cluster's control plane, using the Kubernetes API.

Here are a few optional but recommended tools:

- **[Go](https://go.dev/doc/install)** - A programming language. The `hello-world`
  example App is written in Go, but Fox is able to compile it even without Go
  installed.
- **[VS Code](https://code.visualstudio.com/download)** - A lightweight but powerful
  source code editor. Helpful if you want to explore the `hello-world` app.
- **[Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli)** -
  CLI for communicating with the Azure control plane.
- **[k9s](https://k9scli.io/topics/install/)** - A terminal based UI to interact
  with your Kubernetes clusters. You can use native kubectl commands to
  accomplish the same things, but k9s is a nice convenience and we use it here.
  By the way, the k9s **[homepage](https://k9scli.io/)** is probably the cleverest of any
  company in the k8s space, succeeding in that endeavor at many levels.
