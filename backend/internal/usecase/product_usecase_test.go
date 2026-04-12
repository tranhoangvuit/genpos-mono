package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	gatewaymock "github.com/genpick/genpos-mono/backend/internal/domain/gateway/mock"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

func Test_ProductUsecase_ListProducts(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	cases := map[string]struct {
		input   input.ListProductsInput
		setup   func(*gatewaymock.MockTenantDB, *gatewaymock.MockProductReader)
		want    []*entity.Product
		wantErr bool
		errCode string
	}{
		"returns products": {
			input: input.ListProductsInput{OrgID: "org-1", PageSize: 10},
			setup: func(tdb *gatewaymock.MockTenantDB, pr *gatewaymock.MockProductReader) {
				tdb.EXPECT().
					ReadWithTenant(gomock.Any(), "org-1", gomock.Any()).
					DoAndReturn(func(ctx context.Context, _ string, fn func(context.Context) error) error {
						return fn(ctx)
					})
				pr.EXPECT().
					ListProducts(gomock.Any(), gateway.ListProductsParams{Limit: 10, Offset: 0}).
					Return([]*entity.Product{
						{ID: "p1", OrgID: "org-1", Name: "Widget", SKU: "WGT-1", PriceCents: 1099, Active: true, CreatedAt: now, UpdatedAt: now},
					}, nil)
			},
			want: []*entity.Product{
				{ID: "p1", OrgID: "org-1", Name: "Widget", SKU: "WGT-1", PriceCents: 1099, Active: true, CreatedAt: now, UpdatedAt: now},
			},
		},
		"returns empty list": {
			input: input.ListProductsInput{OrgID: "org-empty", PageSize: 10},
			setup: func(tdb *gatewaymock.MockTenantDB, pr *gatewaymock.MockProductReader) {
				tdb.EXPECT().
					ReadWithTenant(gomock.Any(), "org-empty", gomock.Any()).
					DoAndReturn(func(ctx context.Context, _ string, fn func(context.Context) error) error {
						return fn(ctx)
					})
				pr.EXPECT().
					ListProducts(gomock.Any(), gateway.ListProductsParams{Limit: 10, Offset: 0}).
					Return([]*entity.Product{}, nil)
			},
			want: []*entity.Product{},
		},
		"defaults page_size when zero": {
			input: input.ListProductsInput{OrgID: "org-1", PageSize: 0},
			setup: func(tdb *gatewaymock.MockTenantDB, pr *gatewaymock.MockProductReader) {
				tdb.EXPECT().
					ReadWithTenant(gomock.Any(), "org-1", gomock.Any()).
					DoAndReturn(func(ctx context.Context, _ string, fn func(context.Context) error) error {
						return fn(ctx)
					})
				pr.EXPECT().
					ListProducts(gomock.Any(), gateway.ListProductsParams{Limit: 20, Offset: 0}).
					Return([]*entity.Product{}, nil)
			},
			want: []*entity.Product{},
		},
		"resets page_size over 100 to default": {
			input: input.ListProductsInput{OrgID: "org-1", PageSize: 200},
			setup: func(tdb *gatewaymock.MockTenantDB, pr *gatewaymock.MockProductReader) {
				tdb.EXPECT().
					ReadWithTenant(gomock.Any(), "org-1", gomock.Any()).
					DoAndReturn(func(ctx context.Context, _ string, fn func(context.Context) error) error {
						return fn(ctx)
					})
				pr.EXPECT().
					ListProducts(gomock.Any(), gateway.ListProductsParams{Limit: 20, Offset: 0}).
					Return([]*entity.Product{}, nil)
			},
			want: []*entity.Product{},
		},
		"empty org_id returns bad request": {
			input:   input.ListProductsInput{OrgID: ""},
			setup:   func(_ *gatewaymock.MockTenantDB, _ *gatewaymock.MockProductReader) {},
			wantErr: true,
			errCode: errors.CodeBadRequest,
		},
		"reader error is propagated": {
			input: input.ListProductsInput{OrgID: "org-1", PageSize: 10},
			setup: func(tdb *gatewaymock.MockTenantDB, pr *gatewaymock.MockProductReader) {
				tdb.EXPECT().
					ReadWithTenant(gomock.Any(), "org-1", gomock.Any()).
					DoAndReturn(func(ctx context.Context, _ string, fn func(context.Context) error) error {
						return fn(ctx)
					})
				pr.EXPECT().
					ListProducts(gomock.Any(), gomock.Any()).
					Return(nil, errors.Internal("database error"))
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockTenantDB := gatewaymock.NewMockTenantDB(ctrl)
			mockProductReader := gatewaymock.NewMockProductReader(ctrl)
			tc.setup(mockTenantDB, mockProductReader)

			uc := usecase.NewProductUsecase(mockTenantDB, mockProductReader)
			got, err := uc.ListProducts(context.Background(), tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errCode != "" {
					if code := errors.GetCode(err); code != tc.errCode {
						t.Errorf("error code: want %s, got %s", tc.errCode, code)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			opts := []cmp.Option{
				cmpopts.EquateApproxTime(time.Second),
			}
			if diff := cmp.Diff(tc.want, got, opts...); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
