# Kubernetes CRDs
## kubefox.xigxog.io/v1alpha1





### AppDeployment

AppDeployment is the Schema for the AppDeployments API



| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `apiVersion` | string | `kubefox.xigxog.io/v1alpha1` | |
| `kind` | string | `AppDeployment` | |
| `metadata` | <div style="white-space:nowrap">[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#objectmeta-v1-meta)<div> | <div style="max-width:30rem">Refer to Kubernetes API documentation for fields of `metadata`.</div> | <div style="white-space:nowrap"></div> |
| `spec` | <div style="white-space:nowrap">[AppDeploymentSpec](#appdeploymentspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `status` | <div style="white-space:nowrap">[AppDeploymentStatus](#appdeploymentstatus)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `details` | <div style="white-space:nowrap">[AppDeploymentDetails](#appdeploymentdetails)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



























### Environment





| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `apiVersion` | string | `kubefox.xigxog.io/v1alpha1` | |
| `kind` | string | `Environment` | |
| `metadata` | <div style="white-space:nowrap">[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#objectmeta-v1-meta)<div> | <div style="max-width:30rem">Refer to Kubernetes API documentation for fields of `metadata`.</div> | <div style="white-space:nowrap"></div> |
| `spec` | <div style="white-space:nowrap">[EnvironmentSpec](#environmentspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `data` | <div style="white-space:nowrap">[Data](#data)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `details` | <div style="white-space:nowrap">[DataDetails](#datadetails)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |









### HTTPAdapter





| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `apiVersion` | string | `kubefox.xigxog.io/v1alpha1` | |
| `kind` | string | `HTTPAdapter` | |
| `metadata` | <div style="white-space:nowrap">[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#objectmeta-v1-meta)<div> | <div style="max-width:30rem">Refer to Kubernetes API documentation for fields of `metadata`.</div> | <div style="white-space:nowrap"></div> |
| `spec` | <div style="white-space:nowrap">[HTTPAdapterSpec](#httpadapterspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `details` | <div style="white-space:nowrap">[Details](#details)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |














### Platform

Platform is the Schema for the Platforms API



| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `apiVersion` | string | `kubefox.xigxog.io/v1alpha1` | |
| `kind` | string | `Platform` | |
| `metadata` | <div style="white-space:nowrap">[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#objectmeta-v1-meta)<div> | <div style="max-width:30rem">Refer to Kubernetes API documentation for fields of `metadata`.</div> | <div style="white-space:nowrap"></div> |
| `spec` | <div style="white-space:nowrap">[PlatformSpec](#platformspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `status` | <div style="white-space:nowrap">[PlatformStatus](#platformstatus)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `details` | <div style="white-space:nowrap">[PlatformDetails](#platformdetails)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |














### ReleaseManifest





| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `apiVersion` | string | `kubefox.xigxog.io/v1alpha1` | |
| `kind` | string | `ReleaseManifest` | |
| `metadata` | <div style="white-space:nowrap">[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#objectmeta-v1-meta)<div> | <div style="max-width:30rem">Refer to Kubernetes API documentation for fields of `metadata`.</div> | <div style="white-space:nowrap"></div> |
| `spec` | <div style="white-space:nowrap">[ReleaseManifestSpec](#releasemanifestspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `data` | <div style="white-space:nowrap">[Data](#data)<div> | <div style="max-width:30rem">Data is the merged values of the Environment and VirtualEnvironment Data.</div> | <div style="white-space:nowrap">required</div> |











### VirtualEnvironment





| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `apiVersion` | string | `kubefox.xigxog.io/v1alpha1` | |
| `kind` | string | `VirtualEnvironment` | |
| `metadata` | <div style="white-space:nowrap">[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#objectmeta-v1-meta)<div> | <div style="max-width:30rem">Refer to Kubernetes API documentation for fields of `metadata`.</div> | <div style="white-space:nowrap"></div> |
| `spec` | <div style="white-space:nowrap">[VirtualEnvironmentSpec](#virtualenvironmentspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `data` | <div style="white-space:nowrap">[Data](#data)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `details` | <div style="white-space:nowrap">[DataDetails](#datadetails)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `status` | <div style="white-space:nowrap">[VirtualEnvironmentStatus](#virtualenvironmentstatus)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |







## Types






### AppDeploymentDetails

AppDeploymentDetails defines additional details of AppDeployment

<p style="font-size:.6rem;">
Used by:<br>

- <a href=#appdeployment>AppDeployment</a><br>
- <a href=#appdeploymentmanifest>AppDeploymentManifest</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `title` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `description` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `components` | <div style="white-space:nowrap">map{string, [Details](#details)}<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### AppDeploymentManifest



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#releasemanifestspec>ReleaseManifestSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `metadata` | <div style="white-space:nowrap">[ObjectRef](#objectref)<div> | <div style="max-width:30rem">Refer to Kubernetes API documentation for fields of `metadata`.</div> | <div style="white-space:nowrap">required</div> |
| `spec` | <div style="white-space:nowrap">[AppDeploymentSpec](#appdeploymentspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `details` | <div style="white-space:nowrap">[AppDeploymentDetails](#appdeploymentdetails)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### AppDeploymentSpec

AppDeploymentSpec defines the desired state of AppDeployment

<p style="font-size:.6rem;">
Used by:<br>

- <a href=#appdeployment>AppDeployment</a><br>
- <a href=#appdeploymentmanifest>AppDeploymentManifest</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `appName` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `version` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem">Version of the defined App. Use of semantic versioning is recommended. Once set the AppDeployment spec becomes immutable.</div> | <div style="white-space:nowrap"></div> |
| `commit` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required, pattern: ^[a-z0-9]{40}$</div> |
| `commitTime` | <div style="white-space:nowrap">[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#time-v1-meta)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `branch` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `tag` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `repoURL` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `containerRegistry` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `imagePullSecretName` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `components` | <div style="white-space:nowrap">map{string, ComponentDefinition}<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |



### AppDeploymentStatus

AppDeploymentStatus defines the observed state of AppDeployment

<p style="font-size:.6rem;">
Used by:<br>

- <a href=#appdeployment>AppDeployment</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `conditions` | <div style="white-space:nowrap">[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#condition-v1-meta) array<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `problems` | <div style="white-space:nowrap">[Problems](#problems)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |






### BrokerSpec



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#platformspec>PlatformSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `podSpec` | <div style="white-space:nowrap">[PodSpec](#podspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `containerSpec` | <div style="white-space:nowrap">[ContainerSpec](#containerspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### ComponentDefinition



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#appdeploymentspec>AppDeploymentSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `type` | <div style="white-space:nowrap">enum[`DBAdapter`, `KubeFox`, `HTTPAdapter`]<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `routes` | <div style="white-space:nowrap">[RouteSpec](#routespec) array<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `defaultHandler` | <div style="white-space:nowrap">boolean<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `envVarSchema` | <div style="white-space:nowrap">[EnvVarSchema](#envvarschema)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `dependencies` | <div style="white-space:nowrap">map{string, [Dependency](#dependency)}<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `hash` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required, pattern: ^[a-z0-9]{32}$</div> |
| `image` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### ComponentStatus



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#platformstatus>PlatformStatus</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `ready` | <div style="white-space:nowrap">boolean<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `name` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `hash` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `type` | <div style="white-space:nowrap">[ComponentType](#componenttype)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `podName` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `podIP` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `nodeName` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `nodeIP` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |






### ContainerSpec



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#brokerspec>BrokerSpec</a><br>
- <a href=#httpsrvspec>HTTPSrvSpec</a><br>
- <a href=#natsspec>NATSSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `resources` | <div style="white-space:nowrap">[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core)<div> | <div style="max-width:30rem">Compute Resources required by this container. Cannot be updated. [More info](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).</div> | <div style="white-space:nowrap"></div> |
| `livenessProbe` | <div style="white-space:nowrap">[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#probe-v1-core)<div> | <div style="max-width:30rem">Periodic probe of container liveness. Container will be restarted if the probe fails. Cannot be updated. [More info](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes).</div> | <div style="white-space:nowrap"></div> |
| `readinessProbe` | <div style="white-space:nowrap">[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#probe-v1-core)<div> | <div style="max-width:30rem">Periodic probe of container service readiness. Container will be removed from service endpoints if the probe fails. Cannot be updated. [More info](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes).</div> | <div style="white-space:nowrap"></div> |
| `startupProbe` | <div style="white-space:nowrap">[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#probe-v1-core)<div> | <div style="max-width:30rem">StartupProbe indicates that the Pod has successfully initialized. If specified, no other probes are executed until this completes successfully. If this probe fails, the Pod will be restarted, just as if the livenessProbe failed. This can be used to provide different probe parameters at the beginning of a Pod's lifecycle, when it might take a long time to load data or warm a cache, than during steady-state operation. This cannot be updated. [More info](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes).</div> | <div style="white-space:nowrap"></div> |



### Data



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#environment>Environment</a><br>
- <a href=#environmentmanifest>EnvironmentManifest</a><br>
- <a href=#releasemanifest>ReleaseManifest</a><br>
- <a href=#virtualenvironment>VirtualEnvironment</a><br>
- <a href=#virtualenvironmentmanifest>VirtualEnvironmentManifest</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `vars` | <div style="white-space:nowrap">map{string, [Val](#val)}<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `secrets` | <div style="white-space:nowrap">map{string, [Val](#val)}<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### DataDetails



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#environment>Environment</a><br>
- <a href=#environmentmanifest>EnvironmentManifest</a><br>
- <a href=#virtualenvironment>VirtualEnvironment</a><br>
- <a href=#virtualenvironmentmanifest>VirtualEnvironmentManifest</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `title` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `description` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `vars` | <div style="white-space:nowrap">map{string, [Details](#details)}<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `secrets` | <div style="white-space:nowrap">map{string, [Details](#details)}<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |












### Dependency



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#componentdefinition>ComponentDefinition</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `type` | <div style="white-space:nowrap">enum[`DBAdapter`, `KubeFox`, `HTTPAdapter`]<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |



### Details



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#appdeploymentdetails>AppDeploymentDetails</a><br>
- <a href=#datadetails>DataDetails</a><br>
- <a href=#httpadapter>HTTPAdapter</a><br>
- <a href=#httpadaptermanifest>HTTPAdapterManifest</a><br>
- <a href=#platformdetails>PlatformDetails</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `title` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `description` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### EnvHistoryLimits



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#envreleasepolicy>EnvReleasePolicy</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `count` | <div style="white-space:nowrap">integer<div> | <div style="max-width:30rem">Maximum number of Releases to keep in history. Once the limit is reached the oldest Release in history will be deleted. Age is based on archiveTime. Pointer is used to distinguish between not set and false.</div> | <div style="white-space:nowrap">min: 0, default: 10</div> |
| `ageDays` | <div style="white-space:nowrap">integer<div> | <div style="max-width:30rem">Maximum age of the Release to keep in history. Once the limit is reached the oldest Release in history will be deleted. Age is based on archiveTime. Set to 0 to disable. Pointer is used to distinguish between not set and false.</div> | <div style="white-space:nowrap">min: 0</div> |



### EnvReleasePolicy



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#environmentspec>EnvironmentSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `type` | <div style="white-space:nowrap">enum[`Stable`, `Testing`]<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">default: Stable</div> |
| `activationDeadlineSeconds` | <div style="white-space:nowrap">integer<div> | <div style="max-width:30rem">If the pending Release cannot be activated before the activation deadline it will be considered failed and the Release will automatically rolled back to the current active Release. Pointer is used to distinguish between not set and false.</div> | <div style="white-space:nowrap">min: 3, default: 300</div> |
| `historyLimits` | <div style="white-space:nowrap">[EnvHistoryLimits](#envhistorylimits)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### EnvSchema



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#envtemplate>EnvTemplate</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `vars` | <div style="white-space:nowrap">[EnvVarSchema](#envvarschema)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `secrets` | <div style="white-space:nowrap">[EnvVarSchema](#envvarschema)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### EnvTemplate



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#httpadaptertemplate>HTTPAdapterTemplate</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `template` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `envSchema` | <div style="white-space:nowrap">[EnvSchema](#envschema)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `tree` | <div style="white-space:nowrap">[Tree](#tree)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `parseErr` | <div style="white-space:nowrap">[error](#error)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |













### EnvironmentManifest



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#releasemanifestspec>ReleaseManifestSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `metadata` | <div style="white-space:nowrap">[ObjectRef](#objectref)<div> | <div style="max-width:30rem">Refer to Kubernetes API documentation for fields of `metadata`.</div> | <div style="white-space:nowrap">required</div> |
| `spec` | <div style="white-space:nowrap">[EnvironmentSpec](#environmentspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `data` | <div style="white-space:nowrap">[Data](#data)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `details` | <div style="white-space:nowrap">[DataDetails](#datadetails)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### EnvironmentSpec



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#environment>Environment</a><br>
- <a href=#environmentmanifest>EnvironmentManifest</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `releasePolicy` | <div style="white-space:nowrap">[EnvReleasePolicy](#envreleasepolicy)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |






### EventsSpec



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#platformspec>PlatformSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `timeoutSeconds` | <div style="white-space:nowrap">integer<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">min: 3, default: 30</div> |
| `maxSize` | <div style="white-space:nowrap">[Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/)<div> | <div style="max-width:30rem">Large events reduce performance and increase memory usage. Default 5Mi. Maximum 16Mi.</div> | <div style="white-space:nowrap"></div> |










### HTTPAdapterManifest



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#releasemanifestadapters>ReleaseManifestAdapters</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `metadata` | <div style="white-space:nowrap">[ObjectRef](#objectref)<div> | <div style="max-width:30rem">Refer to Kubernetes API documentation for fields of `metadata`.</div> | <div style="white-space:nowrap">required</div> |
| `spec` | <div style="white-space:nowrap">[HTTPAdapterSpec](#httpadapterspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `details` | <div style="white-space:nowrap">[Details](#details)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### HTTPAdapterSpec



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#httpadapter>HTTPAdapter</a><br>
- <a href=#httpadaptermanifest>HTTPAdapterManifest</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `url` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required, format: uri</div> |
| `headers` | <div style="white-space:nowrap">map{string, string}<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `insecureSkipVerify` | <div style="white-space:nowrap">boolean<div> | <div style="max-width:30rem">InsecureSkipVerify controls whether the Adapter verifies the server's certificate chain and host name. If InsecureSkipVerify is true, any certificate presented by the server and any host name in that certificate is accepted. In this mode, TLS is susceptible to machine-in-the-middle attacks.</div> | <div style="white-space:nowrap">default: false</div> |
| `followRedirects` | <div style="white-space:nowrap">enum[`Never`, `Always`, `SameHost`]<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">default: Never</div> |






### HTTPSrvPorts



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#httpsrvservice>HTTPSrvService</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `http` | <div style="white-space:nowrap">integer<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">min: 1, max: 65535, default: 80</div> |
| `https` | <div style="white-space:nowrap">integer<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">min: 1, max: 65535, default: 443</div> |



### HTTPSrvService



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#httpsrvspec>HTTPSrvSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `type` | <div style="white-space:nowrap">enum[`ClusterIP`, `NodePort`, `LoadBalancer`]<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">default: ClusterIP</div> |
| `ports` | <div style="white-space:nowrap">[HTTPSrvPorts](#httpsrvports)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `labels` | <div style="white-space:nowrap">map{string, string}<div> | <div style="max-width:30rem">Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. [More info](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels).</div> | <div style="white-space:nowrap"></div> |
| `annotations` | <div style="white-space:nowrap">map{string, string}<div> | <div style="max-width:30rem">Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. [More info](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations).</div> | <div style="white-space:nowrap"></div> |



### HTTPSrvSpec



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#platformspec>PlatformSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `podSpec` | <div style="white-space:nowrap">[PodSpec](#podspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `containerSpec` | <div style="white-space:nowrap">[ContainerSpec](#containerspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `service` | <div style="white-space:nowrap">[HTTPSrvService](#httpsrvservice)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |






### LoggerSpec



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#platformspec>PlatformSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `level` | <div style="white-space:nowrap">enum[`debug`, `info`, `warn`, `error`]<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `format` | <div style="white-space:nowrap">enum[`json`, `console`]<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### NATSSpec



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#platformspec>PlatformSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `podSpec` | <div style="white-space:nowrap">[PodSpec](#podspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `containerSpec` | <div style="white-space:nowrap">[ContainerSpec](#containerspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |






### ObjectRef



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#appdeploymentmanifest>AppDeploymentManifest</a><br>
- <a href=#environmentmanifest>EnvironmentManifest</a><br>
- <a href=#httpadaptermanifest>HTTPAdapterManifest</a><br>
- <a href=#virtualenvironmentmanifest>VirtualEnvironmentManifest</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `namespace` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `name` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `uid` | <div style="white-space:nowrap">[UID](#uid)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `resourceVersion` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required, minLength: 1</div> |
| `generation` | <div style="white-space:nowrap">integer<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |




### PlatformDetails

PlatformDetails defines additional details of Platform

<p style="font-size:.6rem;">
Used by:<br>

- <a href=#platform>Platform</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `title` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `description` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### PlatformSpec

PlatformSpec defines the desired state of Platform

<p style="font-size:.6rem;">
Used by:<br>

- <a href=#platform>Platform</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `events` | <div style="white-space:nowrap">[EventsSpec](#eventsspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `broker` | <div style="white-space:nowrap">[BrokerSpec](#brokerspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `httpsrv` | <div style="white-space:nowrap">[HTTPSrvSpec](#httpsrvspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `nats` | <div style="white-space:nowrap">[NATSSpec](#natsspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `logger` | <div style="white-space:nowrap">[LoggerSpec](#loggerspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### PlatformStatus

PlatformStatus defines the observed state of Platform

<p style="font-size:.6rem;">
Used by:<br>

- <a href=#platform>Platform</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `components` | <div style="white-space:nowrap">[ComponentStatus](#componentstatus) array<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `conditions` | <div style="white-space:nowrap">[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#condition-v1-meta) array<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### PodSpec



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#brokerspec>BrokerSpec</a><br>
- <a href=#httpsrvspec>HTTPSrvSpec</a><br>
- <a href=#natsspec>NATSSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `labels` | <div style="white-space:nowrap">map{string, string}<div> | <div style="max-width:30rem">Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. [More info](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels).</div> | <div style="white-space:nowrap"></div> |
| `annotations` | <div style="white-space:nowrap">map{string, string}<div> | <div style="max-width:30rem">Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. [More info](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations).</div> | <div style="white-space:nowrap"></div> |
| `nodeSelector` | <div style="white-space:nowrap">map{string, string}<div> | <div style="max-width:30rem">NodeSelector is a selector which must be true for the pod to fit on a node. Selector which must match a node's labels for the pod to be scheduled on that node. [More info](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/).</div> | <div style="white-space:nowrap"></div> |
| `nodeName` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem">NodeName is a request to schedule this pod onto a specific node. If it is non-empty, the scheduler simply schedules this pod onto that node, assuming that it fits resource requirements.</div> | <div style="white-space:nowrap"></div> |
| `affinity` | <div style="white-space:nowrap">[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#affinity-v1-core)<div> | <div style="max-width:30rem">If specified, the pod's scheduling constraints</div> | <div style="white-space:nowrap"></div> |
| `tolerations` | <div style="white-space:nowrap">[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#toleration-v1-core) array<div> | <div style="max-width:30rem">If specified, the pod's tolerations.</div> | <div style="white-space:nowrap"></div> |



### Problem

ObservedTime is added here instead of api package to prevent k8s.io dependencies from getting pulled into Kit.

<p style="font-size:.6rem;">
Used by:<br>

- <a href=#releasestatus>ReleaseStatus</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `type` | <div style="white-space:nowrap">[ProblemType](#problemtype)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `message` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `causes` | <div style="white-space:nowrap">[ProblemSource](#problemsource) array<div> | <div style="max-width:30rem">Resources and attributes causing problem.</div> | <div style="white-space:nowrap"></div> |
| `observedTime` | <div style="white-space:nowrap">[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#time-v1-meta)<div> | <div style="max-width:30rem">ObservedTime at which the problem was recorded.</div> | <div style="white-space:nowrap"></div> |



### ProblemSource



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#problem>Problem</a><br>
- <a href=#problem>Problem</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `kind` | <div style="white-space:nowrap">[ProblemSourceKind](#problemsourcekind)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `name` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `observedGeneration` | <div style="white-space:nowrap">integer<div> | <div style="max-width:30rem">ObservedGeneration represents the .metadata.generation of the ProblemSource that the problem was generated from. For instance, if the ProblemSource .metadata.generation is currently 12, but the observedGeneration is 9, the problem is out of date with respect to the current state of the instance.</div> | <div style="white-space:nowrap"></div> |
| `path` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem">Path of source object attribute causing problem.</div> | <div style="white-space:nowrap"></div> |
| `value` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem">Value causing problem. Pointer is used to distinguish between not set and empty string.</div> | <div style="white-space:nowrap"></div> |












### Release



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#releasestatus>ReleaseStatus</a><br>
- <a href=#virtualenvironmentspec>VirtualEnvironmentSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `apps` | <div style="white-space:nowrap">map{string, [ReleaseApp](#releaseapp)}<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |







### ReleaseManifestAdapters



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#releasemanifestspec>ReleaseManifestSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `http` | <div style="white-space:nowrap">[HTTPAdapterManifest](#httpadaptermanifest) array<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### ReleaseManifestSpec



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#releasemanifest>ReleaseManifest</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `releaseId` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required, minLength: 1</div> |
| `environment` | <div style="white-space:nowrap">[EnvironmentManifest](#environmentmanifest)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `virtualEnvironment` | <div style="white-space:nowrap">[VirtualEnvironmentManifest](#virtualenvironmentmanifest)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `appDeployments` | <div style="white-space:nowrap">[AppDeploymentManifest](#appdeploymentmanifest) array<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `adapters` | <div style="white-space:nowrap">[ReleaseManifestAdapters](#releasemanifestadapters)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### ReleasePolicy



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#virtualenvironmentspec>VirtualEnvironmentSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `type` | <div style="white-space:nowrap">enum[`Stable`, `Testing`]<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `activationDeadlineSeconds` | <div style="white-space:nowrap">integer<div> | <div style="max-width:30rem">If the pending Release cannot be activated before the activation deadline it will be considered failed and the Release will automatically rolled back to the current active Release. Pointer is used to distinguish between not set and false.</div> | <div style="white-space:nowrap">min: 3</div> |
| `historyLimits` | <div style="white-space:nowrap">[HistoryLimits](#historylimits)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### ReleaseStatus



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#virtualenvironmentstatus>VirtualEnvironmentStatus</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `apps` | <div style="white-space:nowrap">map{string, [ReleaseApp](#releaseapp)}<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `id` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required, minLength: 1</div> |
| `releaseManifest` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `requestTime` | <div style="white-space:nowrap">[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#time-v1-meta)<div> | <div style="max-width:30rem">Time at which the VirtualEnvironment was updated to use the Release.</div> | <div style="white-space:nowrap"></div> |
| `activationTime` | <div style="white-space:nowrap">[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#time-v1-meta)<div> | <div style="max-width:30rem">Time at which the Release became active. If not set the Release was never activated.</div> | <div style="white-space:nowrap"></div> |
| `archiveTime` | <div style="white-space:nowrap">[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#time-v1-meta)<div> | <div style="max-width:30rem">Time at which the Release was archived to history.</div> | <div style="white-space:nowrap"></div> |
| `archiveReason` | <div style="white-space:nowrap">enum[`PendingDeadlineExceeded`, `RolledBack`, `Superseded`]<div> | <div style="max-width:30rem">Reason Release was archived.</div> | <div style="white-space:nowrap"></div> |
| `problems` | <div style="white-space:nowrap">[Problem](#problem) array<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |






### RouteSpec



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#componentdefinition>ComponentDefinition</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `id` | <div style="white-space:nowrap">integer<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `rule` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap">required</div> |
| `priority` | <div style="white-space:nowrap">integer<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `envVarSchema` | <div style="white-space:nowrap">[EnvVarSchema](#envvarschema)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### Val



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#data>Data</a><br>
</p>








### VirtualEnvironmentManifest



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#releasemanifestspec>ReleaseManifestSpec</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `metadata` | <div style="white-space:nowrap">[ObjectRef](#objectref)<div> | <div style="max-width:30rem">Refer to Kubernetes API documentation for fields of `metadata`.</div> | <div style="white-space:nowrap">required</div> |
| `spec` | <div style="white-space:nowrap">[VirtualEnvironmentSpec](#virtualenvironmentspec)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `data` | <div style="white-space:nowrap">[Data](#data)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `details` | <div style="white-space:nowrap">[DataDetails](#datadetails)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### VirtualEnvironmentSpec



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#virtualenvironment>VirtualEnvironment</a><br>
- <a href=#virtualenvironmentmanifest>VirtualEnvironmentManifest</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `environment` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem">Name of the Environment this VirtualEnvironment is part of. This field is immutable.</div> | <div style="white-space:nowrap">required, minLength: 1</div> |
| `release` | <div style="white-space:nowrap">[Release](#release)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `releasePolicy` | <div style="white-space:nowrap">[ReleasePolicy](#releasepolicy)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |



### VirtualEnvironmentStatus



<p style="font-size:.6rem;">
Used by:<br>

- <a href=#virtualenvironment>VirtualEnvironment</a><br>
</p>

| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
| `dataChecksum` | <div style="white-space:nowrap">string<div> | <div style="max-width:30rem">DataChecksum is a hash value of the Data object. The Environment Data object is merged before the hash is created. It can be used to check for changes to the Data object.</div> | <div style="white-space:nowrap"></div> |
| `pendingReleaseFailed` | <div style="white-space:nowrap">boolean<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `activeRelease` | <div style="white-space:nowrap">[ReleaseStatus](#releasestatus)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `pendingRelease` | <div style="white-space:nowrap">[ReleaseStatus](#releasestatus)<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `releaseHistory` | <div style="white-space:nowrap">[ReleaseStatus](#releasestatus) array<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |
| `conditions` | <div style="white-space:nowrap">[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#condition-v1-meta) array<div> | <div style="max-width:30rem"></div> | <div style="white-space:nowrap"></div> |


