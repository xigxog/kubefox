Establish a remote Kubernetes cluster on the Microsoft Azure cloud platform
using the Azure CLI. Keep in mind that creating the specified resources may
result in costs. Instructions at the end of the quickstart will guide you in
tearing down all the created resources.

```{ .shell .copy }
az login
```
Next set the required variables for this quickstart on Azure.

```{ .shell .copy }
export AZ_LOCATION=eastus2 && \
  export AZ_RESOURCE_GROUP=kf-quickstart-infra-eus2-rg && \
  export AZ_AKS_NAME=kf-quickstart-eus2-aks-01
```

Now you will create a Resource Group for the AKS cluster, and then deploy
Azure Kubernetes Service (AKS) to the group. The cluster provisioning will
take several minutes to complete.

```{ .shell .copy }
az group create --location $AZ_LOCATION --name $AZ_RESOURCE_GROUP && \
  az aks create \
    --resource-group $AZ_RESOURCE_GROUP \
    --tier free \
    --name $AZ_AKS_NAME \
    --location $AZ_LOCATION \
    --generate-ssh-keys \
    --node-count 1 \
    --node-vm-size "Standard_B2s"
```
??? example "Output"

    ```json
    {
      "id": "/subscriptions/00000000-0000-0000-0000-00000000/resourceGroups/kf-quickstart-infra-eus2-rg",
      "location": "eastus2",
      "managedBy": null,
      "name": "kf-quickstart-infra-eus2-rg",
      "properties": {
        "provisioningState": "Succeeded"
      },
      "tags": null,
      "type": "Microsoft.Resources/resourceGroups"
    }

    (... and much more ...)

    ```

Once your AKS cluster is ready add the cluster to your kubectl configuration
to securely communicate with the Kube API.

```{ .shell .copy }
az aks get-credentials \
  --resource-group $AZ_RESOURCE_GROUP  \
  --name $AZ_AKS_NAME
```

??? info "A different object named ... already exists in your kubeconfig file messages"

    If you see messages like these, you've probably run the Quickstart on Azure previously.  Just answer "y" to overwrite as shown below.

    ```text

    A different object named kf-quickstart-eus2-aks-01 already exists in your kubeconfig file.
    Overwrite? (y/n): y

    A different object named clusterUser_kf-quickstart-infra-eus2-rg_kf-quickstart-eus2-aks-01 already exists in your kubeconfig file. Overwrite? (y/n): y

    Merged "kf-quickstart-eus2-aks-01" as current context in /Users/Steven/.kube/config

    ```

The last resource to create is the Azure Container Registry (ACR). This is
used to store the KubeFox Component container images.

```{ .shell .copy }
export AZ_ACR_NAME="acr$RANDOM" && \
  az acr create --name $AZ_ACR_NAME --sku Basic --admin-enabled true --resource-group $AZ_RESOURCE_GROUP
```
Set the registry endpoints and token to access the registry. These
environment variables are used by Fox to push container images to ACR.

```{ .shell .copy }
export FOX_REGISTRY_ADDRESS=$(az acr show-endpoints \
    --name $AZ_ACR_NAME \
    --resource-group $AZ_RESOURCE_GROUP \
    --output tsv \
    --query loginServer) && \
  export FOX_REGISTRY_TOKEN=$(az acr login \
    --name $AZ_ACR_NAME \
    --expose-token \
    --output tsv \
    --query accessToken) && \
  export FOX_REGISTRY_USERNAME="00000000-0000-0000-0000-000000000000"
```
