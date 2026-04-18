package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	gatewaymock "github.com/genpick/genpos-mono/backend/internal/domain/gateway/mock"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type orderMocks struct {
	ctrl     *gomock.Controller
	tenantDB *gatewaymock.MockTenantDB
	reader   *gatewaymock.MockOrderReader
}

func newOrderMocks(t *testing.T) *orderMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	return &orderMocks{
		ctrl:     ctrl,
		tenantDB: gatewaymock.NewMockTenantDB(ctrl),
		reader:   gatewaymock.NewMockOrderReader(ctrl),
	}
}

func (m *orderMocks) newUsecase() usecase.OrderUsecase {
	return usecase.NewOrderUsecase(m.tenantDB, m.reader)
}

func (m *orderMocks) stubPassthroughRead() {
	m.tenantDB.EXPECT().
		ReadWithTenant(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ string, fn func(context.Context) error) error {
			return fn(ctx)
		}).AnyTimes()
}

func Test_OrderUsecase_ListOrders(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)

	cases := map[string]struct {
		in          input.ListDailySalesInput
		setup       func(*orderMocks)
		want        []*entity.OrderSummary
		wantErr     bool
		wantErrCode string
	}{
		"returns orders for org + date range": {
			in: input.ListDailySalesInput{
				OrgID:    "org-1",
				StoreID:  "",
				DateFrom: from,
				DateTo:   to,
			},
			setup: func(m *orderMocks) {
				m.stubPassthroughRead()
				m.reader.EXPECT().
					ListByDateRange(gomock.Any(), gateway.ListOrdersParams{
						DateFrom: from,
						DateTo:   to,
						StoreID:  "",
					}).
					Return([]*entity.OrderSummary{
						{ID: "o1", OrderNumber: "ORD-001", Total: "10.0000"},
						{ID: "o2", OrderNumber: "ORD-002", Total: "20.0000"},
					}, nil)
			},
			want: []*entity.OrderSummary{
				{ID: "o1", OrderNumber: "ORD-001", Total: "10.0000"},
				{ID: "o2", OrderNumber: "ORD-002", Total: "20.0000"},
			},
		},
		"forwards store filter": {
			in: input.ListDailySalesInput{
				OrgID:    "org-1",
				StoreID:  "store-7",
				DateFrom: from,
				DateTo:   to,
			},
			setup: func(m *orderMocks) {
				m.stubPassthroughRead()
				m.reader.EXPECT().
					ListByDateRange(gomock.Any(), gateway.ListOrdersParams{
						DateFrom: from,
						DateTo:   to,
						StoreID:  "store-7",
					}).
					Return([]*entity.OrderSummary{}, nil)
			},
			want: []*entity.OrderSummary{},
		},
		"rejects empty org id": {
			in:          input.ListDailySalesInput{DateFrom: from, DateTo: to},
			setup:       func(_ *orderMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeBadRequest,
		},
		"rejects zero date_from": {
			in: input.ListDailySalesInput{
				OrgID:  "org-1",
				DateTo: to,
			},
			setup:       func(_ *orderMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeBadRequest,
		},
		"rejects reversed range": {
			in: input.ListDailySalesInput{
				OrgID:    "org-1",
				DateFrom: to,
				DateTo:   from,
			},
			setup:       func(_ *orderMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeBadRequest,
		},
		"propagates reader error": {
			in: input.ListDailySalesInput{
				OrgID:    "org-1",
				DateFrom: from,
				DateTo:   to,
			},
			setup: func(m *orderMocks) {
				m.stubPassthroughRead()
				m.reader.EXPECT().
					ListByDateRange(gomock.Any(), gomock.Any()).
					Return(nil, errors.Internal("db boom"))
			},
			wantErr:     true,
			wantErrCode: errors.CodeInternal,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := newOrderMocks(t)
			tc.setup(m)

			got, err := m.newUsecase().ListOrders(context.Background(), tc.in)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.wantErrCode != "" && errors.GetCode(err) != tc.wantErrCode {
					t.Errorf("error code: want %s, got %s", tc.wantErrCode, errors.GetCode(err))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("orders mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
