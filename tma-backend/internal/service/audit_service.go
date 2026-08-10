package service

import (
	"context"
	"log/slog"

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
		if err := s.repo.AddLog(context.Background(), log); err != nil {
			slog.Error("audit log write failed",
				slog.String("action_type", actionType),
				slog.String("target_type", targetType),
				slog.String("error", err.Error()),
			)
		}
	}()
}

func (s *AuditService) GetLogs(ctx context.Context, f repository.AuditFilter) ([]domain.AdminActionLog, int, error) {
	return s.repo.GetLogs(ctx, f)
}
