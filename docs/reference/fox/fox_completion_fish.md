## fox completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	fox completion fish | source

To load completions for every new session, execute once:

	fox completion fish > ~/.config/fish/completions/fox.fish

You will need to start a new shell for this setup to take effect.


```
fox completion fish [flags]
```

### Options

```
  -h, --help              help for fish
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
  -v, --verbose                    enable verbose output
```

### SEE ALSO

* [fox completion](fox_completion.md)	 - Generate the autocompletion script for the specified shell

