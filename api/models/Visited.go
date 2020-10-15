package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Visited struct {
	ID         uint32    `gorm:"primary_key;auto_increment" json:"id"`
	PlaceID    string    `gorm:"size:45;not null;unique" json:"place_id"`
	UserID     string    `gorm:"not null" json:"user_id"`
	CreateTime time.Time `gorm:"not null" json:"create_time"`
	UpdateTime time.Time `gorm:"default:CURRENT_TIMESTAMP;not null" json:"update_time"`
}

func (v *Visited) GetVisiteds(db *gorm.DB, uid string) (*[]Visited, error) {
	var err error
	visited := []Visited{}
	err = db.Debug().Model(&Favorites{}).Where("user_id = ?", uid).Find(&visited).Error
	if err != nil {
		return &[]Visited{}, err
	}
	return &visited, err
}

func (v *Visited) SaveVisited(db *gorm.DB) (*Visited, error) {
	var err error
	err = db.Debug().Create(&v).Error
	if err != nil {
		return &Visited{}, err
	}
	return v, nil
}

func (v *Visited) InitData(placeID string, userID string) {
	v.ID = 0
	v.PlaceID = placeID
	v.UserID = userID
	v.CreateTime = time.Now()
	v.UpdateTime = time.Now()
}

func (v *Visited) DeleteVisited(db *gorm.DB, place_id string, uid string) (int64, error) {
	db = db.Debug().Where("place_id = ? and user_id = ?", place_id, uid).Take(&Visited{}).Delete(&Visited{})
	if db.Error != nil {
		return 0, db.Error
	}

	return db.RowsAffected, nil
}
