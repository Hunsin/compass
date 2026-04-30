package statistics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Hunsin/compass/lib/oops"
	statsLib "github.com/Hunsin/compass/lib/statistics"
	pb "github.com/Hunsin/compass/protocols/gen/go/statistics/v1"
)

// mockMarginTransactionStream implements grpc.ServerStreamingServer[pb.MarginTransaction] for tests.
type mockMarginTransactionStream struct {
	grpc.ServerStreamingServer[pb.MarginTransaction]
	sent []*pb.MarginTransaction
	ctx  context.Context
}

func (m *mockMarginTransactionStream) Send(tx *pb.MarginTransaction) error {
	m.sent = append(m.sent, tx)
	return nil
}

func (m *mockMarginTransactionStream) Context() context.Context { return m.ctx }

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

func int64Ptr(v int64) *int64 { return &v }

func validTx() *pb.MarginTransaction {
	ts := timestamppb.New(time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC))
	return &pb.MarginTransaction{
		Date:                        ts,
		MarginPurchaseBuy:           int64Ptr(100),
		MarginPurchaseRedemption:    int64Ptr(50),
		MarginPurchaseCashRepayment: int64Ptr(10),
		MarginPurchaseBalance:       int64Ptr(500),
		MarginPurchaseLimit:         int64Ptr(1000),
		ShortSale:                   int64Ptr(80),
		ShortSaleRedemption:         int64Ptr(40),
		ShortSaleStockRepayment:     int64Ptr(5),
		ShortSaleBalance:            int64Ptr(300),
		ShortSaleLimit:              int64Ptr(800),
		QuotaNextDay:                int64Ptr(200),
	}
}

func TestCreateMarginTransactions(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.CreateMarginTransactionsRequest
		stub     func(*statsLib.MockModel)
		wantCode codes.Code
	}{
		{
			name:     "missing exchange",
			req:      &pb.CreateMarginTransactionsRequest{Symbol: strPtr("2330"), MarginTransactions: []*pb.MarginTransaction{validTx()}},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing symbol",
			req:      &pb.CreateMarginTransactionsRequest{Exchange: strPtr("twse"), MarginTransactions: []*pb.MarginTransaction{validTx()}},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "empty margin_transactions",
			req:      &pb.CreateMarginTransactionsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330")},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "missing date in transaction",
			req: &pb.CreateMarginTransactionsRequest{
				Exchange:           strPtr("twse"),
				Symbol:             strPtr("2330"),
				MarginTransactions: []*pb.MarginTransaction{{}},
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "security not found",
			req: &pb.CreateMarginTransactionsRequest{
				Exchange:           strPtr("twse"),
				Symbol:             strPtr("2330"),
				MarginTransactions: []*pb.MarginTransaction{validTx()},
			},
			stub: func(m *statsLib.MockModel) {
				m.EXPECT().CreateMarginTransactions(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(oops.NotFound("security 2330 not found"))
			},
			wantCode: codes.NotFound,
		},
		{
			name: "already exists",
			req: &pb.CreateMarginTransactionsRequest{
				Exchange:           strPtr("twse"),
				Symbol:             strPtr("2330"),
				MarginTransactions: []*pb.MarginTransaction{validTx()},
			},
			stub: func(m *statsLib.MockModel) {
				m.EXPECT().CreateMarginTransactions(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(oops.AlreadyExists("already exists"))
			},
			wantCode: codes.AlreadyExists,
		},
		{
			name: "internal error",
			req: &pb.CreateMarginTransactionsRequest{
				Exchange:           strPtr("twse"),
				Symbol:             strPtr("2330"),
				MarginTransactions: []*pb.MarginTransaction{validTx()},
			},
			stub: func(m *statsLib.MockModel) {
				m.EXPECT().CreateMarginTransactions(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(oops.Internal(errors.New("db down")))
			},
			wantCode: codes.Internal,
		},
		{
			name: "success",
			req: &pb.CreateMarginTransactionsRequest{
				Exchange:           strPtr("twse"),
				Symbol:             strPtr("2330"),
				MarginTransactions: []*pb.MarginTransaction{validTx()},
			},
			stub: func(m *statsLib.MockModel) {
				m.EXPECT().CreateMarginTransactions(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			wantCode: codes.OK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := statsLib.NewMockModel(t)
			if tc.stub != nil {
				tc.stub(m)
			}

			_, err := New(m).CreateMarginTransactions(context.Background(), tc.req)
			assertCode(t, err, tc.wantCode)
		})
	}
}

func TestGetMarginTransactions(t *testing.T) {
	from := timestamppb.New(time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))
	before := timestamppb.New(time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		req      *pb.GetMarginTransactionsRequest
		stub     func(*statsLib.MockModel)
		wantCode codes.Code
		wantSent int
	}{
		{
			name:     "missing exchange",
			req:      &pb.GetMarginTransactionsRequest{Symbol: strPtr("2330"), From: from, Before: before},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing symbol",
			req:      &pb.GetMarginTransactionsRequest{Exchange: strPtr("twse"), From: from, Before: before},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing from",
			req:      &pb.GetMarginTransactionsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), Before: before},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing before",
			req:      &pb.GetMarginTransactionsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), From: from},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "from not before before",
			req:      &pb.GetMarginTransactionsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), From: before, Before: from},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "security not found",
			req:  &pb.GetMarginTransactionsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), From: from, Before: before},
			stub: func(m *statsLib.MockModel) {
				m.EXPECT().GetMarginTransactions(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, oops.NotFound("security 2330 not found"))
			},
			wantCode: codes.NotFound,
		},
		{
			name: "internal error",
			req:  &pb.GetMarginTransactionsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), From: from, Before: before},
			stub: func(m *statsLib.MockModel) {
				m.EXPECT().GetMarginTransactions(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, oops.Internal(errors.New("db down")))
			},
			wantCode: codes.Internal,
		},
		{
			name: "success with results",
			req:  &pb.GetMarginTransactionsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), From: from, Before: before},
			stub: func(m *statsLib.MockModel) {
				m.EXPECT().GetMarginTransactions(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]*pb.MarginTransaction{validTx()}, nil)
			},
			wantCode: codes.OK,
			wantSent: 1,
		},
		{
			name: "success with no results",
			req:  &pb.GetMarginTransactionsRequest{Exchange: strPtr("twse"), Symbol: strPtr("2330"), From: from, Before: before},
			stub: func(m *statsLib.MockModel) {
				m.EXPECT().GetMarginTransactions(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil)
			},
			wantCode: codes.OK,
			wantSent: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := statsLib.NewMockModel(t)
			if tc.stub != nil {
				tc.stub(m)
			}

			stream := &mockMarginTransactionStream{ctx: context.Background()}
			err := New(m).GetMarginTransactions(tc.req, stream)
			assertCode(t, err, tc.wantCode)
			if tc.wantCode == codes.OK && len(stream.sent) != tc.wantSent {
				t.Errorf("sent %d transactions, want %d", len(stream.sent), tc.wantSent)
			}
		})
	}
}
