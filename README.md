# gRPC: Proto файл и кодогенерация

## Файлы проекта

- [proto/order.proto](proto/order.proto) — контракт
- [generated/order/order.pb.go](generated/order/order.pb.go) — сгенерированные структуры
- [generated/order/order_grpc.pb.go](generated/order/order_grpc.pb.go) — сгенерированный интерфейс + клиент
- [server/main.go](server/main.go) — сервер
- [client/main.go](client/main.go) — клиент
- [Makefile](Makefile) — `make server`, `make client`, `make proto`

## 1. Proto файл — контракт ([proto/order.proto](proto/order.proto))

```protobuf
// Protobuf придумал Google в 2001
// для внутреннего использования.
// gRPC тоже Google (внутри назывался Stubby).
//
// proto1 — внутренний, не публичный
// proto2 — 2008, открыли исходники
//   поля были required/optional:
//   required string name = 1;
//   не заполнил required → ошибка
//   добавил новое required поле →
//   старые клиенты не знают о нём →
//   ошибка
// proto3 — 2016, стандарт
//   все поля необязательные
//   клиент не заполнил поле →
//   сервер получит дефолт (0, "")
syntax = "proto3";

// чтобы message (типы) не конфликтовали
// между разными proto файлами
// (order.Status vs delivery.Status)
package order;

// Go import path + package name
option go_package = "generated/order";

// message = структура данных
// 1, 2, 3 = номера полей
message GetOrderRequest {
    string order_id = 1;
}

message GetOrderResponse {
    string id = 1;
    string status = 2;
    float total = 3;
}

// service = набор методов
service OrderService {
    // rpc = один метод
    rpc GetOrder(GetOrderRequest)
        returns (GetOrderResponse);
}
```

## 2. Кодогенерация

```bash
# Go — язык от Google (2009)
# типизированный, для бэкенда
# синтаксис минимальный (C-подобный)
# встроенная конкурентность (goroutines)
# сборщик мусора (не надо free/delete)

# go install — устанавливаем Go-пакет
# глобально на всю ОС (в $GOPATH/bin)
# protoc-gen-go — плагин для protoc,
# генерирует Go-структуры из message
go install \
  google.golang.org/protobuf/cmd/protoc-gen-go@latest

# protoc-gen-go-grpc — плагин для protoc,
# генерирует интерфейс сервера и клиент
# из service
go install \
  google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Генерируем
# protoc — компилятор Protocol Buffers
# парсит .proto файл и передаёт
# плагинам для генерации кода
# плагины есть для Go, Java, Python,
# C++, Rust, и других языков
#
# установка:
#   brew install protobuf      (macOS)
#   apt install protobuf-compiler (Linux)

# куда класть .pb.go (структуры)
# куда класть _grpc.pb.go (сервис)
protoc \
  --go_out=. \
  --go-grpc_out=. \
  order.proto
```

```
На выходе:

./generated/order/order.pb.go      — структуры
./generated/order/order_grpc.pb.go — интерфейс + клиент
(см. generated/order/)

Руками не трогаем.
Перегенерируем при изменении proto.
```

Что внутри [generated/order/order.pb.go](generated/order/order.pb.go):

```go
// "order" взято из "generated/order"
package order

type GetOrderRequest struct {
    // protobuf передаёт не имена полей,
    // а номера (1, 2, 3)
    // "bytes,1" = тип bytes, поле номер 1
    // менять номера нельзя — сломает
    // старых клиентов
    OrderId string `protobuf:"bytes,1"`
}

type GetOrderResponse struct {
    Id     string  `protobuf:"bytes,1"`
    Status string  `protobuf:"bytes,2"`
    Total  float32 `protobuf:"fixed32,3"`
}
```

Импорт в другом сервисе:

```
Сгенерированный код хранят
в отдельном репозитории (gen-proto).
И сервер, и клиент импортируют
его как Go-модуль.
```

```go
// полный путь = module из go.mod +
// go_package из proto файла
import pb "github.com/user/myapp/generated/order"

// client создаётся так:
// conn, _ := grpc.Dial("localhost:50051")
// client := pb.NewOrderServiceClient(conn)

resp, _ := client.GetOrder(
    ctx, &pb.GetOrderRequest{...},
)
```

## 3. Сервер — имплементация ([server/main.go](server/main.go))

```go
type server struct {
    // Обязательный embed — без него
    // не скомпилируется.
    // Выглядит как костыль: заглушка
    // в продакшн-коде. Но gRPC-Go
    // это требует.
    //
    // gRPC родился в Google (монорепо),
    // где все сервисы обновляются разом.
    // Заглушка страхует: добавили новый
    // rpc — старые серверы не ломаются,
    // новый метод вернёт Unimplemented.
    //
    // В обычных проектах менее актуально:
    // версия proto фиксирована в go.mod.
    // Но таковы правила фреймворка.
    pb.UnimplementedOrderServiceServer
}

func (s *server) GetOrder(
    ctx context.Context,
    req *pb.GetOrderRequest,
) (*pb.GetOrderResponse, error) {
    // Обычная Go логика
    // Сходили в базу, собрали ответ
    return &pb.GetOrderResponse{
        Id:     req.OrderId,
        Status: "paid",
        Total:  1500.0,
    }, nil
}
```

## Запуск сервера

```go
func main() {
    lis, _ := net.Listen("tcp", ":50051")

    s := grpc.NewServer()
    pb.RegisterOrderServiceServer(s, &server{})

    s.Serve(lis)
}
```

## 4. Клиент — вызов как функция ([client/main.go](client/main.go))

```go
func main() {
    conn, _ := grpc.Dial(
        "localhost:50051",
        grpc.WithInsecure(),
    )
    defer conn.Close()

    client := pb.NewOrderServiceClient(conn)

    resp, _ := client.GetOrder(
        context.Background(),
        &pb.GetOrderRequest{
            OrderId: "order-123",
        },
    )

    // "paid"
    fmt.Println(resp.Status)
    // 1500
    fmt.Println(resp.Total)
}
```

```
Никаких URL-ов.
Никакого маршалинга JSON.
Всё типизировано.
Компилятор проверяет.
```

## Итого

```
proto файл
  ↓
protoc (кодогенерация)
  ↓
┌──────────────┬──────────────┐
│ order.pb.go  │ order_grpc.  │
│ (структуры)  │ pb.go        │
│              │ (интерфейс)  │
└──────┬───────┴──────┬───────┘
       │              │
   ┌───▼───┐    ┌─────▼─────┐
   │Сервер │    │  Клиент   │
   │GetOrder│    │GetOrder() │
   │  impl │    │как функция│
   └───────┘    └───────────┘

Поменял proto →
  перегенерировал →
  компилятор покажет
  где сломалось.
```
