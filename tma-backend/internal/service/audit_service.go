package service

import (
	"context"

	"github.com/google/uuid"
	"tma-backend/internal/domain"
	"tma-backend/internal/repository"
)

type AuditService struct {
	repo *repository.AdminRepo
}

func NewAuditService(repo *repository.AdminRepo) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(ctx context.Context, adminID uuid.UUID, actionType, targetType string, targetID uuid.UUID, details interface{}) {
	log := &domain.AdminActionLog{
		AdminID:    adminID,
		ActionType: actionType,
		TargetType: targetType,
		TargetID:   &targetID,
	}

	if d, ok := details.(map[string]interface{}); ok {
		log.Details = d
	}

	if ip, ok := ctx.Value("ip_address").(string); ok {
		log.IPAddress = &ip
	}

	go func() {
		s.repo.AddLog(context.Background(), log)
	}()
}

func (s *AuditService) GetLogs(ctx context.Context, f repository.AuditFilter) ([]domain.AdminActionLog, int, error) {
	return s.repo.GetLogs(ctx, f)
}
