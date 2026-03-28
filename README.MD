# Subscription Service API

REST-сервис для агрегации данных об онлайн подписках пользователей. Реализован на Go с использованием PostgreSQL.

## Функциональность

- **CRUDL операции** над подписками:
  - Создание подписки
  - Получение подписки по ID
  - Обновление подписки
  - Удаление подписки
  - Список всех подписок (с пагинацией)
- **Агрегация данных**: подсчет суммарной стоимости подписок за выбранный период
- **Фильтрация** по ID пользователя и названию сервиса
- **Swagger документация** API
- **Логирование** в формате JSON
- **Docker Compose** для простого развертывания

## Требования

- Docker и Docker Compose
- Или Go 1.25+ и PostgreSQL 15+ (для локальной разработки)

## Быстрый запуск

### С помощью Docker (рекомендуется)

```bash
# Клонируем репозиторий
git clone https://github.com/your-username/subscription-service.git
cd subscription-service

# Запускаем сервис
docker compose up --build
```
## Как протестировать API

### Способ 1: Через Swagger UI (рекомендуется)

1. Запустите сервис: `docker compose up`
2. Откройте браузер: http://localhost:8080/swagger/index.html
3. Нажмите на нужный эндпоинт
4. Нажмите "Try it out"
5. Заполните параметры
6. Нажмите "Execute"

### Способ 2: Через командную строку (Windows CMD)

Создайте файл `test_api.cmd`:

```cmd
@echo off
echo ========================================
echo 1. СОЗДАНИЕ ПОДПИСКИ
echo ========================================
curl -X POST http://localhost:8080/api/subscriptions -H "Content-Type: application/json" -d "{\"service_name\":\"Yandex Plus\",\"price\":400,\"user_id\":\"60601fee-2bf1-4721-ae6f-7636e79a0cba\",\"start_date\":\"07-2025\"}"
echo.

echo ========================================
echo 2. ПОЛУЧЕНИЕ ВСЕХ ПОДПИСОК
echo ========================================
curl http://localhost:8080/api/subscriptions
echo.

echo ========================================
echo 3. СУММАРНАЯ СТОИМОСТЬ ЗА 2025 ГОД
echo ========================================
curl "http://localhost:8080/api/subscriptions/summary?start_date=01-2025&end_date=12-2025"
echo.

echo ========================================
echo 4. СУММАРНАЯ СТОИМОСТЬ С ФИЛЬТРАЦИЕЙ ПО СЕРВИСУ
echo ========================================
curl "http://localhost:8080/api/subscriptions/summary?start_date=01-2025&end_date=12-2025&service_name=Yandex"
echo.

pause