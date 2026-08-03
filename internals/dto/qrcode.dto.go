package dto

import (
	"time"

	"github.com/othie12/scanner-api/internals/db/models"
)

type QrcodeScanDTO struct {
	models.Qrcode
	ScannedAt *time.Time   `json:"scanned_at"`
	ScannedBy *models.User `json:"scanned_by"`
}
