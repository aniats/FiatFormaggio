# Интерсептор

Допустим, изначально наша функция выглядит вот так:
```go
func GetDeposits(ctx context.Context, userID int) ([]Deposit, error) {
    return repository.GetDeposits(userID)
}
```

Интерсептор её оборачивает и получается что-то вроде:
```go
// Трейсинг -> Метрики -> Профилирование -> Сама функция
```

У нас есть interceptor.Chain(), который мы вызываем на самом верхнем уровне. По факту, у нас строится цепочка вида:
```go
// Исходная функция
  originalFunction := GetDeposits

  // После Chain получается:
  tracingMiddleware(
      metricsMiddleware(
          profilingMiddleware(
              originalFunction  // в самом конце наша функция
          )
      )
  )
```

Т.е. по шагам:
```go
func tracingMiddleware(next Handler) Handler {
      return func(ctx, input) {
          // Создание span
          result, err := next(ctx, input)  // Следующий уровень
  middleware
          // Окончание span
      }
  }
```

по факту next -> metricsMiddleware, а не сама наша функция

```go
func metricsMiddleware(next Handler) Handler {
      return func(ctx, input) {
          start := time.Now()
          result, err := next(ctx, input)  // Следующий уровень
  middleware
          recordMetrics(time.Since(start))
      }
  }
```

дальше profilingMiddleware

```go
func profilingMiddleware(next Handler) Handler {
      return func(ctx, input) {
          // Профайлер начинается
          result, err := next(ctx, input)  // Вот тут наконец-то наша функция
          // Профайлер кончается
      }
  }
```

### Конфигурация
```go
type InterceptorConfig struct {
    EnableTracing   bool   // Включить трейсинг
    EnableMetrics   bool   // Включить метрики  
    EnableProfiling bool   // Включить профилирование
    ServiceName     string // Имя сервиса
}
```

### Какие middleware есть сейчас

- Трейсинг - создаёт span'ы для OpenTelemetry: начинает span с именем операции, вытаскивает UserID, ChatID, Command из инпута, записывает ошибки, если они есть
- Метрики - считает время выполнения: засекает время начала, после выполнения записывает длительность
- Профилирование - добавляет метки для pprof, помечает операцию для профайлера

### Цепочка выполнения

Когда вызывается обернутая функция:

1. Профилирование добавляет метки
2. Метрики запускают таймер  
3. Трейсинг создаёт span
4. ЗДЕСЬ выполняется сама функция
5. Трейсинг заканчивает span (записывает ошибки если есть)
6. Метрики записывают время
7. Профилирование завершается


Сначала создается интерсептор на уровне сервисов, на этапе создания каждого из них
```go
interceptor := middleware.NewInterceptor(
    middleware.DefaultConfig("finance-bot"), 
    "finance-bot"
)
```

Каждая функция (handler) оборачивается в враппер, Chain - цепочка, т.е. как бы по цепочке принимаются и обрабатываются middleware
```go
wrappedHandler := interceptor.Chain(myHandler, "get_deposits")
```

```go
type Bot struct {
    interceptor *middleware.Interceptor
}

func NewBot(...) *Bot {
    interceptor := middleware.NewInterceptor(
        middleware.DefaultConfig("bot"), 
        "finance-bot"
    )
    
    return &Bot{
        interceptor: interceptor,
        // ...
    }
}
```

### Ошибки и пропагация до самого верха

Например, для такой цепочки вызовов:
```
Bot.sendDepositsCommand -> FinanceService.GetDeposits -> Repository.GetDeposits
```

Каждый метод обёрнут интерсептором, то при ошибке в Repository, т.е. в самом низу:

1. **Repository span** записывает ошибку SQL
2. **Service span** записывает ту же ошибку (она всплыла наверх)  
3. **Bot span** записывает ту же ошибку (она опять всплыла)

Таким способом видно, где ошибка возникла и через какие слои прошла. Внутри мы достаем и передаем эти поля:
```go
UserID  int64
ChatID  int64
Command string
```

