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

// var MySQLDB *gorm.DB
var DB *gorm.DB

func mysqlConnect() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.MysqlConfig.Username, config.MysqlConfig.Password, config.MysqlConfig.Host,
		config.MysqlConfig.Port, config.MysqlConfig.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return err
	}
	DB = db
	return nil
}

// connect to oracle db
func oracleConnect() error {
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

	DB = db
	return nil
}

func Connect() error {
	dbToUse := "oracle" // oracle | mysql
	if dbToUse == "oracle" {
		return oracleConnect()
	} else {
		return mysqlConnect()
	}
}

func Migrate() error {
	err := DB.AutoMigrate(
		&models.User{},
		&models.ScanLog{},
		&models.FailedScanLog{},
		// &models.Qrcode{},
	)

	return err
}

func FindOffset(page, limit int) int {
	if page < 1 {
		return 0
	}
	return (page - 1) * limit
}
