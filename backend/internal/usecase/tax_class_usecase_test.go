package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	gatewaymock "github.com/genpick/genpos-mono/backend/internal/domain/gateway/mock"
	pkgerrors "github.com/genpick/genpos-mono/backend/pkg/errors"

	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
)

type taxClassMocks struct {
	ctrl     *gomock.Controller
	tenantDB *gatewaymock.MockTenantDB
	reader   *gatewaymock.MockTaxClassReader
	writer   *gatewaymock.MockTaxClassWriter
}

func newTaxClassMocks(t *testing.T) *taxClassMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	return &taxClassMocks{
		ctrl:     ctrl,
		tenantDB: gatewaymock.NewMockTenantDB(ctrl),
		reader:   gatewaymock.NewMockTaxClassReader(ctrl),
		writer:   gatewaymock.NewMockTaxClassWriter(ctrl),
	}
}

func (m *taxClassMocks) newUsecase() usecase.TaxClassUsecase {
	return usecase.NewTaxClassUsecase(m.tenantDB, m.reader, m.writer)
}

func (m *taxClassMocks) stubRead() {
	m.tenantDB.EXPECT().
		ReadWithTenant(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ string, fn func(context.Context) error) error {
			return fn(ctx)
		}).AnyTimes()
}

func (m *taxClassMocks) stubWrite() {
	m.tenantDB.EXPECT().
		WithTenant(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ string, fn func(context.Context) error) error {
			return fn(ctx)
		}).AnyTimes()
}

func Test_TaxClassUsecase_ListTaxClasses(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	cases := map[string]struct {
		orgID       string
		setup       func(*taxClassMocks)
		want        []*entity.TaxClass
		wantErrCode string
	}{
		"returns the org's tax classes with their rate links": {
			orgID: "org-1",
			setup: func(m *taxClassMocks) {
				m.stubRead()
				m.reader.EXPECT().List(gomock.Any()).Return([]*entity.TaxClass{
					{
						ID: "tc-1", OrgID: "org-1", Name: "Standard",
						IsDefault: true, SortOrder: 0,
						CreatedAt: now, UpdatedAt: now,
						Rates: []*entity.TaxClassRate{
							{ID: "tcr-1", TaxRateID: "tr-vat", Sequence: 0, IsCompound: false},
						},
					},
				}, nil)
			},
			want: []*entity.TaxClass{
				{
					ID: "tc-1", OrgID: "org-1", Name: "Standard",
					IsDefault: true, SortOrder: 0,
					CreatedAt: now, UpdatedAt: now,
					Rates: []*entity.TaxClassRate{
						{ID: "tcr-1", TaxRateID: "tr-vat", Sequence: 0, IsCompound: false},
					},
				},
			},
		},
		"empty org id is rejected": {
			orgID:       "",
			setup:       func(*taxClassMocks) {},
			wantErrCode: pkgerrors.CodeBadRequest,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := newTaxClassMocks(t)
			tc.setup(m)
			uc := m.newUsecase()
			got, err := uc.ListTaxClasses(context.Background(), tc.orgID)
			if tc.wantErrCode != "" {
				if err == nil {
					t.Fatalf("want error code %s, got nil", tc.wantErrCode)
				}
				if pkgerrors.GetCode(err) != tc.wantErrCode {
					t.Errorf("error code: want %s, got %s (err=%v)", tc.wantErrCode, pkgerrors.GetCode(err), err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_TaxClassUsecase_CreateTaxClass(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	created := &entity.TaxClass{ID: "tc-new", OrgID: "org-1", Name: "Standard", IsDefault: true, CreatedAt: now, UpdatedAt: now}

	cases := map[string]struct {
		in          input.CreateTaxClassInput
		setup       func(*taxClassMocks)
		want        *entity.TaxClass
		wantErrCode string
	}{
		"creates with rates and clears existing default first": {
			in: input.CreateTaxClassInput{
				OrgID: "org-1",
				Class: input.TaxClassInput{
					Name: "  Standard  ", Description: "", IsDefault: true,
					Rates: []input.TaxClassRateInput{
						{TaxRateID: "tr-vat", Sequence: 0, IsCompound: false},
					},
				},
			},
			setup: func(m *taxClassMocks) {
				m.stubWrite()
				gomock.InOrder(
					m.writer.EXPECT().ClearDefaults(gomock.Any()).Return(nil),
					m.writer.EXPECT().
						Create(gomock.Any(), gateway.CreateTaxClassParams{
							OrgID: "org-1", Name: "Standard", Description: "", IsDefault: true,
							Rates: []gateway.TaxClassRateParams{{TaxRateID: "tr-vat", Sequence: 0, IsCompound: false}},
						}).
						Return(created, nil),
				)
			},
			want: created,
		},
		"non-default skips ClearDefaults": {
			in: input.CreateTaxClassInput{
				OrgID: "org-1",
				Class: input.TaxClassInput{Name: "Reduced"},
			},
			setup: func(m *taxClassMocks) {
				m.stubWrite()
				m.writer.EXPECT().
					Create(gomock.Any(), gateway.CreateTaxClassParams{
						OrgID: "org-1", Name: "Reduced", Description: "", IsDefault: false,
						Rates: []gateway.TaxClassRateParams{},
					}).
					Return(&entity.TaxClass{ID: "tc-r", OrgID: "org-1", Name: "Reduced"}, nil)
			},
			want: &entity.TaxClass{ID: "tc-r", OrgID: "org-1", Name: "Reduced"},
		},
		"empty org id rejected": {
			in:          input.CreateTaxClassInput{Class: input.TaxClassInput{Name: "X"}},
			setup:       func(*taxClassMocks) {},
			wantErrCode: pkgerrors.CodeBadRequest,
		},
		"empty name rejected": {
			in:          input.CreateTaxClassInput{OrgID: "org-1", Class: input.TaxClassInput{Name: "  "}},
			setup:       func(*taxClassMocks) {},
			wantErrCode: pkgerrors.CodeBadRequest,
		},
		"duplicate tax_rate in same class rejected": {
			in: input.CreateTaxClassInput{
				OrgID: "org-1",
				Class: input.TaxClassInput{
					Name: "Bundle",
					Rates: []input.TaxClassRateInput{
						{TaxRateID: "tr-vat", Sequence: 0},
						{TaxRateID: "tr-vat", Sequence: 1},
					},
				},
			},
			setup:       func(*taxClassMocks) {},
			wantErrCode: pkgerrors.CodeBadRequest,
		},
		"empty tax_rate_id rejected": {
			in: input.CreateTaxClassInput{
				OrgID: "org-1",
				Class: input.TaxClassInput{
					Name:  "Bad",
					Rates: []input.TaxClassRateInput{{TaxRateID: "", Sequence: 0}},
				},
			},
			setup:       func(*taxClassMocks) {},
			wantErrCode: pkgerrors.CodeBadRequest,
		},
		"writer error propagates": {
			in: input.CreateTaxClassInput{OrgID: "org-1", Class: input.TaxClassInput{Name: "X"}},
			setup: func(m *taxClassMocks) {
				m.stubWrite()
				m.writer.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db down"))
			},
			wantErrCode: pkgerrors.CodeInternal,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := newTaxClassMocks(t)
			tc.setup(m)
			uc := m.newUsecase()
			got, err := uc.CreateTaxClass(context.Background(), tc.in)
			if tc.wantErrCode != "" {
				if err == nil {
					t.Fatalf("want error code %s, got nil", tc.wantErrCode)
				}
				if pkgerrors.GetCode(err) != tc.wantErrCode {
					t.Errorf("error code: want %s, got %s (err=%v)", tc.wantErrCode, pkgerrors.GetCode(err), err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_TaxClassUsecase_DeleteTaxClass(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		in          input.DeleteTaxClassInput
		setup       func(*taxClassMocks)
		wantErrCode string
	}{
		"deletes via writer": {
			in: input.DeleteTaxClassInput{ID: "tc-1", OrgID: "org-1"},
			setup: func(m *taxClassMocks) {
				m.stubWrite()
				m.writer.EXPECT().SoftDelete(gomock.Any(), "tc-1").Return(nil)
			},
		},
		"missing id rejected": {
			in:          input.DeleteTaxClassInput{OrgID: "org-1"},
			setup:       func(*taxClassMocks) {},
			wantErrCode: pkgerrors.CodeBadRequest,
		},
		"writer not-found surfaces as not-found": {
			in: input.DeleteTaxClassInput{ID: "tc-missing", OrgID: "org-1"},
			setup: func(m *taxClassMocks) {
				m.stubWrite()
				m.writer.EXPECT().SoftDelete(gomock.Any(), "tc-missing").
					Return(pkgerrors.NotFound("tax class not found"))
			},
			wantErrCode: pkgerrors.CodeNotFound,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := newTaxClassMocks(t)
			tc.setup(m)
			uc := m.newUsecase()
			err := uc.DeleteTaxClass(context.Background(), tc.in)
			if tc.wantErrCode != "" {
				if err == nil {
					t.Fatalf("want error code %s, got nil", tc.wantErrCode)
				}
				if pkgerrors.GetCode(err) != tc.wantErrCode {
					t.Errorf("error code: want %s, got %s", tc.wantErrCode, pkgerrors.GetCode(err))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
