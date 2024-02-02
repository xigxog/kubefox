# Protocol Buffers




<a name="kubefoxprotov1component"></a>

### Component



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) |  |  |
| app | [string](#string) |  |  |
| name | [string](#string) |  |  |
| commit | [string](#string) |  |  |
| id | [string](#string) |  |  |
| broker_id | [string](#string) |  |  |






<a name="kubefoxprotov1event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| parent_id | [string](#string) |  |  |
| type | [string](#string) |  |  |
| category | [Category](#kubefoxprotov1category) |  |  |
| create_time | [int64](#int64) |  | Unix time in µs |
| ttl | [int64](#int64) |  | TTL in µs |
| context | [EventContext](#kubefoxprotov1eventcontext) |  |  |
| source | [Component](#kubefoxprotov1component) |  |  |
| target | [Component](#kubefoxprotov1component) |  |  |
| params | [Event.ParamsEntry](#kubefoxprotov1eventparamsentry) | repeated |  |
| values | [Event.ValuesEntry](#kubefoxprotov1eventvaluesentry) | repeated |  |
| content_type | [string](#string) |  |  |
| content | [bytes](#bytes) |  |  |






<a name="kubefoxprotov1eventparamsentry"></a>

### Event.ParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [google.protobuf.Value](#googleprotobufvalue) |  |  |






<a name="kubefoxprotov1eventvaluesentry"></a>

### Event.ValuesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [google.protobuf.Value](#googleprotobufvalue) |  |  |






<a name="kubefoxprotov1eventcontext"></a>

### EventContext



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| platform | [string](#string) |  |  |
| virtual_environment | [string](#string) |  |  |
| app_deployment | [string](#string) |  |  |
| release_manifest | [string](#string) |  |  |






<a name="kubefoxprotov1matchedevent"></a>

### MatchedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#kubefoxprotov1event) |  |  |
| route_id | [int64](#int64) |  |  |
| env | [MatchedEvent.EnvEntry](#kubefoxprotov1matchedeventenventry) | repeated |  |






<a name="kubefoxprotov1matchedeventenventry"></a>

### MatchedEvent.EnvEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [google.protobuf.Value](#googleprotobufvalue) |  |  |





 <!-- end messages -->


<a name="kubefoxprotov1category"></a>

### Category


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| MESSAGE | 1 |  |
| REQUEST | 2 |  |
| RESPONSE | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kubefoxprotov1broker"></a>

### Broker


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Subscribe | [Event](#kubefoxprotov1event) stream | [MatchedEvent](#kubefoxprotov1matchedevent) stream |  |

 <!-- end services -->


