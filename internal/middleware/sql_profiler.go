package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type SQLProfiler struct {
	tracer           trace.Tracer
	enableProfiling  bool
	enableTracing    bool
	profileThreshold time.Duration
}

func NewSQLProfiler(enableProfiling, enableTracing bool, threshold time.Duration) *SQLProfiler {
	return &SQLProfiler{
		tracer:           otel.Tracer(domain.AppName + "-sql"),
		enableProfiling:  enableProfiling,
		enableTracing:    enableTracing,
		profileThreshold: threshold,
	}
}

type ProfiledDB struct {
	*sql.DB
	profiler *SQLProfiler
}

type ProfiledTx struct {
	*sql.Tx
	profiler *SQLProfiler
}

type ProfiledStmt struct {
	*sql.Stmt
	profiler *SQLProfiler
	query    string
}

func (sp *SQLProfiler) WrapDB(db *sql.DB) *ProfiledDB {
	return &ProfiledDB{
		DB:       db,
		profiler: sp,
	}
}

func (pdb *ProfiledDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return pdb.QueryContext(context.Background(), query, args...)
}

func (pdb *ProfiledDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return pdb.profiler.profileQuery(ctx, "Query", query, func() (*sql.Rows, error) {
		return pdb.DB.QueryContext(ctx, query, args...)
	})
}

func (pdb *ProfiledDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return pdb.QueryRowContext(context.Background(), query, args...)
}

func (pdb *ProfiledDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	row, _ := pdb.profiler.profileQueryRow(ctx, "QueryRow", query, func() *sql.Row {
		return pdb.DB.QueryRowContext(ctx, query, args...)
	})
	return row
}

func (pdb *ProfiledDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return pdb.ExecContext(context.Background(), query, args...)
}

func (pdb *ProfiledDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return pdb.profiler.profileExec(ctx, "Exec", query, func() (sql.Result, error) {
		return pdb.DB.ExecContext(ctx, query, args...)
	})
}

func (pdb *ProfiledDB) Prepare(query string) (*ProfiledStmt, error) {
	return pdb.PrepareContext(context.Background(), query)
}

func (pdb *ProfiledDB) PrepareContext(ctx context.Context, query string) (*ProfiledStmt, error) {
	stmt, err := pdb.profiler.profilePrepare(ctx, "Prepare", query, func() (*sql.Stmt, error) {
		return pdb.DB.PrepareContext(ctx, query)
	})

	if err != nil {
		return nil, err
	}

	return &ProfiledStmt{
		Stmt:     stmt,
		profiler: pdb.profiler,
		query:    query,
	}, nil
}

func (pdb *ProfiledDB) Begin() (*ProfiledTx, error) {
	return pdb.BeginTx(context.Background(), nil)
}

func (pdb *ProfiledDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*ProfiledTx, error) {
	tx, err := pdb.profiler.profileTx(ctx, "BeginTx", func() (*sql.Tx, error) {
		return pdb.DB.BeginTx(ctx, opts)
	})

	if err != nil {
		return nil, err
	}

	return &ProfiledTx{
		Tx:       tx,
		profiler: pdb.profiler,
	}, nil
}

func (ptx *ProfiledTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return ptx.QueryContext(context.Background(), query, args...)
}

func (ptx *ProfiledTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return ptx.profiler.profileQuery(ctx, "Tx.Query", query, func() (*sql.Rows, error) {
		return ptx.Tx.QueryContext(ctx, query, args...)
	})
}

func (ptx *ProfiledTx) QueryRow(query string, args ...interface{}) *sql.Row {
	return ptx.QueryRowContext(context.Background(), query, args...)
}

func (ptx *ProfiledTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	row, _ := ptx.profiler.profileQueryRow(ctx, "Tx.QueryRow", query, func() *sql.Row {
		return ptx.Tx.QueryRowContext(ctx, query, args...)
	})
	return row
}

func (ptx *ProfiledTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return ptx.ExecContext(context.Background(), query, args...)
}

func (ptx *ProfiledTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return ptx.profiler.profileExec(ctx, "Tx.Exec", query, func() (sql.Result, error) {
		return ptx.Tx.ExecContext(ctx, query, args...)
	})
}

func (ptx *ProfiledTx) Prepare(query string) (*ProfiledStmt, error) {
	return ptx.PrepareContext(context.Background(), query)
}

func (ptx *ProfiledTx) PrepareContext(ctx context.Context, query string) (*ProfiledStmt, error) {
	stmt, err := ptx.profiler.profilePrepare(ctx, "Tx.Prepare", query, func() (*sql.Stmt, error) {
		return ptx.Tx.PrepareContext(ctx, query)
	})

	if err != nil {
		return nil, err
	}

	return &ProfiledStmt{
		Stmt:     stmt,
		profiler: ptx.profiler,
		query:    query,
	}, nil
}

func (ps *ProfiledStmt) Query(args ...interface{}) (*sql.Rows, error) {
	return ps.QueryContext(context.Background(), args...)
}

func (ps *ProfiledStmt) QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	return ps.profiler.profileQuery(ctx, "Stmt.Query", ps.query, func() (*sql.Rows, error) {
		return ps.Stmt.QueryContext(ctx, args...)
	})
}

func (ps *ProfiledStmt) QueryRow(args ...interface{}) *sql.Row {
	return ps.QueryRowContext(context.Background(), args...)
}

func (ps *ProfiledStmt) QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row {
	row, _ := ps.profiler.profileQueryRow(ctx, "Stmt.QueryRow", ps.query, func() *sql.Row {
		return ps.Stmt.QueryRowContext(ctx, args...)
	})
	return row
}

func (ps *ProfiledStmt) Exec(args ...interface{}) (sql.Result, error) {
	return ps.ExecContext(context.Background(), args...)
}

func (ps *ProfiledStmt) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	return ps.profiler.profileExec(ctx, "Stmt.Exec", ps.query, func() (sql.Result, error) {
		return ps.Stmt.ExecContext(ctx, args...)
	})
}

func (sp *SQLProfiler) profileQuery(ctx context.Context, operation, query string, fn func() (*sql.Rows, error)) (*sql.Rows, error) {
	start := time.Now()
	ctx, span := sp.startSpan(ctx, operation, query)
	defer span.End()

	var labels pprof.LabelSet
	if sp.enableProfiling {
		labels = pprof.Labels(
			"operation", operation,
			"query_type", sp.getQueryType(query),
			"table", sp.extractTableName(query),
		)
		ctx = pprof.WithLabels(ctx, labels)
	}

	var result *sql.Rows
	var err error

	if sp.enableProfiling {
		pprof.Do(ctx, labels, func(ctx context.Context) {
			result, err = fn()
		})
	} else {
		result, err = fn()
	}

	sp.recordMetrics(operation, query, time.Since(start), err)
	sp.finishSpan(span, err)

	return result, err
}

func (sp *SQLProfiler) profileQueryRow(ctx context.Context, operation, query string, fn func() *sql.Row) (*sql.Row, error) {
	start := time.Now()
	ctx, span := sp.startSpan(ctx, operation, query)
	defer span.End()

	var labels pprof.LabelSet
	if sp.enableProfiling {
		labels = pprof.Labels(
			"operation", operation,
			"query_type", sp.getQueryType(query),
			"table", sp.extractTableName(query),
		)
		ctx = pprof.WithLabels(ctx, labels)
	}

	var result *sql.Row

	if sp.enableProfiling {
		pprof.Do(ctx, labels, func(ctx context.Context) {
			result = fn()
		})
	} else {
		result = fn()
	}

	sp.recordMetrics(operation, query, time.Since(start), nil)
	sp.finishSpan(span, nil)

	return result, nil
}

func (sp *SQLProfiler) profileExec(ctx context.Context, operation, query string, fn func() (sql.Result, error)) (sql.Result, error) {
	start := time.Now()
	ctx, span := sp.startSpan(ctx, operation, query)
	defer span.End()

	var labels pprof.LabelSet
	if sp.enableProfiling {
		labels = pprof.Labels(
			"operation", operation,
			"query_type", sp.getQueryType(query),
			"table", sp.extractTableName(query),
		)
		ctx = pprof.WithLabels(ctx, labels)
	}

	var result sql.Result
	var err error

	if sp.enableProfiling {
		pprof.Do(ctx, labels, func(ctx context.Context) {
			result, err = fn()
		})
	} else {
		result, err = fn()
	}

	sp.recordMetrics(operation, query, time.Since(start), err)
	sp.finishSpan(span, err)

	return result, err
}

func (sp *SQLProfiler) profilePrepare(ctx context.Context, operation, query string, fn func() (*sql.Stmt, error)) (*sql.Stmt, error) {
	start := time.Now()
	ctx, span := sp.startSpan(ctx, operation, query)
	defer span.End()

	result, err := fn()

	sp.recordMetrics(operation, query, time.Since(start), err)
	sp.finishSpan(span, err)

	return result, err
}

func (sp *SQLProfiler) profileTx(ctx context.Context, operation string, fn func() (*sql.Tx, error)) (*sql.Tx, error) {
	start := time.Now()
	ctx, span := sp.startSpan(ctx, operation, "")
	defer span.End()

	result, err := fn()

	sp.recordMetrics(operation, "", time.Since(start), err)
	sp.finishSpan(span, err)

	return result, err
}

func (sp *SQLProfiler) startSpan(ctx context.Context, operation, query string) (context.Context, trace.Span) {
	if !sp.enableTracing {
		return ctx, trace.SpanFromContext(ctx)
	}

	spanName := fmt.Sprintf("sql.%s", operation)
	ctx, span := sp.tracer.Start(ctx, spanName)

	if query != "" {
		span.SetAttributes(
			attribute.String("db.statement", sp.sanitizeQuery(query)),
			attribute.String("db.operation", sp.getQueryType(query)),
			attribute.String("db.sql.table", sp.extractTableName(query)),
		)
	}

	return ctx, span
}

func (sp *SQLProfiler) finishSpan(span trace.Span, err error) {
	if !sp.enableTracing {
		return
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func (sp *SQLProfiler) recordMetrics(operation, query string, duration time.Duration, err error) {
	labels := []string{operation}
	if query != "" {
		labels = append(labels, sp.getQueryType(query))
	}

	metrics.RecordRequest(strings.Join(labels, "."), duration)
}

func (sp *SQLProfiler) getQueryType(query string) string {
	query = strings.TrimSpace(strings.ToUpper(query))

	if strings.HasPrefix(query, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(query, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(query, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(query, "DELETE") {
		return "DELETE"
	} else if strings.HasPrefix(query, "CREATE") {
		return "CREATE"
	} else if strings.HasPrefix(query, "DROP") {
		return "DROP"
	} else if strings.HasPrefix(query, "ALTER") {
		return "ALTER"
	}

	return "OTHER"
}

func (sp *SQLProfiler) extractTableName(query string) string {
	query = strings.TrimSpace(strings.ToUpper(query))

	words := strings.Fields(query)

	for i, word := range words {
		if (word == "FROM" || word == "INTO" || word == "UPDATE" || word == "TABLE") && i+1 < len(words) {
			tableName := words[i+1]
			tableName = strings.TrimSuffix(tableName, ",")
			tableName = strings.TrimSuffix(tableName, "(")
			return strings.ToLower(tableName)
		}
	}

	return "unknown"
}

func (sp *SQLProfiler) sanitizeQuery(query string) string {
	if len(query) > 1000 {
		return query[:1000] + "..."
	}
	return query
}

func ProfileDatabase(db *sql.DB, enableProfiling, enableTracing bool, threshold time.Duration) *ProfiledDB {
	profiler := NewSQLProfiler(enableProfiling, enableTracing, threshold)
	return profiler.WrapDB(db)
}
