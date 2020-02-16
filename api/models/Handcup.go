package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type HandcupInfo struct {
	ID             uint32    `gorm:"primary_key;auto_increment" json:"id"`
	GoogleId       string    `gorm:"size:45;not null;unique" json:"google_id"`
	PlaceId        string    `gorm:"size:45;not null;unique" json:"place_id"`
	Name           string    `gorm:"size:45;not null;unique" json:"name"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	Rating         float32   `json:"rating"`
	ImageReference string    `json:"image_reference"`
	ImageWidth     int       `json:"image_width"`
	ImageHeight    int       `json:"image_height"`
	ImageUrl       string    `gorm:"size:150;" json:"image_url"`
	CreateTime     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"create_time"`
	UpdateTime     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"update_time"`
}

func (h *HandcupInfo) FindLatestID(db *gorm.DB) uint32 {
	var latestID uint32
	row := db.Debug().Table("handcup_infos").Select("MAX(id)").Row()
	row.Scan(&latestID)

	return latestID
}

func (h *HandcupInfo) SaveHandcupInfo(db *gorm.DB) (*HandcupInfo, error) {
	var err error
	isExist := db.Raw("SELECT place_id FROM handcup_infos WHERE place_id = ?", h.PlaceId).Scan(&h)
	fmt.Println("place_id影響列數:", isExist.RowsAffected)
	// If this place_id not exist
	if isExist.RowsAffected == 0 {
		err = db.Debug().Create(&h).Error
		if err != nil {
			return &HandcupInfo{}, err
		}
	}
	// fmt.Print(h)
	return h, nil
}

func (h *HandcupInfo) FindAllHandcupInfo(db *gorm.DB) (*[]HandcupInfo, error) {
	var err error
	HandcupInfoList := []HandcupInfo{}
	err = db.Debug().Model(&HandcupInfo{}).Limit(100).Find(&HandcupInfoList).Error
	if err != nil {
		return &[]HandcupInfo{}, err
	}
	return &HandcupInfoList, err
}

func (h *HandcupInfo) FindHandcupInfoByID(db *gorm.DB, hid uint32) (HandcupInfo, error) {
	var err error
	handcupInfo := HandcupInfo{}
	err = db.Debug().Table("handcup_infos").Where("id = ?", hid).Find(&handcupInfo).Error
	if err != nil {
		return handcupInfo, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return handcupInfo, errors.New("HandcupInfo Not Found")
	}

	return handcupInfo, err
}

func (h *HandcupInfo) FindHandcupInfoByPlaceID(db *gorm.DB, pid string) (HandcupInfo, error) {
	var err error
	handcupInfo := HandcupInfo{}
	err = db.Debug().Table("handcup_infos").Where("place_id = ?", pid).Find(&handcupInfo).Error
	if err != nil {
		return handcupInfo, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return handcupInfo, errors.New("HandcupInfo Not Found")
	}

	return handcupInfo, err
}

func (h *HandcupInfo) UpdateAHandcupInfo(db *gorm.DB, hid uint32) (*HandcupInfo, error) {
	var err error
	db = db.Debug().Model(&HandcupInfo{}).Where("id = ?", hid).Take(&HandcupInfo{}).UpdateColumns(
		map[string]interface{}{
			"google_id":       h.GoogleId,
			"name":            h.Name,
			"latitude":        h.Latitude,
			"longitude":       h.Longitude,
			"rating":          h.Rating,
			"image_reference": h.ImageReference,
			"image_width":     h.ImageWidth,
			"image_height":    h.ImageHeight,
			"imageUrl":        h.ImageUrl,
			"update_time":     time.Now(),
		},
	)
	if db.Error != nil {
		return &HandcupInfo{}, db.Error
	}
	// This is the display the updated handcupInfo
	err = db.Debug().Model(&HandcupInfo{}).Where("id = ?", hid).Take(&h).Error
	if err != nil {
		return &HandcupInfo{}, err
	}
	return h, nil
}
