package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/othie12/scanner-api/config"
	database "github.com/othie12/scanner-api/internals/db"
	"github.com/othie12/scanner-api/internals/db/models"
	"github.com/othie12/scanner-api/internals/dto"
	"github.com/othie12/scanner-api/utils"
	"gorm.io/gorm"
)

type QrcodeService struct {
	HTTPClient *http.Client
}

func NewQrcodeService() *QrcodeService {
	return &QrcodeService{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *QrcodeService) Scan(fdn string, scannerId any) (*models.Qrcode, int, error) {
	var rslt models.Qrcode
	if fdn == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("FDN is empty")
	}

	var user models.User
	if getUserErr := database.DB.First(&user, scannerId).Error; getUserErr != nil {
		if getUserErr != gorm.ErrRecordNotFound {
			log.Printf("Error Finding User with Id: %v: %s\n", scannerId, getUserErr.Error())
			return nil, http.StatusInternalServerError, fmt.Errorf("We encountered a Database Error, please try again")
		}

		return nil, http.StatusBadRequest, fmt.Errorf("User account Error. Make sure you are logged in")
	}

	var scannedLog models.ScanLog
	getLogErr := database.DB.Where(&models.ScanLog{Fdn: fdn}).First(&scannedLog).Error
	if getLogErr != nil && getLogErr != gorm.ErrRecordNotFound {
		log.Printf("Error scanning corresponding scanLog for fdn: %s: %s\n", fdn, getLogErr.Error())
		return nil, http.StatusInternalServerError, fmt.Errorf("We encountered a Database Error, please try again")
	}

	if getLogErr == nil {
		failedScanLog := models.FailedScanLog{
			Fdn:         fdn,
			ScannedById: user.ID,
			Reason:      "DUPLICATE",
		}

		if saveLogErr := database.DB.Create(&failedScanLog).Error; saveLogErr != nil {
			log.Printf("Error saving failed scanLog for fdn '%s' by '%s': %s\n", fdn, user.Username, saveLogErr.Error())
			return nil, http.StatusInternalServerError, fmt.Errorf("We encountered a Database Error, please try again")
		}
		return nil, http.StatusConflict, fmt.Errorf("Duplicate Scan")
	}

	if err := database.DB.Where("fdn = ?", fdn).First(&rslt).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Printf("Error finding QrCode with fdn: %s: %s\n", fdn, err.Error())
			return nil, http.StatusInternalServerError, fmt.Errorf("We encountered a Database Error, please try again")
		}

		failedScanLog := models.FailedScanLog{
			Fdn:         fdn,
			ScannedById: user.ID,
			Reason:      "DOES_NOT_EXIST",
		}

		if saveLogErr := database.DB.Create(&failedScanLog).Error; saveLogErr != nil {
			log.Printf("Error saving failed scanLog for fdn '%s' by '%s': %s\n", fdn, user.Username, saveLogErr.Error())
			return nil, http.StatusInternalServerError, fmt.Errorf("We encountered a Database Error, please try again")
		}

		return nil, http.StatusNotFound, fmt.Errorf("Record with This FDN doesn't exist")
	}

	scanLog := models.ScanLog{
		Fdn:         fdn,
		ScannedById: user.ID,
	}

	if saveLogErr := database.DB.Create(&scanLog).Error; saveLogErr != nil {
		log.Printf("Error saving scanLog for fdn '%s' by '%s': %s\n", fdn, user.Username, saveLogErr.Error())
		return nil, http.StatusInternalServerError, fmt.Errorf("We encountered a Database Error, please try again")
	}

	return &rslt, http.StatusOK, nil
}

func (s *QrcodeService) SendSMS(ctx context.Context, message string, contacts ...string) error {
	// 1. Create request payload
	jsonData, err := json.Marshal(
		map[string]interface{}{
			"message":         message,
			"contact_numbers": contacts,
		})

	if err != nil {
		return err
	}

	payload := bytes.NewBuffer(jsonData)

	url := fmt.Sprintf("%s/api/send", config.ServerConfig.SMSServiceBaseUrl)

	// 2. Create request and set headers
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		url,
		payload,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("apiKey", config.ServerConfig.SMSServiceApiKey)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return fmt.Errorf("reading response not ok body returned ERR: %s\n", err.Error())
		}

		return fmt.Errorf("sms to %s failed with status code: %d\n%s\n\n", strings.Join(contacts, ", "), resp.StatusCode, string(body))
	}

	// Read everything into memory
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	bodyString := string(bodyBytes)
	log.Println(bodyString)
	return nil
}

func (s *QrcodeService) All(page int) ([]dto.QrcodeScanDTO, int, error) {
	limit := 10
	var items []models.Qrcode

	tx := database.DB.Limit(limit).Offset(database.FindOffset(page, limit)).Order("id DESC").Find(&items)
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
		if err := database.DB.Model(&models.ScanLog{}).Order("id DESC").Where("created_at > ?", utils.AddTimeToDate(dto.DateFrom, false)).Pluck("fdn", &scannedFdns).Error; err != nil {
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

	tx := database.DB.Where(queryCondition, queryArgs...).Find(&items)
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

	tx := database.DB.
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

	tx := database.DB.Where("id = ?", id).First(&item)

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
	if err := database.DB.Preload("ScannedBy").Where(&models.ScanLog{Fdn: *item.Fdn}).First(&scannedLog).Error; err != nil {
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
