package models

import (
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type HistoryResult struct {
	ID          uint32      `gorm:"primary_key;auto_increment" json:"id"`
	HandcupId   string      `gorm:"size:45;not null;unique" json:"handcup_id"`
	HandcupInfo HandcupInfo `json:"handcupInfo"`
	Distance    uint32      `gorm:"size:45;not null;unique" json:"distance"`
	CreateTime  time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"create_time"`
	UpdateTime  time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"update_time"`
}

func (h *HistoryResult) Prepare() {
	h.ID = 0
	h.HandcupId = html.EscapeString(strings.TrimSpace(h.HandcupId))
	h.HandcupInfo = HandcupInfo{}
	h.Distance = 100
	h.CreateTime = time.Now()
	h.UpdateTime = time.Now()
}

func (h *HistoryResult) SaveHistoryResult(db *gorm.DB) (*HistoryResult, error) {
	var err error
	err = db.Debug().Model(&HistoryResult{}).Create(&h).Error
	if err != nil {
		return &HistoryResult{}, err
	}
	if h.ID != 0 {
		err = db.Debug().Model(&HandcupInfo{}).Where("id = ?", h.HandcupId).Take(&h.HandcupInfo).Error
		if err != nil {
			return &HistoryResult{}, err
		}
	}
	return h, nil
}

func (h *HistoryResult) FindAllHistoryResults(db *gorm.DB) (*[]HistoryResult, error) {
	var err error
	HistoryResults := []HistoryResult{}
	err = db.Debug().Model(&HistoryResult{}).Limit(100).Find(&HistoryResults).Error
	if err != nil {
		return &[]HistoryResult{}, err
	}
	if len(HistoryResults) > 0 {
		for i, _ := range HistoryResults {
			err := db.Debug().Model(&HandcupInfo{}).Where("id = ?", HistoryResults[i].HandcupId).Take(&HistoryResults[i].HandcupInfo).Error
			if err != nil {
				return &[]HistoryResult{}, err
			}
		}
	}
	return &HistoryResults, nil
}

func (h *HistoryResult) FindHistoryResultByID(db *gorm.DB, pid uint64) (*HistoryResult, error) {
	var err error
	err = db.Debug().Model(&HistoryResult{}).Where("id = ?", pid).Take(&h).Error
	if err != nil {
		return &HistoryResult{}, err
	}
	if h.ID != 0 {
		err = db.Debug().Model(&HandcupInfo{}).Where("id = ?", h.HandcupId).Take(&h.HandcupInfo).Error
		if err != nil {
			return &HistoryResult{}, err
		}
	}
	return h, nil
}

func (h *HistoryResult) UpdateAHistoryResult(db *gorm.DB, uid uint32) (*HistoryResult, error) {
	var err error

	err = db.Debug().Model(&HistoryResult{}).Where("id = ?", h.ID).Updates(HistoryResult{Distance: h.Distance, UpdateTime: time.Now()}).Error
	if err != nil {
		return &HistoryResult{}, err
	}
	if h.ID != 0 {
		err = db.Debug().Model(&HandcupInfo{}).Where("id = ?", h.HandcupId).Take(&h.HandcupInfo).Error
		if err != nil {
			return &HistoryResult{}, err
		}
	}
	return h, nil
}
