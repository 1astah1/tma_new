package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Bot struct {
	token      string
	httpClient *http.Client
}

func NewBot(token string) *Bot {
	return &Bot{
		token:      token,
		httpClient: &http.Client{},
	}
}

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type CallbackQuery struct {
	ID      string  `json:"id"`
	From    User    `json:"from"`
	Message *Message `json:"message,omitempty"`
	Data    string  `json:"data"`
}

type User struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	Username     string `json:"username"`
}

func (b *Bot) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var update Update
	if err := json.Unmarshal(body, &update); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	switch {
	case update.Message != nil:
		b.handleMessage(update.Message)
	case update.CallbackQuery != nil:
		b.handleCallback(update.CallbackQuery)
	}

	w.WriteHeader(http.StatusOK)
}

func (b *Bot) RegisterRoutes(r chi.Router) {
	r.Post("/webhook/"+b.token, b.WebhookHandler)
}

func (b *Bot) handleMessage(msg *Message) {
	text := strings.TrimSpace(msg.Text)

	switch text {
	case "/start":
		b.sendWelcome(msg.Chat.ID)
	case "/shop":
		b.sendShopButton(msg.Chat.ID)
	default:
		b.sendMessage(msg.Chat.ID, "Используйте команды: /start, /shop")
	}
}

func (b *Bot) handleCallback(cb *CallbackQuery) {
	data := cb.Data
	chatID := cb.Message.Chat.ID

	switch {
	case strings.HasPrefix(data, "open_order:"):
		orderID := strings.TrimPrefix(data, "open_order:")
		b.sendMessage(chatID, fmt.Sprintf("Открываем заказ #%s...", orderID[:8]))
	case data == "open_shop":
		b.sendShopButton(chatID)
	case data == "open_orders":
		b.sendMessage(chatID, "Откройте приложение для просмотра заказов: [ссылка]")
	default:
		b.sendMessage(chatID, "Команда не распознана")
	}
}

func (b *Bot) sendWelcome(chatID int64) {
	text := `🎮 Добро пожаловать в GameStore!

Здесь вы можете купить игры, валюту и подписки
для PlayStation и Xbox по выгодным ценам!

🔹 Мгновенная выдача ключей
🔹 Активация на ваш аккаунт
🔹 Поддержка 24/7`

	buttons := [][]map[string]interface{}{
		{
			{"text": "🛒 Открыть магазин", "web_app": map[string]string{"url": "http://localhost:5173"}},
		},
		{
			{"text": "📋 Мои заказы", "callback_data": "open_orders"},
		},
	}

	b.sendMessageWithButtons(chatID, text, buttons)
}

func (b *Bot) sendShopButton(chatID int64) {
	text := "🛒 Нажмите кнопку ниже, чтобы открыть магазин:"
	buttons := [][]map[string]interface{}{
		{
			{"text": "🛒 Открыть магазин", "web_app": map[string]string{"url": "http://localhost:5173"}},
		},
	}
	b.sendMessageWithButtons(chatID, text, buttons)
}

func (b *Bot) sendMessage(chatID int64, text string) {
	b.sendMessageWithButtons(chatID, text, nil)
}

func (b *Bot) sendMessageWithButtons(chatID int64, text string, buttons [][]map[string]interface{}) {
	if b.token == "" {
		log.Printf("[BOT] Chat %d: %s\n", chatID, text)
		return
	}

	body := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	if len(buttons) > 0 {
		body["reply_markup"] = map[string]interface{}{
			"inline_keyboard": buttons,
		}
	}

	data, _ := json.Marshal(body)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)

	resp, err := b.httpClient.Post(url, "application/json", strings.NewReader(string(data)))
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return
	}
	defer resp.Body.Close()
}
