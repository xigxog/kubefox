# Install Fox Cli

The fox cli is a tool used to interact with KubeFox, initialize your repositories
and prepare your Apps for deployment and release.  It is modified fairly
frequently, so don't frequently upgrading it!

=== "Install using Go"

    ```{ .shell .copy }
    go install github.com/xigxog/fox@latest
    ```

=== "Install using Bash"

    ```{ .shell .copy }
    curl -sL "https://github.com/xigxog/fox/releases/latest/download/fox-$(uname -s | tr 'A-Z' 'a-z')-amd64.tar.gz" | \
      tar xvz - -C /tmp && \
      sudo mv /tmp/fox /usr/local/bin/fox
    ```

=== "Install Manually"

    Download the [latest Fox release](https://github.com/xigxog/fox/releases/latest){:target="_blank"} for your OS and extract the `fox` binary to a directory on your path.
