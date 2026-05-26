package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tma-backend/internal/domain"
	"tma-backend/internal/service/mocks"
)

func newTestProductService() (*ProductService, *mocks.MockProductStore) {
	repo := &mocks.MockProductStore{}
	svc := NewProductService(repo)
	return svc, repo
}

func TestCreateProduct_Success(t *testing.T) {
	svc, repo := newTestProductService()

	product := &domain.Product{
		Title:           "Test Product",
		Price:           99.99,
		DeliveryMethods: []string{"key", "activation"},
		Status:          domain.ProductStatusActive,
	}

	ctx := context.Background()
	repo.On("Create", ctx, product).Return(nil)

	err := svc.Create(ctx, product)
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestCreateProduct_MissingTitle(t *testing.T) {
	svc, _ := newTestProductService()

	product := &domain.Product{
		Title:           "",
		Price:           99.99,
		DeliveryMethods: []string{"key"},
	}

	ctx := context.Background()
	err := svc.Create(ctx, product)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidInput, err)
}

func TestCreateProduct_NegativePrice(t *testing.T) {
	svc, _ := newTestProductService()

	product := &domain.Product{
		Title:           "Test Product",
		Price:           -10,
		DeliveryMethods: []string{"key"},
	}

	ctx := context.Background()
	err := svc.Create(ctx, product)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidInput, err)
}

func TestCreateProduct_ZeroPrice(t *testing.T) {
	svc, _ := newTestProductService()

	product := &domain.Product{
		Title:           "Test Product",
		Price:           0,
		DeliveryMethods: []string{"key"},
	}

	ctx := context.Background()
	err := svc.Create(ctx, product)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidInput, err)
}

func TestCreateProduct_NoDeliveryMethods(t *testing.T) {
	svc, _ := newTestProductService()

	product := &domain.Product{
		Title:           "Test Product",
		Price:           99.99,
		DeliveryMethods: []string{},
	}

	ctx := context.Background()
	err := svc.Create(ctx, product)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidInput, err)
}

func TestCreateProduct_InvalidDiscount(t *testing.T) {
	svc, _ := newTestProductService()

	product := &domain.Product{
		Title:           "Test Product",
		Price:           99.99,
		DeliveryMethods: []string{"key"},
		DiscountPercent: 150,
	}

	ctx := context.Background()
	err := svc.Create(ctx, product)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidInput, err)
}

func TestCreateProduct_NegativeDiscount(t *testing.T) {
	svc, _ := newTestProductService()

	product := &domain.Product{
		Title:           "Test Product",
		Price:           99.99,
		DeliveryMethods: []string{"key"},
		DiscountPercent: -10,
	}

	ctx := context.Background()
	err := svc.Create(ctx, product)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidInput, err)
}

func TestUpdateProduct(t *testing.T) {
	svc, repo := newTestProductService()

	product := &domain.Product{
		ID:    uuid.New(),
		Title: "Updated Product",
		Price: 149.99,
	}

	ctx := context.Background()
	repo.On("Update", ctx, product).Return(nil)

	err := svc.Update(ctx, product)
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestDeleteProduct(t *testing.T) {
	svc, repo := newTestProductService()

	id := uuid.New()
	ctx := context.Background()
	repo.On("Delete", ctx, id).Return(nil)

	err := svc.Delete(ctx, id)
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestGetByID(t *testing.T) {
	svc, repo := newTestProductService()

	id := uuid.New()
	expected := &domain.Product{
		ID:    id,
		Title: "Test Product",
		Price: 99.99,
	}

	ctx := context.Background()
	repo.On("GetByID", ctx, id).Return(expected, nil)

	result, err := svc.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Title, result.Title)
	assert.Equal(t, expected.Price, result.Price)

	repo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	svc, repo := newTestProductService()

	id := uuid.New()
	ctx := context.Background()
	repo.On("GetByID", ctx, id).Return((*domain.Product)(nil), domain.ErrNotFound)

	_, err := svc.GetByID(ctx, id)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotFound, err)

	repo.AssertExpectations(t)
}
