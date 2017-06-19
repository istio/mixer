// Code generated by protoc-gen-go.
// source: TemplateExtensions.proto
// DO NOT EDIT!

/*
Package istio_mixer_v1_config_template is a generated protocol buffer package.

It is generated from these files:
	TemplateExtensions.proto

It has these top-level messages:
	Expr
*/
package istio_mixer_v1_config_template

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/protoc-gen-go/descriptor"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type TemplateVariety int32

const (
	TemplateVariety_TEMPLATE_VARIETY_UNSPECIFIED TemplateVariety = 0
	TemplateVariety_TEMPLATE_VARIETY_CHECK       TemplateVariety = 1
	TemplateVariety_TEMPLATE_VARIETY_REPORT      TemplateVariety = 2
)

var TemplateVariety_name = map[int32]string{
	0: "TEMPLATE_VARIETY_UNSPECIFIED",
	1: "TEMPLATE_VARIETY_CHECK",
	2: "TEMPLATE_VARIETY_REPORT",
}
var TemplateVariety_value = map[string]int32{
	"TEMPLATE_VARIETY_UNSPECIFIED": 0,
	"TEMPLATE_VARIETY_CHECK":       1,
	"TEMPLATE_VARIETY_REPORT":      2,
}

func (x TemplateVariety) String() string {
	return proto.EnumName(TemplateVariety_name, int32(x))
}
func (TemplateVariety) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Expr struct {
}

func (m *Expr) Reset()                    { *m = Expr{} }
func (m *Expr) String() string            { return proto.CompactTextString(m) }
func (*Expr) ProtoMessage()               {}
func (*Expr) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

var E_TemplateVariety = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.FileOptions)(nil),
	ExtensionType: (*TemplateVariety)(nil),
	Field:         72295727,
	Name:          "istio.mixer.v1.config.template.template_variety",
	Tag:           "varint,72295727,opt,name=template_variety,enum=istio.mixer.v1.config.template.TemplateVariety",
	Filename:      "TemplateExtensions.proto",
}

var E_TemplateName = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.FileOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         72295729,
	Name:          "istio.mixer.v1.config.template.template_name",
	Tag:           "bytes,72295729,opt,name=template_name",
	Filename:      "TemplateExtensions.proto",
}

func init() {
	proto.RegisterType((*Expr)(nil), "istio.mixer.v1.config.template.Expr")
	proto.RegisterEnum("istio.mixer.v1.config.template.TemplateVariety", TemplateVariety_name, TemplateVariety_value)
	proto.RegisterExtension(E_TemplateVariety)
	proto.RegisterExtension(E_TemplateName)
}

func init() { proto.RegisterFile("TemplateExtensions.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 254 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x08, 0x49, 0xcd, 0x2d,
	0xc8, 0x49, 0x2c, 0x49, 0x75, 0xad, 0x28, 0x49, 0xcd, 0x2b, 0xce, 0xcc, 0xcf, 0x2b, 0xd6, 0x2b,
	0x28, 0xca, 0x2f, 0xc9, 0x17, 0x92, 0xcb, 0x2c, 0x2e, 0xc9, 0xcc, 0xd7, 0xcb, 0xcd, 0xac, 0x48,
	0x2d, 0xd2, 0x2b, 0x33, 0xd4, 0x4b, 0xce, 0xcf, 0x4b, 0xcb, 0x4c, 0xd7, 0x2b, 0x81, 0xaa, 0x97,
	0x52, 0x48, 0xcf, 0xcf, 0x4f, 0xcf, 0x49, 0xd5, 0x07, 0xab, 0x4e, 0x2a, 0x4d, 0xd3, 0x4f, 0x49,
	0x2d, 0x4e, 0x2e, 0xca, 0x2c, 0x28, 0xc9, 0x2f, 0x82, 0x98, 0xa0, 0xc4, 0xc6, 0xc5, 0xe2, 0x5a,
	0x51, 0x50, 0xa4, 0x95, 0xc3, 0xc5, 0x0f, 0xb3, 0x25, 0x2c, 0xb1, 0x28, 0x33, 0xb5, 0xa4, 0x52,
	0x48, 0x81, 0x4b, 0x26, 0xc4, 0xd5, 0x37, 0xc0, 0xc7, 0x31, 0xc4, 0x35, 0x3e, 0xcc, 0x31, 0xc8,
	0xd3, 0x35, 0x24, 0x32, 0x3e, 0xd4, 0x2f, 0x38, 0xc0, 0xd5, 0xd9, 0xd3, 0xcd, 0xd3, 0xd5, 0x45,
	0x80, 0x41, 0x48, 0x8a, 0x4b, 0x0c, 0x43, 0x85, 0xb3, 0x87, 0xab, 0xb3, 0xb7, 0x00, 0xa3, 0x90,
	0x34, 0x97, 0x38, 0x86, 0x5c, 0x90, 0x6b, 0x80, 0x7f, 0x50, 0x88, 0x00, 0x93, 0x55, 0x16, 0x97,
	0x00, 0xcc, 0x8d, 0xf1, 0x65, 0x50, 0xeb, 0x64, 0xf4, 0x20, 0x8e, 0xd5, 0x83, 0x39, 0x56, 0xcf,
	0x2d, 0x33, 0x27, 0xd5, 0xbf, 0xa0, 0x04, 0xe4, 0x5f, 0x89, 0xf5, 0xa7, 0xf6, 0x28, 0x29, 0x30,
	0x6a, 0xf0, 0x19, 0xe9, 0xeb, 0xe1, 0xf7, 0xb3, 0x1e, 0x9a, 0x37, 0xac, 0xcc, 0xb8, 0x78, 0xe1,
	0x76, 0xe5, 0x25, 0xe6, 0xa6, 0x12, 0xb0, 0x68, 0x23, 0xc4, 0x22, 0xce, 0x24, 0x36, 0xb0, 0xb4,
	0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x0e, 0x34, 0x11, 0xe7, 0x7e, 0x01, 0x00, 0x00,
}