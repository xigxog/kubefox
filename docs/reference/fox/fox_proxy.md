## fox proxy

Port forward local port to broker's HTTP server adapter

### Synopsis

The proxy command will inspect the Kubernetes cluster and find an available
broker to proxy a local port to. This port can then be used to make HTTP
requests to the broker's HTTP server adapter. This is especially useful during
development and testing.

The optional flags 'virtual-env' and 'app-deployment' can be set which will
automatically inject the values as context to requests sent through the proxy. 
The context can still be overridden manually by setting the header or query 
param on the original request.

```
fox proxy <PORT> [flags]
```

### Examples

```
# Port forward local port 8080 and wait if no brokers are available.
fox proxy 8080 --wait 5m

# Port forward local port 8080 and inject 'my-env' and 'my-dep' context.
fox proxy 8080 --virtual-env my-env --app-deployment my-dep

	http://127.0.0.1:8080/hello                 # uses my-env and my-deployment
	http://127.0.0.1:8080/hello?kf-env=your-env # uses your-env and my-dep
	http://127.0.0.1:8080/hello?kf-dep=your-dep # uses my-env and your-dep
```

### Options

```
  -d, --app-deployment string   deployment to add to proxied requests
      --dry-run                 submit server-side request without persisting the resource
  -h, --help                    help for proxy
  -n, --namespace string        namespace of KubeFox Platform
  -p, --platform string         name of KubeFox Platform to utilize
  -e, --virtual-env string      environment to add to proxied requests
      --wait duration           wait up to the specified time for components to be ready
```

### Options inherited from parent commands

```
  -a, --app string                 path to directory containing KubeFox App
  -i, --info                       enable info output
  -o, --output string              output format, one of ["json", "yaml"] (default "yaml")
      --registry-address string    address of your container registry
      --registry-token string      access token for your container registry
      --registry-username string   username for your container registry
  -m, --timeout duration           timeout for command (default 5m0s)
  -v, --verbose                    enable verbose output
```

### SEE ALSO

* [fox](fox.md)	 - CLI for interacting with KubeFox

