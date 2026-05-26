package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"tma-backend/internal/domain"
)

type NotificationService struct {
	botToken   string
	botUsername string
	httpClient *http.Client
}

func NewNotificationService(botToken string) *NotificationService {
	return &NotificationService{
		botToken:   botToken,
		httpClient: &http.Client{},
	}
}

func (s *NotificationService) SetBotUsername(username string) {
	s.botUsername = username
}

type InlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

func (s *NotificationService) SendMessage(ctx context.Context, chatID int64, text string, buttons [][]InlineButton) error {
	if s.botToken == "" {
		slog.Info("notification (dev mode)", slog.Int64("chat_id", chatID), slog.String("text", text))
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
			order.ID.String()[:8], getProductTitle(order), derefFloat64(order.PaymentAmount))
		buttons = append(buttons, []InlineButton{{Text: "🛒 Открыть заказ", CallbackData: "open_order:" + order.ID.String()}})

	case domain.OrderStatusPaymentVerification:
		text = fmt.Sprintf("📤 Чек получен!\n\nЗаказ #%s — ваш платеж передан на проверку.", order.ID.String()[:8])

	case domain.OrderStatusPaid:
		if order.DeliveryMethod == domain.DeliveryMethodActivation {
			text = fmt.Sprintf("✅ Оплата подтверждена!\n\nЗаказ #%s — модератор скоро начнёт активацию.", order.ID.String()[:8])
		} else {
			text = fmt.Sprintf("✅ Оплата подтверждена!\n\nЗаказ #%s — спасибо за покупку!", order.ID.String()[:8])
		}

	case domain.OrderStatusWaitingActivation:
		text = fmt.Sprintf("📋 Заказ #%s в очереди на активацию.\n\nМодератор скоро начнёт работу.", order.ID.String()[:8])

	case domain.OrderStatusKeyIssued:
		text = fmt.Sprintf("🔑 Ключ выдан!\n\nЗаказ #%s — ваш ключ готов! Проверьте в приложении.", order.ID.String()[:8])

	case domain.OrderStatusAwaitingCredentials:
		text = "📝 Требуются данные аккаунта\n\n🔐 Введите логин и пароль в приложении. Все данные шифруются (AES-256)."

	case domain.OrderStatusCredentialsReceived:
		text = fmt.Sprintf("✅ Данные получены!\n\nЗаказ #%s — модератор проверяет данные и готовится к активации. Ожидайте уведомления.", order.ID.String()[:8])

	case domain.OrderStatusAwaiting2FA:
		text = "🔐 Модератор готов войти в ваш аккаунт!\n\nОтправьте код подтверждения из письма или приложения.\n⏳ У вас есть 10 минут."

	case domain.OrderStatusActivating:
		text = "⚙️ Активация в процессе!\n\nМодератор активирует товар. Это может занять несколько минут. Не закрывайте приложение."

	case domain.OrderStatusActivated:
		text = "✅ Активация завершена!\n\nТовар успешно активирован на вашем аккаунте!"

	case domain.OrderStatusCompleted:
		text = fmt.Sprintf("🎉 Заказ #%s завершён!\n\nТовар успешно активирован! Спасибо за покупку!", order.ID.String()[:8])

	case domain.OrderStatusCancelled:
		text = fmt.Sprintf("❌ Заказ #%s отменен.\n\nПричина: %s",
			order.ID.String()[:8], derefString(order.CancelledReason))

	case domain.OrderStatusCredentialsInvalid:
		text = fmt.Sprintf("❌ Данные аккаунта неверны.\n\nЗаказ #%s\nПричина: %s\n\nПожалуйста, проверьте данные и отправьте их снова в приложении.",
			order.ID.String()[:8], derefString(order.CancelledReason))

	case domain.OrderStatusInvalid2FA:
		text = fmt.Sprintf("❌ Код подтверждения неверен.\n\nЗаказ #%s\nПричина: %s\n\nОтправьте правильный код в приложении.",
			order.ID.String()[:8], derefString(order.CancelledReason))

	case domain.OrderStatusRefundRequested, domain.OrderStatusRefunded:
		text = "💳 Возврат средств обрабатывается."
	}

	if text != "" {
		go func() {
			if err := s.SendMessage(ctx, order.User.TelegramID, text, buttons); err != nil {
				slog.Error("Failed to send notification", slog.String("error", err.Error()))
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

func derefFloat64(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
