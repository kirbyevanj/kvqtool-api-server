package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kirbyevanj/kvqtool-kvq-models/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func NewPostgresDB(dsn string, logger *slog.Logger) (*bun.DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	var err error
	for i := range 10 {
		if err = db.PingContext(context.Background()); err == nil {
			logger.Info("connected to postgres")
			return db, nil
		}
		logger.Warn("postgres not ready, retrying", "attempt", i+1, "err", err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("postgres connection failed after 10 attempts: %w", err)
}

func RunMigrations(ctx context.Context, db *bun.DB, logger *slog.Logger) error {
	tableModels := []interface{}{
		(*models.User)(nil),
		(*models.Project)(nil),
		(*models.VirtualFolder)(nil),
		(*models.Resource)(nil),
		(*models.WorkflowDefinition)(nil),
		(*models.Job)(nil),
	}

	for _, model := range tableModels {
		if _, err := db.NewCreateTable().Model(model).IfNotExists().WithForeignKeys().Exec(ctx); err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}

	if err := seedPlaceholderUser(ctx, db); err != nil {
		return fmt.Errorf("seed user: %w", err)
	}

	logger.Info("database migrations complete")
	return nil
}

var PlaceholderUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func seedPlaceholderUser(ctx context.Context, db *bun.DB) error {
	user := &models.User{
		ID:          PlaceholderUserID,
		Email:       "default@kvqtool.local",
		DisplayName: "Default User",
	}
	_, err := db.NewInsert().Model(user).On("CONFLICT (id) DO NOTHING").Exec(ctx)
	return err
}
