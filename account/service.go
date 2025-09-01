package account

import (
	"context"
	"github.com/segmentio/ksuid"
)

type Service interface {
	PostAccount(ctx context.Context, name string) (*Account, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	GetAccounts(ctx context.Context, skip, take uint64) ([]Account, error)
}

type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type accountService struct {
	repository Repository
}

// GetAccount implements Service.
func (a *accountService) GetAccount(ctx context.Context, id string) (*Account, error) {
	return a.repository.GetAccountByID(ctx, id)
}

// GetAccounts implements Service.
func (a *accountService) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	return a.repository.ListAccounts(ctx, skip, take)
}

// PostAccount implements Service.
func (a *accountService) PostAccount(ctx context.Context, name string) (*Account, error) {
	acc := &Account{
		ID:   ksuid.New().String(),
		Name: name,
	}
	if err := a.repository.PutAccount(ctx, *acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func NewService(r Repository) Service {
	return &accountService{repository: r}
}
