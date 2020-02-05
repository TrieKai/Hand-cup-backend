package models

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
)

type HistoryRequest struct {
	ID           uint32      `gorm:"primary_key;auto_increment" json:"id"`
	GroupId      uint32      `grom:"not null;" json:"group_id"`
	HandcupId    uint32      `gorm:"not null;" json:"handcup_id"`
	HandcupInfo  HandcupInfo `json:"handcupInfo"`
	ReqLatitude  float64     `grom:"not null;" json:"req_latitude"`
	ReqLongitude float64     `grom:"not null;" json:"req_longitude"`
	Distance     uint        `gorm:"not null;" json:"distance"`
	CreateTime   time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"create_time"`
	UpdateTime   time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"update_time"`
}

func (h *HistoryRequest) initData(db *gorm.DB, latestGroupID uint32, latestID uint32) {
	h.GroupId = latestGroupID
	h.HandcupId = latestID + 1
	h.HandcupInfo = HandcupInfo{}
	h.CreateTime = time.Now()
	h.UpdateTime = time.Now()
}

func (h *HistoryRequest) FindLatestGroupID(db *gorm.DB) uint32 {
	var max uint32
	row := db.Table("history_requests").Select("MAX(group_id)").Row()
	row.Scan(&max)
	log.Println(max)

	return max
}

func (h *HistoryRequest) SaveHistoryRequest(db *gorm.DB, latestGroupID uint32, latestID uint32) (*HistoryRequest, error) {
	var err error

	h.initData(db, latestGroupID, latestID)

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
