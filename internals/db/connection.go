package database

import (
	"fmt"
	"log"
	"time"

	oracle "github.com/godoes/gorm-oracle"
	"github.com/othie12/scanner-api/config"
	"github.com/othie12/scanner-api/internals/db/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var MySQLDB *gorm.DB
var OracleDB *gorm.DB

func MysqlConnect() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.MysqlConfig.Username, config.MysqlConfig.Password, config.MysqlConfig.Host,
		config.MysqlConfig.Port, config.MysqlConfig.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return err
	}
	MySQLDB = db
	return nil
}

// connect to oracle db
func OracleConnect() error {
	// Connect to mysql for now to handle testing
	OracleDB = MySQLDB
	return nil
	///////////////////////////////////////////////////////////////////////////

	// dsn := "oracle://sms_user:secret123@192.168.1.20:1521/ORCLPDB1"
	dsn := fmt.Sprintf("oracle://%s:%s@%s:%s/%s",
		config.OracleConfig.Username, config.OracleConfig.Password, config.OracleConfig.Host,
		config.OracleConfig.Port, config.OracleConfig.DBName)

	db, err := gorm.Open(oracle.Open(dsn), &gorm.Config{})

	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return err
	}

	log.Println("Connected to Oracle")

	OracleDB = db
	return nil
}

func Migrate() error {
	err := MySQLDB.AutoMigrate(
		&models.User{},
		&models.ScanLog{},
	)

	return err
}

func FindOffset(page, limit int) int {
	if page < 1 {
		return 0
	}
	return (page - 1) * limit
}
