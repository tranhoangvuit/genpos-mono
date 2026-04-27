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
	ctrl        *gomock.Controller
	tenantDB    *gatewaymock.MockTenantDB
	reader      *gatewaymock.MockOrderReader
	writer      *gatewaymock.MockOrderWriter
	storeReader *gatewaymock.MockOrgStoreReader
}

func newOrderMocks(t *testing.T) *orderMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	return &orderMocks{
		ctrl:        ctrl,
		tenantDB:    gatewaymock.NewMockTenantDB(ctrl),
		reader:      gatewaymock.NewMockOrderReader(ctrl),
		writer:      gatewaymock.NewMockOrderWriter(ctrl),
		storeReader: gatewaymock.NewMockOrgStoreReader(ctrl),
	}
}

func (m *orderMocks) newUsecase() usecase.OrderUsecase {
	return usecase.NewOrderUsecase(m.tenantDB, m.reader, m.writer, m.storeReader)
}

func (m *orderMocks) stubPassthroughRead() {
	m.tenantDB.EXPECT().
		ReadWithTenant(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ string, fn func(context.Context) error) error {
			return fn(ctx)
		}).AnyTimes()
}

func (m *orderMocks) stubPassthroughWrite() {
	m.tenantDB.EXPECT().
		WithTenant(gomock.Any(), gomock.Any(), gomock.Any()).
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

func Test_OrderUsecase_CreateOrder(t *testing.T) {
	t.Parallel()

	baseInput := func() input.CreateOrderInput {
		return input.CreateOrderInput{
			OrgID:       "org-1",
			Source:      "pos",
			ExternalID:  "ext-1",
			OrderNumber: "HD20260427000001",
			UserID:      "user-1",
			Subtotal:    "100",
			Total:       "100",
			LineItems: []input.CreateOrderLineItemInput{
				{
					ProductName: "Coffee",
					Quantity:    "1",
					UnitPrice:   "100",
					LineTotal:   "100",
				},
			},
			Payments: []input.CreateOrderPaymentInput{
				{PaymentMethodID: "pm-cash", Amount: "100"},
			},
		}
	}

	cases := map[string]struct {
		in          input.CreateOrderInput
		setup       func(*orderMocks)
		want        *entity.Order
		wantErr     bool
		wantErrCode string
	}{
		"creates a new order when no idempotency match": {
			in: baseInput(),
			setup: func(m *orderMocks) {
				m.stubPassthroughWrite()
				m.reader.EXPECT().
					GetByExternalID(gomock.Any(), "pos", "ext-1").
					Return(nil, errors.NotFound("order not found"))
				m.storeReader.EXPECT().
					FirstStoreID(gomock.Any(), "org-1").
					Return("store-1", nil)
				m.writer.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, p gateway.CreateOrderParams) (*entity.Order, error) {
						if p.StoreID != "store-1" {
							t.Errorf("StoreID: want store-1, got %s", p.StoreID)
						}
						if p.Status != "completed" {
							t.Errorf("Status: want completed, got %s", p.Status)
						}
						if p.CompletedAt.IsZero() {
							t.Error("CompletedAt: want non-zero default for status=completed")
						}
						return &entity.Order{ID: "o-1", OrderNumber: p.OrderNumber, Source: p.Source, ExternalID: p.ExternalID}, nil
					})
				m.reader.EXPECT().
					ListLineItems(gomock.Any(), "o-1").
					Return([]*entity.OrderLineItem{{ID: "li-1"}}, nil)
				m.reader.EXPECT().
					ListPayments(gomock.Any(), "o-1").
					Return([]*entity.OrderPayment{{ID: "pay-1"}}, nil)
			},
			want: &entity.Order{
				ID:          "o-1",
				OrderNumber: "HD20260427000001",
				Source:      "pos",
				ExternalID:  "ext-1",
				LineItems:   []*entity.OrderLineItem{{ID: "li-1"}},
				Payments:    []*entity.OrderPayment{{ID: "pay-1"}},
			},
		},
		"returns existing order on idempotency hit without re-inserting": {
			in: baseInput(),
			setup: func(m *orderMocks) {
				m.stubPassthroughWrite()
				m.reader.EXPECT().
					GetByExternalID(gomock.Any(), "pos", "ext-1").
					Return(&entity.Order{ID: "o-existing", OrderNumber: "HD20260427000001", Source: "pos", ExternalID: "ext-1"}, nil)
				m.reader.EXPECT().
					ListLineItems(gomock.Any(), "o-existing").
					Return([]*entity.OrderLineItem{{ID: "li-x"}}, nil)
				m.reader.EXPECT().
					ListPayments(gomock.Any(), "o-existing").
					Return([]*entity.OrderPayment{{ID: "pay-x"}}, nil)
				// writer.Create and storeReader.FirstStoreID must not be called
			},
			want: &entity.Order{
				ID:          "o-existing",
				OrderNumber: "HD20260427000001",
				Source:      "pos",
				ExternalID:  "ext-1",
				LineItems:   []*entity.OrderLineItem{{ID: "li-x"}},
				Payments:    []*entity.OrderPayment{{ID: "pay-x"}},
			},
		},
		"falls back to store from request when provided": {
			in: func() input.CreateOrderInput {
				in := baseInput()
				in.StoreID = "store-explicit"
				in.ExternalID = "" // skip idempotency lookup
				return in
			}(),
			setup: func(m *orderMocks) {
				m.stubPassthroughWrite()
				// FirstStoreID must NOT be called when StoreID is supplied
				m.writer.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, p gateway.CreateOrderParams) (*entity.Order, error) {
						if p.StoreID != "store-explicit" {
							t.Errorf("StoreID: want store-explicit, got %s", p.StoreID)
						}
						return &entity.Order{ID: "o-2"}, nil
					})
				m.reader.EXPECT().
					ListLineItems(gomock.Any(), "o-2").
					Return(nil, nil)
				m.reader.EXPECT().
					ListPayments(gomock.Any(), "o-2").
					Return(nil, nil)
			},
			want: &entity.Order{ID: "o-2"},
		},
		"rejects empty org id": {
			in:          input.CreateOrderInput{OrderNumber: "HD1", LineItems: []input.CreateOrderLineItemInput{{ProductName: "x"}}},
			setup:       func(_ *orderMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeBadRequest,
		},
		"rejects missing order number": {
			in: func() input.CreateOrderInput {
				in := baseInput()
				in.OrderNumber = ""
				return in
			}(),
			setup:       func(_ *orderMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeBadRequest,
		},
		"rejects no line items": {
			in: func() input.CreateOrderInput {
				in := baseInput()
				in.LineItems = nil
				return in
			}(),
			setup:       func(_ *orderMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeBadRequest,
		},
		"rejects pos source without user id": {
			in: func() input.CreateOrderInput {
				in := baseInput()
				in.UserID = ""
				return in
			}(),
			setup:       func(_ *orderMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeBadRequest,
		},
		"propagates writer error": {
			in: baseInput(),
			setup: func(m *orderMocks) {
				m.stubPassthroughWrite()
				m.reader.EXPECT().
					GetByExternalID(gomock.Any(), "pos", "ext-1").
					Return(nil, errors.NotFound("order not found"))
				m.storeReader.EXPECT().
					FirstStoreID(gomock.Any(), "org-1").
					Return("store-1", nil)
				m.writer.EXPECT().
					Create(gomock.Any(), gomock.Any()).
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

			got, err := m.newUsecase().CreateOrder(context.Background(), tc.in)

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
				t.Errorf("order mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
