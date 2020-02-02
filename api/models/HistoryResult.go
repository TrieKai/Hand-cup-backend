package models

import (
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type HistoryRequest struct {
	ID           uint32      `gorm:"primary_key;auto_increment" json:"id"`
	GroupId      uint32      `grom:"not null;" json:"group_id"`
	HandcupId    string      `gorm:"size:45;not null;unique" json:"handcup_id"`
	HandcupInfo  HandcupInfo `json:"handcupInfo"`
	ReqLatitude  float64     `grom:"not null;" json:"req_latitude"`
	ReqLongitute float64     `grom:"not null;" json:"req_longitude"`
	Distance     uint32      `gorm:"not null;" json:"distance"`
	CreateTime   time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"create_time"`
	UpdateTime   time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"update_time"`
}

func (h *HistoryRequest) Prepare() {
	h.ID = 0
	h.HandcupId = html.EscapeString(strings.TrimSpace(h.HandcupId))
	h.HandcupInfo = HandcupInfo{}
	h.Distance = 100
	h.CreateTime = time.Now()
	h.UpdateTime = time.Now()
}

func (h *HistoryRequest) SaveHistoryRequest(db *gorm.DB) (*HistoryRequest, error) {
	var err error
	err = db.Debug().Model(&HistoryRequest{}).Create(&h).Error
	if err != nil {
		return &HistoryRequest{}, err
	}
	if h.ID != 0 {
		err = db.Debug().Model(&HandcupInfo{}).Where("id = ?", h.HandcupId).Take(&h.HandcupInfo).Error
		if err != nil {
			return &HistoryRequest{}, err
		}
	}
	return h, nil
}

func (h *HistoryRequest) FindAllHistoryRequests(db *gorm.DB) (*[]HistoryRequest, error) {
	var err error
	HistoryRequests := []HistoryRequest{}
	err = db.Debug().Model(&HistoryRequest{}).Limit(100).Find(&HistoryRequests).Error
	if err != nil {
		return &[]HistoryRequest{}, err
	}
	if len(HistoryRequests) > 0 {
		for i, _ := range HistoryRequests {
			err := db.Debug().Model(&HandcupInfo{}).Where("id = ?", HistoryRequests[i].HandcupId).Take(&HistoryRequests[i].HandcupInfo).Error
			if err != nil {
				return &[]HistoryRequest{}, err
			}
		}
	}
	return &HistoryRequests, nil
}

func (h *HistoryRequest) FindHistoryRequestByID(db *gorm.DB, pid uint64) (*HistoryRequest, error) {
	var err error
	err = db.Debug().Model(&HistoryRequest{}).Where("id = ?", pid).Take(&h).Error
	if err != nil {
		return &HistoryRequest{}, err
	}
	if h.ID != 0 {
		err = db.Debug().Model(&HandcupInfo{}).Where("id = ?", h.HandcupId).Take(&h.HandcupInfo).Error
		if err != nil {
			return &HistoryRequest{}, err
		}
	}
	return h, nil
}

func (h *HistoryRequest) UpdateAHistoryRequest(db *gorm.DB, uid uint32) (*HistoryRequest, error) {
	var err error

	err = db.Debug().Model(&HistoryRequest{}).Where("id = ?", h.ID).Updates(HistoryRequest{Distance: h.Distance, UpdateTime: time.Now()}).Error
	if err != nil {
		return &HistoryRequest{}, err
	}
	if h.ID != 0 {
		err = db.Debug().Model(&HandcupInfo{}).Where("id = ?", h.HandcupId).Take(&h.HandcupInfo).Error
		if err != nil {
			return &HistoryRequest{}, err
		}
	}
	return h, nil
}
