# Документация по архитектуре Middleware

## Обзор

**Unified Interceptor** дает управление методами с возможностями трассировки, сбора метрик и профилирования. Система построена на основе механизма рефлексии Go, построено на основе оберток.

## Основные компоненты

### 1. Interceptor

Главный оркестратор, который управляет всей функциональностью middleware:

```go
type Interceptor struct {
    tracer      trace.Tracer
    middlewares []Middleware
    config      *InterceptorConfig
}
```

**Ключевые возможности:**
- **Интеграция с OpenTelemetry**: трассировка
- **Сбор метрик**: отслеживание длительности запросов и операций
- **Профилирование производительности**: CPU профилирование с pprof метками
- **SQL профилирование**: инструментирование запросов к базе данных
- **Конфигурируемость**: включение/отключение функций по окружению

### 2. Система конфигурации

```go
type InterceptorConfig struct {
    EnableTracing      bool
    EnableMetrics      bool
    EnableProfiling    bool
    EnableSQLProfiling bool
    ServiceName        string
    ProfileSQLQueries  bool
    ProfileThreshold   time.Duration
}
```

Конфигурация по умолчанию включает все функции с порогом профилирования 50мс.

## Архитектурные паттерны

### 1. Перехват методов на основе рефлексии

Система использует пакет `reflect` Go для динамической обёртки функций:

```go
func (ui *Interceptor) InterceptFunc(fn interface{}) interface{} {
    fnValue := reflect.ValueOf(fn)
    fnType := reflect.TypeOf(fn)
    
    return reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
        // Извлечение контекста из аргументов
        ctx := ui.extractContext(args)
        
        // Создание обёртки обработчика
        handler := func(ctx context.Context, input interface{}) (interface{}, error) {
            // Вызов оригинальной функции
            results := fnValue.Call(args)
            // Обработка результатов...
        }
        
        // Применение цепочки middleware
        wrappedHandler := ui.Chain(handler, operationName)
        result, err := wrappedHandler(ctx, args)
        
        // Преобразование обратно к оригинальным типам возврата
        return returnValues
    }).Interface()
}
```

**Как это работает:**
1. Принимает любую функцию как `interface{}`
2. Использует рефлексию для анализа сигнатуры функции
3. Создаёт новую функцию с идентичной сигнатурой
4. Обёртывает оригинальную функцию цепочкой middleware
5. Возвращает обёрнутую функцию

### 2. Паттерн Service Wrapper

Для полного инструментирования сервиса:

```go
type ServiceWrapper struct {
    service     interface{}
    interceptor *Interceptor
    serviceName string
    methodCache map[string]reflect.Value
}
```

**Процесс:**
1. **Анализ сервиса**: используется рефлексия для обнаружения всех методов
2. **Кеширование методов**: предварительно обёртываются все методы для производительности
3. **Динамическая диспетчеризация**: направляет вызовы методов через обёрнутые версии

```go
func (sw *ServiceWrapper) wrapAllMethods() {
    serviceType := reflect.TypeOf(sw.service)
    
    // Проходим по всем методам сервиса
    for i := 0; i < serviceType.NumMethod(); i++ {
        method := serviceType.Method(i)
        methodName := fmt.Sprintf("%s.%s", sw.serviceName, method.Name)
        
        // Получаем оригинальный метод
        originalMethod := serviceValue.Method(i)
        // Обёртываем его middleware
        wrappedMethod := sw.interceptor.InterceptFunc(originalMethod.Interface(), methodName)
        // КЕШИРОВАНИЕ: сохраняем обёрнутый метод в кеше для быстрого доступа
        sw.methodCache[method.Name] = reflect.ValueOf(wrappedMethod)
    }
}
```

### 3. Паттерн цепочки Middleware


```go
type Middleware func(next Handler) Handler
type Handler func(ctx context.Context, input interface{}) (interface{}, error)
```

**Выполнение цепочки:**
```go
func (ui *Interceptor) Chain(handler Handler, operationName string) Handler {
    ctx = context.WithValue(ctx, "operation_name", operationName)
    
    currentHandler := handler
    for _, middleware := range ui.middlewares {
        currentHandler = middleware(currentHandler)
    }
    
    return currentHandler
}
```

## Компоненты Middleware

### 1. Middleware трассировки

**Назначение**: трассировка OpenTelemetry
**Реализация**: `/internal/middleware/unified_interceptor.go:245-266`

```go
func (ui *Interceptor) tracingMiddleware(next Handler) Handler {
    return func(ctx context.Context, input interface{}) (interface{}, error) {
        opName := ui.extractOperationName(ctx)
        
        ctx, span := ui.tracer.Start(ctx, opName)
        defer span.End()
        
        // Извлечение атрибутов из входных данных (UserID, ChatID, Command)
        attrs := ui.extractAttributes(input)
        span.SetAttributes(attrs...)
        
        result, err := next(ctx, input)
        
        if err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, err.Error())
        }
        
        return result, err
    }
}
```

**Возможности:**
- Автоматическое создание span с именами операций
- Извлечение атрибутов контекста (user.id, chat.id, command)
- Запись ошибок и установка статуса
- Сохранение иерархии span

### 2. Middleware метрик

**Назначение**: Сбор метрик производительности
**Реализация**: `/internal/middleware/unified_interceptor.go:268-281`

```go
func (ui *Interceptor) metricsMiddleware(next Handler) Handler {
    return func(ctx context.Context, input interface{}) (interface{}, error) {
        start := time.Now()
        
        result, err := next(ctx, input)
        
        duration := time.Since(start)
        opName := ui.extractOperationName(ctx)
        
        metrics.RecordRequest(opName, duration)
        
        return result, err
    }
}
```

**Отслеживает:**
- Длительность запросов по операциям
- Частоту операций
- Тренды производительности

### 3. Middleware профилирования

**Назначение**: CPU профилирование с контекстом операций
**Реализация**: `/internal/middleware/unified_interceptor.go:283-312`

```go
func (ui *Interceptor) profilingMiddleware(next Handler) Handler {
    return func(ctx context.Context, input interface{}) (interface{}, error) {
        opName := ui.extractOperationName(ctx)
        
        // Создание меток профилирования
        labels := pprof.Labels("operation", opName)
        ctx = pprof.WithLabels(ctx, labels)
        
        var result interface{}
        var err error
        
        // Выполнение с контекстом профилирования
        pprof.Do(ctx, labels, func(labeledCtx context.Context) {
            result, err = next(labeledCtx, input)
        })
        
        // Отслеживание медленных операций
        if time.Since(start) > ui.config.ProfileThreshold {
            // Дополнительное профилирование для медленных операций
        }
        
        return result, err
    }
}
```

**Возможности:**
- Контекстные метки профилирования
- Обнаружение медленных операций
- CPU профили для конкретных операций

## Паттерны использования

### 1. Обёртка обработчиков бота

В обработчиках Telegram бота:

```go
func (h *BrokerageAccountCreationHandler) HandleStep(ctx context.Context, bot *Bot, session *UserSession, msg *Message) error {
    handler := func(ctx context.Context, input interface{}) (interface{}, error) {
        return nil, h.processStep(ctx, bot, session, msg)
    }
    
    params := map[string]interface{}{
        "user_id":      int64(session.UserID),
        "session_step": string(session.CurrentStep),
        "session_type": "create_brokerage_account",
    }
    
    wrappedHandler := bot.interceptor.Chain(handler, "BrokerageAccountCreationHandler.HandleStep")
    _, err := wrappedHandler(ctx, params)
    return err
}
```

### 2. Перехват методов сервиса

```go
// Обёртка всего сервиса
wrappedService := interceptor.InterceptService(financeService, "FinanceService")

// Обёртка отдельных методов
wrappedMethod := interceptor.InterceptFunc(service.CreateDeposit, "CreateDeposit")
```

### 3. Перехват базы данных

```go
// Обёртка соединения с БД для SQL профилирования
wrappedDB := interceptor.InterceptSQL(db)
```

## Распространение контекста

Система сохраняет и улучшает Go контексты по всей цепочке вызовов:

1. **Извлечение контекста**: Автоматически находит `context.Context` в аргументах функции
2. **Улучшение контекста**: Добавляет имена операций и метки профилирования
3. **Распространение контекста**: Передаёт улучшенный контекст через цепочку middleware
4. **Извлечение атрибутов**: Извлекает релевантные данные для трассировки (UserID, ChatID, Command)

## Продвинутые возможности

### 1. Автоматическое извлечение атрибутов

```go
func (ui *Interceptor) extractAttributes(input interface{}) []attribute.KeyValue {
    // Использует рефлексию для поиска общих полей:
    // - UserID, UserId, ID
    // - ChatID, ChatId
    // - Command, Cmd
    
    if userIDField := ui.findField(value, "UserID", "UserId", "ID"); userIDField.IsValid() {
        attrs = append(attrs, attribute.Int64("user.id", userID))
    }
    // ... дополнительное извлечение атрибутов
}
```

### 2. SQL профилирование запросов

Интеграция с SQL профайлер middleware для отслеживания операций с базой данных.

### 3. Распространение ошибок

Сохраняет типы ошибок и контекст, добавляя информацию трассировки:

```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
}
```

------

1. **Кеширование методов**: Предварительно обёрнутые методы кешируются для избежания накладных расходов рефлексии
2. **Ленивая загрузка**: Цепочка middleware строится один раз на тип операции
3. **Конфигурируемые пороги**: Профилирование активируется только для медленных операций

Middleware интегрируется с:
- **OpenTelemetry**: для распределённой трассировки
- **Пользовательские метрики**: для мониторинга специфичного для приложения  
- **pprof**: для CPU профилирования
- **SQL Profiler**: для анализа запросов к базе данных

## Примеры использования

### Пример 1: Инициализация interceptor'а

```go
// Создание конфигурации
config := &InterceptorConfig{
    EnableTracing:      true,
    EnableMetrics:      true, 
    EnableProfiling:    true,
    EnableSQLProfiling: true,
    ServiceName:        "FiatFormaggio",
    ProfileThreshold:   50 * time.Millisecond,
}

// Создание interceptor'а
interceptor := NewInterceptor(config)
```

### Пример 2: Обёртка сервиса с кешированием методов

```go
// У нас есть финансовый сервис
type FinanceService struct {
    repo repository.Repository
}

func (fs *FinanceService) CreateDeposit(ctx context.Context, req *CreateDepositRequest) (*Deposit, error) {
    // логика создания депозита
    return deposit, nil
}

func (fs *FinanceService) GetDepositsByUserID(ctx context.Context, userID int64) ([]*Deposit, error) {
    // логика получения депозитов
    return deposits, nil
}

// Обёртывание всего сервиса - ВСЕ МЕТОДЫ КЕШИРУЮТСЯ АВТОМАТИЧЕСКИ
financeService := &FinanceService{repo: repo}
wrappedService := interceptor.InterceptService(financeService, "FinanceService")

// Теперь все вызовы методов проходят через middleware:
// - CreateDeposit -> трассировка + метрики + профилирование
// - GetDepositsByUserID -> трассировка + метрики + профилирование
```

### Пример 3: Использование кешированных методов

```go
// ServiceWrapper имеет кеш методов
type ServiceWrapper struct {
    service     interface{}
    interceptor *Interceptor  
    serviceName string
    methodCache map[string]reflect.Value // ВОТ ЗДЕСЬ КЕШИРОВАНИЕ!
}

// Вызов кешированного метода
func (sw *ServiceWrapper) CallMethod(methodName string, args ...interface{}) ([]interface{}, error) {
    // Ищем в кеше обёрнутый метод
    if wrappedMethod, exists := sw.methodCache[methodName]; exists {
        // Используем уже обёрнутый метод из кеша - НЕТ НАКЛАДНЫХ РАСХОДОВ НА РЕФЛЕКСИЮ!
        argValues := make([]reflect.Value, len(args))
        for i, arg := range args {
            argValues[i] = reflect.ValueOf(arg)
        }
        
        // Вызываем кешированный обёрнутый метод
        results := wrappedMethod.Call(argValues)
        // ...обработка результатов
    }
    
    return nil, errors.NewTechnicalError(errors.CodeUnknownHandler, fmt.Sprintf("method %s not found", methodName))
}
```

### Пример 4: Обёртка отдельной функции

```go
// Оригинальная функция
func calculateTotalBalance(ctx context.Context, userID int64) (float64, error) {
    // логика расчёта баланса
    return balance, nil
}

// Обёртывание одной функции
wrappedCalculateBalance := interceptor.InterceptFunc(calculateTotalBalance, "CalculateTotalBalance")

// Теперь функция с middleware
balance, err := wrappedCalculateBalance.(func(context.Context, int64) (float64, error))(ctx, userID)
```

### Пример 5: Обёртка обработчика бота с трассировкой

```go
func (h *DepositCreationHandler) HandleStep(ctx context.Context, bot *Bot, session *UserSession, msg *Message) error {
    // Создаём handler-функцию для middleware
    handler := func(ctx context.Context, input interface{}) (interface{}, error) {
        return nil, h.processDepositStep(ctx, bot, session, msg)
    }
    
    // Параметры для трассировки - извлекутся автоматически в extractAttributes
    params := map[string]interface{}{
        "user_id":      int64(session.UserID),     // автоматически станет user.id в span
        "session_step": string(session.CurrentStep),
        "session_type": "create_deposit",
        "chat_id":      session.ChatID,            // автоматически станет chat.id в span
    }
    
    // Применяем цепочку middleware
    wrappedHandler := bot.interceptor.Chain(handler, "DepositCreationHandler.HandleStep")
    _, err := wrappedHandler(ctx, params)
    return err
}
```

### Пример 6: SQL профилирование

```go
// Оригинальное соединение с БД
db, err := sql.Open("postgres", connectionString)
if err != nil {
    return err
}

// Обёртывание базы данных для SQL профилирования
wrappedDB := interceptor.InterceptSQL(db)

// Теперь все SQL запросы будут профилироваться
rows, err := wrappedDB.(*sql.DB).QueryContext(ctx, "SELECT * FROM deposits WHERE user_id = $1", userID)
```

### Пример 7: Извлечение атрибутов для трассировки

```go
// Структура запроса
type CreateDepositRequest struct {
    UserID   int64  `json:"user_id"`   // автоматически извлечётся как user.id
    ChatID   int64  `json:"chat_id"`   // автоматически извлечётся как chat.id  
    Command  string `json:"command"`   // автоматически извлечётся как command
    Amount   float64 `json:"amount"`
    Currency string  `json:"currency"`
}

// Когда этот запрос передаётся в middleware, автоматически создаются атрибуты:
// - user.id = 12345
// - chat.id = 67890  
// - command = "create_deposit"
```

### Пример 8: Конфигурация для разных сред

```go
// Среда разработки - всё включено
devConfig := &InterceptorConfig{
    EnableTracing:      true,
    EnableMetrics:      true,
    EnableProfiling:    true,
    EnableSQLProfiling: true,
    ProfileThreshold:   10 * time.Millisecond, // низкий порог
}

// Продакшн - профилирование отключено для производительности
prodConfig := &InterceptorConfig{
    EnableTracing:      true,
    EnableMetrics:      true, 
    EnableProfiling:    false, // ОТКЛЮЧЕНО в продакшне
    EnableSQLProfiling: true,
    ProfileThreshold:   100 * time.Millisecond, // высокий порог
}
```
