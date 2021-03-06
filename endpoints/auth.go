package endpoints

import (
	"errors"
	"fmt"
	"time"

	"regexp"

	"github.com/crufter/1b-user-service/domain"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

var reg = regexp.MustCompile("^[0-9a-z-]+$")

func (e Endpoints) createToken(tx *gorm.DB, userId string) (*domain.AccessToken, error) {
	token := domain.AccessToken{
		Id:        domain.Sid.MustGenerate(),
		Token:     uuid.NewV4().String(),
		UserId:    userId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return &token, domain.NewAccessTokenDao(tx).Create(token)
}

func (e *Endpoints) Register(email, name, password string) (*domain.User, *domain.AccessToken, error) {
	if email == "" {
		return nil, nil, errors.New("Email can not be empty.")
	}
	if password == "" {
		return nil, nil, errors.New("Password can not be empty.")
	}
	if name == "" {
		return nil, nil, errors.New("Name can't be empty.")
	}
	if !reg.Match([]byte(name)) {
		return nil, nil, errors.New("Allowed name characters: lowercase letters, numbers and dash")
	}
	user, err := domain.NewUserDao(e.db).GetByEmail(email)
	if err == nil && user.Id != "" {
		return nil, nil, errors.New("This email is already registered. Try to log in.")
	}
	passw, err := bcrypt.GenerateFromPassword([]byte(password), 5)
	if err != nil {
		return nil, nil, err
	}
	user = domain.User{
		Id:       domain.Sid.MustGenerate(),
		Email:    email,
		Nick:     name,
		Password: string(passw),
	}
	tx := e.db.Begin()
	userDao := domain.NewUserDao(tx)
	if err := userDao.Create(user); err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("error creating new user %s", err.Error())
	}
	tokenDao := domain.NewAccessTokenDao(tx)
	token := domain.AccessToken{
		Id:        domain.Sid.MustGenerate(),
		Token:     uuid.NewV4().String(),
		UserId:    user.Id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := tokenDao.Create(token); err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("error creating token for user: %s", err.Error())
	}
	err = tx.Commit().Error
	if err != nil {
		return nil, nil, err
	}
	return &user, &token, nil
}

func (e *Endpoints) Login(email, password string) (*domain.User, *domain.AccessToken, error) {
	if email == "" {
		return nil, nil, errors.New("Email can not be empty.")
	}
	if password == "" {
		return nil, nil, errors.New("Password can not be empty.")
	}
	user, err := domain.NewUserDao(e.db).GetByEmail(email)
	if err != nil {
		return nil, nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, nil, errors.New("Could not log in")
	}
	tokenDao := domain.NewAccessTokenDao(e.db)
	token := domain.AccessToken{
		Id:        domain.Sid.MustGenerate(),
		Token:     uuid.NewV4().String(),
		UserId:    user.Id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := tokenDao.Create(token); err != nil {
		return nil, nil, fmt.Errorf("error creating token for user: %s", err.Error())
	}
	return &user, &token, nil
}

// GetUser by token
func (e *Endpoints) GetUser(tokenId string) (*domain.User, error) {
	// first get token
	tokenDao := domain.NewAccessTokenDao(e.db)
	t, err := tokenDao.GetByToken(tokenId)
	if err != nil {
		return nil, fmt.Errorf("token (%s) is associated to no users", tokenId)
	}
	userDao := domain.NewUserDao(e.db)
	u, _ := userDao.GetById(t.UserId)
	return &u, nil
}

func (e *Endpoints) ChangePassword(user *domain.User, oldPassword, newPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return errors.New("Old password is incorrect")
	}
	npassw, err := bcrypt.GenerateFromPassword([]byte(newPassword), 5)
	if err != nil {
		return err
	}
	user.Password = string(npassw)
	return e.db.Save(user).Error
}
