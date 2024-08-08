// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.25.3
// source: proto/finalitygadget.proto

package proto

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

type BlockInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// block_hash is the hash of the block
	BlockHash string `protobuf:"bytes,1,opt,name=block_hash,json=blockHash,proto3" json:"block_hash,omitempty"`
	// block_height is the height of the block
	BlockHeight uint64 `protobuf:"varint,2,opt,name=block_height,json=blockHeight,proto3" json:"block_height,omitempty"`
	// block_timestamp is the unix timestamp of the block
	BlockTimestamp uint64 `protobuf:"varint,3,opt,name=block_timestamp,json=blockTimestamp,proto3" json:"block_timestamp,omitempty"`
}

func (x *BlockInfo) Reset() {
	*x = BlockInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_finalitygadget_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockInfo) ProtoMessage() {}

func (x *BlockInfo) ProtoReflect() protoreflect.Message {
	mi := &file_proto_finalitygadget_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockInfo.ProtoReflect.Descriptor instead.
func (*BlockInfo) Descriptor() ([]byte, []int) {
	return file_proto_finalitygadget_proto_rawDescGZIP(), []int{0}
}

func (x *BlockInfo) GetBlockHash() string {
	if x != nil {
		return x.BlockHash
	}
	return ""
}

func (x *BlockInfo) GetBlockHeight() uint64 {
	if x != nil {
		return x.BlockHeight
	}
	return 0
}

func (x *BlockInfo) GetBlockTimestamp() uint64 {
	if x != nil {
		return x.BlockTimestamp
	}
	return 0
}

type InsertBlockResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *InsertBlockResponse) Reset() {
	*x = InsertBlockResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_finalitygadget_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InsertBlockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InsertBlockResponse) ProtoMessage() {}

func (x *InsertBlockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_finalitygadget_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InsertBlockResponse.ProtoReflect.Descriptor instead.
func (*InsertBlockResponse) Descriptor() ([]byte, []int) {
	return file_proto_finalitygadget_proto_rawDescGZIP(), []int{1}
}

type GetBlockStatusByHeightRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// block_height is the height of the block
	BlockHeight uint64 `protobuf:"varint,1,opt,name=block_height,json=blockHeight,proto3" json:"block_height,omitempty"`
}

func (x *GetBlockStatusByHeightRequest) Reset() {
	*x = GetBlockStatusByHeightRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_finalitygadget_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBlockStatusByHeightRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBlockStatusByHeightRequest) ProtoMessage() {}

func (x *GetBlockStatusByHeightRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_finalitygadget_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBlockStatusByHeightRequest.ProtoReflect.Descriptor instead.
func (*GetBlockStatusByHeightRequest) Descriptor() ([]byte, []int) {
	return file_proto_finalitygadget_proto_rawDescGZIP(), []int{2}
}

func (x *GetBlockStatusByHeightRequest) GetBlockHeight() uint64 {
	if x != nil {
		return x.BlockHeight
	}
	return 0
}

type GetBlockStatusByHashRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// block_hash is the hash of the block
	BlockHash string `protobuf:"bytes,1,opt,name=block_hash,json=blockHash,proto3" json:"block_hash,omitempty"`
}

func (x *GetBlockStatusByHashRequest) Reset() {
	*x = GetBlockStatusByHashRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_finalitygadget_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBlockStatusByHashRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBlockStatusByHashRequest) ProtoMessage() {}

func (x *GetBlockStatusByHashRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_finalitygadget_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBlockStatusByHashRequest.ProtoReflect.Descriptor instead.
func (*GetBlockStatusByHashRequest) Descriptor() ([]byte, []int) {
	return file_proto_finalitygadget_proto_rawDescGZIP(), []int{3}
}

func (x *GetBlockStatusByHashRequest) GetBlockHash() string {
	if x != nil {
		return x.BlockHash
	}
	return ""
}

type GetBlockStatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// is_finalized is true if the block is finalized
	IsFinalized bool `protobuf:"varint,1,opt,name=is_finalized,json=isFinalized,proto3" json:"is_finalized,omitempty"`
}

func (x *GetBlockStatusResponse) Reset() {
	*x = GetBlockStatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_finalitygadget_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBlockStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBlockStatusResponse) ProtoMessage() {}

func (x *GetBlockStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_finalitygadget_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBlockStatusResponse.ProtoReflect.Descriptor instead.
func (*GetBlockStatusResponse) Descriptor() ([]byte, []int) {
	return file_proto_finalitygadget_proto_rawDescGZIP(), []int{4}
}

func (x *GetBlockStatusResponse) GetIsFinalized() bool {
	if x != nil {
		return x.IsFinalized
	}
	return false
}

type GetLatestBlockRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetLatestBlockRequest) Reset() {
	*x = GetLatestBlockRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_finalitygadget_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetLatestBlockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetLatestBlockRequest) ProtoMessage() {}

func (x *GetLatestBlockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_finalitygadget_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetLatestBlockRequest.ProtoReflect.Descriptor instead.
func (*GetLatestBlockRequest) Descriptor() ([]byte, []int) {
	return file_proto_finalitygadget_proto_rawDescGZIP(), []int{5}
}

var File_proto_finalitygadget_proto protoreflect.FileDescriptor

var file_proto_finalitygadget_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x74, 0x79,
	0x67, 0x61, 0x64, 0x67, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x76, 0x0a, 0x09, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x49, 0x6e, 0x66, 0x6f,
	0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x12,
	0x21, 0x0a, 0x0c, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67,
	0x68, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0e, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x15, 0x0a, 0x13, 0x49,
	0x6e, 0x73, 0x65, 0x72, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x42, 0x0a, 0x1d, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x42, 0x79, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x0c, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x68, 0x65, 0x69,
	0x67, 0x68, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x22, 0x3c, 0x0a, 0x1b, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x79, 0x48, 0x61, 0x73, 0x68, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x68,
	0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x48, 0x61, 0x73, 0x68, 0x22, 0x3b, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x21,
	0x0a, 0x0c, 0x69, 0x73, 0x5f, 0x66, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x69, 0x73, 0x46, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x7a, 0x65,
	0x64, 0x22, 0x17, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x4c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x32, 0xc9, 0x02, 0x0a, 0x0e, 0x46,
	0x69, 0x6e, 0x61, 0x6c, 0x69, 0x74, 0x79, 0x47, 0x61, 0x64, 0x67, 0x65, 0x74, 0x12, 0x3b, 0x0a,
	0x0b, 0x49, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x10, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x1a,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x49, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x5d, 0x0a, 0x16, 0x47, 0x65,
	0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x79, 0x48, 0x65,
	0x69, 0x67, 0x68, 0x74, 0x12, 0x24, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x74,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x79, 0x48, 0x65, 0x69,
	0x67, 0x68, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x59, 0x0a, 0x14, 0x47, 0x65, 0x74,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x79, 0x48, 0x61, 0x73,
	0x68, 0x12, 0x22, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x79, 0x48, 0x61, 0x73, 0x68, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x47, 0x65,
	0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x40, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x4c, 0x61, 0x74, 0x65, 0x73,
	0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x1c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x47,
	0x65, 0x74, 0x4c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x10, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x61, 0x62, 0x79, 0x6c, 0x6f, 0x6e, 0x6c, 0x61, 0x62, 0x73,
	0x2d, 0x69, 0x6f, 0x2f, 0x66, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x74, 0x79, 0x2d, 0x67, 0x61, 0x64,
	0x67, 0x65, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_proto_finalitygadget_proto_rawDescOnce sync.Once
	file_proto_finalitygadget_proto_rawDescData = file_proto_finalitygadget_proto_rawDesc
)

func file_proto_finalitygadget_proto_rawDescGZIP() []byte {
	file_proto_finalitygadget_proto_rawDescOnce.Do(func() {
		file_proto_finalitygadget_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_finalitygadget_proto_rawDescData)
	})
	return file_proto_finalitygadget_proto_rawDescData
}

var file_proto_finalitygadget_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_proto_finalitygadget_proto_goTypes = []interface{}{
	(*BlockInfo)(nil),                     // 0: proto.BlockInfo
	(*InsertBlockResponse)(nil),           // 1: proto.InsertBlockResponse
	(*GetBlockStatusByHeightRequest)(nil), // 2: proto.GetBlockStatusByHeightRequest
	(*GetBlockStatusByHashRequest)(nil),   // 3: proto.GetBlockStatusByHashRequest
	(*GetBlockStatusResponse)(nil),        // 4: proto.GetBlockStatusResponse
	(*GetLatestBlockRequest)(nil),         // 5: proto.GetLatestBlockRequest
}
var file_proto_finalitygadget_proto_depIdxs = []int32{
	0, // 0: proto.FinalityGadget.InsertBlock:input_type -> proto.BlockInfo
	2, // 1: proto.FinalityGadget.GetBlockStatusByHeight:input_type -> proto.GetBlockStatusByHeightRequest
	3, // 2: proto.FinalityGadget.GetBlockStatusByHash:input_type -> proto.GetBlockStatusByHashRequest
	5, // 3: proto.FinalityGadget.GetLatestBlock:input_type -> proto.GetLatestBlockRequest
	1, // 4: proto.FinalityGadget.InsertBlock:output_type -> proto.InsertBlockResponse
	4, // 5: proto.FinalityGadget.GetBlockStatusByHeight:output_type -> proto.GetBlockStatusResponse
	4, // 6: proto.FinalityGadget.GetBlockStatusByHash:output_type -> proto.GetBlockStatusResponse
	0, // 7: proto.FinalityGadget.GetLatestBlock:output_type -> proto.BlockInfo
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_finalitygadget_proto_init() }
func file_proto_finalitygadget_proto_init() {
	if File_proto_finalitygadget_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_finalitygadget_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockInfo); i {
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
		file_proto_finalitygadget_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InsertBlockResponse); i {
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
		file_proto_finalitygadget_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBlockStatusByHeightRequest); i {
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
		file_proto_finalitygadget_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBlockStatusByHashRequest); i {
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
		file_proto_finalitygadget_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBlockStatusResponse); i {
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
		file_proto_finalitygadget_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetLatestBlockRequest); i {
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
			RawDescriptor: file_proto_finalitygadget_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_finalitygadget_proto_goTypes,
		DependencyIndexes: file_proto_finalitygadget_proto_depIdxs,
		MessageInfos:      file_proto_finalitygadget_proto_msgTypes,
	}.Build()
	File_proto_finalitygadget_proto = out.File
	file_proto_finalitygadget_proto_rawDesc = nil
	file_proto_finalitygadget_proto_goTypes = nil
	file_proto_finalitygadget_proto_depIdxs = nil
}
