package port

const (
	// ChannelLogger название канала логирования
	ChannelLogger = "logger"
	// ChannelTelegram название канала Telegram
	ChannelTelegram = "telegram"
)

// NotificationChannel определяет порт для отправки уведомлений через конкретный канал
type NotificationChannel interface {
	// Send отправляет уже отформатированное уведомление через данный канал
	Send(formattedMessage string) error
	// Channel возвращает название канала
	Channel() string
}

// NotificationSender определяет порт для отправки уведомлений через различные каналы
type NotificationSender interface {
	// Send отправляет уведомление через указанный канал
	Send(channel string, formattedMessage string) error
	// RegisterChannel регистрирует новый канал для отправки уведомлений
	RegisterChannel(channel NotificationChannel)
}
