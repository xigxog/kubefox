## fox deploy

Deploy KubeFox App using the component code from the currently checked out Git commit

```
fox deploy [flags]
```

### Options

```
  -t, --create-tag         create Git tag using the AppDeployment version
      --dry-run            submit server-side request without persisting the resource
  -g, --generate           only generate AppDeployment and exit
  -h, --help               help for deploy
  -d, --name string        name to use for AppDeployment, defaults to <APP NAME>-<VERSION | GIT REF | GIT COMMIT>
  -n, --namespace string   namespace of KubeFox Platform
  -p, --platform string    name of KubeFox Platform to utilize
  -s, --version string     version to assign to the AppDeployment, making it immutable
      --wait duration      wait up to the specified time for components to be ready
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

