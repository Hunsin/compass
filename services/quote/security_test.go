package quote

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Hunsin/compass/lib/oops"
	quoteLib "github.com/Hunsin/compass/lib/quote"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
)

// mockSecuritySendStream implements grpc.ServerStreamingServer[pb.Security] for tests.
type mockSecuritySendStream struct {
	grpc.ServerStreamingServer[pb.Security]
	sent []*pb.Security
	ctx  context.Context
}

func (m *mockSecuritySendStream) Send(s *pb.Security) error {
	m.sent = append(m.sent, s)
	return nil
}

func (m *mockSecuritySendStream) Context() context.Context { return m.ctx }

// mockSecurityRecvStream implements grpc.ClientStreamingServer[pb.Security, emptypb.Empty].
type mockSecurityRecvStream struct {
	grpc.ClientStreamingServer[pb.Security, emptypb.Empty]
	messages []*pb.Security
	idx      int
	ctx      context.Context
	closed   bool
}

func (m *mockSecurityRecvStream) Recv() (*pb.Security, error) {
	if m.idx >= len(m.messages) {
		return nil, io.EOF
	}
	msg := m.messages[m.idx]
	m.idx++
	return msg, nil
}

func (m *mockSecurityRecvStream) SendAndClose(_ *emptypb.Empty) error {
	m.closed = true
	return nil
}

func (m *mockSecurityRecvStream) Context() context.Context { return m.ctx }

func TestCreateSecurities(t *testing.T) {
	validSec := &pb.Security{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Name: strPtr("TSMC")}

	tests := []struct {
		name       string
		messages   []*pb.Security
		stub       func(*quoteLib.MockModel)
		wantCode   codes.Code
		wantClosed bool
	}{
		{
			name:     "missing exchange field",
			messages: []*pb.Security{{Symbol: strPtr("2330"), Name: strPtr("TSMC")}},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing symbol field",
			messages: []*pb.Security{{Exchange: strPtr("twse"), Name: strPtr("TSMC")}},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing name field",
			messages: []*pb.Security{{Exchange: strPtr("twse"), Symbol: strPtr("2330")}},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "exchange not found",
			messages: []*pb.Security{validSec},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().CreateSecurities(mock.Anything, mock.Anything).Return(oops.NotFound("exchange not found"))
			},
			wantCode: codes.NotFound,
		},
		{
			name:     "already exists",
			messages: []*pb.Security{validSec},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().CreateSecurities(mock.Anything, mock.Anything).Return(oops.AlreadyExists("security already exists"))
			},
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "internal error",
			messages: []*pb.Security{validSec},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().CreateSecurities(mock.Anything, mock.Anything).Return(errors.New("db down"))
			},
			wantCode: codes.Internal,
		},
		{
			name:     "success",
			messages: []*pb.Security{validSec},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().CreateSecurities(mock.Anything, mock.Anything).Return(nil)
			},
			wantCode:   codes.OK,
			wantClosed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := quoteLib.NewMockModel(t)
			if tc.stub != nil {
				tc.stub(m)
			}
			svc := New(m, zerolog.Nop())
			stream := &mockSecurityRecvStream{messages: tc.messages, ctx: context.Background()}
			err := svc.CreateSecurities(stream)
			assertCode(t, err, tc.wantCode)
			if tc.wantClosed && !stream.closed {
				t.Error("expected SendAndClose to be called")
			}
		})
	}
}

func TestGetSecurities(t *testing.T) {
	exch, sym, name := "twse", "2330", "TSMC"

	tests := []struct {
		name     string
		req      *pb.Exchange
		stub     func(*quoteLib.MockModel)
		wantCode codes.Code
		wantSent int
	}{
		{
			name:     "missing abbr",
			req:      &pb.Exchange{},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "exchange not found",
			req:  &pb.Exchange{Abbr: strPtr("twse")},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().GetSecurities(mock.Anything, mock.Anything).Return(nil, oops.NotFound("exchange not found"))
			},
			wantCode: codes.NotFound,
		},
		{
			name: "internal error",
			req:  &pb.Exchange{Abbr: strPtr("twse")},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().GetSecurities(mock.Anything, mock.Anything).Return(nil, errors.New("db down"))
			},
			wantCode: codes.Internal,
		},
		{
			name: "success with securities",
			req:  &pb.Exchange{Abbr: strPtr("twse")},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().GetSecurities(mock.Anything, mock.Anything).
					Return([]*pb.Security{{Exchange: &exch, Symbol: &sym, Name: &name}}, nil)
			},
			wantCode: codes.OK,
			wantSent: 1,
		},
		{
			name: "exchange exists but no securities",
			req:  &pb.Exchange{Abbr: strPtr("twse")},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().GetSecurities(mock.Anything, mock.Anything).Return(nil, nil)
			},
			wantCode: codes.OK,
			wantSent: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := quoteLib.NewMockModel(t)
			if tc.stub != nil {
				tc.stub(m)
			}
			svc := New(m, zerolog.Nop())
			stream := &mockSecuritySendStream{ctx: context.Background()}
			err := svc.GetSecurities(tc.req, stream)
			assertCode(t, err, tc.wantCode)
			if tc.wantCode == codes.OK && len(stream.sent) != tc.wantSent {
				t.Errorf("sent %d securities, want %d", len(stream.sent), tc.wantSent)
			}
		})
	}
}
