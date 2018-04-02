// Code generated by protoc-gen-go. DO NOT EDIT.
// source: trade/trade.proto

/*
Package trade is a generated protocol buffer package.

Package trade defines the fields of data types which shared
between services.

It is generated from these files:
	trade/trade.proto

It has these top-level messages:
	Daily
	Security
	Market
	Null
*/
package trade

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
	Date   string  `protobuf:"bytes,1,opt,name=date" json:"date,omitempty"`
	Open   float64 `protobuf:"fixed64,2,opt,name=open" json:"open,omitempty"`
	High   float64 `protobuf:"fixed64,3,opt,name=high" json:"high,omitempty"`
	Low    float64 `protobuf:"fixed64,4,opt,name=low" json:"low,omitempty"`
	Close  float64 `protobuf:"fixed64,5,opt,name=close" json:"close,omitempty"`
	Volume uint64  `protobuf:"varint,6,opt,name=volume" json:"volume,omitempty"`
	Avg    float64 `protobuf:"fixed64,7,opt,name=avg" json:"avg,omitempty"`
}

func (m *Daily) Reset()                    { *m = Daily{} }
func (m *Daily) String() string            { return proto.CompactTextString(m) }
func (*Daily) ProtoMessage()               {}
func (*Daily) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

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
	Market string `protobuf:"bytes,1,opt,name=market" json:"market,omitempty"`
	Isin   string `protobuf:"bytes,2,opt,name=isin" json:"isin,omitempty"`
	Symbol string `protobuf:"bytes,3,opt,name=symbol" json:"symbol,omitempty"`
	Name   string `protobuf:"bytes,4,opt,name=name" json:"name,omitempty"`
	Type   string `protobuf:"bytes,5,opt,name=type" json:"type,omitempty"`
	Listed string `protobuf:"bytes,6,opt,name=listed" json:"listed,omitempty"`
}

func (m *Security) Reset()                    { *m = Security{} }
func (m *Security) String() string            { return proto.CompactTextString(m) }
func (*Security) ProtoMessage()               {}
func (*Security) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

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
	Code     string `protobuf:"bytes,1,opt,name=code" json:"code,omitempty"`
	Name     string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Currency string `protobuf:"bytes,3,opt,name=currency" json:"currency,omitempty"`
}

func (m *Market) Reset()                    { *m = Market{} }
func (m *Market) String() string            { return proto.CompactTextString(m) }
func (*Market) ProtoMessage()               {}
func (*Market) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

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

// A Null is nothing.
type Null struct {
}

func (m *Null) Reset()                    { *m = Null{} }
func (m *Null) String() string            { return proto.CompactTextString(m) }
func (*Null) ProtoMessage()               {}
func (*Null) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func init() {
	proto.RegisterType((*Daily)(nil), "trade.Daily")
	proto.RegisterType((*Security)(nil), "trade.Security")
	proto.RegisterType((*Market)(nil), "trade.Market")
	proto.RegisterType((*Null)(nil), "trade.Null")
}

func init() { proto.RegisterFile("trade/trade.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 266 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x91, 0xc1, 0x4e, 0x84, 0x30,
	0x14, 0x45, 0xd3, 0x19, 0xc0, 0x69, 0x57, 0xda, 0x18, 0xd3, 0xb8, 0x9a, 0xb0, 0x62, 0xa5, 0x0b,
	0x7f, 0xc1, 0xa5, 0xba, 0xa8, 0x5f, 0xd0, 0x81, 0x17, 0xa6, 0xb1, 0x50, 0x52, 0x0a, 0x86, 0x9f,
	0x70, 0xe7, 0xff, 0x9a, 0xf7, 0xda, 0x10, 0x37, 0xe4, 0xdc, 0x9b, 0xde, 0x72, 0x02, 0xe2, 0x2e,
	0x06, 0xd3, 0xc1, 0x33, 0x3d, 0x9f, 0xa6, 0xe0, 0xa3, 0x97, 0x25, 0x85, 0xfa, 0x97, 0x89, 0xf2,
	0xd5, 0x58, 0xb7, 0x49, 0x29, 0x8a, 0xce, 0x44, 0x50, 0xec, 0xcc, 0x1a, 0xae, 0x89, 0xb1, 0xf3,
	0x13, 0x8c, 0xea, 0x70, 0x66, 0x0d, 0xd3, 0xc4, 0xd8, 0x5d, 0x6d, 0x7f, 0x55, 0xc7, 0xd4, 0x21,
	0xcb, 0x5b, 0x71, 0x74, 0xfe, 0x5b, 0x15, 0x54, 0x21, 0xca, 0x7b, 0x51, 0xb6, 0xce, 0xcf, 0xa0,
	0x4a, 0xea, 0x52, 0x90, 0x0f, 0xa2, 0x5a, 0xbd, 0x5b, 0x06, 0x50, 0xd5, 0x99, 0x35, 0x85, 0xce,
	0x09, 0xf7, 0x66, 0xed, 0xd5, 0x4d, 0xda, 0x9b, 0xb5, 0xaf, 0x7f, 0x98, 0x38, 0x7d, 0x42, 0xbb,
	0x04, 0x1b, 0x37, 0x9c, 0x0d, 0x26, 0x7c, 0x41, 0xcc, 0x72, 0x39, 0xa1, 0x8a, 0x9d, 0x6d, 0xd2,
	0xe3, 0x9a, 0x18, 0xcf, 0xce, 0xdb, 0x70, 0xf1, 0x8e, 0x04, 0xb9, 0xce, 0x09, 0xcf, 0x8e, 0x66,
	0x00, 0x72, 0xe4, 0x9a, 0x18, 0xbb, 0xb8, 0x4d, 0xc9, 0x91, 0x6b, 0x62, 0xdc, 0x3b, 0x3b, 0x47,
	0xe8, 0x48, 0x91, 0xeb, 0x9c, 0xea, 0x37, 0x51, 0xbd, 0xef, 0x6f, 0x6d, 0x7d, 0xb7, 0x7f, 0x28,
	0xe4, 0xfd, 0xf6, 0xc3, 0xbf, 0xdb, 0x1f, 0xc5, 0xa9, 0x5d, 0x42, 0x80, 0xb1, 0xdd, 0xb2, 0xcb,
	0x9e, 0xeb, 0x4a, 0x14, 0x1f, 0x8b, 0x73, 0x97, 0x8a, 0x7e, 0xc6, 0xcb, 0x5f, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xed, 0xf7, 0xe7, 0x5f, 0xa1, 0x01, 0x00, 0x00,
}
