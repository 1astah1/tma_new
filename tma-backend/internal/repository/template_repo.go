package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tma-backend/internal/domain"
)

type TemplateRepo struct {
	db *sqlx.DB
}

func NewTemplateRepo(db *sqlx.DB) *TemplateRepo {
	return &TemplateRepo{db: db}
}

func (r *TemplateRepo) List(ctx context.Context, category string) ([]domain.ChatTemplate, error) {
	var templates []domain.ChatTemplate
	query := "SELECT * FROM chat_templates WHERE is_active = TRUE"
	args := []interface{}{}
	if category != "" {
		query += " AND category = $1"
		args = append(args, category)
	}
	query += " ORDER BY category, title"
	err := r.db.SelectContext(ctx, &templates, query, args...)
	if err != nil {
		return nil, err
	}
	if templates == nil {
		templates = []domain.ChatTemplate{}
	}
	return templates, nil
}

func (r *TemplateRepo) ListAll(ctx context.Context) ([]domain.ChatTemplate, error) {
	var templates []domain.ChatTemplate
	err := r.db.SelectContext(ctx, &templates, "SELECT * FROM chat_templates ORDER BY category, title")
	if err != nil {
		return nil, err
	}
	if templates == nil {
		templates = []domain.ChatTemplate{}
	}
	return templates, nil
}

func (r *TemplateRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ChatTemplate, error) {
	var t domain.ChatTemplate
	err := r.db.GetContext(ctx, &t, "SELECT * FROM chat_templates WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepo) Create(ctx context.Context, t *domain.ChatTemplate) error {
	return r.db.GetContext(ctx, t,
		`INSERT INTO chat_templates (title, message, category, is_active)
		 VALUES ($1, $2, $3, $4) RETURNING *`,
		t.Title, t.Message, t.Category, t.IsActive)
}

func (r *TemplateRepo) Update(ctx context.Context, t *domain.ChatTemplate) error {
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE chat_templates SET title=:title, message=:message, category=:category,
		 is_active=:is_active, updated_at=NOW() WHERE id=:id`, t)
	return err
}

func (r *TemplateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM chat_templates WHERE id = $1", id)
	return err
}

func (r *TemplateRepo) ListFAQ(ctx context.Context) ([]domain.FAQItem, error) {
	var items []domain.FAQItem
	err := r.db.SelectContext(ctx, &items,
		"SELECT * FROM faq_items WHERE is_active = TRUE ORDER BY sort_order, created_at")
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.FAQItem{}
	}
	return items, nil
}

func (r *TemplateRepo) ListFAQAll(ctx context.Context) ([]domain.FAQItem, error) {
	var items []domain.FAQItem
	err := r.db.SelectContext(ctx, &items, "SELECT * FROM faq_items ORDER BY sort_order, created_at")
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.FAQItem{}
	}
	return items, nil
}

func (r *TemplateRepo) GetFAQByID(ctx context.Context, id uuid.UUID) (*domain.FAQItem, error) {
	var f domain.FAQItem
	err := r.db.GetContext(ctx, &f, "SELECT * FROM faq_items WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *TemplateRepo) CreateFAQ(ctx context.Context, f *domain.FAQItem) error {
	return r.db.GetContext(ctx, f,
		`INSERT INTO faq_items (question, answer, is_active, sort_order)
		 VALUES ($1, $2, $3, $4) RETURNING *`,
		f.Question, f.Answer, f.IsActive, f.SortOrder)
}

func (r *TemplateRepo) UpdateFAQ(ctx context.Context, f *domain.FAQItem) error {
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE faq_items SET question=:question, answer=:answer, is_active=:is_active,
		 sort_order=:sort_order, updated_at=NOW() WHERE id=:id`, f)
	return err
}

func (r *TemplateRepo) DeleteFAQ(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM faq_items WHERE id = $1", id)
	return err
}
