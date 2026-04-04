# Лабораторная работа №2(10). Веб-разработка: FastAPI (Python) vs Gin (Go)
**Студент:** *Платов Артем Русланович*\
**Группа:** *220032-11*\
**Вариант:** *6*\
**Сложность:** *Средняя*
---

## Task6. Сравнить скорость ответа FastAPI и Gin под нагрузкой (wrk/ab).

### Endpoints
- /ping
  - port: 8000/8001 
  - description: {"message": "pong"}
- /json
  - port: 8000/8001
  - description: JSON с uuid, timestamp, nested data
- /echo
  - port: 8000/8001
  - description: Возврат POST тела
- /slow
  - port: 8000/8001 
  - description: 50ms задержка

### Запуск серверов(PowerShell):
#### Окно 1 — FastAPI:
cd D:/.../fastapi_server\
py main.py

#### Окно 2 — Gin:

cd D:/.../gin_server\
go run main.go

### Запуск бенчмарка:
cd D:/.../task6_mid\
python benchmark.py

Скрипт автоматически:\
Проверит что оба сервера работают
Запустит тесты /ping, /json, /slow, /echo
Покажет сравнение RPS, latency (p50, p90, p99)
Определит победителя для каждого endpoint

### Запуск тестов:
#### Go (Gin)
cd gin_server\
go test -v ./...

#### Python (FastAPI)
cd fastapi_server\
pip install -r requirements.txt\
pytest -v test_main.py

---

## Task8. Добавить Swagger-документацию для FastAPI и OpenAPI для Gin.

### Запуск серверов:
#### FastAPI
cd task8_mid/fastapi_server\
py -3 -m pip install -r requirements.txt\
py main.py

#### Gin (новый терминал)
cd task8_mid/gin_server\
go mod download\
go run main.go

### Запуск тестов:
#### Go (Gin)
cd gin_server\
go test -v ./...

#### Python (FastAPI)
cd fastapi_server\
pip install -r requirements.txt\
pytest -v test_main.py

---

## Task2. Добавить middleware для логирования в Go.

### Запуск сервера:
cd task2_mid\
go run main.go

### Запуск тестов:
cd gin_server\
go test -v ./...


### Логи выводятся в:

stdout (консоль)\
server.log (файл)\

#### Формат логов:
[GET] /hello?name=John | Status: 200 | Latency: 15µs | IP: 127.0.0.1