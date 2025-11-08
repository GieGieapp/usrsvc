package usecase

import (
	"context"

	"usrsvc/internal/domain"
)

type userUC struct{ repo domain.UserRepository }

func NewUserUC(r domain.UserRepository) domain.UserUsecase { return &userUC{repo: r} }

func (u *userUC) List(ctx context.Context, search string, page, size int) ([]domain.Customer, int32, error) {
	if size <= 0 {
		size = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * size
	return u.repo.ListCustomers(ctx, search, size, offset)
}

func (u *userUC) Get(ctx context.Context, id int32) (*domain.Customer, error) {
	return u.repo.GetCustomer(ctx, id)
}

func (u *userUC) Create(ctx context.Context, c domain.Customer) (int32, error) {
	return u.repo.CreateCustomer(ctx, c)
}

func (u *userUC) Update(ctx context.Context, id int32, c domain.Customer) error {
	return u.repo.UpdateCustomer(ctx, id, c)
}

func (u *userUC) Delete(ctx context.Context, id int32) error {
	return u.repo.DeleteCustomer(ctx, id)
}

func (u *userUC) ListNationality(ctx context.Context) ([]domain.Nationality, error) {
	return u.repo.ListNationalities(ctx)
}
