package service

import (
	"strings"

	"github.com/google/uuid"
	"tma-backend/internal/domain"
)

var codMW4ProductID = uuid.MustParse("63c124f6-af26-4cac-b5da-3be19905398a")

const codMW4DisplayTitle = "Call of Duty®: Modern Warfare® 4 (PS/XBOX/PC)"

func applyProductDisplayOverrides(p *domain.Product) {
	if p == nil {
		return
	}
	if p.ID == codMW4ProductID {
		p.Title = codMW4DisplayTitle
		return
	}
	if strings.Contains(p.Title, "Modern Warfare") && strings.Contains(p.Title, "(Windows)") {
		p.Title = strings.ReplaceAll(p.Title, "(Windows)", "(PS/XBOX/PC)")
	}
}

func applyProductDisplayOverridesList(products []domain.Product) []domain.Product {
	for i := range products {
		applyProductDisplayOverrides(&products[i])
	}
	return products
}
