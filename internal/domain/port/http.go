package port

// HTTPServer определяет порт для HTTP сервера
type HTTPServer interface {
	Start() error
}
