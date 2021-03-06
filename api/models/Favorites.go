package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Favorites struct {
	ID         uint32    `gorm:"primary_key;auto_increment" json:"id"`
	PlaceID    string    `gorm:"size:45;not null;unique" json:"place_id"`
	UserID     string    `gorm:"not null" json:"user_id"`
	CreateTime time.Time `gorm:"not null" json:"create_time"`
	UpdateTime time.Time `gorm:"default:CURRENT_TIMESTAMP;not null" json:"update_time"`
}

func (f *Favorites) GetFavorites(db *gorm.DB, uid string) (*[]Favorites, error) {
	var err error
	favorites := []Favorites{}
	err = db.Debug().Model(&Favorites{}).Where("user_id = ?", uid).Find(&favorites).Error
	if err != nil {
		return &[]Favorites{}, err
	}
	return &favorites, err
}

func (f *Favorites) SaveFavorite(db *gorm.DB) (*Favorites, error) {
	var err error
	err = db.Debug().Create(&f).Error
	if err != nil {
		return &Favorites{}, err
	}
	return f, nil
}

func (f *Favorites) InitData(placeID string, uid string) {
	f.ID = 0
	f.PlaceID = placeID
	f.UserID = uid
	f.CreateTime = time.Now()
	f.UpdateTime = time.Now()
}

func (f *Favorites) DeleteFavorite(db *gorm.DB, place_id string, uid string) (int64, error) {
	db = db.Debug().Model(&User{}).Where("place_id = ? and user_id = ?", place_id, uid).Take(&Favorites{}).Delete(&Favorites{})
	if db.Error != nil {
		return 0, db.Error
	}

	return db.RowsAffected, nil
}
