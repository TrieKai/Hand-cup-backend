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

// func (h *HandcupInfo) Prepare() {
// 	h.ID = 0
// 	h.GoogleId = html.EscapeString(strings.TrimSpace(h.GoogleId))
// 	h.PlaceId = html.EscapeString(strings.TrimSpace(h.GoogleId))
// 	h.Name = html.EscapeString(strings.TrimSpace(h.Name))
// 	h.Latitude = 0
// 	h.Longitude = 0
// 	h.Rating = 0
// 	h.ImageReference = html.EscapeString(strings.TrimSpace(h.ImageReference))
// 	h.ImageWidth = 0
// 	h.ImageHeight = 0
// 	h.ImageUrl = html.EscapeString(strings.TrimSpace(h.ImageUrl))
// 	h.CreateTime = time.Now()
// 	h.UpdateTime = time.Now()
// }

func (h *HandcupInfo) FindLatestID(db *gorm.DB) uint32 {
	var latestID uint32
	row := db.Debug().Table("handcup_infos").Select("MAX(id)").Row()
	row.Scan(&latestID)

	return latestID
}

func (h *HandcupInfo) SaveHandcupInfo(db *gorm.DB) (*HandcupInfo, error) {
	var err error
	isExist := db.Raw("SELECT place_id FROM handcup_infos WHERE place_id = ?", h.PlaceId).Scan(&h)
	fmt.Println("place_id是否存在", isExist.RowsAffected)
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

func (h *HandcupInfo) FindHandcupInfoByID(db *gorm.DB, hid uint32) (*HandcupInfo, error) {
	var err error
	err = db.Debug().Model(HandcupInfo{}).Where("id = ?", hid).Take(&h).Error
	if err != nil {
		return &HandcupInfo{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &HandcupInfo{}, errors.New("HandcupInfo Not Found")
	}
	return h, err
}

func (h *HandcupInfo) UpdateAHandcupInfo(db *gorm.DB, hid uint32) (*HandcupInfo, error) {
	var err error
	db = db.Debug().Model(&HandcupInfo{}).Where("id = ?", hid).Take(&HandcupInfo{}).UpdateColumns(
		map[string]interface{}{
			"name":      h.Name,
			"imageUrl":  h.ImageUrl,
			"rating":    h.Rating,
			"update_at": time.Now(),
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
