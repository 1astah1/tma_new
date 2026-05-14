package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"tma-backend/internal/domain"
)

type NotificationService struct {
	botToken  string
	httpClient *http.Client
}

type InlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

func NewNotificationService(botToken string) *NotificationService {
	return &NotificationService{
		botToken:   botToken,
		httpClient: &http.Client{},
	}
}

func (s *NotificationService) SendMessage(ctx context.Context, chatID int64, text string, buttons [][]InlineButton) error {
	if s.botToken == "" {
		log.Printf("[NOTIFICATION] To chat %d: %s\n", chatID, text)
		return nil
	}

	body := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
		"parse_mode": "HTML",
	}

	if len(buttons) > 0 {
		body["reply_markup"] = map[string]interface{}{
			"inline_keyboard": buttons,
		}
	}

	data, _ := json.Marshal(body)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.botToken)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (s *NotificationService) SendOrderStatusUpdate(ctx context.Context, order *domain.Order) {
	if order.User == nil {
		return
	}

	var text string
	buttons := [][]InlineButton{}

	switch order.Status {
	case domain.OrderStatusNew, domain.OrderStatusWaitingPayment:
		text = fmt.Sprintf("🆕 Заказ #%s создан!\n\nТовар: %s\nСумма: %.2f ₽\n\n📌 Для оплаты используйте реквизиты и загрузите чек в приложении.",
			order.ID.String()[:8], getProductTitle(order), order.PaymentAmount)
		buttons = append(buttons, []InlineButton{{Text: "🛒 Открыть заказ", CallbackData: "open_order:" + order.ID.String()}})

	case domain.OrderStatusPaymentVerification:
		text = fmt.Sprintf("📤 Чек получен!\n\nЗаказ #%s — ваш платеж передан на проверку.", order.ID.String()[:8])

	case domain.OrderStatusPaid:
		text = fmt.Sprintf("✅ Оплата подтверждена!\n\nЗаказ #%s — спасибо за покупку!", order.ID.String()[:8])

	case domain.OrderStatusKeyIssued:
		text = fmt.Sprintf("🔑 Ключ выдан!\n\nЗаказ #%s — ваш ключ готов! Проверьте в приложении.", order.ID.String()[:8])

	case domain.OrderStatusAwaitingCredentials:
		text = "📝 Требуются данные аккаунта\n\n🔐 Все данные шифруются (AES-256). Доступны только администратору."
		buttons = append(buttons, []InlineButton{{Text: "📝 Ввести данные", CallbackData: "submit_data:" + order.ID.String()}})

	case domain.OrderStatusAwaiting2FA:
		text = "🔐 Администратор готов войти в ваш аккаунт!\n\nПожалуйста, отправьте код подтверждения.\n⏳ Код нужно отправить в течение 5 минут."
		buttons = append(buttons, []InlineButton{{Text: "🔑 Отправить код", CallbackData: "submit_code:" + order.ID.String()}})

	case domain.OrderStatusActivating:
		text = "✅ Код получен! Администратор приступает к активации."

	case domain.OrderStatusActivated, domain.OrderStatusCompleted:
		text = fmt.Sprintf("✅ Активация завершена!\n\nТовар успешно активирован на вашем аккаунте!")
		buttons = append(buttons, []InlineButton{{Text: "🛒 В магазин", CallbackData: "open_shop"}})

	case domain.OrderStatusCancelled:
		text = fmt.Sprintf("❌ Заказ #%s отменен.\n\nПричина: %s",
			order.ID.String()[:8], derefString(order.CancelledReason))

	case domain.OrderStatusRefundRequested, domain.OrderStatusRefunded:
		text = "💳 Возврат средств обрабатывается."
	}

	if text != "" {
		go func() {
			if err := s.SendMessage(ctx, order.User.TelegramID, text, buttons); err != nil {
				log.Printf("Failed to send notification: %v", err)
			}
		}()
	}
}

func getProductTitle(order *domain.Order) string {
	if order.Product != nil {
		return order.Product.Title
	}
	return "Товар"
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
