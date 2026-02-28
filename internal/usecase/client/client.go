package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
	ucDomain "github.com/plagora/backend/internal/domain/usecase"
)

type clientUseCase struct {
	repo repository.ClientRepository
}

func New(repo repository.ClientRepository) ucDomain.ClientUseCase {
	return &clientUseCase{repo: repo}
}

func (c *clientUseCase) GetAll(ctx context.Context) ([]*entity.Client, error) {
	return c.repo.FindAll(ctx)
}

func (c *clientUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Client, error) {
	cl, err := c.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("client %s not found: %w", id, err)
	}
	return cl, nil
}

func (c *clientUseCase) Create(ctx context.Context, input ucDomain.CreateClientInput) (*entity.Client, error) {
	cl := &entity.Client{
		ID:        uuid.New(),
		Name:      input.Name,
		Phone:     input.Phone,
		Email:     input.Email,
		Notes:     input.Notes,
		CreatedAt: time.Now(),
	}
	if err := c.repo.Create(ctx, cl); err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}
	return cl, nil
}

func (c *clientUseCase) Update(ctx context.Context, id uuid.UUID, input ucDomain.UpdateClientInput) (*entity.Client, error) {
	cl, err := c.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("client %s not found: %w", id, err)
	}
	cl.Name = input.Name
	cl.Phone = input.Phone
	cl.Email = input.Email
	cl.Notes = input.Notes
	if err := c.repo.Update(ctx, cl); err != nil {
		return nil, fmt.Errorf("updating client: %w", err)
	}
	return cl, nil
}

func (c *clientUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return c.repo.Delete(ctx, id)
}
