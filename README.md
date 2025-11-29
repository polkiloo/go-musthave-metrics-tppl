## Профилирование и нагрузка

В репозитории определены make targets, которые запускают соответствующие бенчмарки, сохраняют `pprof` в директорию `profiles` и, при необходимости, эмулируют внешнюю нагрузку.

### profile-storage — in-memory хранилище

```bash
# профиль до оптимизаций
PROFILE_DIR=profiles/base make profile-storage

# профиль после оптимизаций
PROFILE_DIR=profiles/optimized make profile-storage

# сравнение результатов
go tool pprof -top -diff_base=profiles/base/storage.pprof profiles/optimized/storage.pprof
```

### profile-collector — сбор системных метрик

```bash
# профиль до оптимизаций
PROFILE_DIR=profiles/base make profile-collector

# профиль после оптимизаций
PROFILE_DIR=profiles/optimized make profile-collector

# сравнение результатов
go tool pprof -top -diff_base=profiles/base/collector.pprof profiles/optimized/collector.pprof
```


### profile-network — сетевой путь обновления метрик

```bash
# базовый профиль до оптимизаций
PROFILE_DIR=profiles/base make profile-network

# повторный запуск после оптимизаций
PROFILE_DIR=profiles/optimized make profile-network

# анализ изменений
go tool pprof -top -diff_base=profiles/base/network.pprof profiles/optimized/network.pprof
```
Пример отчёта `pprof -top -diff_base` после оптимизации сетевого пути:

```
File: handler.test
Type: alloc_space
Time: 2025-11-12 14:39:06 MSK
Showing nodes accounting for 4140.70kB, 134.61% of 3076.09kB total
Dropped 6 nodes (cum <= 15.38kB)
      flat  flat%   sum%        cum   cum%
 1545.77kB 50.25% 50.25%  1545.77kB 50.25%  regexp.onePassCopy
 1039.10kB 33.78% 84.03%  1039.10kB 33.78%  regexp/syntax.(*compiler).inst (inline)
  528.17kB 17.17% 101.20%   528.17kB 17.17%  net/http.init.func15
     514kB 16.71% 117.91%      514kB 16.71%  bufio.NewReaderSize (inline)
  513.69kB 16.70% 134.61%   513.69kB 16.70%  regexp.mergeRuneSets.func2 (inline)
 -512.05kB 16.65% 117.96%  -512.05kB 16.65%  regexp/syntax.(*parser).newRegexp (inline)
  512.05kB 16.65% 134.61%   512.05kB 16.65%  regexp/syntax.simplify1 (inline)
 -512.03kB 16.65% 117.96%  -512.03kB 16.65%  strings.NewReplacer (inline)
  512.01kB 16.64% 134.61%   512.01kB 16.64%  runtime.(*timers).addHeap
         0     0% 134.61%      514kB 16.71%  bufio.NewReader (inline)
         0     0% 134.61%  -512.03kB 16.65%  github.com/gin-gonic/gin.init
         0     0% 134.61%  1558.83kB 50.68%  github.com/go-playground/validator/v10.init
         0     0% 134.61%  1024.72kB 33.31%  github.com/go-playground/validator/v10.init.0
         0     0% 134.61%   528.17kB 17.17%  net/http.(*Request).write
         0     0% 134.61%      514kB 16.71%  net/http.(*conn).serve
         0     0% 134.61%   528.17kB 17.17%  net/http.(*persistConn).writeLoop
         0     0% 134.61%   528.17kB 17.17%  net/http.(*transferWriter).doBodyCopy
         0     0% 134.61%   528.17kB 17.17%  net/http.(*transferWriter).writeBody
         0     0% 134.61%   528.17kB 17.17%  net/http.getCopyBuf (inline)
         0     0% 134.61%      514kB 16.71%  net/http.newBufioReader
         0     0% 134.61%  2583.55kB 83.99%  regexp.Compile (inline)
         0     0% 134.61%  2583.55kB 83.99%  regexp.MustCompile
         0     0% 134.61%  2583.55kB 83.99%  regexp.compile
         0     0% 134.61%  2059.46kB 66.95%  regexp.compileOnePass
         0     0% 134.61%   513.69kB 16.70%  regexp.makeOnePass
         0     0% 134.61%   513.69kB 16.70%  regexp.makeOnePass.func1
         0     0% 134.61%   513.69kB 16.70%  regexp.mergeRuneSets
         0     0% 134.61%   512.05kB 16.65%  regexp/syntax.(*Regexp).Simplify
         0     0% 134.61%  1039.10kB 33.78%  regexp/syntax.(*compiler).compile
         0     0% 134.61%  1039.10kB 33.78%  regexp/syntax.(*compiler).rune
         0     0% 134.61%   524.09kB 17.04%  regexp/syntax.Compile
         0     0% 134.61%  -512.05kB 16.65%  regexp/syntax.Parse (inline)
         0     0% 134.61%  -512.05kB 16.65%  regexp/syntax.parse
         0     0% 134.61%   512.01kB 16.64%  runtime.(*scavengerState).sleep
         0     0% 134.61%   512.01kB 16.64%  runtime.(*timer).maybeAdd
         0     0% 134.61%   512.01kB 16.64%  runtime.(*timer).modify
         0     0% 134.61%   512.01kB 16.64%  runtime.(*timer).reset (inline)
         0     0% 134.61%   512.01kB 16.64%  runtime.bgscavenge
         0     0% 134.61%  2071.52kB 67.34%  runtime.doInit (inline)
         0     0% 134.61%  2071.52kB 67.34%  runtime.doInit1
         0     0% 134.61%  2071.52kB 67.34%  runtime.main
         0     0% 134.61%      513kB 16.68%  runtime.mcall
         0     0% 134.61%     -513kB 16.68%  runtime.mstart
         0     0% 134.61%     -513kB 16.68%  runtime.mstart0
         0     0% 134.61%     -513kB 16.68%  runtime.mstart1
         0     0% 134.61%      513kB 16.68%  runtime.park_m
         0     0% 134.61%   528.17kB 17.17%  sync.(*Pool).Get
```
Данный дифф демонстрирует оптимизацию sign middleware, в рамках которой было добавлено повторное использование буферов запросов и ответов. Данная оптимизация позволила избежать выделения ресурсов для каждого запроса, сохраняя при этом проверку подписи и подписывание ответов.

- Бенчмарк: `BenchmarkGinHandlerUpdatesJSONNetwork` (пакет `internal/handler`).
- Результат: `$(PROFILE_DIR)/network.pprof`, журнал нагрузки `$(PROFILE_DIR)/network_hey.txt` и лог сервера `$(PROFILE_DIR)/network_server.log` (создаются, если не установлен `SKIP_HEY=1`).
- Цель автоматически поднимает `cmd/server` на адресе из `HEY_URL` (по умолчанию `http://localhost:8080`), ждёт готовности эндпоинта `/`, затем отправляет JSON батч метрик из `HEY_PAYLOAD` (по умолчанию `testdata/network_batch.json`) через `hey`.
- При ошибке подключения `hey` выводит подробности и протокол сервера, а временные файлы удаляются. Чтобы снять только бенчмарк без запуска сервера и нагрузки, установите `SKIP_HEY=1`.

### Параметры запуска

Каждая цель поддерживает следующие переменные окружения:

- `PROFILE_DIR` — путь к директории, в которую сохраняются профили (по умолчанию `profiles`).
- `PROFILE_BENCH_COUNT` — число запусков бенчмарка для усреднения (по умолчанию `1`).
- `NETWORK_BENCHTIME`, `COLLECTOR_BENCHTIME`, `STORAGE_BENCHTIME` — длительность соответствующего бенчмарка.
- `HEY_URL`, `HEY_PAYLOAD`, `HEY_REQUESTS`, `HEY_CONCURRENCY` — параметры нагрузки утилиты hey в цели `profile-network`.
- `SKIP_HEY=1` — пропустить сетевую нагрузку через hey и сохранить только профили бенчмарка.

Файл `testdata/network_batch.json` содержит готовый набор метрик в формате JSON.