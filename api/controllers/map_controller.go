package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"handCup-project-backend/api/models"
	"handCup-project-backend/api/responses"
	formaterror "handCup-project-backend/api/utils"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"googlemaps.github.io/maps"
)

type handleMapParms struct {
	nextToken    string
	location     maps.LatLng
	distance     uint
	w            http.ResponseWriter
	r            *http.Request
	respDataList []models.HandcupRespData
}

type handleUpdateMapParms struct {
	nextToken    string
	location     maps.LatLng
	w            http.ResponseWriter
	r            *http.Request
	groupId      uint32
	respDataList []models.HandcupRespData
}

type saveResultsParms struct {
	w                 http.ResponseWriter
	r                 *http.Request
	location          maps.LatLng
	handcupIdResponse []models.HandcupIdResponse
	respDataList      []models.HandcupRespData
}

type reqData struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Distance  uint    `json:"distance"`
}

type fakeCoordinate struct {
	lat float64
	lng float64
}

var keyword = "飲料店"

func (server *Server) GetHandcupList(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var reqData reqData
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		log.Println("JSON Unmarshal:", err)
	}
	log.Println("Requset body 內容:", reqData)
	// 24.9927061, 121.4491151 24.9888971, 121.4481381
	HistoryRequest := models.HistoryRequest{}
	HistoryRequest.ReqLatitude = reqData.Latitude
	HistoryRequest.ReqLongitude = reqData.Longitude
	HistoryRequest.Distance = reqData.Distance

	hisReqResp, err := HistoryRequest.CheckHistoryReq(server.DB)
	if err != nil {
		panic(err)
	}

	log.Println("檢查資料庫歷史紀錄:", hisReqResp)
	if len(hisReqResp) == 0 {
		// 如果 HistoryRequest 內沒有紀錄
		handleMapParms := handleMapParms{w: w, r: r}
		handleMapParms.location.Lat = reqData.Latitude
		handleMapParms.location.Lng = reqData.Longitude
		handleMapParms.distance = 300 // !?暫定300m
		// handleMapParms.distance = reqData.Distance
		server.handleGoogleMap(handleMapParms)
	} else {
		// 如果 HistoryRequest 內有紀錄
		server.handleHistoryReq(saveResultsParms{w: w, r: r, location: maps.LatLng{Lat: reqData.Latitude, Lng: reqData.Longitude}, handcupIdResponse: hisReqResp})
	}
}

func (server *Server) handleGoogleMap(parms handleMapParms) {
	c := server.googleAPIAuth()

	location := &maps.LatLng{Lat: parms.location.Lat, Lng: parms.location.Lng}
	distance := parms.distance
	r := &maps.NearbySearchRequest{
		Location: location,
		Radius:   distance,
		Keyword:  keyword,
	}
	// log.Println(parms.location.Lat, parms.location.Lng, parms.distance)
	if len(parms.nextToken) != 0 {
		r.PageToken = parms.nextToken
	}
	// Call Google map API
	resp, err := c.NearbySearch(context.Background(), r)
	// log.Println(resp)
	if err != nil {
		log.Println("Request nearby search fatal error: %s", err)
	}

	handcupInfo := models.HandcupInfo{}
	HistoryRequest := models.HistoryRequest{}
	respDataList := []models.HandcupRespData{}
	HistoryRequest.ReqLatitude = parms.location.Lat
	HistoryRequest.ReqLongitude = parms.location.Lng
	latestGroupID := HistoryRequest.FindLatestGroupID(server.DB)

	// 如果有值就傳承下去
	if len(parms.respDataList) != 0 {
		respDataList = parms.respDataList
	}

	for _, s := range resp.Results {
		handcupInfo = server.handleHandcupInfoData(s) // 將 Google map API 的值塞入 handcupInfo 中
		handcupInfo.ID = handcupInfo.FindLatestID(server.DB) + 1

		// 處理要回傳給前端的資料
		respData := models.HandcupRespData{
			PlaceId:      handcupInfo.PlaceId,
			Name:         handcupInfo.Name,
			Latitude:     handcupInfo.Latitude,
			Longitude:    handcupInfo.Longitude,
			Rating:       handcupInfo.Rating,
			RatingsTotal: handcupInfo.RatingsTotal,
			ImageUrl:     handcupInfo.ImageUrl,
		}
		// 先去搜尋 DB 內有無飲料店資料，並取出 views 回傳
		FHIBPResp, FHIBPError := handcupInfo.FindHandcupInfoByPlaceID(server.DB, handcupInfo.PlaceId)
		if FHIBPError != nil {
			respData.Views = 1
		} else {
			respData.Views = FHIBPResp.Views + 1
		}
		respDataList = append(respDataList, respData) // 把資料塞進 respDataList 中

		// 將 Google map API 的資料存入 DB [handcup_infos]
		handcupInfoCreated, err := handcupInfo.SaveHandcupInfo(server.DB)
		if err != nil {
			formattedError := formaterror.FormatError(err.Error())
			responses.ERROR(parms.w, http.StatusInternalServerError, formattedError)
			return
		}

		latestHisReqID := HistoryRequest.FindLatestHisReqID(server.DB)
		HistoryRequest.InitData(latestHisReqID, latestGroupID+1, handcupInfoCreated.ID, r.Radius, r.Keyword)
		// 將 Google API 的資料存入 DB [history_requests]
		HistoryRequest.SaveHistoryReq(server.DB)
	}
	// Recall next page with nextToken
	if resp.NextPageToken != "" {
		server.handleGoogleMap(handleMapParms{
			nextToken:    resp.NextPageToken,
			location:     *location,
			distance:     distance,
			w:            parms.w,
			r:            parms.r,
			respDataList: respDataList,
		})
	} else {
		// parms.w.Header().Set("Access-Control-Allow-Origin", "*")
		responses.JSON(parms.w, http.StatusOK, respDataList)
	}
}

func (server *Server) handleUpdateGoogleMap(parms handleUpdateMapParms) {
	HistoryRequest := models.HistoryRequest{}
	handcupInfo := models.HandcupInfo{}

	g := HistoryRequest.GetGroupHisReqByGId(server.DB, parms.groupId)
	log.Println("重新要一次GOOGLE API! 經度: ", g.ReqLatitude, "緯度: ", g.ReqLongitude)
	r := &maps.NearbySearchRequest{
		Location: &maps.LatLng{Lat: g.ReqLatitude, Lng: g.ReqLongitude},
		Radius:   g.Distance,
		Keyword:  keyword,
	}
	if len(parms.nextToken) != 0 {
		r.PageToken = parms.nextToken
	}
	c := server.googleAPIAuth()
	resp, err := c.NearbySearch(context.Background(), r)
	if err != nil {
		log.Println("Request nearby search fatal error: %s", err)
	}

	respDataList := []models.HandcupRespData{}

	// 如果有值就傳承下去
	if len(parms.respDataList) != 0 {
		respDataList = parms.respDataList
	}

	for _, s := range resp.Results {
		handcupInfo = server.handleHandcupInfoData(s) // 將 Google map API 的值塞入 handcupInfo 中
		h, _ := handcupInfo.FindHandcupInfoByPlaceID(server.DB, s.PlaceID)
		// 處理要回傳給前端的資料
		respData := models.HandcupRespData{
			PlaceId:      handcupInfo.PlaceId,
			Name:         handcupInfo.Name,
			Latitude:     handcupInfo.Latitude,
			Longitude:    handcupInfo.Longitude,
			Rating:       handcupInfo.Rating,
			RatingsTotal: handcupInfo.RatingsTotal,
			Views:        h.Views, // DB 內的 views
			ImageUrl:     handcupInfo.ImageUrl,
		}
		respDataList = append(respDataList, respData) // 把資料塞進 respDataList 中

		// 如果資料庫內 [handcup_infos] 有這飲料店資訊
		if h.ID != 0 {
			handcupInfo.ID = h.ID
			handcupInfoUpdated, err := handcupInfo.UpdateAHandcupInfo(server.DB, h.ID)
			if err != nil {
				formattedError := formaterror.FormatError(err.Error())
				responses.ERROR(parms.w, http.StatusInternalServerError, formattedError)
				return
			}

			log.Println("我改飲料店資訊了喔", handcupInfoUpdated)
			HistoryRequest.ID = HistoryRequest.GetIDByHandcupID(server.DB, handcupInfoUpdated.ID)
			HistoryRequest.GroupId = g.GroupId
			HistoryRequest.HandcupId = handcupInfoUpdated.ID
			HistoryRequest.ReqLatitude = g.ReqLatitude
			HistoryRequest.ReqLongitude = g.ReqLongitude
			HistoryRequest.Distance = g.Distance
			HistoryRequest.Keyword = r.Keyword
			HistoryRequest.UpdateAHistoryRequest(server.DB, g.GroupId)
		} else {
			handcupInfo.ID = handcupInfo.FindLatestID(server.DB) + 1
			// 將 Google API 的資料存入 DB [handcup_infos]
			handcupInfoCreated, err := handcupInfo.SaveHandcupInfo(server.DB)
			if err != nil {
				formattedError := formaterror.FormatError(err.Error())
				responses.ERROR(parms.w, http.StatusInternalServerError, formattedError)
				return
			}

			log.Println("欸這一區有新的飲料店，已經新增了喔", handcupInfoCreated)
			HistoryRequest.ReqLatitude = parms.location.Lat  // 請求的緯度
			HistoryRequest.ReqLongitude = parms.location.Lng // 請求的經度
			latestHisReqID := HistoryRequest.FindLatestHisReqID(server.DB)
			HistoryRequest.InitData(latestHisReqID, g.GroupId, handcupInfoCreated.ID, r.Radius, r.Keyword)
			// 將 Google API 的資料存入 DB [history_requests]
			HistoryRequest.SaveHistoryReq(server.DB)
		}
	}

	// Recall next page with nextToken
	if resp.NextPageToken != "" {
		server.handleUpdateGoogleMap(handleUpdateMapParms{
			nextToken:    resp.NextPageToken,
			location:     parms.location,
			w:            parms.w,
			r:            parms.r,
			groupId:      parms.groupId,
			respDataList: respDataList,
		})
	} else {
		// parms.w.Header().Set("Access-Control-Allow-Origin", "*")
		responses.JSON(parms.w, http.StatusOK, respDataList)
	}
}

func (server *Server) handleHistoryReq(parms saveResultsParms) {
	handcupInfo := models.HandcupInfo{}
	var timeIsExpire bool = false
	_ = timeIsExpire
	respDataList := []models.HandcupRespData{}

	// 如果有值就傳承下去
	if len(parms.respDataList) != 0 {
		respDataList = parms.respDataList
	}

	maxUpdateTime := parms.handcupIdResponse[0].UpdateTime
	// To find max update_time from ReqHistory
	for _, h := range parms.handcupIdResponse {
		if h.UpdateTime.After(maxUpdateTime) {
			maxUpdateTime = h.UpdateTime
		}
	}
	log.Println("此 group 最新更新時間: ", maxUpdateTime)
	thresholdTime := maxUpdateTime.AddDate(0, 0, 7) // Add 7 days
	if time.Now().After(thresholdTime) {
		// 只要 ReqHistory 最新的 update_time 超過七天就需要更新資訊
		log.Println("Group_id: ", parms.handcupIdResponse[0].GroupId, "資料超過七天啦!")
		t := handleUpdateMapParms{location: parms.location, r: parms.r, w: parms.w, groupId: parms.handcupIdResponse[0].GroupId}
		server.handleUpdateGoogleMap(t) // Call handleUpdateGoogleMap func
	} else {
		// 反之直接撈 DB 資料
		for _, h := range parms.handcupIdResponse {
			log.Println("Group ID: ", h.GroupId)
			log.Println("Handcup ID: ", h.HandcupId)
			log.Println("Updata time: ", h.UpdateTime)
			resp, err := handcupInfo.FindHandcupInfoByID(server.DB, h.HandcupId) // Get handcup infomation by handcup_id
			if err != nil {
				log.Println("為甚麼會錯: ", err)
			}
			respDataList = append(respDataList, resp) // 把資料塞進 respDataList 中
		}
		responses.JSON(parms.w, http.StatusOK, respDataList)
	}
}

func (server *Server) handleHandcupInfoData(s maps.PlacesSearchResult) models.HandcupInfo {
	handcupInfo := models.HandcupInfo{
		GoogleId:     s.ID,
		PlaceId:      s.PlaceID,
		Name:         s.Name,
		Latitude:     s.Geometry.Location.Lat,
		Longitude:    s.Geometry.Location.Lng,
		Rating:       s.Rating,
		RatingsTotal: s.UserRatingsTotal,
		Views:        1, // 起始值
	}

	if s.Photos != nil {
		handcupInfo.ImageReference = s.Photos[0].PhotoReference
		handcupInfo.ImageWidth = s.Photos[0].Width
		handcupInfo.ImageHeight = s.Photos[0].Height
	} else {
		handcupInfo.ImageReference = ""
		handcupInfo.ImageWidth = 0
		handcupInfo.ImageHeight = 0
	}
	handcupInfo.ImageUrl = server.requestPhoto(handcupInfo.ImageReference)

	return handcupInfo
}

func (server *Server) requestPhoto(ref string) string {
	if ref == "" {
		return ""
	}

	key, err := server.loadGoogleKey()
	if err != nil {
		log.Println("Load API key fatal error: %s", err)
	}

	maxWidth := 400
	requsetBaseURL := "https://maps.googleapis.com/maps/api/place/photo?"
	requestURL := fmt.Sprintf("%smaxwidth=%d&photoreference=%s&key=%s", requsetBaseURL, maxWidth, ref, key)
	resp, err := http.Get(requestURL)
	if err != nil {
		log.Println("http.Get => %v", err.Error())
	}

	// The Request in the Response is the last URL the
	finalURL := resp.Request.URL.String()
	log.Println("The photo url you ended up at is: \n", finalURL)

	return finalURL
}

func (server *Server) GetPlaceDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	c := server.googleAPIAuth()
	d, err := c.PlaceDetails(context.Background(), &maps.PlaceDetailsRequest{PlaceID: vars["placeId"]})
	if err != nil {
		println("Place detail error:", err)
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
	}

	responses.JSON(w, http.StatusOK, d)
}
