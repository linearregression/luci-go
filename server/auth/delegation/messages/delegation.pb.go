// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/luci/luci-go/server/auth/delegation/messages/delegation.proto

/*
Package messages is a generated protocol buffer package.

It is generated from these files:
	github.com/luci/luci-go/server/auth/delegation/messages/delegation.proto

It has these top-level messages:
	DelegationToken
	Subtoken
*/
package messages

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Subtoken_Kind int32

const (
	// This is to catch old tokens that don't have 'kind' field yet.
	//
	// Tokens of this kind are interpreted as 'BEARER_DELEGATION_TOKEN' for now,
	// for compatibility. But eventually (when all backends are updated), they
	// will become invalid (and there will be no way to generate them). This is
	// needed to avoid old servers accidentally interpret tokens of kind != 0 as
	// BEARER_DELEGATION_TOKEN tokens.
	Subtoken_UNKNOWN_KIND Subtoken_Kind = 0
	// The token of this kind can be sent in X-Delegation-Token-V1 HTTP header.
	// The services will check all restrictions of the token, and will
	// authenticate requests as coming from 'delegated_identity'.
	Subtoken_BEARER_DELEGATION_TOKEN Subtoken_Kind = 1
)

var Subtoken_Kind_name = map[int32]string{
	0: "UNKNOWN_KIND",
	1: "BEARER_DELEGATION_TOKEN",
}
var Subtoken_Kind_value = map[string]int32{
	"UNKNOWN_KIND":            0,
	"BEARER_DELEGATION_TOKEN": 1,
}

func (x Subtoken_Kind) String() string {
	return proto.EnumName(Subtoken_Kind_name, int32(x))
}
func (Subtoken_Kind) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 0} }

// Signed serialized Subtoken.
//
// This message is just an envelope that carries the serialized Subtoken message
// and its signature.
//
// Next ID: 6.
type DelegationToken struct {
	// Identity of a service that signed this token.
	//
	// It can be a 'service:<app-id>' string or 'user:<service-account-email>'
	// string.
	//
	// In both cases the appropriate certificate store will be queried (via SSL)
	// for the public key to use for signature verification.
	SignerId string `protobuf:"bytes,2,opt,name=signer_id,json=signerId" json:"signer_id,omitempty"`
	// ID of a key used for making the signature.
	//
	// There can be multiple active keys at any moment in time: one used for new
	// signatures, and one being rotated out (but still valid for verification).
	//
	// The lifetime of the token indirectly depends on the lifetime of the signing
	// key, which is 24h. So delegation tokens can't live longer than 24h.
	SigningKeyId string `protobuf:"bytes,3,opt,name=signing_key_id,json=signingKeyId" json:"signing_key_id,omitempty"`
	// The signature: PKCS1_v1_5+SHA256(serialized_subtoken, signing_key_id).
	Pkcs1Sha256Sig []byte `protobuf:"bytes,4,opt,name=pkcs1_sha256_sig,json=pkcs1Sha256Sig,proto3" json:"pkcs1_sha256_sig,omitempty"`
	// Serialized Subtoken message. It's signature is stored in pkcs1_sha256_sig.
	SerializedSubtoken []byte `protobuf:"bytes,5,opt,name=serialized_subtoken,json=serializedSubtoken,proto3" json:"serialized_subtoken,omitempty"`
}

func (m *DelegationToken) Reset()                    { *m = DelegationToken{} }
func (m *DelegationToken) String() string            { return proto.CompactTextString(m) }
func (*DelegationToken) ProtoMessage()               {}
func (*DelegationToken) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *DelegationToken) GetSignerId() string {
	if m != nil {
		return m.SignerId
	}
	return ""
}

func (m *DelegationToken) GetSigningKeyId() string {
	if m != nil {
		return m.SigningKeyId
	}
	return ""
}

func (m *DelegationToken) GetPkcs1Sha256Sig() []byte {
	if m != nil {
		return m.Pkcs1Sha256Sig
	}
	return nil
}

func (m *DelegationToken) GetSerializedSubtoken() []byte {
	if m != nil {
		return m.SerializedSubtoken
	}
	return nil
}

// Identifies who delegates what authority to whom where.
//
// Next ID: 9.
type Subtoken struct {
	// What kind of token is this.
	//
	// Defines how it can be used. See comments for Kind enum.
	Kind Subtoken_Kind `protobuf:"varint,8,opt,name=kind,enum=messages.Subtoken_Kind" json:"kind,omitempty"`
	// Identifier of this subtoken as generated by the token server.
	//
	// Used for logging and tracking purposes.
	SubtokenId int64 `protobuf:"varint,4,opt,name=subtoken_id,json=subtokenId" json:"subtoken_id,omitempty"`
	// Identity whose authority is delegated.
	//
	// A string of the form "user:<email>".
	DelegatedIdentity string `protobuf:"bytes,1,opt,name=delegated_identity,json=delegatedIdentity" json:"delegated_identity,omitempty"`
	// Who requested this token.
	//
	// This can match delegated_identity if the user is delegating their own
	// identity or it can be a different id if the token is actually
	// an impersonation token.
	RequestorIdentity string `protobuf:"bytes,7,opt,name=requestor_identity,json=requestorIdentity" json:"requestor_identity,omitempty"`
	// When the token was generated (and when it becomes valid).
	//
	// Number of seconds since epoch (Unix timestamp).
	CreationTime int64 `protobuf:"varint,2,opt,name=creation_time,json=creationTime" json:"creation_time,omitempty"`
	// How long the token is considered valid (in seconds).
	ValidityDuration int32 `protobuf:"varint,3,opt,name=validity_duration,json=validityDuration" json:"validity_duration,omitempty"`
	// Who can present this token.
	//
	// Each item can be an identity string (e.g. "user:<email>"), a "group:<name>"
	// string, or special "*" string which means "Any bearer can use the token".
	Audience []string `protobuf:"bytes,5,rep,name=audience" json:"audience,omitempty"`
	// What services should accept this token.
	//
	// List of services (specified as service identities, e.g. "service:app-id")
	// that should accept this token. May also contain special "*" string, which
	// means "All services".
	Services []string `protobuf:"bytes,6,rep,name=services" json:"services,omitempty"`
}

func (m *Subtoken) Reset()                    { *m = Subtoken{} }
func (m *Subtoken) String() string            { return proto.CompactTextString(m) }
func (*Subtoken) ProtoMessage()               {}
func (*Subtoken) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Subtoken) GetKind() Subtoken_Kind {
	if m != nil {
		return m.Kind
	}
	return Subtoken_UNKNOWN_KIND
}

func (m *Subtoken) GetSubtokenId() int64 {
	if m != nil {
		return m.SubtokenId
	}
	return 0
}

func (m *Subtoken) GetDelegatedIdentity() string {
	if m != nil {
		return m.DelegatedIdentity
	}
	return ""
}

func (m *Subtoken) GetRequestorIdentity() string {
	if m != nil {
		return m.RequestorIdentity
	}
	return ""
}

func (m *Subtoken) GetCreationTime() int64 {
	if m != nil {
		return m.CreationTime
	}
	return 0
}

func (m *Subtoken) GetValidityDuration() int32 {
	if m != nil {
		return m.ValidityDuration
	}
	return 0
}

func (m *Subtoken) GetAudience() []string {
	if m != nil {
		return m.Audience
	}
	return nil
}

func (m *Subtoken) GetServices() []string {
	if m != nil {
		return m.Services
	}
	return nil
}

func init() {
	proto.RegisterType((*DelegationToken)(nil), "messages.DelegationToken")
	proto.RegisterType((*Subtoken)(nil), "messages.Subtoken")
	proto.RegisterEnum("messages.Subtoken_Kind", Subtoken_Kind_name, Subtoken_Kind_value)
}

func init() {
	proto.RegisterFile("github.com/luci/luci-go/server/auth/delegation/messages/delegation.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 431 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x92, 0x4d, 0x6f, 0xd3, 0x40,
	0x10, 0x86, 0x71, 0x3e, 0x8a, 0x3b, 0x84, 0xe0, 0x2e, 0x87, 0x5a, 0xf4, 0x40, 0x14, 0x38, 0x44,
	0xaa, 0x6a, 0x8b, 0xa2, 0x72, 0x2f, 0x8a, 0x05, 0xc6, 0xc8, 0x91, 0x9c, 0x20, 0x8e, 0x2b, 0xc7,
	0x3b, 0x72, 0x46, 0x49, 0xec, 0xe2, 0x5d, 0x57, 0x0a, 0xff, 0x8b, 0xdf, 0xc6, 0x15, 0xed, 0xc6,
	0x9b, 0x72, 0xb1, 0xbc, 0xef, 0xf3, 0x48, 0x9e, 0x7d, 0x3d, 0xf0, 0xb5, 0x24, 0xb5, 0x69, 0xd7,
	0x41, 0x51, 0xef, 0xc3, 0x5d, 0x5b, 0x90, 0x79, 0xdc, 0x94, 0x75, 0x28, 0xb1, 0x79, 0xc4, 0x26,
	0xcc, 0x5b, 0xb5, 0x09, 0x05, 0xee, 0xb0, 0xcc, 0x15, 0xd5, 0x55, 0xb8, 0x47, 0x29, 0xf3, 0x12,
	0xe5, 0x7f, 0x59, 0xf0, 0xd0, 0xd4, 0xaa, 0x66, 0xae, 0x45, 0xd3, 0x3f, 0x0e, 0xbc, 0x9a, 0x9f,
	0xf0, 0xaa, 0xde, 0x62, 0xc5, 0xae, 0xe0, 0x5c, 0x52, 0x59, 0x61, 0xc3, 0x49, 0xf8, 0xbd, 0x89,
	0x33, 0x3b, 0xcf, 0xdc, 0x63, 0x10, 0x0b, 0xf6, 0x1e, 0xc6, 0xfa, 0x9d, 0xaa, 0x92, 0x6f, 0xf1,
	0xa0, 0x8d, 0xbe, 0x31, 0x46, 0x5d, 0x9a, 0xe0, 0x21, 0x16, 0x6c, 0x06, 0xde, 0xc3, 0xb6, 0x90,
	0x1f, 0xb8, 0xdc, 0xe4, 0xb7, 0x77, 0x9f, 0xb8, 0xa4, 0xd2, 0x1f, 0x4c, 0x9c, 0xd9, 0x28, 0x1b,
	0x9b, 0x7c, 0x69, 0xe2, 0x25, 0x95, 0x2c, 0x84, 0xd7, 0x12, 0x1b, 0xca, 0x77, 0xf4, 0x1b, 0x05,
	0x97, 0xed, 0x5a, 0xe9, 0x19, 0xfc, 0xa1, 0x91, 0xd9, 0x13, 0x5a, 0x76, 0xe4, 0xdb, 0xc0, 0x75,
	0xbc, 0xde, 0xf4, 0x6f, 0x0f, 0x5c, 0x1b, 0xb1, 0x6b, 0x18, 0x6c, 0xa9, 0x12, 0xbe, 0x3b, 0x71,
	0x66, 0xe3, 0xdb, 0xcb, 0xc0, 0xde, 0x2e, 0xb0, 0x46, 0x90, 0x50, 0x25, 0x32, 0x23, 0xb1, 0xb7,
	0xf0, 0xc2, 0x7e, 0x45, 0x4f, 0xaf, 0xa7, 0xea, 0x67, 0x60, 0xa3, 0x58, 0xb0, 0x1b, 0x60, 0x5d,
	0x61, 0x28, 0x38, 0x09, 0xac, 0x14, 0xa9, 0x83, 0xef, 0x98, 0x5b, 0x5e, 0x9c, 0x48, 0xdc, 0x01,
	0xad, 0x37, 0xf8, 0xab, 0x45, 0xa9, 0xea, 0xe6, 0x49, 0x7f, 0x7e, 0xd4, 0x4f, 0xe4, 0xa4, 0xbf,
	0x83, 0x97, 0x45, 0x83, 0xa6, 0x6d, 0xae, 0x68, 0x8f, 0xa6, 0xe0, 0x7e, 0x36, 0xb2, 0xe1, 0x8a,
	0xf6, 0xc8, 0xae, 0xe1, 0xe2, 0x31, 0xdf, 0x91, 0x20, 0x75, 0xe0, 0xa2, 0x6d, 0x0c, 0x30, 0x3d,
	0x0f, 0x33, 0xcf, 0x82, 0x79, 0x97, 0xb3, 0x37, 0xe0, 0xe6, 0xad, 0x20, 0xac, 0x0a, 0xf4, 0x87,
	0x93, 0xbe, 0xfe, 0x5b, 0xf6, 0xac, 0x99, 0x5e, 0x0e, 0x2a, 0x50, 0xfa, 0x67, 0x47, 0x66, 0xcf,
	0xd3, 0x3b, 0x18, 0xe8, 0x5a, 0x98, 0x07, 0xa3, 0x1f, 0x69, 0x92, 0x2e, 0x7e, 0xa6, 0x3c, 0x89,
	0xd3, 0xb9, 0xf7, 0x8c, 0x5d, 0xc1, 0xe5, 0xe7, 0xe8, 0x3e, 0x8b, 0x32, 0x3e, 0x8f, 0xbe, 0x47,
	0x5f, 0xee, 0x57, 0xf1, 0x22, 0xe5, 0xab, 0x45, 0x12, 0xa5, 0x9e, 0xb3, 0x3e, 0x33, 0x2b, 0xf4,
	0xf1, 0x5f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x49, 0xa7, 0x07, 0xba, 0x8e, 0x02, 0x00, 0x00,
}
