package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	"go.uber.org/zap"
)

func connectDatabase(sqlConn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", sqlConn)
	if err != nil {
		return nil, err
	}
	isConnected := false
	for i := 0; i < 12; i++ {
		err := db.Ping()
		if err == nil {
			logger.Info("Connected to database")
			isConnected = true
			break
		}
		logger.Error("failed connecting to database", zap.Error(err), zap.String("connection string", sqlConn))
		logger.Info("Retrying in 5 seconds")
		time.Sleep(5 * time.Second)
	}
	if !isConnected {
		return nil, errors.New("unable to connect to db")
	}
	migrateDB(db)
	return db, nil
}

func migrateDB(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Error("failed creating postgres driver for migration", zap.Error(err))

		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file:///migrations",
		"postgres", driver)
	if err != nil {
		logger.Error("failed reading migration files", zap.Error(err))
		return err
	}
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		logger.Error("failed migration", zap.Error(err))
		return err
	}
	return nil
}
