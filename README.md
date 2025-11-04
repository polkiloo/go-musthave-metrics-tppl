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
Type: alloc_space
Time: 2025-11-05 00:41:02 MSK
Showing nodes accounting for -1547.29kB, 25.09% of 6166kB total
Dropped 9 nodes (cum <= 30.83kB)
      flat  flat%   sum%        cum   cum%
-1032.02kB 16.74% 16.74% -1032.02kB 16.74%  io.init.func1
-1027.35kB 16.66% 33.40% -1027.35kB 16.66%  regexp/syntax.(*compiler).inst (inline)
  516.64kB  8.38% 25.02%   516.64kB  8.38%  runtime.procresize
 -514.63kB  8.35% 33.37%  -514.63kB  8.35%  regexp.mergeRuneSets.func2 (inline)
    -514kB  8.34% 41.70%     -514kB  8.34%  bufio.NewReaderSize
  512.05kB  8.30% 33.40%   512.05kB  8.30%  regexp/syntax.simplify1 (inline)
  512.02kB  8.30% 25.09%   512.02kB  8.30%  net.ipToSockaddr
         0     0% 25.09% -1032.02kB 16.74%  bufio.(*Writer).Flush
         0     0% 25.09% -1029.92kB 16.70%  github.com/go-playground/validator/v10.init
         0     0% 25.09% -1032.02kB 16.74%  io.Copy (inline)
         0     0% 25.09% -1032.02kB 16.74%  io.CopyN
         0     0% 25.09% -1032.02kB 16.74%  io.copyBuffer
         0     0% 25.09% -1032.02kB 16.74%  io.discard.ReadFrom
```

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