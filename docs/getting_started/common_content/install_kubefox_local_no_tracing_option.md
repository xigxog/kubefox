If you're using a Mac, we need to load the correct version of KubeFox for your
processor type (Intel or Mac). If you're uncertain of how to determine what
processor you're running, expand the "Mac Processor Determination" section
below.

<!-- Note:  Had to add the silly "../../../../" prefix to the paths.  Simply put,     -->
<!-- mkdocs does an http get starting at the location of the file in the directory.   -->
<!-- That creates a problem because the location is variable if snippets are          -->
<!-- employed.  So the option was to locate images where the current doc is -         -->
<!-- which sucks in my mind.  They should be centrally located, and one should        -->
<!-- be able to provide an assets path or the like.                                   -->

<!-- So I finally gave up.  The silly prefix will (effectively) start the             -->
<!-- search at the root.                                                              -->

<!-- Additionally, a block comment was not respected by mkdocs, so each of these      -->
<!-- lines had to be independently commented.                                         -->


??? info "Mac Processor Determination"

    Click the Apple icon at the top left of any window and select "About this Mac".

    <figure markdown>
      <img src="../../../../images/screenshots/generic/mac-about-this-mac.png" width="300">
    </figure>

    This is an example of the output if you're running Apple silicon (an "M" processor):

    <figure markdown>
      <img src="../../../../images/screenshots/generic/mac-apple-silicon-about.jpeg" width="300">
    </figure>

    And this is an example of the output if you're running an Intel processor:

    <figure markdown>
      <img src="../../../../images/screenshots/generic/mac-intel-silicon-about.png">
    </figure>


If (and only if) you're running a Mac with Apple silicon (M1, M2 or M3), select the "Apple Silicon" tab.  Otherwise, select the "Not Apple Silicon" tab.

=== "Not Apple Silicon"

    Install KubeFox on local kind cluster.

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


=== "Apple Silicon"

    Install Kubefox on local kind cluster (Apple Silicon).


    ```{ .shell .copy }
    helm upgrade kubefox kubefox \
    --repo https://xigxog.github.io/helm-charts \
    --create-namespace --namespace kubefox-system \
    --install --wait --set image.tag=main
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
