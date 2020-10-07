package models

import (
	"errors"
	"html"
	"log"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uint32    `gorm:"primary_key;auto_increment" json:"id"`
	UserID    string    `gorm:"size:255;not null;unique" json:"userId"`
	Name      string    `gorm:"size:255" json:"name"`
	Email     string    `gorm:"size:100;not null;unique" json:"email"`
	Password  string    `gorm:"size:100;not null;" json:"password"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (u *User) BeforeSave() error {
	hashedPassword, err := Hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) Prepare() {
	u.ID = 0
	u.Name = html.EscapeString(strings.TrimSpace(u.Name))
	u.Email = html.EscapeString(strings.TrimSpace(u.Email))
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
}

func (u *User) Validate(action string) error {
	switch strings.ToLower(action) {
	case "update":
		if u.Password == "" {
			return errors.New("Required Password")
		}
		return nil

	case "login":
		if u.Password == "" {
			return errors.New("Required Password")
		}
		if u.UserID == "" {
			return errors.New("Required UserID")
		}
		if u.Email == "" {
			return errors.New("Required Email")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("Invalid Email")
		}
		return nil

	default:
		if u.UserID == "" {
			return errors.New("Required UserID")
		}
		if u.Password == "" {
			return errors.New("Required Password")
		}
		if u.Email == "" {
			return errors.New("Required Email")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("Invalid Email")
		}
		return nil
	}
}

func (u *User) SaveUser(db *gorm.DB) (*User, error) {
	var err error
	err = db.Debug().Create(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

func (u *User) FindAllUsers(db *gorm.DB) (*[]User, error) {
	var err error
	users := []User{}
	err = db.Debug().Model(&User{}).Limit(100).Find(&users).Error
	if err != nil {
		return &[]User{}, err
	}
	return &users, err
}

func (u *User) FindUserByID(db *gorm.DB, uid string) (*User, error) {
	var err error
	err = db.Debug().Model(User{}).Where("user_id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &User{}, errors.New("User Not Found")
	}
	return u, err
}

func (u *User) UpdateAUser(db *gorm.DB, uid string) (*User, error) {
	db = db.Debug().Model(&User{}).Where("user_id = ?", uid).Take(&User{}).UpdateColumns(
		map[string]interface{}{
			"name":      u.Name,
			"update_at": time.Now(),
		},
	)
	if db.Error != nil {
		return &User{}, db.Error
	}
	// This is the display the updated user
	err := db.Debug().Model(&User{}).Where("user_id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

func (u *User) UpdatePassword(db *gorm.DB, uid string) (*User, error) {
	// To hash the password
	err := u.BeforeSave()
	if err != nil {
		log.Println(err)
	}
	db = db.Debug().Model(&User{}).Where("user_id = ?", uid).Take(&User{}).UpdateColumns(
		map[string]interface{}{
			"password":  u.Password,
			"update_at": time.Now(),
		},
	)
	if db.Error != nil {
		return &User{}, db.Error
	}
	// This is the display the updated user
	err = db.Debug().Model(&User{}).Where("user_id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

func (u *User) ResetPassword(db *gorm.DB, email string) (*User, error) {
	// To hash the password
	err := u.BeforeSave()
	if err != nil {
		log.Println(err)
	}
	db = db.Debug().Model(&User{}).Where("email = ?", email).Take(&User{}).UpdateColumns(
		map[string]interface{}{
			"password":  u.Password,
			"update_at": time.Now(),
		},
	)
	if db.Error != nil || db.RowsAffected == 0 {
		return &User{}, db.Error
	}
	// This is the display the updated user
	err = db.Debug().Model(&User{}).Where("email = ?", email).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

func (u *User) DeleteAUser(db *gorm.DB, uid string) (int64, error) {
	db = db.Debug().Model(&User{}).Where("user_id = ?", uid).Take(&User{}).Delete(&User{})
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
