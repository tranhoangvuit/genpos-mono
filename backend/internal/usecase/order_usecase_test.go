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
	ctrl         *gomock.Controller
	tenantDB     *gatewaymock.MockTenantDB
	reader       *gatewaymock.MockOrderReader
	writer       *gatewaymock.MockOrderWriter
	storeReader  *gatewaymock.MockOrgStoreReader
	memberReader *gatewaymock.MockMemberReader
}

func newOrderMocks(t *testing.T) *orderMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	return &orderMocks{
		ctrl:         ctrl,
		tenantDB:     gatewaymock.NewMockTenantDB(ctrl),
		reader:       gatewaymock.NewMockOrderReader(ctrl),
		writer:       gatewaymock.NewMockOrderWriter(ctrl),
		storeReader:  gatewaymock.NewMockOrgStoreReader(ctrl),
		memberReader: gatewaymock.NewMockMemberReader(ctrl),
	}
}

func (m *orderMocks) newUsecase() usecase.OrderUsecase {
	return usecase.NewOrderUsecase(m.tenantDB, m.reader, m.writer, m.storeReader, m.memberReader)
}

func (m *orderMocks) allowStoreAccess() {
	m.memberReader.EXPECT().
		HasStoreAccess(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(true, nil).
		AnyTimes()
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
				m.allowStoreAccess()
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
				m.reader.EXPECT().
					ListOrderAdjustments(gomock.Any(), "o-1").
					Return(nil, nil)
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
				m.reader.EXPECT().
					ListOrderAdjustments(gomock.Any(), "o-existing").
					Return(nil, nil)
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
				m.allowStoreAccess()
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
				m.reader.EXPECT().
					ListOrderAdjustments(gomock.Any(), "o-2").
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
				m.allowStoreAccess()
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
		"recomputes aggregates from line taxes and adjustments when supplied": {
			in: func() input.CreateOrderInput {
				in := baseInput()
				in.StoreID = "store-explicit"
				in.ExternalID = "" // skip idempotency lookup
				in.Subtotal = "9999"
				in.TaxTotal = "9999"
				in.DiscountTotal = "9999"
				in.Total = "9999"
				in.LineItems = []input.CreateOrderLineItemInput{{
					ProductName: "Coffee",
					Quantity:    "2",
					UnitPrice:   "100",
					LineTotal:   "200",
					Taxes: []input.CreateOrderLineItemTaxInput{
						{Sequence: 1, NameSnapshot: "VAT", RateSnapshot: "0.10", Amount: "20"},
					},
					Adjustments: []input.CreateOrderLineAdjustmentInput{
						{Sequence: 1, Kind: "discount", SourceType: "manual", NameSnapshot: "Promo", CalculationType: "fixed_amount", Amount: "-30", AppliesBeforeTax: true},
					},
				}}
				return in
			}(),
			setup: func(m *orderMocks) {
				m.stubPassthroughWrite()
				m.allowStoreAccess()
				m.writer.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, p gateway.CreateOrderParams) (*entity.Order, error) {
						// Subtotal = qty*unit = 200; tax = 20 (sum of taxes[]);
						// discount = 30 (abs(sum(neg adjustments))); total = 200+20-30 = 190.
						if p.Subtotal != "200.0000" {
							t.Errorf("Subtotal: want 200.0000, got %s", p.Subtotal)
						}
						if p.TaxTotal != "20.0000" {
							t.Errorf("TaxTotal: want 20.0000, got %s", p.TaxTotal)
						}
						if p.DiscountTotal != "30.0000" {
							t.Errorf("DiscountTotal: want 30.0000, got %s", p.DiscountTotal)
						}
						if p.Total != "190.0000" {
							t.Errorf("Total: want 190.0000, got %s", p.Total)
						}
						if len(p.LineItems) != 1 || p.LineItems[0].TaxAmount != "20.0000" || p.LineItems[0].DiscountAmount != "30.0000" {
							t.Errorf("line aggregates not recomputed: %+v", p.LineItems[0])
						}
						return &entity.Order{ID: "o-3"}, nil
					})
				m.reader.EXPECT().
					ListLineItems(gomock.Any(), "o-3").
					Return(nil, nil)
				m.reader.EXPECT().
					ListPayments(gomock.Any(), "o-3").
					Return(nil, nil)
				m.reader.EXPECT().
					ListOrderAdjustments(gomock.Any(), "o-3").
					Return(nil, nil)
			},
			want: &entity.Order{ID: "o-3"},
		},
		"applies negative order adjustment exactly once": {
			in: func() input.CreateOrderInput {
				in := baseInput()
				in.StoreID = "store-explicit"
				in.ExternalID = ""
				in.Subtotal = "0"
				in.TaxTotal = "0"
				in.DiscountTotal = "0"
				in.Total = "0"
				in.LineItems = []input.CreateOrderLineItemInput{{
					ProductName: "Coffee",
					Quantity:    "2",
					UnitPrice:   "100",
					LineTotal:   "200",
					Taxes: []input.CreateOrderLineItemTaxInput{{
						Sequence: 1, NameSnapshot: "VAT", RateSnapshot: "0.1000",
						TaxableBase: "200", Amount: "20",
					}},
				}}
				in.Adjustments = []input.CreateOrderAdjustmentInput{{
					Sequence: 1, Kind: "discount", SourceType: "manual",
					NameSnapshot: "Order off", CalculationType: "fixed_amount",
					Amount: "-10", AppliesBeforeTax: true, ProrateStrategy: "pro_rata_taxable_base",
				}}
				return in
			}(),
			setup: func(m *orderMocks) {
				m.stubPassthroughWrite()
				m.allowStoreAccess()
				m.writer.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, p gateway.CreateOrderParams) (*entity.Order, error) {
						// sub = 200, tax = 20, order discount = -10.
						// total must be sub + tax + order_adj = 200 + 20 - 10 = 210.
						// discount_total reports the absolute reduction = 10.
						if p.Subtotal != "200.0000" {
							t.Errorf("Subtotal: want 200.0000, got %s", p.Subtotal)
						}
						if p.TaxTotal != "20.0000" {
							t.Errorf("TaxTotal: want 20.0000, got %s", p.TaxTotal)
						}
						if p.DiscountTotal != "10.0000" {
							t.Errorf("DiscountTotal: want 10.0000, got %s", p.DiscountTotal)
						}
						if p.Total != "210.0000" {
							t.Errorf("Total: want 210.0000 (no double-subtract), got %s", p.Total)
						}
						return &entity.Order{ID: "o-neg"}, nil
					})
				m.reader.EXPECT().ListLineItems(gomock.Any(), "o-neg").Return(nil, nil)
				m.reader.EXPECT().ListPayments(gomock.Any(), "o-neg").Return(nil, nil)
				m.reader.EXPECT().ListOrderAdjustments(gomock.Any(), "o-neg").Return(nil, nil)
			},
			want: &entity.Order{ID: "o-neg"},
		},
		"rejects mixed children across line items": {
			in: func() input.CreateOrderInput {
				in := baseInput()
				in.StoreID = "store-explicit"
				in.ExternalID = ""
				in.LineItems = []input.CreateOrderLineItemInput{
					{
						ProductName: "Latte", Quantity: "1", UnitPrice: "100", LineTotal: "110",
						Taxes: []input.CreateOrderLineItemTaxInput{{
							Sequence: 1, NameSnapshot: "VAT", RateSnapshot: "0.1000",
							TaxableBase: "100", Amount: "10",
						}},
					},
					{ProductName: "Bagel", Quantity: "1", UnitPrice: "50", LineTotal: "50"},
				}
				return in
			}(),
			setup:       func(_ *orderMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeBadRequest,
		},
		"keeps caller aggregates when no children provided": {
			in: func() input.CreateOrderInput {
				in := baseInput()
				in.StoreID = "store-explicit"
				in.ExternalID = ""
				in.Subtotal = "111"
				in.TaxTotal = "11"
				in.DiscountTotal = "5"
				in.Total = "117"
				return in
			}(),
			setup: func(m *orderMocks) {
				m.stubPassthroughWrite()
				m.allowStoreAccess()
				m.writer.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, p gateway.CreateOrderParams) (*entity.Order, error) {
						if p.Subtotal != "111" || p.TaxTotal != "11" || p.DiscountTotal != "5" || p.Total != "117" {
							t.Errorf("legacy aggregates were rewritten: %+v", p)
						}
						return &entity.Order{ID: "o-4"}, nil
					})
				m.reader.EXPECT().
					ListLineItems(gomock.Any(), "o-4").
					Return(nil, nil)
				m.reader.EXPECT().
					ListPayments(gomock.Any(), "o-4").
					Return(nil, nil)
				m.reader.EXPECT().
					ListOrderAdjustments(gomock.Any(), "o-4").
					Return(nil, nil)
			},
			want: &entity.Order{ID: "o-4"},
		},
		"rejects pos order when user is not assigned to the resolved store": {
			in: func() input.CreateOrderInput {
				in := baseInput()
				in.StoreID = "store-explicit"
				in.ExternalID = ""
				return in
			}(),
			setup: func(m *orderMocks) {
				m.stubPassthroughWrite()
				m.memberReader.EXPECT().
					HasStoreAccess(gomock.Any(), "user-1", "store-explicit").
					Return(false, nil)
				// writer.Create must not be called
			},
			wantErr:     true,
			wantErrCode: errors.CodeForbidden,
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
