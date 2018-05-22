// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/Hunsin/compass/trade/trade.proto

package trade

/*
Package trade defines the fields of data types which shared
between services.
*/

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

// A Daily represents the daily trading data of a Security on
// a specific date.
type Daily struct {
	Date                 string   `protobuf:"bytes,1,opt,name=date" json:"date,omitempty"`
	Open                 float64  `protobuf:"fixed64,2,opt,name=open" json:"open,omitempty"`
	High                 float64  `protobuf:"fixed64,3,opt,name=high" json:"high,omitempty"`
	Low                  float64  `protobuf:"fixed64,4,opt,name=low" json:"low,omitempty"`
	Close                float64  `protobuf:"fixed64,5,opt,name=close" json:"close,omitempty"`
	Volume               uint64   `protobuf:"varint,6,opt,name=volume" json:"volume,omitempty"`
	Avg                  float64  `protobuf:"fixed64,7,opt,name=avg" json:"avg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Daily) Reset()         { *m = Daily{} }
func (m *Daily) String() string { return proto.CompactTextString(m) }
func (*Daily) ProtoMessage()    {}
func (*Daily) Descriptor() ([]byte, []int) {
	return fileDescriptor_trade_ad51c53e484e6da4, []int{0}
}
func (m *Daily) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Daily.Unmarshal(m, b)
}
func (m *Daily) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Daily.Marshal(b, m, deterministic)
}
func (dst *Daily) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Daily.Merge(dst, src)
}
func (m *Daily) XXX_Size() int {
	return xxx_messageInfo_Daily.Size(m)
}
func (m *Daily) XXX_DiscardUnknown() {
	xxx_messageInfo_Daily.DiscardUnknown(m)
}

var xxx_messageInfo_Daily proto.InternalMessageInfo

func (m *Daily) GetDate() string {
	if m != nil {
		return m.Date
	}
	return ""
}

func (m *Daily) GetOpen() float64 {
	if m != nil {
		return m.Open
	}
	return 0
}

func (m *Daily) GetHigh() float64 {
	if m != nil {
		return m.High
	}
	return 0
}

func (m *Daily) GetLow() float64 {
	if m != nil {
		return m.Low
	}
	return 0
}

func (m *Daily) GetClose() float64 {
	if m != nil {
		return m.Close
	}
	return 0
}

func (m *Daily) GetVolume() uint64 {
	if m != nil {
		return m.Volume
	}
	return 0
}

func (m *Daily) GetAvg() float64 {
	if m != nil {
		return m.Avg
	}
	return 0
}

// A Security represents a financial instrument in a Market.
type Security struct {
	Market               string   `protobuf:"bytes,1,opt,name=market" json:"market,omitempty"`
	Isin                 string   `protobuf:"bytes,2,opt,name=isin" json:"isin,omitempty"`
	Symbol               string   `protobuf:"bytes,3,opt,name=symbol" json:"symbol,omitempty"`
	Name                 string   `protobuf:"bytes,4,opt,name=name" json:"name,omitempty"`
	Type                 string   `protobuf:"bytes,5,opt,name=type" json:"type,omitempty"`
	Listed               string   `protobuf:"bytes,6,opt,name=listed" json:"listed,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Security) Reset()         { *m = Security{} }
func (m *Security) String() string { return proto.CompactTextString(m) }
func (*Security) ProtoMessage()    {}
func (*Security) Descriptor() ([]byte, []int) {
	return fileDescriptor_trade_ad51c53e484e6da4, []int{1}
}
func (m *Security) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Security.Unmarshal(m, b)
}
func (m *Security) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Security.Marshal(b, m, deterministic)
}
func (dst *Security) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Security.Merge(dst, src)
}
func (m *Security) XXX_Size() int {
	return xxx_messageInfo_Security.Size(m)
}
func (m *Security) XXX_DiscardUnknown() {
	xxx_messageInfo_Security.DiscardUnknown(m)
}

var xxx_messageInfo_Security proto.InternalMessageInfo

func (m *Security) GetMarket() string {
	if m != nil {
		return m.Market
	}
	return ""
}

func (m *Security) GetIsin() string {
	if m != nil {
		return m.Isin
	}
	return ""
}

func (m *Security) GetSymbol() string {
	if m != nil {
		return m.Symbol
	}
	return ""
}

func (m *Security) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Security) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Security) GetListed() string {
	if m != nil {
		return m.Listed
	}
	return ""
}

// A Market represents an exchange where financial instruments
// are traded.
type Market struct {
	Code                 string   `protobuf:"bytes,1,opt,name=code" json:"code,omitempty"`
	Name                 string   `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Currency             string   `protobuf:"bytes,3,opt,name=currency" json:"currency,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Market) Reset()         { *m = Market{} }
func (m *Market) String() string { return proto.CompactTextString(m) }
func (*Market) ProtoMessage()    {}
func (*Market) Descriptor() ([]byte, []int) {
	return fileDescriptor_trade_ad51c53e484e6da4, []int{2}
}
func (m *Market) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Market.Unmarshal(m, b)
}
func (m *Market) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Market.Marshal(b, m, deterministic)
}
func (dst *Market) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Market.Merge(dst, src)
}
func (m *Market) XXX_Size() int {
	return xxx_messageInfo_Market.Size(m)
}
func (m *Market) XXX_DiscardUnknown() {
	xxx_messageInfo_Market.DiscardUnknown(m)
}

var xxx_messageInfo_Market proto.InternalMessageInfo

func (m *Market) GetCode() string {
	if m != nil {
		return m.Code
	}
	return ""
}

func (m *Market) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Market) GetCurrency() string {
	if m != nil {
		return m.Currency
	}
	return ""
}

func init() {
	proto.RegisterType((*Daily)(nil), "trade.Daily")
	proto.RegisterType((*Security)(nil), "trade.Security")
	proto.RegisterType((*Market)(nil), "trade.Market")
}

func init() {
	proto.RegisterFile("github.com/Hunsin/compass/trade/trade.proto", fileDescriptor_trade_ad51c53e484e6da4)
}

var fileDescriptor_trade_ad51c53e484e6da4 = []byte{
	// 280 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x91, 0xc1, 0x6a, 0xc3, 0x20,
	0x1c, 0xc6, 0xb1, 0x6d, 0xb2, 0xea, 0x69, 0xc8, 0x18, 0xb2, 0x53, 0xe9, 0xa9, 0x30, 0x68, 0x0f,
	0x7b, 0x85, 0x1d, 0x76, 0xd8, 0x2e, 0xee, 0x09, 0xac, 0x91, 0x54, 0xa6, 0x31, 0xa8, 0xc9, 0xc8,
	0x4b, 0xec, 0xb6, 0xf7, 0x1d, 0xff, 0xbf, 0x12, 0x76, 0x09, 0xbf, 0xef, 0xc3, 0xcf, 0xfc, 0x48,
	0xd8, 0x73, 0x6f, 0xf3, 0x6d, 0xba, 0x9e, 0x75, 0xf0, 0x97, 0xb7, 0x69, 0x48, 0x76, 0xb8, 0xe8,
	0xe0, 0x47, 0x95, 0xd2, 0x25, 0x47, 0xd5, 0x99, 0xf2, 0x3c, 0x8f, 0x31, 0xe4, 0xc0, 0x1b, 0x0c,
	0xc7, 0x5f, 0xc2, 0x9a, 0x57, 0x65, 0xdd, 0xc2, 0x39, 0xdb, 0x75, 0x2a, 0x1b, 0x41, 0x0e, 0xe4,
	0x44, 0x25, 0x32, 0x74, 0x61, 0x34, 0x83, 0xd8, 0x1c, 0xc8, 0x89, 0x48, 0x64, 0xe8, 0x6e, 0xb6,
	0xbf, 0x89, 0x6d, 0xe9, 0x80, 0xf9, 0x3d, 0xdb, 0xba, 0xf0, 0x2d, 0x76, 0x58, 0x01, 0xf2, 0x07,
	0xd6, 0x68, 0x17, 0x92, 0x11, 0x0d, 0x76, 0x25, 0xf0, 0x47, 0xd6, 0xce, 0xc1, 0x4d, 0xde, 0x88,
	0xf6, 0x40, 0x4e, 0x3b, 0x59, 0x13, 0xec, 0xd5, 0xdc, 0x8b, 0xbb, 0xb2, 0x57, 0x73, 0x7f, 0xfc,
	0x21, 0x6c, 0xff, 0x69, 0xf4, 0x14, 0x6d, 0x5e, 0x60, 0xe6, 0x55, 0xfc, 0x32, 0xb9, 0xca, 0xd5,
	0x04, 0x2a, 0x36, 0xd9, 0xa2, 0x47, 0x25, 0x32, 0x9c, 0x4d, 0x8b, 0xbf, 0x06, 0x87, 0x82, 0x54,
	0xd6, 0x04, 0x67, 0x07, 0xe5, 0x0d, 0x3a, 0x52, 0x89, 0x0c, 0x5d, 0x5e, 0xc6, 0xe2, 0x48, 0x25,
	0x32, 0xec, 0x9d, 0x4d, 0xd9, 0x74, 0xa8, 0x48, 0x65, 0x4d, 0xc7, 0x77, 0xd6, 0x7e, 0xac, 0x6f,
	0xd5, 0xa1, 0x5b, 0x3f, 0x14, 0xf0, 0x7a, 0xfb, 0xe6, 0xdf, 0xed, 0x4f, 0x6c, 0xaf, 0xa7, 0x18,
	0xcd, 0xa0, 0x97, 0xea, 0xb2, 0xe6, 0x6b, 0x8b, 0x3f, 0xe1, 0xe5, 0x2f, 0x00, 0x00, 0xff, 0xff,
	0x15, 0x9e, 0xc2, 0x94, 0xb3, 0x01, 0x00, 0x00,
}
