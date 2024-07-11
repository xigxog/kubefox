Setup a Kubernetes cluster on your workstation using kind and Docker. Kind
is an excellent tool specifically designed for quickly establishing a
cluster for testing purposes.

```{ .shell .copy }
kind create cluster --wait 5m
```

<!-- Be aware that I had to shift the ```text tag left to  -->
<!-- prevent it from showing in the output. -->
??? example "Output"
        
    ```text
        Creating cluster "kind" ...
        ✓ Ensuring node image (kindest/node:v1.27.3) 🖼
        ✓ Preparing nodes 📦
        ✓ Writing configuration 📜
        ✓ Starting control-plane 🕹️
        ✓ Installing CNI 🔌
        ✓ Installing StorageClass 💾
        ✓ Waiting ≤ 5m0s for control-plane = Ready ⏳
        • Ready after 15s 💚
        Set kubectl context to "kind-kind"
        You can now use your cluster with:

        kubectl cluster-info --context kind-kind

        Have a nice day! 👋
    ```