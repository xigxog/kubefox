## fox completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(fox completion bash)

To load completions for every new session, execute once:

#### Linux:

	fox completion bash > /etc/bash_completion.d/fox

#### macOS:

	fox completion bash > $(brew --prefix)/etc/bash_completion.d/fox

You will need to start a new shell for this setup to take effect.


```
fox completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
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

* [fox completion](fox_completion.md)	 - Generate the autocompletion script for the specified shell

