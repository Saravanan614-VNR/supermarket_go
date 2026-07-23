/*
 * Contract ID: CTR-006
 * Service Name: SupermarketService
 * Description: GORM & MySQL Database Connection Setup and Pool Initialization.
 *              Integrates custom logger routing SQL traces directly to Zap.
 */

package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormZapLogger redirects GORM's internal logging calls to our structured Zap logger
type GormZapLogger struct {
	ZapLogger     *zap.Logger
	SlowThreshold time.Duration
}

// LogMode implements gormlogger.Interface
func (l *GormZapLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	// Let Zap control active log levels dynamically
	return l
}

// Info implements gormlogger.Interface
func (l *GormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.ZapLogger.Sugar().Infof(msg, data...)
}

// Warn implements gormlogger.Interface
func (l *GormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.ZapLogger.Sugar().Warnf(msg, data...)
}

// Error implements gormlogger.Interface
func (l *GormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.ZapLogger.Sugar().Errorf(msg, data...)
}

// Trace implements gormlogger.Interface for printing SQL statements, durations, and rows affected.
func (l *GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		l.ZapLogger.Error("Database Query Failure",
			zap.Error(err),
			zap.Duration("elapsed_ms", elapsed),
			zap.Int64("rows_affected", rows),
			zap.String("sql", sql),
		)
		return
	}

	if elapsed > l.SlowThreshold {
		l.ZapLogger.Warn("Database Slow Query Warning",
			zap.Duration("elapsed_ms", elapsed),
			zap.Int64("rows_affected", rows),
			zap.String("sql", sql),
		)
		return
	}

	l.ZapLogger.Debug("Database SQL Trace",
		zap.Duration("elapsed_ms", elapsed),
		zap.Int64("rows_affected", rows),
		zap.String("sql", sql),
	)
}

/*
 * ⚠ DATABASE — Per-service Schema Rule:
 * We connect to the "supermarket_supermarket" database schema mapped for SupermarketService.
 * Tables are pre-created by migrations/V1__initial_schema.sql (or init.sql).
 * Do NOT use GORM's db.AutoMigrate() in production environments to avoid locks and drift.
 * Ensure tables are created prior to booting the app.
 */

// NewDatabase initializes a production-ready GORM MySQL connection pool.
func NewDatabase(cfg *Config, zapLog *zap.Logger) (*gorm.DB, error) {
	// Custom logger threshold of 200ms to flag slow queries
	gormLog := &GormZapLogger{
		ZapLogger:     zapLog,
		SlowThreshold: 200 * time.Millisecond,
	}

	zapLog.Info("Connecting to MySQL Database...", zap.String("dsn", maskDSN(cfg.MYSQL_DSN)))

	// Open GORM MySQL Connection
	db, err := gorm.Open(mysql.Open(cfg.MYSQL_DSN), &gorm.Config{
		Logger: gormLog,
		// Disable default transactions for improved performance on simple writes
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Retrieve underlying sql.DB instance to configure connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve underlying sql.DB: %w", err)
	}

	// Configure DB Connection Pool (Production-ready settings)
	sqlDB.SetMaxOpenConns(cfg.DB_MAX_OPEN_CONNS)
	sqlDB.SetMaxIdleConns(cfg.DB_MAX_IDLE_CONNS)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DB_CONN_MAX_LIFETIME_MINS) * time.Minute)

	// Validate connectivity with immediate Ping
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	zapLog.Info("MySQL Database connection pool initialized successfully",
		zap.Int("max_open_conns", cfg.DB_MAX_OPEN_CONNS),
		zap.Int("max_idle_conns", cfg.DB_MAX_IDLE_CONNS),
		zap.Duration("conn_max_lifetime", time.Duration(cfg.DB_CONN_MAX_LIFETIME_MINS)*time.Minute),
	)

	return db, nil
}

// maskDSN hides credentials in DSN strings for security during logs
func maskDSN(dsn string) string {
	if dsn == "" {
		return ""
	}
	parts := strings.Split(dsn, "@")
	if len(parts) < 2 {
		return dsn // Not a standard DSN format, return as is
	}
	userPassParts := strings.Split(parts[0], ":")
	if len(userPassParts) == 2 {
		return fmt.Sprintf("%s:****@%s", userPassParts[0], parts[1])
	}
	return fmt.Sprintf("****@%s", parts[1])
}