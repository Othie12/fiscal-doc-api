package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username  string `json:"username"`
	UserLevel string `json:"user_level" gorm:"default:normal"` // normal / admin

	Password string `json:"password"`
}

func (u *User) RemoveSensitiveData() {
	u.Password = ""
}

// CREATE TABLE `qrcodes` (
//   `id` int(10) UNSIGNED NOT NULL,
//   `code` varchar(255) NOT NULL,
//   `manufacturer` varchar(255) NOT NULL,
//   `customer` varchar(255) NOT NULL,
//   `product` varchar(255) NOT NULL,
//   `time` timestamp NOT NULL DEFAULT current_timestamp(),
//   `fdn` varchar(30) DEFAULT NULL,
//   `cust_tin` varchar(30) DEFAULT NULL,
//   `mfr_tin` varchar(30) DEFAULT NULL
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

type Qrcode struct {
	Id           uint      `json:"id" gorm:"primarykey"`
	Code         string    `json:"code" gorm:"not null"`
	Manufacturer string    `json:"manufacturer" gorm:"not null"`
	Customer     string    `json:"customer" gorm:"not null"`
	Product      string    `json:"product" gorm:"not null"`
	Time         time.Time `json:"time" gorm:"not null"`
	Fdn          *string   `json:"fdn"`
	CustTin      *string   `json:"cust_tin"`
	MfrTin       *string   `json:"mfr_tin"`
}

type ScanLog struct {
	gorm.Model
	Fdn         string `json:"fdn" gorm:"unique"`
	ScannedById uint   `json:"scanned_by_id"`

	ScannedBy *User `json:"scanned_by" gorm:"foreignKey:ScannedById;references:ID;constraint:OnDelete:SET NULL;"`
}

// ////////////////////////// Custom Date component ////////////////////////////
type Date struct {
	time.Time
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format("2006-01-02"))
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

func (d Date) Value() (driver.Value, error) {
	return d.Time.Format("2006-01-02"), nil
}

func (d *Date) Scan(value interface{}) error {
	if value == nil {
		return errors.New("can't scan nil into Date")
	}
	t, err := time.Parse("2006-01-02", value.(string))
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}
