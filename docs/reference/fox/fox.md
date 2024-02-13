## fox

CLI for interacting with KubeFox

### Synopsis


🦊 Fox is a CLI for interacting with KubeFox. You can use it to build, deploy, 
and release your KubeFox Apps.


### Options

```
  -a, --app string                 path to directory containing KubeFox App
  -h, --help                       help for fox
  -i, --info                       enable info output
  -o, --output string              output format, one of ["json", "yaml"] (default "yaml")
      --registry-address string    address of your container registry
      --registry-token string      access token for your container registry
      --registry-username string   username for your container registry
  -v, --verbose                    enable verbose output
```

### SEE ALSO

* [fox build](fox_build.md)	 - Build and optionally push an OCI image of component
* [fox completion](fox_completion.md)	 - Generate the autocompletion script for the specified shell
* [fox config](fox_config.md)	 - Configure 🦊 Fox
* [fox deploy](fox_deploy.md)	 - Deploy KubeFox App using the component code from the currently checked out Git commit
* [fox docs](fox_docs.md)	 - Generate docs for 🦊 Fox
* [fox init](fox_init.md)	 - Initialize a KubeFox App
* [fox proxy](fox_proxy.md)	 - Port forward local port to broker's HTTP server adapter
* [fox publish](fox_publish.md)	 - Builds, pushes, and deploys KubeFox Apps using the component code from the currently checked out Git commit
* [fox release](fox_release.md)	 - Release specified AppDeployment and VirtualEnvironment
* [fox version](fox_version.md)	 - Show version information of 🦊 Fox

