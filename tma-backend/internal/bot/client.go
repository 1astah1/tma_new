package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"tma-backend/internal/domain"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type Bot struct {
	token      string
	httpClient *http.Client
	db         *sqlx.DB
	adminRepo  *repository.AdminRepo
	orderRepo  *repository.OrderRepo
	userRepo   *repository.UserRepo
	notifSvc   *service.NotificationService
	orderSvc   *service.OrderService

	mu       sync.RWMutex
	orders   map[int64]string
	admins   map[int64]string
	tmaURL   string
	apiURL   string
	botName  string
}

func NewBot(token string, db *sqlx.DB, adminRepo *repository.AdminRepo, orderRepo *repository.OrderRepo, userRepo *repository.UserRepo, notifSvc *service.NotificationService, orderSvc *service.OrderService, tmaURL, apiURL string) *Bot {
	return &Bot{
		token:      token,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		db:         db,
		adminRepo:  adminRepo,
		orderRepo:  orderRepo,
		userRepo:   userRepo,
		notifSvc:   notifSvc,
		orderSvc:   orderSvc,
		orders:     make(map[int64]string),
		admins:     make(map[int64]string),
		tmaURL:     tmaURL,
		apiURL:     apiURL,
	}
}

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

type Message struct {
	MessageID int  `json:"message_id"`
	Chat      Chat `json:"chat"`
	Text      string `json:"text"`
	From      User  `json:"from"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message,omitempty"`
	Data    string   `json:"data"`
}

type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

var platformEmoji = map[string]string{
	"ps4": "🎮 PS4",
	"ps5": "🎮 PS5",
	"xbox": "🟢 Xbox",
}

var typeEmoji = map[string]string{
	"game":         "🎮",
	"currency":     "🪙",
	"subscription": "📦",
}

func (b *Bot) fullImageURL(imgPath string) string {
	if imgPath == "" {
		return ""
	}
	if strings.HasPrefix(imgPath, "https://") {
		return imgPath
	}
	if strings.HasPrefix(imgPath, "/") {
		return b.apiURL + imgPath
	}
	return b.apiURL + "/" + imgPath
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

func (b *Bot) StartPolling(ctx context.Context) {
	if b.token == "" {
		slog.Info("bot polling disabled: no token")
		return
	}

	go func() {
		var offset int
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		slog.Info("bot polling started")
		for {
			select {
			case <-ctx.Done():
				slog.Info("bot polling stopped")
				return
			case <-ticker.C:
				updates, err := b.getUpdates(offset)
				if err != nil {
					slog.Error("bot polling error", slog.String("error", err.Error()))
					continue
				}
				for _, upd := range updates {
					offset = upd.UpdateID + 1
					switch {
					case upd.Message != nil:
						b.handleMessage(upd.Message)
					case upd.CallbackQuery != nil:
						b.handleCallback(upd.CallbackQuery)
					}
				}
			}
		}
	}()
}

func (b *Bot) getUpdates(offset int) ([]Update, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=5", b.token, offset)
	resp, err := b.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if !result.OK {
		return nil, fmt.Errorf("telegram API error")
	}
	return result.Result, nil
}

func (b *Bot) handleMessage(msg *Message) {
	text := strings.TrimSpace(msg.Text)
	chatID := msg.Chat.ID

	if b.isAdminChat(chatID) {
		b.handleAdminMessage(chatID, msg.From, text)
		return
	}

	switch {
	case strings.HasPrefix(text, "/start"):
		b.handleStart(chatID, msg.From, text)
	default:
		b.handleUserMessage(chatID, msg.From, text)
	}
}

func (b *Bot) isAdminChat(chatID int64) bool {
	b.mu.RLock()
	_, ok := b.admins[chatID]
	b.mu.RUnlock()
	if ok {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := b.adminRepo.GetByTelegramID(ctx, chatID)
	if err == nil {
		b.mu.Lock()
		b.admins[chatID] = ""
		b.mu.Unlock()
		return true
	}
	return false
}

// ──────────────── START / ORDER / CART ────────────────

func (b *Bot) handleStart(chatID int64, from User, text string) {
	parts := strings.SplitN(text, " ", 2)
	payload := ""
	if len(parts) > 1 {
		payload = strings.TrimSpace(parts[1])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if payload == "" {
		b.sendWelcome(chatID)
		return
	}

	switch {
	case strings.HasPrefix(payload, "order_"):
		b.handleOrderStart(ctx, chatID, from, strings.TrimPrefix(payload, "order_"))
	case strings.HasPrefix(payload, "cart_"):
		b.handleCartStart(ctx, chatID, from, strings.TrimPrefix(payload, "cart_"))
	default:
		b.sendMessage(chatID, "Команда не распознана. Используйте /start")
	}
}

func (b *Bot) handleOrderStart(ctx context.Context, chatID int64, from User, orderIDStr string) {
	orderUUID, err := uuid.Parse(orderIDStr)
	if err != nil {
		b.sendMessage(chatID, "Ошибка: неверный ID заказа")
		return
	}

	order, err := b.orderRepo.GetByIDWithJoins(ctx, orderUUID)
	if err != nil {
		b.sendMessage(chatID, "Заказ не найден. Проверьте ID или обратитесь в поддержку.")
		return
	}

	b.mu.Lock()
	b.orders[chatID] = orderIDStr
	b.mu.Unlock()

	if order.Product == nil {
		b.sendMessage(chatID, "Заказ не содержит информации о товаре")
		return
	}

	p := order.Product
	platform := platformEmoji[string(p.Platform)]
	if platform == "" {
		platform = string(p.Platform)
	}
	deliveryIcon := "🔑"
	deliveryLabel := "Ключ"
	if order.DeliveryMethod == domain.DeliveryMethodActivation {
		deliveryIcon = "🔐"
		deliveryLabel = "Активация на аккаунт"
	}
	price := fmt.Sprintf("%.0f ₽", *order.PaymentAmount)
	qty := ""
	if order.Quantity > 1 {
		qty = fmt.Sprintf("\n<b>Количество:</b> %d шт", order.Quantity)
	}
	typeIcon := typeEmoji[string(p.Type)]
	if typeIcon == "" {
		typeIcon = "🎮"
	}

	caption := fmt.Sprintf(`%s <b>%s</b>

<b>Платформа:</b> %s
<b>Тип:</b> %s
%s <b>Способ:</b> %s%s

<b>💰 Сумма:</b> %s

🆔 <b>Заказ #%s</b> создан!
👨‍💼 Менеджер скоро свяжется с вами.`, typeIcon, p.Title, platform, typeLabel(p.Type), deliveryIcon, deliveryLabel, qty, price, orderIDStr[:8])

	buttons := b.orderActionButtons(orderIDStr)

	imgURL := ""
	if p.ImageURL != nil {
		imgURL = b.fullImageURL(*p.ImageURL)
	}
	b.sendRichMessage(chatID, imgURL, caption, buttons)

	b.notifyAdminsAboutOrder(ctx, order, chatID, from)
}

func (b *Bot) handleCartStart(ctx context.Context, chatID int64, from User, idsStr string) {
	ids := strings.Split(idsStr, ",")
	type itemInfo struct {
		title string
		qty   int
		price float64
	}
	var items []itemInfo
	total := 0.0
	firstImg := ""

	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		orderUUID, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		order, err := b.orderRepo.GetByIDWithJoins(ctx, orderUUID)
		if err != nil {
			continue
		}
		if order.Product == nil {
			continue
		}
		price := 0.0
		if order.PaymentAmount != nil {
			price = *order.PaymentAmount
		}
		items = append(items, itemInfo{title: order.Product.Title, qty: order.Quantity, price: price})
		total += price
		if firstImg == "" && order.Product.ImageURL != nil {
			firstImg = b.fullImageURL(*order.Product.ImageURL)
		}
	}

	b.mu.Lock()
	if len(ids) > 0 {
		b.orders[chatID] = ids[0]
	}
	b.mu.Unlock()

	var sb strings.Builder
	sb.WriteString("🛒 <b>Ваша корзина</b>\n\n")
	for i, item := range items {
		if i >= 10 {
			sb.WriteString(fmt.Sprintf("… и ещё %d товаров\n", len(items)-i))
			break
		}
		sb.WriteString(fmt.Sprintf("%d. <b>%s</b>", i+1, item.title))
		if item.qty > 1 {
			sb.WriteString(fmt.Sprintf(" × %d", item.qty))
		}
		sb.WriteString(fmt.Sprintf(" — %.0f ₽\n", item.price))
	}
	sb.WriteString(fmt.Sprintf("\n━━━━━━━━━━━━━━━\n<b>💰 Итого:</b> %.0f ₽", total))

	firstOrderID := ""
	if len(ids) > 0 {
		firstOrderID = ids[0]
	}

	buttons := b.cartActionButtons(firstOrderID)

	b.sendRichMessage(chatID, firstImg, sb.String(), buttons)

	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		orderUUID, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		order, err := b.orderRepo.GetByIDWithJoins(ctx, orderUUID)
		if err != nil {
			continue
		}
		b.notifyAdminsAboutOrder(ctx, order, chatID, from)
	}
}

func typeLabel(t domain.ProductType) string {
	switch t {
	case domain.ProductTypeGame:
		return "Игра"
	case domain.ProductTypeCurrency:
		return "Валюта"
	case domain.ProductTypeSubscription:
		return "Подписка"
	}
	return string(t)
}

func (b *Bot) shopWebAppButton() map[string]any {
	return map[string]any{
		"text":    "🛒 Открыть магазин",
		"web_app": map[string]string{"url": b.tmaURL},
	}
}

func (b *Bot) orderActionButtons(orderID string) [][]map[string]any {
	short := orderID[:8]
	return [][]map[string]any{
		{
			{"text": "📋 Статус заказа", "callback_data": "status_" + short},
			{"text": "💬 Чат с менеджером", "callback_data": "chat_" + short},
		},
		{
			b.shopWebAppButton(),
			{"text": "📦 Мои заказы", "callback_data": "open_orders"},
		},
		{
			{"text": "❌ Отменить заказ", "callback_data": "cancel_" + short},
		},
	}
}

func (b *Bot) cartActionButtons(firstOrderID string) [][]map[string]any {
	short := ""
	if len(firstOrderID) >= 8 {
		short = firstOrderID[:8]
	}
	return [][]map[string]any{
		{
			{"text": "💬 Чат с менеджером", "callback_data": "chat_" + short},
		},
		{
			b.shopWebAppButton(),
			{"text": "📦 Мои заказы", "callback_data": "open_orders"},
		},
	}
}

// ──────────────── ADMIN NOTIFICATION ────────────────

func (b *Bot) notifyAdminsAboutOrder(ctx context.Context, order *domain.Order, userChatID int64, from User) {
	admins, err := b.adminRepo.List(ctx)
	if err != nil {
		slog.Error("failed to list admins", slog.String("error", err.Error()))
		return
	}

	price := "—"
	if order.PaymentAmount != nil {
		price = fmt.Sprintf("%.0f ₽", *order.PaymentAmount)
	}
	title := "Товар"
	if order.Product != nil {
		title = order.Product.Title
	}
	username := from.Username
	if username == "" {
		username = fmt.Sprintf("id%d", from.ID)
	}
	orderShort := order.ID.String()[:8]
	delivery := "ключ"
	if order.DeliveryMethod == domain.DeliveryMethodActivation {
		delivery = "активация"
	}

	text := fmt.Sprintf(`🔔 <b>Новый заказ #%s</b>

👤 <b>Пользователь:</b> @%s
🎮 <b>Товар:</b> %s
📦 <b>Способ:</b> %s
💳 <b>Сумма:</b> %s

/reply_%s — ответить пользователю`, orderShort, username, title, delivery, price, orderShort)

	adminsNotified := 0
	for _, admin := range admins {
		if !admin.IsActive {
			continue
		}
		if admin.TelegramID == from.ID {
			continue
		}
		err := b.sendMessage(admin.TelegramID, text)
		if err == nil {
			adminsNotified++
		}
	}

	if adminsNotified == 0 {
		slog.Info("no admins notified for new order", slog.String("order_id", order.ID.String()))
	}
}

// ──────────────── USER MESSAGE FORWARDING ────────────────

func (b *Bot) handleUserMessage(chatID int64, from User, text string) {
	b.mu.RLock()
	orderID, hasOrder := b.orders[chatID]
	b.mu.RUnlock()

	orderShort := "?"
	if hasOrder && len(orderID) >= 8 {
		orderShort = orderID[:8]
	}

	username := from.Username
	if username == "" {
		username = fmt.Sprintf("id%d", from.ID)
	}

	admins, err := b.adminRepo.List(context.Background())
	if err != nil {
		b.sendMessage(chatID, "Не удалось отправить сообщение. Попробуйте позже.")
		return
	}

	orderContext := ""
	if hasOrder {
		orderContext = fmt.Sprintf(" (#%s)", orderShort)
	}

	forwarded := 0
	for _, admin := range admins {
		if !admin.IsActive {
			continue
		}
		txt := fmt.Sprintf("💬 <b>@%s</b>%s:\n%s\n\n/reply_%s ответить", username, orderContext, text, orderShort)
		err := b.sendMessage(admin.TelegramID, txt)
		if err == nil {
			forwarded++
		}
	}

	if forwarded == 0 {
		b.sendMessage(chatID, "Ваше сообщение не доставлено. Все администраторы сейчас недоступны. Попробуйте позже.")
	}
}

// ──────────────── ADMIN COMMANDS ────────────────

func (b *Bot) handleAdminMessage(chatID int64, from User, text string) {
	if strings.HasPrefix(text, "/reply_") {
		b.handleReplyCommand(chatID, text)
		return
	}

	switch text {
	case "/start", "/help":
		b.sendMessage(chatID, `👨‍💼 <b>Панель администратора</b>

/reply_ORDERID текст — ответить пользователю
/list — список активных диалогов
/help — эта справка`)
	case "/list":
		b.handleListCommand(chatID)
	default:
		b.sendMessage(chatID, "Неизвестная команда. /help — справка")
	}
}

func (b *Bot) handleReplyCommand(adminChatID int64, text string) {
	parts := strings.SplitN(text, " ", 3)
	if len(parts) < 3 {
		b.sendMessage(adminChatID, "Формат: /reply_ORDERID текст\nНапример: /reply_A1B2C3 Здравствуйте! Чем могу помочь?")
		return
	}

	orderShort := strings.TrimPrefix(parts[0], "/reply_")
	replyText := parts[2]

	var targetChatID int64
	found := b.findUserChatID(orderShort, &targetChatID)

	if !found {
		b.sendMessage(adminChatID, fmt.Sprintf("Пользователь с заказом #%s не найден в активных диалогах. Возможно, диалог уже закрыт.", orderShort))
		return
	}

	err := b.sendMessage(targetChatID, fmt.Sprintf("👨‍💼 <b>Менеджер (заказ #%s):</b>\n%s", orderShort, replyText))
	if err != nil {
		b.sendMessage(adminChatID, "Ошибка отправки сообщения пользователю.")
		return
	}

	if order, dbErr := b.orderRepo.FindByIDPrefix(context.Background(), orderShort); dbErr == nil {
		admin, _ := b.adminRepo.GetByTelegramID(context.Background(), adminChatID)
		senderID := uuid.Nil
		if admin != nil {
			senderID = admin.ID
		}
		_ = b.orderSvc.SendChatMessage(context.Background(), order.ID, senderID, "admin", replyText)
	}

	b.sendMessage(adminChatID, "✅ Сообщение отправлено пользователю.")
}

func (b *Bot) findUserChatID(orderShort string, result *int64) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for chatID, oid := range b.orders {
		check := oid
		if len(oid) > 8 {
			check = oid[:8]
		}
		if check == orderShort || oid == orderShort {
			*result = chatID
			return true
		}
	}
	return false
}

func (b *Bot) handleListCommand(adminChatID int64) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.orders) == 0 {
		b.sendMessage(adminChatID, "Нет активных диалогов.")
		return
	}

	var lines []string
	lines = append(lines, "<b>Активные диалоги:</b>\n")
	for chatID, orderID := range b.orders {
		short := "?"
		if len(orderID) >= 8 {
			short = orderID[:8]
		}
		lines = append(lines, fmt.Sprintf("• Заказ #%s — user:%d — /reply_%s", short, chatID, short))
	}
	b.sendMessage(adminChatID, strings.Join(lines, "\n"))
}

// ──────────────── CALLBACKS ────────────────

func (b *Bot) handleCallback(cb *CallbackQuery) {
	data := cb.Data
	chatID := cb.Message.Chat.ID
	cbID := cb.ID

	b.answerCallbackQuery(cbID, "")

	switch {
	case strings.HasPrefix(data, "cancel_"):
		b.handleCancelCallback(chatID, strings.TrimPrefix(data, "cancel_"))
	case strings.HasPrefix(data, "status_"):
		orderShort := strings.TrimPrefix(data, "status_")
		b.sendMessage(chatID, fmt.Sprintf("📋 Откройте приложение, чтобы увидеть статус заказа #%s:\n%s/order/%s", orderShort, b.tmaURL, orderShort))
	case strings.HasPrefix(data, "chat_"):
		orderShort := strings.TrimPrefix(data, "chat_")
		b.sendMessage(chatID, fmt.Sprintf("💬 Напишите ваше сообщение — менеджер ответит в ближайшее время.\n\nЗаказ #%s", orderShort))
	case strings.HasPrefix(data, "confirm_cancel_"):
		b.handleConfirmCancel(chatID, strings.TrimPrefix(data, "confirm_cancel_"))
	case data == "open_shop":
		// Web App открывается через кнопку меню — не шлём ссылку в чат.
		return
	case data == "open_orders":
		b.sendMessage(chatID, fmt.Sprintf("📋 Мои заказы:\n%s/orders", b.tmaURL))
	default:
		b.sendMessage(chatID, "Команда не распознана")
	}
}

func (b *Bot) handleCancelCallback(chatID int64, orderShort string) {
	buttons := [][]map[string]any{
		{
			{"text": "✅ Да, отменить", "callback_data": "confirm_cancel_" + orderShort},
			{"text": "❌ Нет", "callback_data": "cancel_no"},
		},
	}
	b.sendMessageWithButtons(chatID,
		fmt.Sprintf("❓ <b>Вы уверены, что хотите отменить заказ #%s?</b>\n\nЕсли товар уже получен, отмена невозможна.", orderShort),
		buttons)
}

func (b *Bot) handleConfirmCancel(chatID int64, orderShort string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b.mu.RLock()
	fullOrderID, ok := b.orders[chatID]
	b.mu.RUnlock()

	if !ok {
		// Try to find order by short ID
		for cid, oid := range b.orders {
			if (len(oid) >= 8 && oid[:8] == orderShort) || oid == orderShort {
				fullOrderID = oid
				chatID = cid
				ok = true
				break
			}
		}
	}

	if !ok {
		b.sendMessage(chatID, "Заказ не найден. Возможно, он уже отменён или завершён.")
		return
	}

	orderUUID, err := uuid.Parse(fullOrderID)
	if err != nil {
		b.sendMessage(chatID, "Ошибка ID заказа.")
		return
	}

	order, err := b.orderRepo.GetByID(ctx, orderUUID)
	if err != nil {
		b.sendMessage(chatID, "Заказ не найден.")
		return
	}
	err = b.orderSvc.CancelOrderByUser(ctx, orderUUID, order.UserID, "Отменён пользователем через бота")
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("Не удалось отменить заказ: %s", err.Error()))
		return
	}

	b.sendMessage(chatID, "✅ Заказ отменён.")
	b.notifSvc.SendOrderStatusUpdate(ctx, &domain.Order{
		ID:     orderUUID,
		Status: domain.OrderStatusCancelled,
		User:   &domain.User{TelegramID: chatID},
	})
}

// ──────────────── SEND METHODS ────────────────

func (b *Bot) sendRichMessage(chatID int64, imgURL, caption string, buttons [][]map[string]any) {
	if imgURL != "" && (strings.HasPrefix(imgURL, "https://") || strings.HasPrefix(imgURL, "http://")) {
		err := b.sendPhoto(chatID, imgURL, caption, buttons)
		if err == nil {
			return
		}
		slog.Error("sendPhoto failed, falling back to text", slog.String("error", err.Error()))
	}
	b.sendMessageWithButtons(chatID, caption, buttons)
}

func (b *Bot) sendPhoto(chatID int64, photoURL, caption string, buttons [][]map[string]any) error {
	if b.token == "" {
		slog.Info("bot photo (dev mode)", slog.Int64("chat_id", chatID), slog.String("caption", caption))
		return nil
	}

	body := map[string]any{
		"chat_id":   chatID,
		"photo":     photoURL,
		"caption":   caption,
		"parse_mode": "HTML",
	}

	if len(buttons) > 0 {
		body["reply_markup"] = map[string]any{
			"inline_keyboard": buttons,
		}
	}

	data, _ := json.Marshal(body)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", b.token)

	resp, err := b.httpClient.Post(url, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: %d %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (b *Bot) answerCallbackQuery(cbID, text string) {
	if b.token == "" {
		return
	}
	body := map[string]any{
		"callback_query_id": cbID,
		"text":              text,
	}
	data, _ := json.Marshal(body)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", b.token)
	b.httpClient.Post(url, "application/json", strings.NewReader(string(data)))
}

func (b *Bot) sendWelcome(chatID int64) {
	text := `🎮 <b>Добро пожаловать в GameStore!</b>

Здесь вы можете купить игры, валюту и подписки
для PlayStation и Xbox по выгодным ценам!

🔹 Мгновенная выдача ключей
🔹 Активация на ваш аккаунт
🔹 Поддержка 24/7`

	buttons := [][]map[string]any{
		{
			b.shopWebAppButton(),
		},
		{
			{"text": "📋 Мои заказы", "callback_data": "open_orders"},
		},
	}

	b.sendMessageWithButtons(chatID, text, buttons)
}

func (b *Bot) sendMessage(chatID int64, text string) error {
	return b.sendMessageWithButtons(chatID, text, nil)
}

func (b *Bot) sendMessageWithButtons(chatID int64, text string, buttons [][]map[string]any) error {
	if b.token == "" {
		slog.Info("bot message (dev mode)", slog.Int64("chat_id", chatID), slog.String("text", text))
		return nil
	}

	body := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	if len(buttons) > 0 {
		body["reply_markup"] = map[string]any{
			"inline_keyboard": buttons,
		}
	}

	data, _ := json.Marshal(body)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)

	resp, err := b.httpClient.Post(url, "application/json", strings.NewReader(string(data)))
	if err != nil {
		slog.Error("Failed to send bot message", slog.String("error", err.Error()))
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		slog.Error("Telegram API error", slog.Int("status", resp.StatusCode), slog.String("body", string(respBody)))
		return fmt.Errorf("telegram API error: %d %s", resp.StatusCode, string(respBody))
	}

	return nil
}
