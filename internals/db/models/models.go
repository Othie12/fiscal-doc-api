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

func (User) TableName() string {
	return "ST_USERS"
}

type Qrcode struct {
	Id              uint      `json:"id" gorm:"primarykey"`
	Code            string    `json:"code" gorm:"not null"`
	Manufacturer    string    `json:"manufacturer" gorm:"not null"`
	Customer        string    `json:"customer" gorm:"not null"`
	Product         string    `json:"product" gorm:"not null"`
	Time            time.Time `json:"time" gorm:"not null"`
	Fdn             *string   `json:"fdn"`
	CustTin         *string   `json:"cust_tin"`
	MfrTin          *string   `json:"mfr_tin"`
	VcVehicleNumber string    `json:"vc_vehicle_number"`
	VcDriverName    string    `json:"vc_driver_name"`
	VcPrdDet        string    `json:"vc_prd_det"`
	VcCustMbl       string    `json:"vc_cust_mbl"`
	VcInvoiceNo     string    `json:"vc_invoice_no"`
	DtInvoiceDate   string    `json:"dt_invoice_date"`
	VcScoulMbl      string    `json:"vc_scoul_mbl"`
}

func (Qrcode) TableName() string {
	return "ST_QRCODES"
}

type ScanLog struct {
	gorm.Model
	Fdn         string `json:"fdn" gorm:"unique"`
	ScannedById uint   `json:"scanned_by_id"`

	ScannedBy *User `json:"scanned_by" gorm:"foreignKey:ScannedById;references:ID;constraint:OnDelete:SET NULL;"`
}

func (ScanLog) TableName() string {
	return "ST_SCAN_LOGS"
}

type FailedScanLog struct {
	gorm.Model
	Fdn         string `json:"fdn"`
	ScannedById uint   `json:"scanned_by_id"`
	Reason      string `json:"reason"`

	ScannedBy *User `json:"scanned_by" gorm:"foreignKey:ScannedById;references:ID;constraint:OnDelete:SET NULL;"`
}

func (FailedScanLog) TableName() string {
	return "ST_FAILED_SCAN_LOGS"
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
