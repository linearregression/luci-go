// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/luci/luci-go/logdog/api/endpoints/coordinator/services/v1/state.proto

package logdog

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// LogStreamState is the log stream state value communicated to services.
type LogStreamState struct {
	// ProtoVersion is the protobuf version for this stream.
	ProtoVersion string `protobuf:"bytes,1,opt,name=proto_version,json=protoVersion" json:"proto_version,omitempty"`
	// The log stream's secret.
	//
	// Note that the secret is returned! This is okay, since this endpoint is only
	// accessible to trusted services. The secret can be cached by services to
	// validate stream information without needing to ping the Coordinator in
	// between each update.
	Secret []byte `protobuf:"bytes,2,opt,name=secret,proto3" json:"secret,omitempty"`
	// The stream index of the log stream's terminal message. If the value is -1,
	// the log is still streaming.
	TerminalIndex int64 `protobuf:"varint,3,opt,name=terminal_index,json=terminalIndex" json:"terminal_index,omitempty"`
	// If the log stream has been archived.
	Archived bool `protobuf:"varint,4,opt,name=archived" json:"archived,omitempty"`
	// If the log stream has been purged.
	Purged bool `protobuf:"varint,5,opt,name=purged" json:"purged,omitempty"`
}

func (m *LogStreamState) Reset()                    { *m = LogStreamState{} }
func (m *LogStreamState) String() string            { return proto.CompactTextString(m) }
func (*LogStreamState) ProtoMessage()               {}
func (*LogStreamState) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *LogStreamState) GetProtoVersion() string {
	if m != nil {
		return m.ProtoVersion
	}
	return ""
}

func (m *LogStreamState) GetSecret() []byte {
	if m != nil {
		return m.Secret
	}
	return nil
}

func (m *LogStreamState) GetTerminalIndex() int64 {
	if m != nil {
		return m.TerminalIndex
	}
	return 0
}

func (m *LogStreamState) GetArchived() bool {
	if m != nil {
		return m.Archived
	}
	return false
}

func (m *LogStreamState) GetPurged() bool {
	if m != nil {
		return m.Purged
	}
	return false
}

func init() {
	proto.RegisterType((*LogStreamState)(nil), "logdog.LogStreamState")
}

func init() {
	proto.RegisterFile("github.com/luci/luci-go/logdog/api/endpoints/coordinator/services/v1/state.proto", fileDescriptor1)
}

var fileDescriptor1 = []byte{
	// 225 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x8e, 0xc1, 0x4a, 0x03, 0x31,
	0x10, 0x86, 0x89, 0xd5, 0xa5, 0x86, 0xb6, 0x87, 0x1c, 0x24, 0x78, 0x5a, 0x14, 0x61, 0x2f, 0x36,
	0x88, 0x4f, 0x21, 0x78, 0x90, 0x2d, 0x78, 0x2d, 0x69, 0x32, 0xa4, 0x03, 0xbb, 0x99, 0x65, 0x92,
	0x5d, 0x7c, 0x24, 0x1f, 0x53, 0x36, 0x5d, 0x7b, 0x19, 0xf8, 0xbe, 0xf9, 0x67, 0xf8, 0xe5, 0x57,
	0xc0, 0x7c, 0x1e, 0x4f, 0x7b, 0x47, 0xbd, 0xe9, 0x46, 0x87, 0x65, 0xbc, 0x06, 0x32, 0x1d, 0x05,
	0x4f, 0xc1, 0xd8, 0x01, 0x0d, 0x44, 0x3f, 0x10, 0xc6, 0x9c, 0x8c, 0x23, 0x62, 0x8f, 0xd1, 0x66,
	0x62, 0x93, 0x80, 0x27, 0x74, 0x90, 0xcc, 0xf4, 0x66, 0x52, 0xb6, 0x19, 0xf6, 0x03, 0x53, 0x26,
	0x55, 0x5d, 0x2e, 0x9f, 0x7e, 0x85, 0xdc, 0x7d, 0x52, 0x38, 0x64, 0x06, 0xdb, 0x1f, 0xe6, 0x80,
	0x7a, 0x96, 0xdb, 0x92, 0x39, 0x4e, 0xc0, 0x09, 0x29, 0x6a, 0x51, 0x8b, 0xe6, 0xbe, 0xdd, 0x14,
	0xf9, 0x7d, 0x71, 0xea, 0x41, 0x56, 0x09, 0x1c, 0x43, 0xd6, 0x37, 0xb5, 0x68, 0x36, 0xed, 0x42,
	0xea, 0x45, 0xee, 0x32, 0x70, 0x8f, 0xd1, 0x76, 0x47, 0x8c, 0x1e, 0x7e, 0xf4, 0xaa, 0x16, 0xcd,
	0xaa, 0xdd, 0xfe, 0xdb, 0x8f, 0x59, 0xaa, 0x47, 0xb9, 0xb6, 0xec, 0xce, 0x38, 0x81, 0xd7, 0xb7,
	0xb5, 0x68, 0xd6, 0xed, 0x95, 0xe7, 0xd7, 0xc3, 0xc8, 0x01, 0xbc, 0xbe, 0x2b, 0x9b, 0x85, 0x4e,
	0x55, 0x29, 0xf0, 0xfe, 0x17, 0x00, 0x00, 0xff, 0xff, 0x47, 0x6f, 0xc6, 0x19, 0x0d, 0x01, 0x00,
	0x00,
}
