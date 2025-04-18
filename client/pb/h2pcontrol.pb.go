// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.19.6
// source: h2pcontrol.proto

package h2pcontrol_manager

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Empty struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Empty) Reset() {
	*x = Empty{}
	mi := &file_h2pcontrol_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_h2pcontrol_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_h2pcontrol_proto_rawDescGZIP(), []int{0}
}

type HeartbeatPong struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Healthy       bool                   `protobuf:"varint,1,opt,name=healthy,proto3" json:"healthy,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HeartbeatPong) Reset() {
	*x = HeartbeatPong{}
	mi := &file_h2pcontrol_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HeartbeatPong) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartbeatPong) ProtoMessage() {}

func (x *HeartbeatPong) ProtoReflect() protoreflect.Message {
	mi := &file_h2pcontrol_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeartbeatPong.ProtoReflect.Descriptor instead.
func (*HeartbeatPong) Descriptor() ([]byte, []int) {
	return file_h2pcontrol_proto_rawDescGZIP(), []int{1}
}

func (x *HeartbeatPong) GetHealthy() bool {
	if x != nil {
		return x.Healthy
	}
	return false
}

type ServerDefinition struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ServerName    string                 `protobuf:"bytes,1,opt,name=server_name,json=serverName,proto3" json:"server_name,omitempty"`
	Port          string                 `protobuf:"bytes,2,opt,name=port,proto3" json:"port,omitempty"`
	Version       string                 `protobuf:"bytes,3,opt,name=version,proto3" json:"version,omitempty"`
	ProtoFiles    []*File                `protobuf:"bytes,4,rep,name=proto_files,json=protoFiles,proto3" json:"proto_files,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ServerDefinition) Reset() {
	*x = ServerDefinition{}
	mi := &file_h2pcontrol_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerDefinition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerDefinition) ProtoMessage() {}

func (x *ServerDefinition) ProtoReflect() protoreflect.Message {
	mi := &file_h2pcontrol_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerDefinition.ProtoReflect.Descriptor instead.
func (*ServerDefinition) Descriptor() ([]byte, []int) {
	return file_h2pcontrol_proto_rawDescGZIP(), []int{2}
}

func (x *ServerDefinition) GetServerName() string {
	if x != nil {
		return x.ServerName
	}
	return ""
}

func (x *ServerDefinition) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

func (x *ServerDefinition) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *ServerDefinition) GetProtoFiles() []*File {
	if x != nil {
		return x.ProtoFiles
	}
	return nil
}

type RegisterRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Server        *ServerDefinition      `protobuf:"bytes,1,opt,name=server,proto3" json:"server,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RegisterRequest) Reset() {
	*x = RegisterRequest{}
	mi := &file_h2pcontrol_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegisterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterRequest) ProtoMessage() {}

func (x *RegisterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_h2pcontrol_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterRequest.ProtoReflect.Descriptor instead.
func (*RegisterRequest) Descriptor() ([]byte, []int) {
	return file_h2pcontrol_proto_rawDescGZIP(), []int{3}
}

func (x *RegisterRequest) GetServer() *ServerDefinition {
	if x != nil {
		return x.Server
	}
	return nil
}

type RegisterResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Result        string                 `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RegisterResponse) Reset() {
	*x = RegisterResponse{}
	mi := &file_h2pcontrol_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegisterResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterResponse) ProtoMessage() {}

func (x *RegisterResponse) ProtoReflect() protoreflect.Message {
	mi := &file_h2pcontrol_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterResponse.ProtoReflect.Descriptor instead.
func (*RegisterResponse) Descriptor() ([]byte, []int) {
	return file_h2pcontrol_proto_rawDescGZIP(), []int{4}
}

func (x *RegisterResponse) GetResult() string {
	if x != nil {
		return x.Result
	}
	return ""
}

type FetchServerDefinition struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Description   string                 `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Addr          string                 `protobuf:"bytes,3,opt,name=addr,proto3" json:"addr,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FetchServerDefinition) Reset() {
	*x = FetchServerDefinition{}
	mi := &file_h2pcontrol_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FetchServerDefinition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FetchServerDefinition) ProtoMessage() {}

func (x *FetchServerDefinition) ProtoReflect() protoreflect.Message {
	mi := &file_h2pcontrol_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FetchServerDefinition.ProtoReflect.Descriptor instead.
func (*FetchServerDefinition) Descriptor() ([]byte, []int) {
	return file_h2pcontrol_proto_rawDescGZIP(), []int{5}
}

func (x *FetchServerDefinition) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FetchServerDefinition) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *FetchServerDefinition) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

type FetchServersResponse struct {
	state         protoimpl.MessageState   `protogen:"open.v1"`
	Servers       []*FetchServerDefinition `protobuf:"bytes,1,rep,name=servers,proto3" json:"servers,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FetchServersResponse) Reset() {
	*x = FetchServersResponse{}
	mi := &file_h2pcontrol_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FetchServersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FetchServersResponse) ProtoMessage() {}

func (x *FetchServersResponse) ProtoReflect() protoreflect.Message {
	mi := &file_h2pcontrol_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FetchServersResponse.ProtoReflect.Descriptor instead.
func (*FetchServersResponse) Descriptor() ([]byte, []int) {
	return file_h2pcontrol_proto_rawDescGZIP(), []int{6}
}

func (x *FetchServersResponse) GetServers() []*FetchServerDefinition {
	if x != nil {
		return x.Servers
	}
	return nil
}

type FetchSpecificServerRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Addr          string                 `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FetchSpecificServerRequest) Reset() {
	*x = FetchSpecificServerRequest{}
	mi := &file_h2pcontrol_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FetchSpecificServerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FetchSpecificServerRequest) ProtoMessage() {}

func (x *FetchSpecificServerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_h2pcontrol_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FetchSpecificServerRequest.ProtoReflect.Descriptor instead.
func (*FetchSpecificServerRequest) Descriptor() ([]byte, []int) {
	return file_h2pcontrol_proto_rawDescGZIP(), []int{7}
}

func (x *FetchSpecificServerRequest) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

type FetchSpecificServerResponse struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	ServerDefinition *FetchServerDefinition `protobuf:"bytes,1,opt,name=server_definition,json=serverDefinition,proto3" json:"server_definition,omitempty"`
	Proto            string                 `protobuf:"bytes,2,opt,name=proto,proto3" json:"proto,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *FetchSpecificServerResponse) Reset() {
	*x = FetchSpecificServerResponse{}
	mi := &file_h2pcontrol_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FetchSpecificServerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FetchSpecificServerResponse) ProtoMessage() {}

func (x *FetchSpecificServerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_h2pcontrol_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FetchSpecificServerResponse.ProtoReflect.Descriptor instead.
func (*FetchSpecificServerResponse) Descriptor() ([]byte, []int) {
	return file_h2pcontrol_proto_rawDescGZIP(), []int{8}
}

func (x *FetchSpecificServerResponse) GetServerDefinition() *FetchServerDefinition {
	if x != nil {
		return x.ServerDefinition
	}
	return nil
}

func (x *FetchSpecificServerResponse) GetProto() string {
	if x != nil {
		return x.Proto
	}
	return ""
}

type StubRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ServerName    string                 `protobuf:"bytes,1,opt,name=Server_name,json=ServerName,proto3" json:"Server_name,omitempty"`
	Version       string                 `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Language      string                 `protobuf:"bytes,3,opt,name=language,proto3" json:"language,omitempty"` // e.g., "python", "java"
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StubRequest) Reset() {
	*x = StubRequest{}
	mi := &file_h2pcontrol_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StubRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StubRequest) ProtoMessage() {}

func (x *StubRequest) ProtoReflect() protoreflect.Message {
	mi := &file_h2pcontrol_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StubRequest.ProtoReflect.Descriptor instead.
func (*StubRequest) Descriptor() ([]byte, []int) {
	return file_h2pcontrol_proto_rawDescGZIP(), []int{9}
}

func (x *StubRequest) GetServerName() string {
	if x != nil {
		return x.ServerName
	}
	return ""
}

func (x *StubRequest) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *StubRequest) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

type StubResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	ZipData       []byte                 `protobuf:"bytes,2,opt,name=zip_data,json=zipData,proto3" json:"zip_data,omitempty"`
	Checksum      string                 `protobuf:"bytes,3,opt,name=checksum,proto3" json:"checksum,omitempty"` // Optional checksum (e.g., SHA256)
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StubResponse) Reset() {
	*x = StubResponse{}
	mi := &file_h2pcontrol_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StubResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StubResponse) ProtoMessage() {}

func (x *StubResponse) ProtoReflect() protoreflect.Message {
	mi := &file_h2pcontrol_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StubResponse.ProtoReflect.Descriptor instead.
func (*StubResponse) Descriptor() ([]byte, []int) {
	return file_h2pcontrol_proto_rawDescGZIP(), []int{10}
}

func (x *StubResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *StubResponse) GetZipData() []byte {
	if x != nil {
		return x.ZipData
	}
	return nil
}

func (x *StubResponse) GetChecksum() string {
	if x != nil {
		return x.Checksum
	}
	return ""
}

type File struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Content       []byte                 `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *File) Reset() {
	*x = File{}
	mi := &file_h2pcontrol_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *File) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*File) ProtoMessage() {}

func (x *File) ProtoReflect() protoreflect.Message {
	mi := &file_h2pcontrol_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use File.ProtoReflect.Descriptor instead.
func (*File) Descriptor() ([]byte, []int) {
	return file_h2pcontrol_proto_rawDescGZIP(), []int{11}
}

func (x *File) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *File) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

var File_h2pcontrol_proto protoreflect.FileDescriptor

const file_h2pcontrol_proto_rawDesc = "" +
	"\n" +
	"\x10h2pcontrol.proto\x12\n" +
	"h2pcontrol\"\a\n" +
	"\x05Empty\")\n" +
	"\rHeartbeatPong\x12\x18\n" +
	"\ahealthy\x18\x01 \x01(\bR\ahealthy\"\x94\x01\n" +
	"\x10ServerDefinition\x12\x1f\n" +
	"\vserver_name\x18\x01 \x01(\tR\n" +
	"serverName\x12\x12\n" +
	"\x04port\x18\x02 \x01(\tR\x04port\x12\x18\n" +
	"\aversion\x18\x03 \x01(\tR\aversion\x121\n" +
	"\vproto_files\x18\x04 \x03(\v2\x10.h2pcontrol.FileR\n" +
	"protoFiles\"G\n" +
	"\x0fRegisterRequest\x124\n" +
	"\x06server\x18\x01 \x01(\v2\x1c.h2pcontrol.ServerDefinitionR\x06server\"*\n" +
	"\x10RegisterResponse\x12\x16\n" +
	"\x06result\x18\x01 \x01(\tR\x06result\"a\n" +
	"\x15FetchServerDefinition\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12 \n" +
	"\vdescription\x18\x02 \x01(\tR\vdescription\x12\x12\n" +
	"\x04addr\x18\x03 \x01(\tR\x04addr\"S\n" +
	"\x14FetchServersResponse\x12;\n" +
	"\aservers\x18\x01 \x03(\v2!.h2pcontrol.FetchServerDefinitionR\aservers\"0\n" +
	"\x1aFetchSpecificServerRequest\x12\x12\n" +
	"\x04addr\x18\x01 \x01(\tR\x04addr\"\x83\x01\n" +
	"\x1bFetchSpecificServerResponse\x12N\n" +
	"\x11server_definition\x18\x01 \x01(\v2!.h2pcontrol.FetchServerDefinitionR\x10serverDefinition\x12\x14\n" +
	"\x05proto\x18\x02 \x01(\tR\x05proto\"d\n" +
	"\vStubRequest\x12\x1f\n" +
	"\vServer_name\x18\x01 \x01(\tR\n" +
	"ServerName\x12\x18\n" +
	"\aversion\x18\x02 \x01(\tR\aversion\x12\x1a\n" +
	"\blanguage\x18\x03 \x01(\tR\blanguage\"Y\n" +
	"\fStubResponse\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x19\n" +
	"\bzip_data\x18\x02 \x01(\fR\azipData\x12\x1a\n" +
	"\bchecksum\x18\x03 \x01(\tR\bchecksum\"4\n" +
	"\x04File\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x18\n" +
	"\acontent\x18\x02 \x01(\fR\acontent2\xfc\x02\n" +
	"\aManager\x12<\n" +
	"\aGetStub\x12\x17.h2pcontrol.StubRequest\x1a\x18.h2pcontrol.StubResponse\x12K\n" +
	"\x0eRegisterServer\x12\x1b.h2pcontrol.RegisterRequest\x1a\x1c.h2pcontrol.RegisterResponse\x129\n" +
	"\tHeartbeat\x12\x11.h2pcontrol.Empty\x1a\x19.h2pcontrol.HeartbeatPong\x12C\n" +
	"\fFetchServers\x12\x11.h2pcontrol.Empty\x1a .h2pcontrol.FetchServersResponse\x12f\n" +
	"\x13FetchSpecificServer\x12&.h2pcontrol.FetchSpecificServerRequest\x1a'.h2pcontrol.FetchSpecificServerResponseB8\n" +
	"\x14io.h2pcontrol.clientB\n" +
	"h2pcontrolP\x01Z\x12h2pcontrol.managerb\x06proto3"

var (
	file_h2pcontrol_proto_rawDescOnce sync.Once
	file_h2pcontrol_proto_rawDescData []byte
)

func file_h2pcontrol_proto_rawDescGZIP() []byte {
	file_h2pcontrol_proto_rawDescOnce.Do(func() {
		file_h2pcontrol_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_h2pcontrol_proto_rawDesc), len(file_h2pcontrol_proto_rawDesc)))
	})
	return file_h2pcontrol_proto_rawDescData
}

var file_h2pcontrol_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_h2pcontrol_proto_goTypes = []any{
	(*Empty)(nil),                       // 0: h2pcontrol.Empty
	(*HeartbeatPong)(nil),               // 1: h2pcontrol.HeartbeatPong
	(*ServerDefinition)(nil),            // 2: h2pcontrol.ServerDefinition
	(*RegisterRequest)(nil),             // 3: h2pcontrol.RegisterRequest
	(*RegisterResponse)(nil),            // 4: h2pcontrol.RegisterResponse
	(*FetchServerDefinition)(nil),       // 5: h2pcontrol.FetchServerDefinition
	(*FetchServersResponse)(nil),        // 6: h2pcontrol.FetchServersResponse
	(*FetchSpecificServerRequest)(nil),  // 7: h2pcontrol.FetchSpecificServerRequest
	(*FetchSpecificServerResponse)(nil), // 8: h2pcontrol.FetchSpecificServerResponse
	(*StubRequest)(nil),                 // 9: h2pcontrol.StubRequest
	(*StubResponse)(nil),                // 10: h2pcontrol.StubResponse
	(*File)(nil),                        // 11: h2pcontrol.File
}
var file_h2pcontrol_proto_depIdxs = []int32{
	11, // 0: h2pcontrol.ServerDefinition.proto_files:type_name -> h2pcontrol.File
	2,  // 1: h2pcontrol.RegisterRequest.server:type_name -> h2pcontrol.ServerDefinition
	5,  // 2: h2pcontrol.FetchServersResponse.servers:type_name -> h2pcontrol.FetchServerDefinition
	5,  // 3: h2pcontrol.FetchSpecificServerResponse.server_definition:type_name -> h2pcontrol.FetchServerDefinition
	9,  // 4: h2pcontrol.Manager.GetStub:input_type -> h2pcontrol.StubRequest
	3,  // 5: h2pcontrol.Manager.RegisterServer:input_type -> h2pcontrol.RegisterRequest
	0,  // 6: h2pcontrol.Manager.Heartbeat:input_type -> h2pcontrol.Empty
	0,  // 7: h2pcontrol.Manager.FetchServers:input_type -> h2pcontrol.Empty
	7,  // 8: h2pcontrol.Manager.FetchSpecificServer:input_type -> h2pcontrol.FetchSpecificServerRequest
	10, // 9: h2pcontrol.Manager.GetStub:output_type -> h2pcontrol.StubResponse
	4,  // 10: h2pcontrol.Manager.RegisterServer:output_type -> h2pcontrol.RegisterResponse
	1,  // 11: h2pcontrol.Manager.Heartbeat:output_type -> h2pcontrol.HeartbeatPong
	6,  // 12: h2pcontrol.Manager.FetchServers:output_type -> h2pcontrol.FetchServersResponse
	8,  // 13: h2pcontrol.Manager.FetchSpecificServer:output_type -> h2pcontrol.FetchSpecificServerResponse
	9,  // [9:14] is the sub-list for method output_type
	4,  // [4:9] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_h2pcontrol_proto_init() }
func file_h2pcontrol_proto_init() {
	if File_h2pcontrol_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_h2pcontrol_proto_rawDesc), len(file_h2pcontrol_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_h2pcontrol_proto_goTypes,
		DependencyIndexes: file_h2pcontrol_proto_depIdxs,
		MessageInfos:      file_h2pcontrol_proto_msgTypes,
	}.Build()
	File_h2pcontrol_proto = out.File
	file_h2pcontrol_proto_goTypes = nil
	file_h2pcontrol_proto_depIdxs = nil
}
