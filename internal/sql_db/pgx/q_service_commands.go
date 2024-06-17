package pgx

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strconv"
	"time"

	"github.com/dimiro1/health"
	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations/*
var fs embed.FS

// Migrator creates migration of postgres database.
// Commands: up/down, number/-number steps (1/-1 steps).
func (p *pgx) Migrator(_ context.Context, command string) (step int, err error) {
	const (
		int999Key    = 999
		upKey        = "up"
		downKey      = "down"
		migrationKey = "migrations"
		dialectKey   = "postgres"
	)

	var steps int
	switch command {
	case upKey:
		steps = int999Key
	case downKey:
		steps = -int999Key
	default:
		steps, err = strconv.Atoi(command)
		if err != nil {
			const msg = "failed to convert migrate argument to digit: %w"
			return 0, fmt.Errorf(msg, err)
		}
	}

	if steps == 0 {
		return 0, nil
	}

	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: fs,
		Root:       migrationKey,
	}

	var direction migrate.MigrationDirection

	if steps > 0 {
		direction = migrate.Up
	} else if steps < 0 {
		direction = migrate.Down
		steps = -steps
	}

	step, err = migrate.ExecMax(p.db, dialectKey, migrations, direction, steps)
	if err != nil {
		const msg = "failed to execute migrations: %w"
		return 0, fmt.Errorf(msg, err)
	}

	if direction == migrate.Down {
		step = -step
	}

	return step, nil
}

// stats returns database statistics.
func (p *pgx) stats() (stats sql.DBStats) {
	return p.db.Stats()
}

// Check checks the availability of postgres instance.
func (p *pgx) Check(ctx context.Context) (res health.Health) {
	var err error

	const (
		idleKey               = "idle"
		inUseKey              = "in_use"
		maxIdleClosedKey      = "max_idle_closed"
		maxIdleTimeClosedKey  = "max_idle_time_closed"
		maxLifeTimeClosedKey  = "max_life_time_closed"
		maxOpenConnectionsKey = "max_open_connections"
		openConnectionsKey    = "open_connections"
		waitCountKey          = "wait_count"
		waitDurationKey       = "wait_duration"
		sqlQuerySelect1       = "SELECT 1"
		durationKey           = "duration"
		errorKey              = "error"
	)

	ctxHealth, cancel := context.WithTimeout(ctx, p.commitTimeout)
	defer func() {
		if err == nil {
			stats := p.stats()
			res.AddInfo(idleKey, stats.Idle)
			res.AddInfo(inUseKey, stats.InUse)
			res.AddInfo(maxIdleClosedKey, stats.MaxIdleClosed)
			res.AddInfo(maxIdleTimeClosedKey, stats.MaxIdleTimeClosed)
			res.AddInfo(maxLifeTimeClosedKey, stats.MaxLifetimeClosed)
			res.AddInfo(maxOpenConnectionsKey, stats.MaxOpenConnections)
			res.AddInfo(openConnectionsKey, stats.OpenConnections)
			res.AddInfo(waitCountKey, stats.WaitCount)
			res.AddInfo(waitDurationKey, stats.WaitDuration)
			res.Up()
		}

		if cancel != nil {
			cancel()
		}
	}()

	start := time.Now()
	_, err = p.db.ExecContext(ctxHealth, sqlQuerySelect1)
	res.AddInfo(durationKey, time.Since(start).String())
	if err != nil {
		res.AddInfo(errorKey, err.Error())
		res.Down()
		return res
	}

	return res
}
