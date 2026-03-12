package quote

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Hunsin/compass/lib/oops"
	quoteLib "github.com/Hunsin/compass/lib/quote"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
)

// mockOHLCVStream implements grpc.ServerStreamingServer[pb.OHLCV] for tests.
type mockOHLCVStream struct {
	grpc.ServerStreamingServer[pb.OHLCV]
	sent []*pb.OHLCV
	ctx  context.Context
}

func (m *mockOHLCVStream) Send(o *pb.OHLCV) error {
	m.sent = append(m.sent, o)
	return nil
}

func (m *mockOHLCVStream) Context() context.Context { return m.ctx }

func dur(secs int64) *durationpb.Duration {
	return &durationpb.Duration{Seconds: secs}
}

func ts(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func TestCreateOHLCVs(t *testing.T) {
	validTs := ts(time.Date(2025, 12, 11, 9, 0, 0, 0, time.UTC))
	open, high, low, close_, vol := 100.0, 110.0, 98.0, 105.0, uint64(1000)
	validOHLCV := &pb.OHLCV{Ts: validTs, Open: &open, High: &high, Low: &low, Close: &close_, Volume: &vol}

	tests := []struct {
		name     string
		req      *pb.CreateOHLCVsRequest
		stub     func(*quoteLib.MockModel)
		wantCode codes.Code
	}{
		{
			name:     "missing exchange",
			req:      &pb.CreateOHLCVsRequest{Symbol: strPtr("2330"), Interval: dur(60), Ohlcv: []*pb.OHLCV{validOHLCV}},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing symbol",
			req:      &pb.CreateOHLCVsRequest{Exchange: strPtr("twse"), Interval: dur(60), Ohlcv: []*pb.OHLCV{validOHLCV}},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing interval",
			req:      &pb.CreateOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Ohlcv: []*pb.OHLCV{validOHLCV}},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "empty ohlcv",
			req:      &pb.CreateOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(60)},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "invalid interval",
			req:      &pb.CreateOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(999), Ohlcv: []*pb.OHLCV{validOHLCV}},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing timestamp in ohlcv",
			req:      &pb.CreateOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(60), Ohlcv: []*pb.OHLCV{{Open: &open}}},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "security not found",
			req:  &pb.CreateOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("9999"), Interval: dur(60), Ohlcv: []*pb.OHLCV{validOHLCV}},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().CreateOHLCVs(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(oops.NotFound("security not found"))
			},
			wantCode: codes.NotFound,
		},
		{
			name: "internal error",
			req:  &pb.CreateOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(60), Ohlcv: []*pb.OHLCV{validOHLCV}},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().CreateOHLCVs(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("db down"))
			},
			wantCode: codes.Internal,
		},
		{
			name: "success 1m",
			req:  &pb.CreateOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(60), Ohlcv: []*pb.OHLCV{validOHLCV}},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().CreateOHLCVs(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			wantCode: codes.OK,
		},
		{
			name: "success 1d",
			req:  &pb.CreateOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(86400), Ohlcv: []*pb.OHLCV{validOHLCV}},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().CreateOHLCVs(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
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
			_, err := svc.CreateOHLCVs(context.Background(), tc.req)
			assertCode(t, err, tc.wantCode)
		})
	}
}

func TestGetOHLCVs(t *testing.T) {
	from := ts(time.Date(2025, 12, 11, 9, 0, 0, 0, time.UTC))
	before := ts(time.Date(2025, 12, 11, 10, 0, 0, 0, time.UTC))
	open, high, low, close_, vol := 100.0, 110.0, 98.0, 105.0, uint64(1000)
	row := &pb.OHLCV{Ts: from, Open: &open, High: &high, Low: &low, Close: &close_, Volume: &vol}

	tests := []struct {
		name     string
		req      *pb.GetOHLCVsRequest
		stub     func(*quoteLib.MockModel)
		wantCode codes.Code
		wantSent int
	}{
		{
			name:     "missing exchange",
			req:      &pb.GetOHLCVsRequest{Symbol: strPtr("2330"), Interval: dur(60), From: from, Before: before},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing symbol",
			req:      &pb.GetOHLCVsRequest{Exchange: strPtr("twse"), Interval: dur(60), From: from, Before: before},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing interval",
			req:      &pb.GetOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), From: from, Before: before},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing from",
			req:      &pb.GetOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(60), Before: before},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing before",
			req:      &pb.GetOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(60), From: from},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "invalid interval",
			req:      &pb.GetOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(999), From: from, Before: before},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "from not before before",
			req:      &pb.GetOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(60), From: before, Before: from},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "security not found",
			req:  &pb.GetOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("9999"), Interval: dur(60), From: from, Before: before},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().GetOHLCVs(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, oops.NotFound("security not found"))
			},
			wantCode: codes.NotFound,
		},
		{
			name: "internal error",
			req:  &pb.GetOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(60), From: from, Before: before},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().GetOHLCVs(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("db down"))
			},
			wantCode: codes.Internal,
		},
		{
			name: "success streams rows",
			req:  &pb.GetOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(60), From: from, Before: before},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().GetOHLCVs(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]*pb.OHLCV{row}, nil)
			},
			wantCode: codes.OK,
			wantSent: 1,
		},
		{
			name: "empty result",
			req:  &pb.GetOHLCVsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Interval: dur(60), From: from, Before: before},
			stub: func(m *quoteLib.MockModel) {
				m.EXPECT().GetOHLCVs(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil)
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
			stream := &mockOHLCVStream{ctx: context.Background()}
			err := svc.GetOHLCVs(tc.req, stream)
			assertCode(t, err, tc.wantCode)
			if tc.wantCode == codes.OK && len(stream.sent) != tc.wantSent {
				t.Errorf("sent %d rows, want %d", len(stream.sent), tc.wantSent)
			}
		})
	}
}
