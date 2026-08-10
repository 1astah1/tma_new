package service

import (
	"context"

	"github.com/google/uuid"
	"tma-backend/internal/domain"
	"tma-backend/internal/repository"
)

type ProductService struct {
	repo   ProductStore
	titles *EnglishTitleResolver
}

func NewProductService(repo ProductStore, titles *EnglishTitleResolver) *ProductService {
	return &ProductService{repo: repo, titles: titles}
}

func (s *ProductService) enrichProduct(ctx context.Context, p *domain.Product) {
	if p == nil {
		return
	}
	applyProductDisplayOverrides(p)
	if s.titles != nil {
		p.Title = s.titles.Resolve(ctx, p.ID, p.Title)
	}
}

func (s *ProductService) enrichProducts(ctx context.Context, products []domain.Product) []domain.Product {
	for i := range products {
		s.enrichProduct(ctx, &products[i])
	}
	return products
}

func (s *ProductService) List(ctx context.Context, f repository.ProductFilter) ([]domain.Product, int, error) {
	products, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	products = s.enrichProducts(ctx, products)
	EnrichListProductPrices(ctx, s.repo, products)
	return products, total, nil
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.enrichProduct(ctx, product)
	return product, nil
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
	if p.DiscountPercent < 0 || p.DiscountPercent > 100 {
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

func (s *ProductService) ActivateAllGames(ctx context.Context) (int64, error) {
	return s.repo.ActivateAllGames(ctx)
}

func (s *ProductService) SyncMetadataFromImports(ctx context.Context, minPrice float64) (int64, error) {
	return s.repo.SyncMetadataFromImports(ctx, minPrice)
}
