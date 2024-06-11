# Install the Fox CLI

The Fox CLI ("Fox") is a tool used to interact with KubeFox and prepare your Apps
for deployment and release.

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

You can check the version of Fox with ```fox version```.

If you were running the Quickstart, you can return to it by clicking **[here](./quickstart.md#prerequisites)**.

If you were running the KubeFox and Hasura Tutorial, you can return to it by clicking **[here](./tutorials/graphql.md#prerequisites)**.
