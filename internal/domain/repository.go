//go:generate mockery --name=UserRepository --output=../mocks --case=underscore
package domain

import "context"

type UserRepository interface {
	ListCustomers(ctx context.Context, search string, limit, offset int) ([]Customer, int32, error)
	GetCustomer(ctx context.Context, id int32) (*Customer, error)
	CreateCustomer(ctx context.Context, c Customer) (int32, error)
	UpdateCustomer(ctx context.Context, id int32, c Customer) error
	DeleteCustomer(ctx context.Context, id int32) error
	ListNationalities(ctx context.Context) ([]Nationality, error)
}
