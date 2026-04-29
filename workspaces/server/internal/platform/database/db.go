package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/NishLy/go-fiber-boilerplate/config"
	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	appLogger "github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbMap = make(map[string]*DBStruct)
var mutex sync.Mutex

type DBStruct struct {
	LastUsed int64
	DB       *gorm.DB
}

func GetIndentifier(ctx context.Context) string {
	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok || tenantID == "" {
		return "default"
	}

	return tenantID
}

// GetDB returns a tenant-specific database connection. It creates a new connection if one does not already exist for the given identifier.
func GetDB(identifier string, withLogger bool) (DBStruct, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if dbStruct, ok := dbMap[identifier]; ok {
		dbStruct.LastUsed = time.Now().Unix()
		return *dbStruct, nil
	}

	config := config.Get()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable&TimeZone=Asia/Jakarta",
		config.DBUSER, config.DBPASS, config.DBHOST, config.DBPORT, config.DBNAME,
	)

	dbConfig := &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		TranslateError:         true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	}

	if withLogger {
		dbConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), dbConfig)

	if err != nil {
		return DBStruct{}, err
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)

	dbMap[identifier] = &DBStruct{
		DB:       db,
		LastUsed: time.Now().Unix(),
	}
	return *dbMap[identifier], nil
}

func CloseDB(identifier string) error {
	mutex.Lock()
	defer mutex.Unlock()
	if dbStruct, ok := dbMap[identifier]; ok {
		sqlDB, err := dbStruct.DB.DB()
		if err != nil {
			return err
		}
		if err := sqlDB.Close(); err != nil {
			return err
		}
		delete(dbMap, identifier)
	}
	return nil
}

func CleanupDBs(maxIdleTime time.Duration) {
	mutex.Lock()
	defer mutex.Unlock()
	currentTime := time.Now().Unix()
	for identifier, dbStruct := range dbMap {
		if currentTime-dbStruct.LastUsed > int64(maxIdleTime.Seconds()) {
			err := CloseDB(identifier)
			if err != nil {
				appLogger.Sugar.Errorf("Failed to close DB for identifier %s: %v", identifier, err)
			}
		}
	}
}

func GetDBFromContext(ctx context.Context) (*gorm.DB, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return nil, apperror.InternalErr(
			fmt.Errorf("no database connection found in context"),
		)
	}
	return db, nil
}

func SetContextWithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, "db", tx)
}
