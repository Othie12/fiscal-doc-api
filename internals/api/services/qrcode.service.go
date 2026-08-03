package services

import (
	"fmt"
	"log"
	"net/http"

	database "github.com/othie12/scanner-api/internals/db"
	"github.com/othie12/scanner-api/internals/db/models"
	"github.com/othie12/scanner-api/internals/dto"
	"github.com/othie12/scanner-api/utils"
	"gorm.io/gorm"
)

type QrcodeService struct {
}

func NewQrcodeService() *QrcodeService {
	return &QrcodeService{}
}

func (s *QrcodeService) Scan(fdn string, scannerId any) (*models.Qrcode, int, error) {
	var rslt models.Qrcode
	if fdn == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("FDN is empty")
	}

	var user models.User
	if getUserErr := database.MySQLDB.First(&user, scannerId).Error; getUserErr != nil {
		if getUserErr != gorm.ErrRecordNotFound {
			log.Printf("Error Finding User with Id: %v: %s\n", scannerId, getUserErr.Error())
			return nil, http.StatusInternalServerError, fmt.Errorf("We encountered a Database Error, please try again")
		}

		return nil, http.StatusBadRequest, fmt.Errorf("User account Error. Make sure you are logged in")
	}

	var scannedLog models.ScanLog
	getLogErr := database.MySQLDB.Where(&models.ScanLog{Fdn: fdn}).First(&scannedLog).Error
	if getLogErr != nil && getLogErr != gorm.ErrRecordNotFound {
		log.Printf("Error scanning corresponding scanLog for fdn: %s: %s\n", fdn, getLogErr.Error())
		return nil, http.StatusInternalServerError, fmt.Errorf("We encountered a Database Error, please try again")
	}

	if getLogErr == nil {
		return nil, http.StatusConflict, fmt.Errorf("Duplicate Scan")
	}

	if err := database.OracleDB.Where("fdn = ?", fdn).First(&rslt).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Printf("Error finding QrCode with fdn: %s: %s\n", fdn, err.Error())
			return nil, http.StatusInternalServerError, fmt.Errorf("We encountered a Database Error, please try again")
		}

		return nil, http.StatusNotFound, fmt.Errorf("Record with This FDN doesn't exist")
	}

	scanLog := models.ScanLog{
		Fdn:         fdn,
		ScannedById: user.ID,
	}

	if saveLogErr := database.MySQLDB.Create(&scanLog).Error; saveLogErr != nil {
		log.Printf("Error saving scanLog for fdn '%s' by '%s': %s\n", fdn, user.Username, saveLogErr.Error())
		return nil, http.StatusInternalServerError, fmt.Errorf("We encountered a Database Error, please try again")
	}

	return &rslt, http.StatusOK, nil
}

func (s *QrcodeService) All(page int) ([]dto.QrcodeScanDTO, int, error) {
	limit := 10
	var items []models.Qrcode

	tx := database.OracleDB.Limit(limit).Offset(database.FindOffset(page, limit)).Order("id DESC").Find(&items)
	if tx.Error != nil {
		log.Printf("Failed to fetch items: %v\n", tx.Error)
		err := fmt.Errorf("An error occured while fetching items.")
		return nil, http.StatusInternalServerError, err
	}

	return s.MorphToDtoMultiple(items), http.StatusAccepted, nil
}

type FilterItemsDTO struct {
	DateFrom string `json:"date_from" validate:"required"`
	DateTo   string `json:"date_to" validate:"required"`
	Status   string `json:"status"`
}

func (s *QrcodeService) GetFiltered(dto FilterItemsDTO) ([]dto.QrcodeScanDTO, int, error) {
	var items []models.Qrcode
	var queryCondition interface{} = "time BETWEEN ? AND ?"
	queryArgs := []interface{}{utils.AddTimeToDate(dto.DateFrom, false), utils.AddTimeToDate(dto.DateTo, true)}

	if dto.Status == "scanned" || dto.Status == "unscanned" {
		var scannedFdns []string
		if err := database.MySQLDB.Model(&models.ScanLog{}).Order("id DESC").Where("created_at > ?", utils.AddTimeToDate(dto.DateFrom, false)).Pluck("fdn", &scannedFdns).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				log.Printf("Error scanning fdns to filter: %s\n", err.Error())
				return nil, http.StatusInternalServerError, fmt.Errorf("Database error occured")
			}
		}

		if dto.Status == "scanned" {
			queryCondition = fmt.Sprintf("%s %s", queryCondition, "AND fdn IN ?")
			queryArgs = append(queryArgs, scannedFdns)
		} else if len(scannedFdns) > 0 {
			queryCondition = fmt.Sprintf("%s %s", queryCondition, "AND fdn NOT IN ?")
			queryArgs = append(queryArgs, scannedFdns)
		}
	}

	tx := database.OracleDB.Where(queryCondition, queryArgs...).Find(&items)
	if tx.Error != nil {
		log.Printf("Failed to fetch items: %v\n", tx.Error)
		err := fmt.Errorf("An error occured while fetching items.")
		return nil, http.StatusInternalServerError, err
	}

	log.Println(queryCondition)
	log.Printf("%v\n", queryArgs...)

	return s.MorphToDtoMultiple(items), http.StatusAccepted, nil
}

func (s *QrcodeService) Search(query string) ([]dto.QrcodeScanDTO, int, error) {
	var items []models.Qrcode

	tx := database.OracleDB.
		Where("fdn LIKE ?", "%"+query+"%").
		Limit(10).Find(&items)

	if tx.Error != nil {
		log.Printf("Failed to fetch items: %v\n", tx.Error)
		err := fmt.Errorf("An error occured while fetching items.")
		return nil, http.StatusInternalServerError, err
	}

	return s.MorphToDtoMultiple(items), http.StatusAccepted, nil
}

func (s *QrcodeService) Find(id any) (*dto.QrcodeScanDTO, int, error) {
	var item models.Qrcode

	tx := database.OracleDB.Where("id = ?", id).First(&item)

	if tx.Error != nil {
		log.Printf("Failed to fetch item: %v\n", tx.Error)
		err := fmt.Errorf("An error occured while fetching item. Perhaps it doesn't exist")
		return nil, http.StatusInternalServerError, err
	}

	morphed := s.MorphToDto(item)
	return &morphed, http.StatusAccepted, nil
}

func (s *QrcodeService) MorphToDto(item models.Qrcode) dto.QrcodeScanDTO {
	if item.Fdn == nil || *item.Fdn == "" {
		return dto.QrcodeScanDTO{Qrcode: item}
	}

	var scannedLog models.ScanLog
	if err := database.MySQLDB.Preload("ScannedBy").Where(&models.ScanLog{Fdn: *item.Fdn}).First(&scannedLog).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Printf("Error scanning corresponding scanLog for fdn: %s: %s\n", *item.Fdn, err.Error())
		}
		return dto.QrcodeScanDTO{Qrcode: item}
	}

	return dto.QrcodeScanDTO{
		Qrcode:    item,
		ScannedAt: &scannedLog.CreatedAt,
		ScannedBy: scannedLog.ScannedBy,
	}
}

func (s *QrcodeService) MorphToDtoMultiple(items []models.Qrcode) []dto.QrcodeScanDTO {
	var results []dto.QrcodeScanDTO
	for _, item := range items {
		results = append(results, s.MorphToDto(item))
	}
	return results
}
