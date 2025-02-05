# Тестирование

Для тестов требуется запущенный Docker контейнер с Postgres, заполненный тестовыми данными.

## 1. Установка переменной окружения с паролем для Postgres

```console
export POSTGRES_PASSWORD='some_pass'
```

## 2. Запустить контейнер

```console
chmod +x cmd/run_docker.sh
./cmd/run_docker.sh
```

## 3. Запуск тестов

```console
go test -v ./...
```
