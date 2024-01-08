## fox release

Release specified AppDeployment and VirtualEnvironment

### Synopsis

The release command activates the routes of the components belonging to the 
specified AppDeployment. This causes genesis events matching components' routes
to be automatically sent to the component with the specified VirtualEnvironment
being injected.

The AppDeployment can be identified by its name, commit, short-commit (first 7 
characters), version, Git tag, or Git branch. 🦊 Fox will inspect the Kubernetes
cluster to find a matching AppDeployment. If more than one AppDeployment is
found you will be prompted to select the desired AppDeployment.

```
fox release <NAME | COMMIT | SHORT COMMIT | VERSION | TAG | BRANCH> [flags]
```

### Examples

```
# Release the AppDeployment named 'main' using the 'dev' Virtual Environment.
fox release main --virtual-env dev

# Release the AppDeployment with version 'v1.2.3' using the 'prod' 
# VirtualEnvironment, creating an DataSnapshot if needed.
fox release v1.2.3 --virtual-env prod --create-snapshot
```

### Options

```
  -c, --create-snapshot      create an immutable snapshot of VirtualEnvironment data and use for Release
      --dry-run              submit server-side request without persisting the resource
  -h, --help                 help for release
  -n, --namespace string     namespace of KubeFox Platform
  -p, --platform string      name of KubeFox Platform to utilize
  -d, --snapshot string      name of DataSnapshot to use for Release
  -e, --virtual-env string   name of VirtualEnvironment to use for Release
      --wait duration        wait up the specified time for components to be ready
```

### Options inherited from parent commands

```
  -a, --app string      path to directory containing KubeFox App
  -i, --info            enable info output
  -o, --output string   output format, one of ["json", "yaml"] (default "yaml")
  -v, --verbose         enable verbose output
```

### SEE ALSO

* [fox](fox.md)	 - CLI for interacting with KubeFox

