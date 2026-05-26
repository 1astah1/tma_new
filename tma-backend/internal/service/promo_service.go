package service

import (
	"errors"
	"tma-backend/internal/repository"
	"github.com/google/uuid"
)

type PromoService struct {
	promoRepo *repository.PromoCodeRepo
}

func NewPromoService(promoRepo *repository.PromoCodeRepo) *PromoService {
	return &PromoService{promoRepo: promoRepo}
}

func (s *PromoService) ValidateAndApply(code string, subtotal float64) (*repository.PromoCode, float64, error) {
	promo, err := s.promoRepo.GetByCode(code)
	if err != nil {
		return nil, 0, errors.New("промокод не найден")
	}
	if !s.promoRepo.IsValid(promo) {
		return nil, 0, errors.New("промокод недействителен")
	}

	var discount float64
	if promo.DiscountFixed > 0 {
		discount = promo.DiscountFixed
	} else if promo.DiscountPercent > 0 {
		discount = subtotal * (promo.DiscountPercent / 100)
	}

	finalPrice := subtotal - discount
	if finalPrice < 0 {
		finalPrice = 0
	}

	return promo, finalPrice, nil
}

func (s *PromoService) ApplyPromo(code string) error {
	promo, err := s.promoRepo.GetByCode(code)
	if err != nil {
		return err
	}
	return s.promoRepo.IncrementUsage(promo.ID)
}

func (s *PromoService) List() ([]repository.PromoCode, error) {
	return s.promoRepo.List()
}

func (s *PromoService) Create(promo *repository.PromoCode) error {
	return s.promoRepo.Create(promo)
}

func (s *PromoService) Update(id string, updates map[string]interface{}) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return s.promoRepo.Update(parsed, updates)
}

func (s *PromoService) Delete(id string) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return s.promoRepo.Delete(parsed)
}

func (s *PromoService) GetByID(id uuid.UUID) (*repository.PromoCode, error) {
	return s.promoRepo.GetByID(id)
}

func (s *PromoService) UpdatePromo(promo *repository.PromoCode) error {
	return s.promoRepo.UpdatePromo(promo)
}
