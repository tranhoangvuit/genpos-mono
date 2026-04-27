package usecase_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	gatewaymock "github.com/genpick/genpos-mono/backend/internal/domain/gateway/mock"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type memberMocks struct {
	ctrl     *gomock.Controller
	tenantDB *gatewaymock.MockTenantDB
	reader   *gatewaymock.MockMemberReader
	writer   *gatewaymock.MockMemberWriter
}

func newMemberMocks(t *testing.T) *memberMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	return &memberMocks{
		ctrl:     ctrl,
		tenantDB: gatewaymock.NewMockTenantDB(ctrl),
		reader:   gatewaymock.NewMockMemberReader(ctrl),
		writer:   gatewaymock.NewMockMemberWriter(ctrl),
	}
}

func (m *memberMocks) newUsecase() usecase.MemberUsecase {
	return usecase.NewMemberUsecase(m.tenantDB, m.reader, m.writer)
}

func (m *memberMocks) stubPassthroughWrite() {
	m.tenantDB.EXPECT().
		WithTenant(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ string, fn func(context.Context) error) error {
			return fn(ctx)
		}).AnyTimes()
}

func Test_MemberUsecase_CreateMember(t *testing.T) {
	t.Parallel()

	baseInput := func() input.CreateMemberInput {
		return input.CreateMemberInput{
			OrgID:    "org-1",
			Name:     "Ana",
			Email:    "ana@example.com",
			Phone:    "+1",
			RoleID:   "role-1",
			Password: "supersecret",
		}
	}

	cases := map[string]struct {
		in              input.CreateMemberInput
		setup           func(*memberMocks)
		want            *entity.Member
		wantErr         bool
		wantErrCode     string
		wantStoreIDsArg []string
	}{
		"all_stores=true wipes store assignments": {
			in: func() input.CreateMemberInput {
				in := baseInput()
				in.AllStores = true
				in.StoreIDs = []string{"store-a", "store-b"} // should be ignored
				return in
			}(),
			setup: func(m *memberMocks) {
				m.stubPassthroughWrite()
				m.writer.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, p gateway.CreateMemberParams) (string, error) {
						if !p.AllStores {
							t.Errorf("AllStores: want true, got false")
						}
						return "u-1", nil
					})
				m.writer.EXPECT().
					ReplaceStores(gomock.Any(), "org-1", "u-1", []string(nil)).
					Return(nil)
				m.reader.EXPECT().
					GetByID(gomock.Any(), "u-1").
					Return(&entity.Member{ID: "u-1", AllStores: true}, nil)
			},
			want: &entity.Member{ID: "u-1", AllStores: true},
		},
		"all_stores=false persists provided list": {
			in: func() input.CreateMemberInput {
				in := baseInput()
				in.AllStores = false
				in.StoreIDs = []string{"store-a", "store-b"}
				return in
			}(),
			setup: func(m *memberMocks) {
				m.stubPassthroughWrite()
				m.writer.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return("u-2", nil)
				m.writer.EXPECT().
					ReplaceStores(gomock.Any(), "org-1", "u-2", []string{"store-a", "store-b"}).
					Return(nil)
				m.reader.EXPECT().
					GetByID(gomock.Any(), "u-2").
					Return(&entity.Member{ID: "u-2", StoreIDs: []string{"store-a", "store-b"}}, nil)
			},
			want: &entity.Member{ID: "u-2", StoreIDs: []string{"store-a", "store-b"}},
		},
		"rejects short password": {
			in: func() input.CreateMemberInput {
				in := baseInput()
				in.Password = "short"
				return in
			}(),
			setup:       func(_ *memberMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeBadRequest,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := newMemberMocks(t)
			tc.setup(m)

			got, err := m.newUsecase().CreateMember(context.Background(), tc.in)

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
				t.Errorf("member mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_MemberUsecase_UpdateMember(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		in      input.UpdateMemberInput
		setup   func(*memberMocks)
		wantErr bool
	}{
		"all_stores=true wipes store assignments on update": {
			in: input.UpdateMemberInput{
				ID:        "u-1",
				OrgID:     "org-1",
				Name:      "Ana",
				RoleID:    "role-1",
				Status:    "active",
				AllStores: true,
				StoreIDs:  []string{"store-a"}, // ignored
			},
			setup: func(m *memberMocks) {
				m.stubPassthroughWrite()
				m.writer.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, p gateway.UpdateMemberParams) error {
						if !p.AllStores {
							t.Errorf("AllStores: want true, got false")
						}
						return nil
					})
				m.writer.EXPECT().
					ReplaceStores(gomock.Any(), "org-1", "u-1", []string(nil)).
					Return(nil)
				m.reader.EXPECT().
					GetByID(gomock.Any(), "u-1").
					Return(&entity.Member{ID: "u-1", AllStores: true}, nil)
			},
		},
		"all_stores=false replaces with explicit list": {
			in: input.UpdateMemberInput{
				ID:        "u-1",
				OrgID:     "org-1",
				Name:      "Ana",
				RoleID:    "role-1",
				Status:    "active",
				AllStores: false,
				StoreIDs:  []string{"store-x"},
			},
			setup: func(m *memberMocks) {
				m.stubPassthroughWrite()
				m.writer.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
				m.writer.EXPECT().
					ReplaceStores(gomock.Any(), "org-1", "u-1", []string{"store-x"}).
					Return(nil)
				m.reader.EXPECT().
					GetByID(gomock.Any(), "u-1").
					Return(&entity.Member{ID: "u-1", StoreIDs: []string{"store-x"}}, nil)
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := newMemberMocks(t)
			tc.setup(m)

			_, err := m.newUsecase().UpdateMember(context.Background(), tc.in)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
