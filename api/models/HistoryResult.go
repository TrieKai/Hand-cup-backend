package models

import (
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
)

type HistoryRequest struct {
	ID           uint32      `gorm:"type:bigint(20) unsigned auto_increment;not null;primary_key" json:"id"`
	GroupId      uint32      `grom:"not null;" json:"group_id"`
	HandcupId    uint32      `gorm:"not null;" json:"handcup_id"`
	HandcupInfo  HandcupInfo `json:"handcupInfo"`
	ReqLatitude  float64     `grom:"not null;" json:"req_latitude"`
	ReqLongitude float64     `grom:"not null;" json:"req_longitude"`
	Distance     uint        `gorm:"not null;" json:"distance"`
	Keyword      string      `gorm:"not null;" json:"keyword"`
	CreateTime   time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"create_time"`
	UpdateTime   time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"update_time"`
}

type CheckHRResponse struct {
	GroupId      uint32    `grom:"not null;" json:"group_id"`
	ReqLatitude  float64   `grom:"not null;" json:"req_latitude"`
	ReqLongitude float64   `grom:"not null;" json:"req_longitude"`
	Distance     uint      `grom:"not null;" json:"distance"`
	UpdateTime   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"update_time"`
}

type HandcupIdResponse struct {
	GroupId    uint32    `grom:"not null;" json:"group_id"`
	HandcupId  uint32    `gorm:"not null;" json:"handcup_id"`
	UpdateTime time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"update_time"`
}

func (h *HistoryRequest) InitData(latestHisReqID uint32, groupID uint32, handcupID uint32, distance uint, keyword string) {
	h.ID = latestHisReqID + 1
	h.GroupId = groupID
	h.HandcupId = handcupID
	h.HandcupInfo = HandcupInfo{}
	h.Distance = distance
	h.Keyword = keyword
	h.CreateTime = time.Now()
	h.UpdateTime = time.Now()
}

func (h *HistoryRequest) FindLatestHisReqID(db *gorm.DB) uint32 {
	var max uint32
	row := db.Table("history_requests").Select("MAX(id)").Row()
	row.Scan(&max)
	log.Println(max)

	return max
}

func (h *HistoryRequest) FindLatestGroupID(db *gorm.DB) uint32 {
	var max uint32
	row := db.Table("history_requests").Select("MAX(group_id)").Row()
	row.Scan(&max)
	log.Println(max)

	return max
}

func (h *HistoryRequest) CheckHistoryReq(db *gorm.DB) ([]HandcupIdResponse, error) {
	var err error
	var respData []CheckHRResponse
	var resp []HandcupIdResponse
	// 計算半徑約100公尺內
	maxLat := h.ReqLatitude + 0.00045
	minLat := h.ReqLatitude - 0.00045
	maxLng := h.ReqLongitude + 0.00045
	minLng := h.ReqLongitude - 0.00045

	db.
		Table("history_requests").
		Select("group_id, req_latitude, req_longitude, distance, update_time").
		Where("(req_latitude BETWEEN ? AND ?) AND (req_longitude BETWEEN ? AND ?)", minLat, maxLat, minLng, maxLng).
		Group("group_id").
		Find(&respData)

	fmt.Println("DB內已搜尋到的資料:", respData)

	// 如果在 History requests 內有資料
	if len(respData) != 0 {
		resp = h.findHandcupId(db, respData[0].GroupId) // 以 GroupId 去找所有的 HandcupId
	} else {
		return nil, err
	}

	return resp, err
}

func (h *HistoryRequest) findHandcupId(db *gorm.DB, groupId uint32) []HandcupIdResponse {
	var resp []HandcupIdResponse
	db.Table("history_requests").Select("group_id, handcup_id, update_time").Where("group_id = ?", groupId).Find(&resp)

	return resp
}

func (h *HistoryRequest) SaveHistoryReq(db *gorm.DB) (*HistoryRequest, error) {
	var err error
	fmt.Println(&h)
	err = db.Debug().Create(&h).Error
	if err != nil {
		return &HistoryRequest{}, err
	}
	if h.GroupId != 0 {
		err = db.Debug().Model(&HandcupInfo{}).Where("id = ?", h.HandcupId).Take(&h.HandcupInfo).Error
		if err != nil {
			return &HistoryRequest{}, err
		}
	}
	return h, nil
}

// func (h *HistoryRequest) FindAllHistoryRequests(db *gorm.DB) (*[]HistoryRequest, error) {
// 	var err error
// 	HistoryRequests := []HistoryRequest{}
// 	err = db.Debug().Model(&HistoryRequest{}).Limit(100).Find(&HistoryRequests).Error
// 	if err != nil {
// 		return &[]HistoryRequest{}, err
// 	}
// 	if len(HistoryRequests) > 0 {
// 		for i, _ := range HistoryRequests {
// 			err := db.Debug().Model(&HandcupInfo{}).Where("id = ?", HistoryRequests[i].HandcupId).Take(&HistoryRequests[i].HandcupInfo).Error
// 			if err != nil {
// 				return &[]HistoryRequest{}, err
// 			}
// 		}
// 	}
// 	return &HistoryRequests, nil
// }

func (h *HistoryRequest) GetGroupHisReqByGId(db *gorm.DB, groupId uint32) CheckHRResponse {
	var respData CheckHRResponse
	db.
		Table("history_requests").
		Select("group_id, req_latitude, req_longitude, distance").
		Where("group_id = ?", groupId).
		First(&respData)

	return respData
}

// func (h *HistoryRequest) GetIDByHandcupID(db *gorm.DB, hId uint32) uint32 {
// 	var id uint32

// 	db.
// 		Table("history_requests").
// 		Select("id").
// 		Where("handcup_id = ?", hId)

// 	return id
// }

// func (h *HistoryRequest) FindHistoryRequestByID(db *gorm.DB, pid uint64) (*HistoryRequest, error) {
// 	var err error
// 	err = db.Debug().Model(&HistoryRequest{}).Where("id = ?", pid).Take(&h).Error
// 	if err != nil {
// 		return &HistoryRequest{}, err
// 	}
// 	if h.ID != 0 {
// 		err = db.Debug().Model(&HandcupInfo{}).Where("id = ?", h.HandcupId).Take(&h.HandcupInfo).Error
// 		if err != nil {
// 			return &HistoryRequest{}, err
// 		}
// 	}
// 	return h, nil
// }

func (h *HistoryRequest) UpdateAHistoryRequest(db *gorm.DB) (*HistoryRequest, error) {
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
