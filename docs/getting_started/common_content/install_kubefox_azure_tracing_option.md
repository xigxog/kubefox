=== "Tracing Active"

    Install KubeFox on Azure with tracing active.

    ```{ .shell .copy }
    helm upgrade kubefox kubefox \
    --repo https://xigxog.github.io/helm-charts \
    --create-namespace --namespace kubefox-system \
    --set telemetry.enabled=true \
    --install --wait
    ```

    ??? example "Output"

        ```text
        Release "kubefox" does not exist. Installing it now.
        NAME: kubefox
        LAST DEPLOYED: Fri May 24 18:18:57 2024
        NAMESPACE: kubefox-system
        STATUS: deployed
        REVISION: 1
        ```

=== "No Tracing"

    Install KubeFox on Azure without tracing active.

    ```{ .shell .copy }
    helm upgrade kubefox kubefox \
    --repo https://xigxog.github.io/helm-charts \
    --create-namespace --namespace kubefox-system \
    --install --wait
    ```

    ??? example "Output"

        ```text
        Release "kubefox" does not exist. Installing it now.
        NAME: kubefox
        LAST DEPLOYED: Fri May 24 18:18:57 2024
        NAMESPACE: kubefox-system
        STATUS: deployed
        REVISION: 1
        ```