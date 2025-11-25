package port

import "net/http"

// HTTPClient определяет порт для выполнения HTTP запросов
type HTTPClient interface {
	// Do отправляет HTTP запрос и возвращает ответ
	Do(req *http.Request) (*http.Response, error)
}
