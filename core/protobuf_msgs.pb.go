// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: protobuf_msgs.proto

package core

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Category int32

const (
	Category_UNKNOWN  Category = 0
	Category_MESSAGE  Category = 1
	Category_REQUEST  Category = 2
	Category_RESPONSE Category = 3
)

// Enum value maps for Category.
var (
	Category_name = map[int32]string{
		0: "UNKNOWN",
		1: "MESSAGE",
		2: "REQUEST",
		3: "RESPONSE",
	}
	Category_value = map[string]int32{
		"UNKNOWN":  0,
		"MESSAGE":  1,
		"REQUEST":  2,
		"RESPONSE": 3,
	}
)

func (x Category) Enum() *Category {
	p := new(Category)
	*p = x
	return p
}

func (x Category) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Category) Descriptor() protoreflect.EnumDescriptor {
	return file_protobuf_msgs_proto_enumTypes[0].Descriptor()
}

func (Category) Type() protoreflect.EnumType {
	return &file_protobuf_msgs_proto_enumTypes[0]
}

func (x Category) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Category.Descriptor instead.
func (Category) EnumDescriptor() ([]byte, []int) {
	return file_protobuf_msgs_proto_rawDescGZIP(), []int{0}
}

type Component struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name     string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Commit   string `protobuf:"bytes,2,opt,name=commit,proto3" json:"commit,omitempty"`
	Id       string `protobuf:"bytes,3,opt,name=id,proto3" json:"id,omitempty"`
	BrokerId string `protobuf:"bytes,4,opt,name=broker_id,json=brokerId,proto3" json:"broker_id,omitempty"`
}

func (x *Component) Reset() {
	*x = Component{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protobuf_msgs_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Component) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Component) ProtoMessage() {}

func (x *Component) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_msgs_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Component.ProtoReflect.Descriptor instead.
func (*Component) Descriptor() ([]byte, []int) {
	return file_protobuf_msgs_proto_rawDescGZIP(), []int{0}
}

func (x *Component) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Component) GetCommit() string {
	if x != nil {
		return x.Commit
	}
	return ""
}

func (x *Component) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Component) GetBrokerId() string {
	if x != nil {
		return x.BrokerId
	}
	return ""
}

type EventContext struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Deployment  string `protobuf:"bytes,1,opt,name=deployment,proto3" json:"deployment,omitempty"`
	Environment string `protobuf:"bytes,2,opt,name=environment,proto3" json:"environment,omitempty"`
	Release     string `protobuf:"bytes,3,opt,name=release,proto3" json:"release,omitempty"`
}

func (x *EventContext) Reset() {
	*x = EventContext{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protobuf_msgs_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventContext) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventContext) ProtoMessage() {}

func (x *EventContext) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_msgs_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventContext.ProtoReflect.Descriptor instead.
func (*EventContext) Descriptor() ([]byte, []int) {
	return file_protobuf_msgs_proto_rawDescGZIP(), []int{1}
}

func (x *EventContext) GetDeployment() string {
	if x != nil {
		return x.Deployment
	}
	return ""
}

func (x *EventContext) GetEnvironment() string {
	if x != nil {
		return x.Environment
	}
	return ""
}

func (x *EventContext) GetRelease() string {
	if x != nil {
		return x.Release
	}
	return ""
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ParentId string   `protobuf:"bytes,2,opt,name=parent_id,json=parentId,proto3" json:"parent_id,omitempty"`
	Type     string   `protobuf:"bytes,3,opt,name=type,proto3" json:"type,omitempty"`
	Category Category `protobuf:"varint,4,opt,name=category,proto3,enum=kubefox.proto.v1.Category" json:"category,omitempty"`
	// Unix time in µs
	CreateTime int64 `protobuf:"varint,5,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	// TTL in µs
	Ttl         int64                      `protobuf:"varint,6,opt,name=ttl,proto3" json:"ttl,omitempty"`
	Context     *EventContext              `protobuf:"bytes,7,opt,name=context,proto3" json:"context,omitempty"`
	Source      *Component                 `protobuf:"bytes,8,opt,name=source,proto3" json:"source,omitempty"`
	Target      *Component                 `protobuf:"bytes,9,opt,name=target,proto3" json:"target,omitempty"`
	Params      map[string]*structpb.Value `protobuf:"bytes,10,rep,name=params,proto3" json:"params,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Values      map[string]*structpb.Value `protobuf:"bytes,11,rep,name=values,proto3" json:"values,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	ContentType string                     `protobuf:"bytes,14,opt,name=content_type,json=contentType,proto3" json:"content_type,omitempty"`
	Content     []byte                     `protobuf:"bytes,15,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protobuf_msgs_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_msgs_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_protobuf_msgs_proto_rawDescGZIP(), []int{2}
}

func (x *Event) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Event) GetParentId() string {
	if x != nil {
		return x.ParentId
	}
	return ""
}

func (x *Event) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Event) GetCategory() Category {
	if x != nil {
		return x.Category
	}
	return Category_UNKNOWN
}

func (x *Event) GetCreateTime() int64 {
	if x != nil {
		return x.CreateTime
	}
	return 0
}

func (x *Event) GetTtl() int64 {
	if x != nil {
		return x.Ttl
	}
	return 0
}

func (x *Event) GetContext() *EventContext {
	if x != nil {
		return x.Context
	}
	return nil
}

func (x *Event) GetSource() *Component {
	if x != nil {
		return x.Source
	}
	return nil
}

func (x *Event) GetTarget() *Component {
	if x != nil {
		return x.Target
	}
	return nil
}

func (x *Event) GetParams() map[string]*structpb.Value {
	if x != nil {
		return x.Params
	}
	return nil
}

func (x *Event) GetValues() map[string]*structpb.Value {
	if x != nil {
		return x.Values
	}
	return nil
}

func (x *Event) GetContentType() string {
	if x != nil {
		return x.ContentType
	}
	return ""
}

func (x *Event) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

type MatchedEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Event   *Event                     `protobuf:"bytes,1,opt,name=event,proto3" json:"event,omitempty"`
	RouteId int64                      `protobuf:"varint,2,opt,name=route_id,json=routeId,proto3" json:"route_id,omitempty"`
	Env     map[string]*structpb.Value `protobuf:"bytes,3,rep,name=env,proto3" json:"env,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *MatchedEvent) Reset() {
	*x = MatchedEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protobuf_msgs_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MatchedEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MatchedEvent) ProtoMessage() {}

func (x *MatchedEvent) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_msgs_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MatchedEvent.ProtoReflect.Descriptor instead.
func (*MatchedEvent) Descriptor() ([]byte, []int) {
	return file_protobuf_msgs_proto_rawDescGZIP(), []int{3}
}

func (x *MatchedEvent) GetEvent() *Event {
	if x != nil {
		return x.Event
	}
	return nil
}

func (x *MatchedEvent) GetRouteId() int64 {
	if x != nil {
		return x.RouteId
	}
	return 0
}

func (x *MatchedEvent) GetEnv() map[string]*structpb.Value {
	if x != nil {
		return x.Env
	}
	return nil
}

var File_protobuf_msgs_proto protoreflect.FileDescriptor

var file_protobuf_msgs_proto_rawDesc = []byte{
	0x0a, 0x13, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x5f, 0x6d, 0x73, 0x67, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x6b, 0x75, 0x62, 0x65, 0x66, 0x6f, 0x78, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x64, 0x0a, 0x09, 0x43, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65,
	0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1b,
	0x0a, 0x09, 0x62, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x62, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x64, 0x22, 0x6a, 0x0a, 0x0c, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x64,
	0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x65,
	0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0b, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x18, 0x0a,
	0x07, 0x72, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x72, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x22, 0xb4, 0x05, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x12,
	0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x12, 0x36, 0x0a, 0x08, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x66, 0x6f, 0x78, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79,
	0x52, 0x08, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x12, 0x1f, 0x0a, 0x0b, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x74,
	0x74, 0x6c, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x74, 0x74, 0x6c, 0x12, 0x38, 0x0a,
	0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e,
	0x2e, 0x6b, 0x75, 0x62, 0x65, 0x66, 0x6f, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x76,
	0x31, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x07,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x12, 0x33, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x66, 0x6f,
	0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x6f,
	0x6e, 0x65, 0x6e, 0x74, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x33, 0x0a, 0x06,
	0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x6b,
	0x75, 0x62, 0x65, 0x66, 0x6f, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x76, 0x31, 0x2e,
	0x43, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x52, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x12, 0x3b, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x18, 0x0a, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x66, 0x6f, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x3b,
	0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x0b, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23,
	0x2e, 0x6b, 0x75, 0x62, 0x65, 0x66, 0x6f, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x76,
	0x31, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x0e, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18,
	0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x1a, 0x51, 0x0a, 0x0b, 0x50, 0x61, 0x72, 0x61,
	0x6d, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x51, 0x0a, 0x0b, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xe3,
	0x01, 0x0a, 0x0c, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12,
	0x2d, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17,
	0x2e, 0x6b, 0x75, 0x62, 0x65, 0x66, 0x6f, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x76,
	0x31, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x19,
	0x0a, 0x08, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x07, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x49, 0x64, 0x12, 0x39, 0x0a, 0x03, 0x65, 0x6e, 0x76,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x6b, 0x75, 0x62, 0x65, 0x66, 0x6f, 0x78,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x65,
	0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x45, 0x6e, 0x76, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x03, 0x65, 0x6e, 0x76, 0x1a, 0x4e, 0x0a, 0x08, 0x45, 0x6e, 0x76, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x2a, 0x3f, 0x0a, 0x08, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79,
	0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x0b, 0x0a,
	0x07, 0x4d, 0x45, 0x53, 0x53, 0x41, 0x47, 0x45, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x52, 0x45,
	0x51, 0x55, 0x45, 0x53, 0x54, 0x10, 0x02, 0x12, 0x0c, 0x0a, 0x08, 0x52, 0x45, 0x53, 0x50, 0x4f,
	0x4e, 0x53, 0x45, 0x10, 0x03, 0x42, 0x20, 0x5a, 0x1e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x78, 0x69, 0x67, 0x78, 0x6f, 0x67, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x66,
	0x6f, 0x78, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protobuf_msgs_proto_rawDescOnce sync.Once
	file_protobuf_msgs_proto_rawDescData = file_protobuf_msgs_proto_rawDesc
)

func file_protobuf_msgs_proto_rawDescGZIP() []byte {
	file_protobuf_msgs_proto_rawDescOnce.Do(func() {
		file_protobuf_msgs_proto_rawDescData = protoimpl.X.CompressGZIP(file_protobuf_msgs_proto_rawDescData)
	})
	return file_protobuf_msgs_proto_rawDescData
}

var file_protobuf_msgs_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protobuf_msgs_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_protobuf_msgs_proto_goTypes = []interface{}{
	(Category)(0),          // 0: kubefox.proto.v1.Category
	(*Component)(nil),      // 1: kubefox.proto.v1.Component
	(*EventContext)(nil),   // 2: kubefox.proto.v1.EventContext
	(*Event)(nil),          // 3: kubefox.proto.v1.Event
	(*MatchedEvent)(nil),   // 4: kubefox.proto.v1.MatchedEvent
	nil,                    // 5: kubefox.proto.v1.Event.ParamsEntry
	nil,                    // 6: kubefox.proto.v1.Event.ValuesEntry
	nil,                    // 7: kubefox.proto.v1.MatchedEvent.EnvEntry
	(*structpb.Value)(nil), // 8: google.protobuf.Value
}
var file_protobuf_msgs_proto_depIdxs = []int32{
	0,  // 0: kubefox.proto.v1.Event.category:type_name -> kubefox.proto.v1.Category
	2,  // 1: kubefox.proto.v1.Event.context:type_name -> kubefox.proto.v1.EventContext
	1,  // 2: kubefox.proto.v1.Event.source:type_name -> kubefox.proto.v1.Component
	1,  // 3: kubefox.proto.v1.Event.target:type_name -> kubefox.proto.v1.Component
	5,  // 4: kubefox.proto.v1.Event.params:type_name -> kubefox.proto.v1.Event.ParamsEntry
	6,  // 5: kubefox.proto.v1.Event.values:type_name -> kubefox.proto.v1.Event.ValuesEntry
	3,  // 6: kubefox.proto.v1.MatchedEvent.event:type_name -> kubefox.proto.v1.Event
	7,  // 7: kubefox.proto.v1.MatchedEvent.env:type_name -> kubefox.proto.v1.MatchedEvent.EnvEntry
	8,  // 8: kubefox.proto.v1.Event.ParamsEntry.value:type_name -> google.protobuf.Value
	8,  // 9: kubefox.proto.v1.Event.ValuesEntry.value:type_name -> google.protobuf.Value
	8,  // 10: kubefox.proto.v1.MatchedEvent.EnvEntry.value:type_name -> google.protobuf.Value
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_protobuf_msgs_proto_init() }
func file_protobuf_msgs_proto_init() {
	if File_protobuf_msgs_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protobuf_msgs_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Component); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protobuf_msgs_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventContext); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protobuf_msgs_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protobuf_msgs_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MatchedEvent); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protobuf_msgs_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protobuf_msgs_proto_goTypes,
		DependencyIndexes: file_protobuf_msgs_proto_depIdxs,
		EnumInfos:         file_protobuf_msgs_proto_enumTypes,
		MessageInfos:      file_protobuf_msgs_proto_msgTypes,
	}.Build()
	File_protobuf_msgs_proto = out.File
	file_protobuf_msgs_proto_rawDesc = nil
	file_protobuf_msgs_proto_goTypes = nil
	file_protobuf_msgs_proto_depIdxs = nil
}