[![Coverage Status](https://coveralls.io/repos/github/beliaev-aa/notifications/badge.svg?branch=master)](https://coveralls.io/github/beliaev-aa/notifications?branch=master)

# notifications

Notification web service для обработки webhook запросов от YouTrack и отправки уведомлений через различные каналы (Telegram, VK Teams, Logger).

## Возможности

- Обработка webhook запросов от YouTrack
- Отправка уведомлений через Telegram, VK Teams и Logger каналы
- Настройка проектов YouTrack и разрешенных каналов уведомлений для каждого проекта
- Приватность проектов: каждый проект использует свой `chat_id` для Telegram и VK Teams
- Управление уведомлениями для черновиков: возможность отключить отправку уведомлений для задач-черновиков на уровне настройки проекта

## Конфигурация

### Структура конфигурации

Приложение использует YAML файл конфигурации (по умолчанию `./config/config.yml`). Путь можно переопределить через переменную окружения `CONFIG_PATH`.

**Примечание:** Параметр `vkteams.insecure_skip_verify` позволяет игнорировать проверку SSL сертификата при подключении к VK Teams API. Это может быть полезно для разработки с самоподписанными сертификатами, но **не рекомендуется для production** окружения из соображений безопасности.

Пример конфигурации:

```yaml
http:
  addr: ":3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5

telegram:
  bot_token: "your_bot_token"  # Глобальный токен бота (обязателен, если используется Telegram)
  timeout: 10                  # Таймаут для HTTP запросов к Telegram API (секунды)

vkteams:
  bot_token: "your_vkteams_bot_token"  # Глобальный токен бота (обязателен, если используется VK Teams)
  timeout: 10                          # Таймаут для HTTP запросов к VK Teams API (секунды)
  api_url: "https://api.vkteams.ru/bot/v1"  # URL API VK Teams (обязателен)
  insecure_skip_verify: false         # Игнорировать проверку SSL сертификата (не рекомендуется для production)

logger:
  level: "debug"

notifications:
  youtrack:
    projects:  # ⚠️ Ключ "projects" обязателен!
      projectName1:  # Имя проекта в нижнем регистре (рекомендуется)
        allowedChannels: [telegram, logger]
        sendDraftNotification: true  # Отправлять уведомления для черновиков (по умолчанию true)
        telegram:
          chat_id: "123456789"  # Обязательно, если telegram в allowedChannels
      projectName2:
        allowedChannels: [logger]
        sendDraftNotification: false  # Не отправлять уведомления для черновиков
      projectName3:
        allowedChannels: [telegram]
        telegram:
          chat_id: "987654321"
        # sendDraftNotification не указан - будет использовано значение по умолчанию (true)
      projectName4:
        allowedChannels: [vkteams, logger]
        sendDraftNotification: false
        vkteams:
          chat_id: "chat123"  # Обязательно, если vkteams в allowedChannels
      projectName5:
        allowedChannels: [telegram, vkteams]
        telegram:
          chat_id: "123456789"
        vkteams:
          chat_id: "chat456"
```

**Важные замечания:**

1. **Ключ `projects` обязателен:** Проекты должны быть указаны под ключом `projects` внутри `youtrack`. Неправильно: `youtrack: DEMO: ...`. Правильно: `youtrack: projects: DEMO: ...`

2. **Нормализация регистра:** Если в YouTrack проект называется "DEMO" (верхний регистр), в конфигурации его можно указать как `demo`, `DEMO` или `Demo` - все варианты будут работать одинаково благодаря нормализации регистра. После загрузки конфигурации все ключи проектов автоматически приводятся к нижнему регистру.

### Настройка проектов

Каждый проект YouTrack должен быть настроен в секции `notifications.youtrack.projects`:

- **`allowedChannels`** - список разрешенных каналов уведомлений:
  - `telegram` - отправка через Telegram
  - `vkteams` - отправка через VK Teams
  - `logger` - логирование уведомлений
- **`sendDraftNotification`** - отправлять ли уведомления для черновиков:
  - `true` - отправлять уведомления для черновиков (значение по умолчанию)
  - `false` - не отправлять уведомления для черновиков
  - Если параметр не указан, используется значение `true` по умолчанию
- **`telegram.chat_id`** - обязателен, если `telegram` в `allowedChannels`
- **`vkteams.chat_id`** - обязателен, если `vkteams` в `allowedChannels`

**Важно:** Имена проектов нормализуются к нижнему регистру при загрузке конфигурации и при обработке webhook. Это означает, что проекты "DEMO", "Demo" и "demo" будут обрабатываться одинаково. В конфигурации можно указать проект в любом регистре, но рекомендуется использовать нижний регистр для единообразия.

### Поведение

- Если настройки для проекта не существуют, то уведомление игнорируется, при этом пишется запись в лог через системный логгер
- Если настройки существуют, то используются только разрешенные каналы
- Для задач-черновиков: если `sendDraftNotification = false`, уведомление игнорируется. По умолчанию уведомления для черновиков отправляются
- Для Telegram канала: **обязательно** используется `chat_id` из настроек проекта (приватность проектов)
- Для VK Teams канала: **обязательно** используется `chat_id` из настроек проекта (приватность проектов)
- VK Teams использует такое же форматирование сообщений, как и Telegram канал

### Логирование и отладка

Приложение логирует важную информацию для отладки:

- **При обработке webhook:** Если проект не найден, в debug-логах будет список доступных проектов:
  ```
  {"level":"debug","msg":"Project not found in configuration","project":"DEMO","project_normalized":"demo","available_projects":["demo","project2"]}
  ```

- **Черновики:** Если уведомление для черновика игнорируется из-за `sendDraftNotification = false`, в логах будет запись:
  ```
  {"level":"info","msg":"Draft notification is disabled for project, ignoring notification","project":"projectName","isDraft":true}
  ```

- **Нормализация регистра:** В логах отображаются как оригинальное, так и нормализованное имя проекта для удобства отладки.

- **Регистрация каналов:** При старте приложения логируется регистрация каждого канала уведомлений:
  ```
  {"level":"info","msg":"Notification channel registered","channel":"logger"}
  {"level":"info","msg":"Notification channel registered","channel":"telegram"}
  {"level":"info","msg":"Notification channel registered","channel":"vkteams"}
  ```

## Переменные окружения

Конфигурацию можно переопределить через переменные окружения (приоритет над YAML):

- `CONFIG_PATH` - путь к файлу конфигурации
- `HTTP_ADDR` - адрес HTTP сервера
- `HTTP_SHUTDOWN_TIMEOUT` - таймаут graceful shutdown (секунды)
- `HTTP_READ_TIMEOUT` - таймаут чтения запроса (секунды)
- `HTTP_WRITE_TIMEOUT` - таймаут записи ответа (секунды)
- `TELEGRAM_BOT_TOKEN` - токен Telegram бота
- `TELEGRAM_TIMEOUT` - таймаут для HTTP запросов к Telegram API (секунды)
- `VKTEAMS_BOT_TOKEN` - токен VK Teams бота
- `VKTEAMS_TIMEOUT` - таймаут для HTTP запросов к VK Teams API (секунды)
- `VKTEAMS_API_URL` - URL API VK Teams (обязателен, например: https://api.vkteams.ru/bot/v1)
- `VKTEAMS_INSECURE_SKIP_VERIFY` - игнорировать проверку SSL сертификата (только `true` или `false`)
- `LOG_LEVEL` - уровень логирования (debug, info, warn, error)

## Настройка webhook в YouTrack

Для работы сервиса необходимо настроить webhook в YouTrack, который будет отправлять уведомления о изменениях в задачах.

### Установка скрипта webhook

1. Откройте настройки проекта в YouTrack (или глобальные настройки для всех проектов)
2. Перейдите в раздел **Администрирование** → **Рабочие процессы** → **Новый рабочий процесс** → **Редактор JavaScript**
3. Создайте новый скрипт или откройте существующий для действия **При изменении**
4. Скопируйте содержимое файла `scripts/youtrack/webhook.js` в редактор скрипта
5. **Важно**: Измените URL в скрипте на адрес вашего сервиса:
   ```javascript
   const WEBHOOK_URL = 'http://your-server:3000/webhook/youtrack';
   ```
   Для локальной разработки с Docker можно использовать:
   ```javascript
   const WEBHOOK_URL = 'http://notification.local:3000/webhook/youtrack';
   ```
6. Сохраните скрипт
7. Примените скрипт к проекту, для которых необходимо отслеживать уведомления через сервис

### Требования к полям в YouTrack

Скрипт требует наличия следующих полей в проекте:
- **Assignee** (тип: User) - исполнитель задачи
- **State** (тип: Enum) - статус задачи
- **Priority** (тип: Enum) - приоритет задачи

Если эти поля отсутствуют в проекте, их необходимо создать перед установкой скрипта.

### Отслеживаемые события

Скрипт отправляет уведомления при следующих изменениях:
- Изменение исполнителя задачи (Assignee)
- Изменение статуса задачи (State)
- Изменение приоритета задачи (Priority)
- Добавление комментария к задаче (Comment)

### Формат данных

Скрипт формирует JSON payload со следующей структурой:
- Информация о проекте (name, presentation)
- Информация о задаче (idReadable, summary, url, state, priority, assignee)
- Информация об авторе изменения (updater)
- Массив изменений (changes) с деталями каждого изменения

## Запуск

```bash
go run cmd/server/main.go
```

## Тестирование

```bash
go test ./...
```

### Особенности реализации

- **Регистронезависимое сравнение проектов:** Все имена проектов нормализуются к нижнему регистру при загрузке конфигурации и при обработке webhook
- **Приватность проектов:** Каждый проект использует свой `chat_id` для Telegram и VK Teams, что обеспечивает изоляцию уведомлений между проектами
- **Управление черновиками:** Настройка `sendDraftNotification` позволяет контролировать отправку уведомлений для задач-черновиков на уровне каждого проекта. По умолчанию уведомления для черновиков отправляются
- **Единое форматирование:** VK Teams канал использует такое же форматирование сообщений, как и Telegram канал
- **Гибкая конфигурация:** Поддержка как YAML файлов, так и переменных окружения (приоритет у ENV)
