# Conditions

## Platform

| Type      | Status | Reason             | Description                                                        |
| --------- | ------ | ------------------ | ------------------------------------------------------------------ |
| Available | True   | PlatformAvailable  | The KubeFox Broker, HTTP Server, and NATS event bus are available. |
|           | False  | BrokerUnavailable  | The KubeFox Broker is not available.                               |
|           |        | HTTPSrvUnavailable | The KubeFox HTTP Server is not available.                          |
|           |        | NATSUnavailable    | The NATS event bus is not available.                               |

## AppDeployment

| Type      | Status | Reason | Description                                   |
| --------- | ------ | ------ | --------------------------------------------- |
| Available | True   |        | All component Pods have Ready condition.      |
|           | False  |        | One or more component Pods is not ready.      |
| Deployed  | True   |        | All components were successfully deployed.    |
|           | False  |        | One or more components could not be deployed. |

## Release

| Type                   | Status | Reason                   | Description                                                                                                                  |
| ---------------------- | ------ | ------------------------ | ---------------------------------------------------------------------------------------------------------------------------- |
| Available              | True   |                          | The AppDeployment and VirtualEnv specified are available.                                                                    |
|                        | False  |                          | The AppDeployment or VirtualEnv specified are not available.                                                                 |
| AppDeploymentAvailable | True   |                          | The AppDeployment specified exists, matches the Release version, and is available.                                           |
|                        | False  | AppDeploymentNotFound    | The AppDeployment does not exist.                                                                                            |
|                        |        | AppDeploymentUnavailable | The AppDeployment has one or more component Pods that are not ready.                                                         |
|                        |        | VersionMismatch          | The AppDeployment and Release versions do not match.                                                                         |
| VirtualEnvAvailable    | True   |                          | The VirtualEnv specified has correctly typed variables, all required variables, and all required adapters present and valid. |
|                        | False  | VarMissing               | The VirtualEnv has one or more required variables missing.                                                                   |
|                        |        | VarConflict              | The VirtualEnv has one or more unique variables in conflict.                                                                 |
|                        |        | VarWrongType             | The VirtualEnv has one or more variables of the incorrect type.                                                              |
|                        |        | AdapterMissing           | The VirtualEnv has one or more Adapters missing.                                                                             |
|                        |        | AdapterInvalid           | The VirtualEnv has one or more Adapters that could not be validated.                                                         |
