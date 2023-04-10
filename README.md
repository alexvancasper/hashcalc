
# Hash calc service

Сервис считает SHA3-512 хэши и складывает их в БД PostgreSQL. Возвращает id и расчитанный hash. 
Логи пишет в GrayLog, статистика собирается в Prometheus и визуализируется в Grafana.

### Как запустить?

Склонировать репозиторий
```sh
git clone <path to repo>
cd final/server
docker-compose up -d
```

### Как запустить тесты?
Предварительно нужно поднять PostgreSQL

```sh
docker run -d --rm \
    --name postgres_test \
    -p 5432:5432 \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=hashdb \
    postgres:latest
```

Удостовериться, что БД запустилась.
```sh
docker ps -a | grep postgres_test
```
Затем
```sh
~/final/server$ make test

```

Остановить PostgresQL
```sh
docker stop postgres_test
```

### Logging levels

```sh
Panic =  0
Fatal =  1
Error =  2
Warn =  3
Info =  4
Debug =  5
Trace =  6
```
### Configuration examples  

#### server

```yaml
server:
name: Compute hash server
host: 0.0.0.0
port: 8090
worker-count: 5
cache-count: 5
db:
# Supported DB type is postgres only
type: postgres
pool-count: 5
host: postgres
port: 5432
user: postgres
pass: postgres
dbname: hashdb
ssl: disable
metric:
host: 0.0.0.0
port: 7755
path: metrics
logging:
provider: graylog
host: graylog
port: 12201
level: 6
```
#### Client
```yaml
client:
name: Compute hash client
host: 0.0.0.0
port: 8080
grpc:
host: hash-calc-service
port: 8090
metric:
host: 0.0.0.0
port: 7766
path: metrics
logging:
provider: graylog
host: graylog
port: 12201
level: 6
```

#### Prometheus
```yaml
scrape_configs:
- job_name: main
scrape_interval: 5s
static_configs:
- targets:
- hash-calc-service:7755
- hash-calc-client:7766
```
#### Migrations
```sql
-- +goose Up
-- +goose StatementBegin
CREATE  TABLE  hashes
(
id serial  primary key,
hash  CHAR(128) UNIQUE
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP  TABLE hashes;
-- +goose StatementEnd
```