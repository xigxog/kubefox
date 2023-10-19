# Protocol Buffers




<a name="kubefoxprotov1component"></a>

### Component



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| commit | [string](#string) |  |  |
| id | [string](#string) |  |  |






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
| deployment | [string](#string) |  |  |
| environment | [string](#string) |  |  |
| release | [string](#string) |  |  |






<a name="kubefoxprotov1matchedevent"></a>

### MatchedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#kubefoxprotov1event) |  |  |
| env | [MatchedEvent.EnvEntry](#kubefoxprotov1matchedeventenventry) | repeated |  |
| route_id | [int64](#int64) |  |  |






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



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |
