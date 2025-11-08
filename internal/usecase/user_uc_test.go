package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"usrsvc/internal/domain"
	"usrsvc/internal/mocks"
)

func TestNewUserUC(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	uc := NewUserUC(repo)

	require.NotNil(t, uc)
	// karena di package yang sama, kita bisa assert ke struct konkret
	u, ok := uc.(*userUC)
	require.True(t, ok)
	assert.Equal(t, repo, u.repo)
}

func Test_userUC_Create(t *testing.T) {
	ctx := context.Background()
	c := domain.Customer{
		NationalityID: 1,
		Name:          "ALFA",
		Dob:           time.Date(1992, 5, 10, 0, 0, 0, 0, time.UTC),
		PhoneNum:      "0811",
		Email:         "alfa@example.com",
		Family:        []domain.FamilyMember{{Relation: "Spouse", Name: "BETA"}},
	}

	t.Run("ok", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		repo.
			On("CreateCustomer", ctx, c).
			Return(int32(123), nil).
			Once()

		uc := NewUserUC(repo)
		id, err := uc.Create(ctx, c)
		require.NoError(t, err)
		assert.Equal(t, int32(123), id)
	})

	t.Run("conflict", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		repo.
			On("CreateCustomer", ctx, c).
			Return(int32(0), domain.ErrConflict).
			Once()

		uc := NewUserUC(repo)
		id, err := uc.Create(ctx, c)
		require.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrConflict))
		assert.Equal(t, int32(0), id)
	})
}

func Test_userUC_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		repo.
			On("GetCustomer", ctx, int32(36)).
			Return(&domain.Customer{ID: 36, Name: "ALFA"}, nil).
			Once()

		uc := NewUserUC(repo)
		got, err := uc.Get(ctx, 36)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int32(36), got.ID)
	})

	t.Run("not_found", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		repo.
			On("GetCustomer", ctx, int32(999)).
			Return((*domain.Customer)(nil), domain.ErrNotFound).
			Once()

		uc := NewUserUC(repo)
		got, err := uc.Get(ctx, 999)
		require.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrNotFound))
		assert.Nil(t, got)
	})
}

func Test_userUC_List(t *testing.T) {
	ctx := context.Background()

	t.Run("ok_with_pagination", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		// page=2,size=10 -> limit=10, offset=10
		repo.
			On("ListCustomers", ctx, "AL", 10, 10).
			Return([]domain.Customer{
				{ID: 36, Name: "ALFA"},
				{ID: 37, Name: "BRAVO"},
			}, int32(42), nil).
			Once()

		uc := NewUserUC(repo)
		rows, total, err := uc.List(ctx, "AL", 2, 10)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Equal(t, int32(42), total)
	})

	t.Run("normalize_when_page_size_invalid", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		// input page<=0,size<=0 → size=10,page=1 → limit=10, offset=0
		repo.
			On("ListCustomers", ctx, "", 10, 0).
			Return([]domain.Customer{}, int32(0), nil).
			Once()

		uc := NewUserUC(repo)
		rows, total, err := uc.List(ctx, "", 0, 0)
		require.NoError(t, err)
		assert.Empty(t, rows)
		assert.Equal(t, int32(0), total)
	})

	t.Run("repo_error", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		repo.
			On("ListCustomers", ctx, "X", 5, 0).
			Return(([]domain.Customer)(nil), int32(0), errors.New("db down")).
			Once()

		uc := NewUserUC(repo)
		rows, total, err := uc.List(ctx, "X", 1, 5)
		require.Error(t, err)
		assert.Nil(t, rows)
		assert.Equal(t, int32(0), total)
	})
}

func Test_userUC_ListNationality(t *testing.T) {
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		repo.
			On("ListNationalities", ctx).
			Return([]domain.Nationality{
				{ID: 1, Name: "Indonesia", Code: strPtr("ID")},
				{ID: 2, Name: "Malaysia", Code: strPtr("MY")},
			}, nil).
			Once()

		uc := NewUserUC(repo)
		out, err := uc.ListNationality(ctx)
		require.NoError(t, err)
		require.Len(t, out, 2)
		assert.Equal(t, "Indonesia", out[0].Name)
	})

	t.Run("repo_error", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		repo.
			On("ListNationalities", ctx).
			Return(([]domain.Nationality)(nil), errors.New("db down")).
			Once()

		uc := NewUserUC(repo)
		out, err := uc.ListNationality(ctx)
		require.Error(t, err)
		assert.Nil(t, out)
	})
}

func Test_userUC_Update(t *testing.T) {
	ctx := context.Background()
	in := domain.Customer{Name: "NEW", Email: "new@example.com"}

	t.Run("ok", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		repo.
			On("UpdateCustomer", ctx, int32(36), in).
			Return(nil).
			Once()

		uc := NewUserUC(repo)
		require.NoError(t, uc.Update(ctx, 36, in))
	})

	t.Run("not_found", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		repo.
			On("UpdateCustomer", ctx, int32(36), in).
			Return(domain.ErrNotFound).
			Once()

		uc := NewUserUC(repo)
		err := uc.Update(ctx, 36, in)
		require.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrNotFound))
	})
}

func Test_userUC_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		repo.
			On("DeleteCustomer", ctx, int32(36)).
			Return(nil).
			Once()

		uc := NewUserUC(repo)
		require.NoError(t, uc.Delete(ctx, 36))
	})

	t.Run("not_found", func(t *testing.T) {
		repo := mocks.NewUserRepository(t)
		repo.
			On("DeleteCustomer", ctx, int32(999)).
			Return(domain.ErrNotFound).
			Once()

		uc := NewUserUC(repo)
		err := uc.Delete(ctx, 999)
		require.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrNotFound))
	})
}

func strPtr(s string) *string { return &s }
