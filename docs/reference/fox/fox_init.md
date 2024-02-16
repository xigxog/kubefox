## fox init

Initialize a KubeFox App

### Synopsis

The init command creates the skelton of a KubeFox App and ensures a Git 
repository is present. It will optionally create simple 'hello-world' app to get
you started.

```
fox init [flags]
```

### Options

```
  -h, --help         help for init
      --quickstart   use defaults to setup KubeFox for quickstart guide
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

