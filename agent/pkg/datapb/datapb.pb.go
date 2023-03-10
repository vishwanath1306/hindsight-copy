// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.12.0
// source: datapb.proto

package datapb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Trigger struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	QueueId     int32    `protobuf:"varint,1,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
	BaseTraceId uint64   `protobuf:"fixed64,2,opt,name=base_trace_id,json=baseTraceId,proto3" json:"base_trace_id,omitempty"`
	TraceIds    []uint64 `protobuf:"fixed64,3,rep,packed,name=trace_ids,json=traceIds,proto3" json:"trace_ids,omitempty"`
}

func (x *Trigger) Reset() {
	*x = Trigger{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datapb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Trigger) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Trigger) ProtoMessage() {}

func (x *Trigger) ProtoReflect() protoreflect.Message {
	mi := &file_datapb_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Trigger.ProtoReflect.Descriptor instead.
func (*Trigger) Descriptor() ([]byte, []int) {
	return file_datapb_proto_rawDescGZIP(), []int{0}
}

func (x *Trigger) GetQueueId() int32 {
	if x != nil {
		return x.QueueId
	}
	return 0
}

func (x *Trigger) GetBaseTraceId() uint64 {
	if x != nil {
		return x.BaseTraceId
	}
	return 0
}

func (x *Trigger) GetTraceIds() []uint64 {
	if x != nil {
		return x.TraceIds
	}
	return nil
}

type TriggerRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Src      string     `protobuf:"bytes,1,opt,name=src,proto3" json:"src,omitempty"`
	Triggers []*Trigger `protobuf:"bytes,2,rep,name=triggers,proto3" json:"triggers,omitempty"`
}

func (x *TriggerRequest) Reset() {
	*x = TriggerRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datapb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TriggerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TriggerRequest) ProtoMessage() {}

func (x *TriggerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_datapb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TriggerRequest.ProtoReflect.Descriptor instead.
func (*TriggerRequest) Descriptor() ([]byte, []int) {
	return file_datapb_proto_rawDescGZIP(), []int{1}
}

func (x *TriggerRequest) GetSrc() string {
	if x != nil {
		return x.Src
	}
	return ""
}

func (x *TriggerRequest) GetTriggers() []*Trigger {
	if x != nil {
		return x.Triggers
	}
	return nil
}

type TriggerReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TriggerReply) Reset() {
	*x = TriggerReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datapb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TriggerReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TriggerReply) ProtoMessage() {}

func (x *TriggerReply) ProtoReflect() protoreflect.Message {
	mi := &file_datapb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TriggerReply.ProtoReflect.Descriptor instead.
func (*TriggerReply) Descriptor() ([]byte, []int) {
	return file_datapb_proto_rawDescGZIP(), []int{2}
}

type BreadcrumbAddress struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   int32  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Addr string `protobuf:"bytes,2,opt,name=addr,proto3" json:"addr,omitempty"`
}

func (x *BreadcrumbAddress) Reset() {
	*x = BreadcrumbAddress{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datapb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BreadcrumbAddress) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BreadcrumbAddress) ProtoMessage() {}

func (x *BreadcrumbAddress) ProtoReflect() protoreflect.Message {
	mi := &file_datapb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BreadcrumbAddress.ProtoReflect.Descriptor instead.
func (*BreadcrumbAddress) Descriptor() ([]byte, []int) {
	return file_datapb_proto_rawDescGZIP(), []int{3}
}

func (x *BreadcrumbAddress) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *BreadcrumbAddress) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

type Breadcrumbs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TraceId uint64  `protobuf:"fixed64,1,opt,name=trace_id,json=traceId,proto3" json:"trace_id,omitempty"`
	Addrs   []int32 `protobuf:"varint,2,rep,packed,name=addrs,proto3" json:"addrs,omitempty"`
}

func (x *Breadcrumbs) Reset() {
	*x = Breadcrumbs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datapb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Breadcrumbs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Breadcrumbs) ProtoMessage() {}

func (x *Breadcrumbs) ProtoReflect() protoreflect.Message {
	mi := &file_datapb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Breadcrumbs.ProtoReflect.Descriptor instead.
func (*Breadcrumbs) Descriptor() ([]byte, []int) {
	return file_datapb_proto_rawDescGZIP(), []int{4}
}

func (x *Breadcrumbs) GetTraceId() uint64 {
	if x != nil {
		return x.TraceId
	}
	return 0
}

func (x *Breadcrumbs) GetAddrs() []int32 {
	if x != nil {
		return x.Addrs
	}
	return nil
}

type BreadcrumbsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Src         string               `protobuf:"bytes,1,opt,name=src,proto3" json:"src,omitempty"`
	Addresses   []*BreadcrumbAddress `protobuf:"bytes,2,rep,name=addresses,proto3" json:"addresses,omitempty"`
	Breadcrumbs []*Breadcrumbs       `protobuf:"bytes,3,rep,name=breadcrumbs,proto3" json:"breadcrumbs,omitempty"`
}

func (x *BreadcrumbsRequest) Reset() {
	*x = BreadcrumbsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datapb_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BreadcrumbsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BreadcrumbsRequest) ProtoMessage() {}

func (x *BreadcrumbsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_datapb_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BreadcrumbsRequest.ProtoReflect.Descriptor instead.
func (*BreadcrumbsRequest) Descriptor() ([]byte, []int) {
	return file_datapb_proto_rawDescGZIP(), []int{5}
}

func (x *BreadcrumbsRequest) GetSrc() string {
	if x != nil {
		return x.Src
	}
	return ""
}

func (x *BreadcrumbsRequest) GetAddresses() []*BreadcrumbAddress {
	if x != nil {
		return x.Addresses
	}
	return nil
}

func (x *BreadcrumbsRequest) GetBreadcrumbs() []*Breadcrumbs {
	if x != nil {
		return x.Breadcrumbs
	}
	return nil
}

type BreadcrumbsReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *BreadcrumbsReply) Reset() {
	*x = BreadcrumbsReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datapb_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BreadcrumbsReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BreadcrumbsReply) ProtoMessage() {}

func (x *BreadcrumbsReply) ProtoReflect() protoreflect.Message {
	mi := &file_datapb_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BreadcrumbsReply.ProtoReflect.Descriptor instead.
func (*BreadcrumbsReply) Descriptor() ([]byte, []int) {
	return file_datapb_proto_rawDescGZIP(), []int{6}
}

var File_datapb_proto protoreflect.FileDescriptor

var file_datapb_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x64, 0x61, 0x74, 0x61, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x64, 0x61, 0x74, 0x61, 0x70, 0x62, 0x22, 0x65, 0x0a, 0x07, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65,
	0x72, 0x12, 0x19, 0x0a, 0x08, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x07, 0x71, 0x75, 0x65, 0x75, 0x65, 0x49, 0x64, 0x12, 0x22, 0x0a, 0x0d,
	0x62, 0x61, 0x73, 0x65, 0x5f, 0x74, 0x72, 0x61, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x06, 0x52, 0x0b, 0x62, 0x61, 0x73, 0x65, 0x54, 0x72, 0x61, 0x63, 0x65, 0x49, 0x64,
	0x12, 0x1b, 0x0a, 0x09, 0x74, 0x72, 0x61, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x06, 0x52, 0x08, 0x74, 0x72, 0x61, 0x63, 0x65, 0x49, 0x64, 0x73, 0x22, 0x4f, 0x0a,
	0x0e, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x10, 0x0a, 0x03, 0x73, 0x72, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x72,
	0x63, 0x12, 0x2b, 0x0a, 0x08, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x70, 0x62, 0x2e, 0x54, 0x72, 0x69,
	0x67, 0x67, 0x65, 0x72, 0x52, 0x08, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x22, 0x0e,
	0x0a, 0x0c, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x37,
	0x0a, 0x11, 0x42, 0x72, 0x65, 0x61, 0x64, 0x63, 0x72, 0x75, 0x6d, 0x62, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x64, 0x64, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x61, 0x64, 0x64, 0x72, 0x22, 0x3e, 0x0a, 0x0b, 0x42, 0x72, 0x65, 0x61, 0x64,
	0x63, 0x72, 0x75, 0x6d, 0x62, 0x73, 0x12, 0x19, 0x0a, 0x08, 0x74, 0x72, 0x61, 0x63, 0x65, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x06, 0x52, 0x07, 0x74, 0x72, 0x61, 0x63, 0x65, 0x49,
	0x64, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x64, 0x64, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x05,
	0x52, 0x05, 0x61, 0x64, 0x64, 0x72, 0x73, 0x22, 0x96, 0x01, 0x0a, 0x12, 0x42, 0x72, 0x65, 0x61,
	0x64, 0x63, 0x72, 0x75, 0x6d, 0x62, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10,
	0x0a, 0x03, 0x73, 0x72, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x72, 0x63,
	0x12, 0x37, 0x0a, 0x09, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x70, 0x62, 0x2e, 0x42, 0x72, 0x65,
	0x61, 0x64, 0x63, 0x72, 0x75, 0x6d, 0x62, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x09,
	0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x12, 0x35, 0x0a, 0x0b, 0x62, 0x72, 0x65,
	0x61, 0x64, 0x63, 0x72, 0x75, 0x6d, 0x62, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13,
	0x2e, 0x64, 0x61, 0x74, 0x61, 0x70, 0x62, 0x2e, 0x42, 0x72, 0x65, 0x61, 0x64, 0x63, 0x72, 0x75,
	0x6d, 0x62, 0x73, 0x52, 0x0b, 0x62, 0x72, 0x65, 0x61, 0x64, 0x63, 0x72, 0x75, 0x6d, 0x62, 0x73,
	0x22, 0x12, 0x0a, 0x10, 0x42, 0x72, 0x65, 0x61, 0x64, 0x63, 0x72, 0x75, 0x6d, 0x62, 0x73, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x32, 0x48, 0x0a, 0x05, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x12, 0x3f, 0x0a,
	0x0d, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x12, 0x16,
	0x2e, 0x64, 0x61, 0x74, 0x61, 0x70, 0x62, 0x2e, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x70, 0x62, 0x2e,
	0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x32, 0x94,
	0x01, 0x0a, 0x0b, 0x43, 0x6f, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x3e,
	0x0a, 0x0c, 0x4c, 0x6f, 0x63, 0x61, 0x6c, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x12, 0x16,
	0x2e, 0x64, 0x61, 0x74, 0x61, 0x70, 0x62, 0x2e, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x70, 0x62, 0x2e,
	0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x45,
	0x0a, 0x0b, 0x42, 0x72, 0x65, 0x61, 0x64, 0x63, 0x72, 0x75, 0x6d, 0x62, 0x73, 0x12, 0x1a, 0x2e,
	0x64, 0x61, 0x74, 0x61, 0x70, 0x62, 0x2e, 0x42, 0x72, 0x65, 0x61, 0x64, 0x63, 0x72, 0x75, 0x6d,
	0x62, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x64, 0x61, 0x74, 0x61,
	0x70, 0x62, 0x2e, 0x42, 0x72, 0x65, 0x61, 0x64, 0x63, 0x72, 0x75, 0x6d, 0x62, 0x73, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x09, 0x5a, 0x07, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_datapb_proto_rawDescOnce sync.Once
	file_datapb_proto_rawDescData = file_datapb_proto_rawDesc
)

func file_datapb_proto_rawDescGZIP() []byte {
	file_datapb_proto_rawDescOnce.Do(func() {
		file_datapb_proto_rawDescData = protoimpl.X.CompressGZIP(file_datapb_proto_rawDescData)
	})
	return file_datapb_proto_rawDescData
}

var file_datapb_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_datapb_proto_goTypes = []interface{}{
	(*Trigger)(nil),            // 0: datapb.Trigger
	(*TriggerRequest)(nil),     // 1: datapb.TriggerRequest
	(*TriggerReply)(nil),       // 2: datapb.TriggerReply
	(*BreadcrumbAddress)(nil),  // 3: datapb.BreadcrumbAddress
	(*Breadcrumbs)(nil),        // 4: datapb.Breadcrumbs
	(*BreadcrumbsRequest)(nil), // 5: datapb.BreadcrumbsRequest
	(*BreadcrumbsReply)(nil),   // 6: datapb.BreadcrumbsReply
}
var file_datapb_proto_depIdxs = []int32{
	0, // 0: datapb.TriggerRequest.triggers:type_name -> datapb.Trigger
	3, // 1: datapb.BreadcrumbsRequest.addresses:type_name -> datapb.BreadcrumbAddress
	4, // 2: datapb.BreadcrumbsRequest.breadcrumbs:type_name -> datapb.Breadcrumbs
	1, // 3: datapb.Agent.RemoteTrigger:input_type -> datapb.TriggerRequest
	1, // 4: datapb.Coordinator.LocalTrigger:input_type -> datapb.TriggerRequest
	5, // 5: datapb.Coordinator.Breadcrumbs:input_type -> datapb.BreadcrumbsRequest
	2, // 6: datapb.Agent.RemoteTrigger:output_type -> datapb.TriggerReply
	2, // 7: datapb.Coordinator.LocalTrigger:output_type -> datapb.TriggerReply
	6, // 8: datapb.Coordinator.Breadcrumbs:output_type -> datapb.BreadcrumbsReply
	6, // [6:9] is the sub-list for method output_type
	3, // [3:6] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_datapb_proto_init() }
func file_datapb_proto_init() {
	if File_datapb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_datapb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Trigger); i {
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
		file_datapb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TriggerRequest); i {
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
		file_datapb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TriggerReply); i {
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
		file_datapb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BreadcrumbAddress); i {
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
		file_datapb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Breadcrumbs); i {
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
		file_datapb_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BreadcrumbsRequest); i {
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
		file_datapb_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BreadcrumbsReply); i {
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
			RawDescriptor: file_datapb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_datapb_proto_goTypes,
		DependencyIndexes: file_datapb_proto_depIdxs,
		MessageInfos:      file_datapb_proto_msgTypes,
	}.Build()
	File_datapb_proto = out.File
	file_datapb_proto_rawDesc = nil
	file_datapb_proto_goTypes = nil
	file_datapb_proto_depIdxs = nil
}
