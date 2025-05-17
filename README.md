# ADV_ASSIK4
## 🎯 Цель задания

Добавить в микросервисы:
- Redis-кеширование для read-heavy endpoint
- MongoDB транзакции
- Управление схемой через Mongo (коллекции уже созданы)

---

## 📦 Сервисы и порты

| Сервис           | Порт   |
|------------------|--------|
| user-service     | 50051  |
| inventory-service| 50052  |
| order-service    | 50053  |
| Redis            | 6379   |
| MongoDB (Replica Set) | 27017 |
| NATS             | 4222   |

---

## ✅ Выполненные требования

### 🔹 Redis Caching
- `user-service`: `GetUserProfile` — кешируется по ключу `user:<id>`
- `inventory-service`: `GetProductByID` — кешируется по ключу `product:<id>`
- TTL кеша — 5 минут
- При обновлении данных кеш инвалидируется (при необходимости)

### 🔹 MongoDB Transactions
- `order-service`: `CreateOrder` работает в транзакции:
  - сохраняет заказ
  - отправляет событие в NATS
  - в случае ошибки откатывается

---

## 🚀 Запуск MongoDB как Replica Set

```bash
docker run --name mongo-rs -p 27017:27017 -d mongo --replSet rs0
docker exec -it mongo-rs mongosh
> rs.initiate()
> cfg = rs.conf()
> cfg.members[0].host = "localhost:27017"
> rs.reconfig(cfg, { force: true })

🧪 Тесты (через BloomRPC)

1️⃣ user-service → GetUserProfile
Service: UserService
Method: GetUserProfile
Request:
{
  "id": "64f8f8e2a1b2c933a9c12345"
}
Ожидаемые логи:
Первый раз: ✅ Отдано из Mongo и закешировано
Повторно: 🔁 Отдано из Redis кеша

2️⃣ inventory-service → GetProductByID
Service: ProductService
Method: GetProductByID
Request:
{
  "id": "67f2a30cb9e23361a8b71238"
}
Ожидаемые логи:
Первый раз: ✅ Отдано из Mongo и закешировано
Повторно: 🔁 Отдано из Redis
3️⃣ order-service → CreateOrder
Service: OrderService
Method: CreateOrder
Request:
{
  "user_id": "6828a457b06e9a781c49fc58",
  "products": [
    {
      "product_id": "67f2a30cb9e23361a8b71238",
      "quantity": 2
    }
  ]
}
Ожидаемые логи:
📤 Событие 'order.created' отправлено в NATS
✅ Транзакция успешно завершена
(в inventory-service): Stock is reduced: 67f2a30cb9e23361a8b71238 : -2
