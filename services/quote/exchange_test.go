package quote

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	quoteLib "github.com/Hunsin/compass/lib/quote"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
)

// mockExchangeStream implements grpc.ServerStreamingServer[pb.Exchange] for tests.
type mockExchangeStream struct {
	grpc.ServerStreamingServer[pb.Exchange]
	sent []*pb.Exchange
	ctx  context.Context
}

func (m *mockExchangeStream) Send(ex *pb.Exchange) error {
	m.sent = append(m.sent, ex)
	return nil
}

func (m *mockExchangeStream) Context() context.Context {
	return m.ctx
}

// assertCode checks that err has the expected gRPC code.
// Use codes.OK to assert no error.
func assertCode(t *testing.T, err error, want codes.Code) {
	t.Helper()
	if want == codes.OK {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		return
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Errorf("expected gRPC status error, got: %v", err)
		return
	}
	if st.Code() != want {
		t.Errorf("code: got %v, want %v", st.Code(), want)
	}
}

func strPtr(s string) *string { return &s }

func TestCreateExchange(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.Exchange
		stub     func(*quoteLib.MockModel)
		wantCode codes.Code
	}{
		{
			name:     "missing abbr",
			req:      &pb.Exchange{Name: strPtr("TWSE"), Timezone: strPtr("Asia/Taipei")},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing name",
			req:      &pb.Exchange{Abbr: strPtr("twse"), Timezone: strPtr("Asia/Taipei")},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing timezone",
			req:      &pb.Exchange{Abbr: strPtr("twse"), Name: strPtr("TWSE")},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "invalid timezone",
			req:      &pb.Exchange{Abbr: strPtr("twse"), Name: strPtr("TWSE"), Timezone: strPtr("Not/AZone")},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "already exists",
			req:  &pb.Exchange{Abbr: strPtr("twse"), Name: strPtr("TWSE"), Timezone: strPtr("Asia/Taipei")},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().CreateExchange(mock.Anything, mock.Anything).Return(quoteLib.ErrAlreadyExists)
			},
			wantCode: codes.AlreadyExists,
		},
		{
			name: "internal error",
			req:  &pb.Exchange{Abbr: strPtr("twse"), Name: strPtr("TWSE"), Timezone: strPtr("Asia/Taipei")},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().CreateExchange(mock.Anything, mock.Anything).Return(errors.New("db down"))
			},
			wantCode: codes.Internal,
		},
		{
			name: "success",
			req:  &pb.Exchange{Abbr: strPtr("twse"), Name: strPtr("TWSE"), Timezone: strPtr("Asia/Taipei")},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().CreateExchange(mock.Anything, mock.Anything).Return(nil)
			},
			wantCode: codes.OK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := quoteLib.NewMockModel(t)
			if tc.stub != nil {
				tc.stub(m)
			}
			svc := New(m, zerolog.Nop())
			_, err := svc.CreateExchange(context.Background(), tc.req)
			assertCode(t, err, tc.wantCode)
		})
	}
}

func TestGetExchanges_Success(t *testing.T) {
	abbr, name, tz := "twse", "TWSE", "Asia/Taipei"
	m := quoteLib.NewMockModel(t)
	m.On("GetExchanges", mock.Anything).Return([]*pb.Exchange{{Abbr: &abbr, Name: &name, Timezone: &tz}}, nil)
	svc := New(m, zerolog.Nop())
	stream := &mockExchangeStream{ctx: context.Background()}

	if err := svc.GetExchanges(&emptypb.Empty{}, stream); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stream.sent) != 1 {
		t.Fatalf("expected 1 exchange, got %d", len(stream.sent))
	}
	if stream.sent[0].GetAbbr() != "twse" {
		t.Errorf("abbr: got %q, want %q", stream.sent[0].GetAbbr(), "twse")
	}
}

func TestGetExchanges_Error(t *testing.T) {
	m := quoteLib.NewMockModel(t)
	m.On("GetExchanges", mock.Anything).Return(nil, errors.New("db down"))
	svc := New(m, zerolog.Nop())
	stream := &mockExchangeStream{ctx: context.Background()}
	assertCode(t, svc.GetExchanges(&emptypb.Empty{}, stream), codes.Internal)
}
