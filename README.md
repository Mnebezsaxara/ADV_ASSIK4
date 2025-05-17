# ADV_ASSIK4

## 🎯 Цель задания

Добавить в микросервисы:
- Redis-кеширование для read-heavy endpoint
- MongoDB транзакции
- Управление схемой через Mongo (коллекции уже созданы)

---

## 📦 Сервисы и порты

| Сервис             | Порт   |
|--------------------|--------|
| user-service       | 50051  |
| inventory-service  | 50052  |
| order-service      | 50053  |
| Redis              | 6379   |
| MongoDB (Replica Set) | 27017 |
| NATS               | 4222   |

---

## ✅ Выполненные требования

### 🔹 Redis Caching
- `user-service`: `GetUserProfile` — кешируется по ключу `user:<id>`
- `inventory-service`: `GetProductByID` — кешируется по ключу `product:<id>`
- TTL кеша — 5 минут
- При необходимости кеш можно инвалидировать

### 🔹 MongoDB Transactions
- `order-service`: `CreateOrder` работает в транзакции:
  - сохраняет заказ
  - отправляет событие в NATS
  - при ошибке — полный rollback

---

## 🚀 Запуск MongoDB Replica Set (локально через Docker)

> ⚠️ Обязательно подключение по внешнему IP (например, `192.168.1.70`) — для взаимодействия с клиентами **вне контейнера** (Compass, mongosh и т.д.)

```bash
docker run -d --name mongo-rs \
  --add-host=host.docker.internal:host-gateway \
  -p 27017:27017 \
  mongo --replSet rs0 --bind_ip_all


Затем в контейнере:

docker exec -it mongo-rs mongosh

> rs.initiate()
> cfg = rs.conf()
> cfg.members[0].host = "192.168.1.70:27017"  # ⚠️ Указать ваш IP из `ipconfig`
> rs.reconfig(cfg, { force: true })
Проверить:

> rs.status()  # должно быть stateStr: "PRIMARY"


🧪 Тесты (через BloomRPC)
1️⃣ user-service → GetUserProfile
{
  "id": "64f8f8e2a1b2c933a9c12345"
}
✅ Первый вызов: "Отдано из Mongo и закешировано"
🔁 Повтор: "Отдано из Redis кеша"

2️⃣ inventory-service → GetProductByID
{
  "id": "67f2a30cb9e23361a8b71238"
}
✅ Первый вызов: "Отдано из Mongo и закешировано"
🔁 Повтор: "Отдано из Redis кеша"

3️⃣ order-service → CreateOrder
{
  "user_id": "6828a457b06e9a781c49fc58",
  "products": [
    {
      "product_id": "67f2a30cb9e23361a8b71238",
      "quantity": 2
    }
  ]
}

📤 В логах order-service:
Событие 'order.created' отправлено в NATS
✅ Транзакция успешно завершена

📉 В логах inventory-service:
Stock is reduced: 67f2a30cb9e23361a8b71238 : -2

📎 Примечания
Все подключения к MongoDB происходят по адресу:
mongodb://192.168.1.70:27017/?replicaSet=rs0

Это обязательно, чтобы Docker, Compass и gRPC-клиенты видели одну и ту же MongoDB

Redis, NATS и gRPC работают без изменений




