# Hello World

In this example we'll compare two simple `hello-world` apps that read the
environment variable `who` and say hello to them. The first is a
"native" Kubernetes app using HTTP servers, Dockerfile, and various Kubernetes
resources. The second is a KubeFox component which will be deployed using
[fox](https://github.com/xigxog/fox), the KubeFox CLI.

To get started you'll need the following installed:

- [Go](https://go.dev/doc/install)
- [Git](https://github.com/git-guides/install-git)
- [Docker](https://docs.docker.com/engine/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

You'll also need access to a Kubernetes cluster. If you'd like to run a
Kubernetes cluster on your workstation for testing we recommend using [kind
(Kubernetes in Docker)](https://kind.sigs.k8s.io/docs/user/quick-start/).

## Native

We have included everything needed to run the app in Kubernetes. This includes:

- Dockerfile to build and package the app into an OCI container.
- Kubernetes Deployments to run the app on the Kubernetes cluster.
- Kubernetes ConfigMap to store environment variables used by the app.
- Kubernetes Service to be able to send requests to the pod.

Take a look at the various files, there is a lot going on.

There are a few steps to deploy the app to Kubernetes. Run all the following
commands from the `native` directory.

First, build the app container images using Docker. If you are using kind
locally you can leave the container registry set to localhost, otherwise replace
it with the container registry you'd like to use.

```shell
export CONTAINER_REGISTRY="localhost"
docker buildx build ./backend --file Dockerfile --tag "$CONTAINER_REGISTRY/hello-world-backend:main"
docker buildx build ./frontend --file Dockerfile --tag "$CONTAINER_REGISTRY/hello-world-frontend:main"
```

Next you'll need to make the container images available to Kubernetes. If you
are using kind you can load the image directly without using a container
registry.

```shell
export KIND_CLUSTER="kind"
kind load docker-image --name "$KIND_CLUSTER" "$CONTAINER_REGISTRY/hello-world-backend:main"
kind load docker-image --name "$KIND_CLUSTER" "$CONTAINER_REGISTRY/hello-world-frontend:main"
```

Otherwise push the image to the container registry.

```shell
docker push "$CONTAINER_REGISTRY/hello-world-backend:main"
docker push "$CONTAINER_REGISTRY/hello-world-frontend:main"
```

Finally, create a Kubernetes namespace and apply the ConfigMaps and Deployment
to run the app on Kubernetes.

```shell
kubectl create namespace hello-world-qa
kubectl apply --namespace hello-world-qa --filename hack/environments/qa.yaml
kubectl apply --namespace hello-world-qa --filename hack/deployments/

# Example output:
# namespace/hello-world created
# configmap/env created
# deployment.apps/hello-world-backend created
# service/hello-world-backend created
# deployment.apps/hello-world-frontend created
# service/hello-world-frontend created
```

If everything worked you should see two pods running. You can check using
kubectl.

```shell
kubectl get pods --namespace hello-world-qa

# Example output:
# NAME                                    READY   STATUS    RESTARTS   AGE
# hello-world-backend-865d6697d5-2vwnw    1/1     Running   0          10s
# hello-world-frontend-5579b569c9-fdsnw   1/1     Running   0          19s
```

Time to test the app. To keep things simple you'll port forward to the pod to
access its HTTP server. Open up a new terminal and run the following to start
the port forward. This will open the port `8888` on your workstation which will
forward requests to the pod.

```shell
kubectl port-forward --namespace hello-world-qa service/hello-world-frontend 8888:http

# Example output:
# Forwarding from 127.0.0.1:8888 -> 3333
# Forwarding from [::1]:8888 -> 3333
```

Finally send a HTTP request to the app.

```shell
curl http://127.0.0.1:8888/qa/hello

# Example output:
#👋 Hello World!
```

It works! But how do you run the app in a different environment so you can
change who to say hello to? You need to update the ConfigMap `env` that contains
the `who` variable. Of course if you change what is running now the `world`
environment will no longer exist. Instead you can create a new namespace and run
the app there with the updated ConfigMap. Try it out.

```shell
kubectl create namespace hello-world-prod
kubectl apply --namespace hello-world-prod --filename hack/environments/prod.yaml
kubectl apply --namespace hello-world-prod --filename hack/deployments/

# Example output:
# namespace/hello-world-prod created
# configmap/env created
# deployment.apps/hello-world-backend created
# service/hello-world-backend created
# deployment.apps/hello-world-frontend created
# service/hello-world-frontend created
```

Now you can test the app in the new environment. Once again open up a new
terminal and run the following to start the port forward but use port `8889`
this time.

```shell
kubectl port-forward --namespace hello-world-prod service/hello-world-frontend 8889:http

# Example output:
# Forwarding from 127.0.0.1:8889 -> 3333
# Forwarding from [::1]:8889 -> 3333
```

Then send a HTTP request to app.

```shell
curl http://127.0.0.1:8889/prod/hello

# Example output:
#👋 Hello Universe!
```

Great! It's using the new environment. Take a look at what is running on
Kubernetes now. You can use a label from the Deployments to show pods from
multiple namespaces.

```shell
kubectl get pods --all-namespaces --selector=app.kubernetes.io/name=hello-world-native

# Example output:
# NAMESPACE          NAME                                            READY   STATUS    RESTARTS   AGE
# hello-world-prod   hello-world-backend-9f67b958d-lwm6t             1/1     Running   0          2m30s
# hello-world-prod   hello-world-frontend-887674586-2q298            1/1     Running   0          2m25s
# hello-world-qa     hello-world-backend-865d6697d5-tpbfr            1/1     Running   0          3m49s
# hello-world-qa     hello-world-frontend-5579b569c9-fdsnw           1/1     Running   0          5m10s
```

## KubeFox

For details on how to run the KubeFox `hello-world` app see the [quickstart
tutorial](https://docs.kubefox.io/quickstart/). It uses the same `hello-world`
app as found here!
