package services

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	database "github.com/othie12/scanner-api/internals/db"
	"github.com/othie12/scanner-api/internals/db/models"
	"github.com/othie12/scanner-api/utils"
	"gorm.io/gorm"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

// Create a new user
// IN => CreateUserDTO
// OUT => user, http status, error
func (s *UserService) Create(dto models.User) (*models.User, int, error) {
	var user models.User

	// Check if username already exists
	if database.DB.Where(models.User{Username: dto.Username}).Take(&user).Error == nil {
		err := fmt.Errorf("A user already exists with the given Username")
		return nil, http.StatusBadRequest, err
	}

	// hash password
	if dto.Password != "" {
		hashedPassword, err := utils.HashPassword(dto.Password)
		if err != nil {
			log.Printf("Password hash failed: %v\n", err)
			err = fmt.Errorf("Failed to hash password")
			return nil, http.StatusInternalServerError, err
		}

		dto.Password = hashedPassword
	}

	// Create user record in the DB
	tx := database.DB.Create(&dto)
	if tx.Error != nil {
		log.Printf("Failed to create user: %v\n", tx.Error)
		err := fmt.Errorf("An error occured while creating user")
		return nil, http.StatusInternalServerError, err
	}

	return &dto, http.StatusAccepted, nil
}

func (s *UserService) All(page int) ([]models.User, int, error) {
	var users []models.User
	limit := 10

	if err := database.DB.Limit(limit).Offset(database.FindOffset(page, limit)).Order("id DESC").Find(&users).Error; err != nil {
		log.Printf("Fetch users Failed with Error: %v\n", err)
		err := fmt.Errorf("An error occured while finding users")
		return nil, http.StatusInternalServerError, err
	}

	return users, http.StatusOK, nil
}

func (s *UserService) Search(query string) ([]models.User, int, error) {
	var users []models.User

	//tx := database.DB.Preload("Creator").Preload("Approver").Where("tag LIKE %?%", query).Or("brand LIKE %?%", query).Find(&users)
	tx := database.DB.
		Where("username LIKE ?", "%"+query+"%").
		Limit(10).Find(&users)

	if tx.Error != nil {
		log.Printf("Failed to fetch users: %v\n", tx.Error)
		err := fmt.Errorf("An error occured while fetching users.")
		return nil, http.StatusInternalServerError, err
	}

	return users, http.StatusAccepted, nil
}

func (s *UserService) Get(id any) (*models.User, int, error) {
	var user models.User

	if err := database.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("User with id %s not found: %v\n", id, err)
			err := fmt.Errorf("User not found")
			return nil, http.StatusNotFound, err
		}

		log.Printf("Fetch single user by Id Error: %v\n", err)
		err := fmt.Errorf("An error occured while finding user")
		return nil, http.StatusInternalServerError, err
	}

	user.RemoveSensitiveData()
	return &user, http.StatusOK, nil
}

func (s *UserService) ChangeUserLevel(id any, newLevel string) (*models.User, int, error) {
	user, status, er := s.Get(id)
	if er != nil {
		return user, status, er
	}

	user.UserLevel = newLevel
	if err := database.DB.Save(user).Error; err != nil {
		log.Printf("Update user level error: %v\n", err)
		err := fmt.Errorf("An error occured while updating user role")
		return nil, http.StatusInternalServerError, err
	}

	user.RemoveSensitiveData()
	return user, http.StatusOK, nil
}

type ChangePwdDTO struct {
	CurrPwd       string `json:"curr_pwd"`
	NewPwd        string `json:"new_pwd"`
	ConfirmNewPwd string `json:"confirm_new_pwd"`
}

func (s *UserService) ChangePwd(id any, shouldCheckHash bool, dto ChangePwdDTO) (*models.User, int, error) {
	var user models.User

	if err := database.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("User with id %s not found: %v\n", id, err)
			err := fmt.Errorf("User not found")
			return nil, http.StatusNotFound, err
		}

		log.Printf("Fetch single user by Id Error: %v\n", err)
		err := fmt.Errorf("An error occured while finding user")
		return nil, http.StatusInternalServerError, err
	}

	if (shouldCheckHash || user.UserLevel == "admin") && !utils.CheckPasswordHash(dto.CurrPwd, user.Password) {
		err := fmt.Errorf("Incorrect password")
		return nil, http.StatusUnauthorized, err
	}

	hashedPassword, err := utils.HashPassword(dto.NewPwd)
	if err != nil {
		log.Printf("Password hash failed: %v\n", err)
		err = fmt.Errorf("Failed to hash password")
		return nil, http.StatusInternalServerError, err
	}

	user.Password = hashedPassword

	if err := database.DB.Save(&user).Error; err != nil {
		log.Printf("Update user password error: %v\n", err)
		err := fmt.Errorf("An error occured while updating password")
		return nil, http.StatusInternalServerError, err
	}

	user.RemoveSensitiveData()
	return &user, http.StatusOK, nil
}
