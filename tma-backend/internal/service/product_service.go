package service

import (
	"context"

	"github.com/google/uuid"
	"tma-backend/internal/domain"
	"tma-backend/internal/repository"
)

type ProductService struct {
	repo *repository.ProductRepo
}

func NewProductService(repo *repository.ProductRepo) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List(ctx context.Context, f repository.ProductFilter) ([]domain.Product, int, error) {
	return s.repo.List(ctx, f)
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) Create(ctx context.Context, p *domain.Product) error {
	if p.Title == "" {
		return domain.ErrInvalidInput
	}
	if p.Price <= 0 {
		return domain.ErrInvalidInput
	}
	if len(p.DeliveryMethods) == 0 {
		return domain.ErrInvalidInput
	}
	return s.repo.Create(ctx, p)
}

func (s *ProductService) Update(ctx context.Context, p *domain.Product) error {
	return s.repo.Update(ctx, p)
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
