//go:generate mockery --name=UserUsecase --output=../mocks --case=underscore
package domain

import "context"

type UserUsecase interface {
	List(ctx context.Context, search string, page, size int) ([]Customer, int32, error)
	Get(ctx context.Context, id int32) (*Customer, error)
	Create(ctx context.Context, c Customer) (int32, error)
	Update(ctx context.Context, id int32, c Customer) error
	Delete(ctx context.Context, id int32) error
	ListNationality(ctx context.Context) ([]Nationality, error)
}
