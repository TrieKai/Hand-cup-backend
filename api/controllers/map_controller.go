package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"handCup-project-backend/api/models"
	"handCup-project-backend/api/responses"
	formaterror "handCup-project-backend/api/utils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"googlemaps.github.io/maps"
)

type handleMapParms struct {
	nextToken string
	location  maps.LatLng
	distance  uint
	w         http.ResponseWriter
	r         *http.Request
}

type handleUpdateMapParms struct {
	nextToken string
	w         http.ResponseWriter
	r         *http.Request
	groupId   uint32
}

type saveResultsParms struct {
	w                 http.ResponseWriter
	r                 *http.Request
	handcupIdResponse []models.HandcupIdResponse
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

var respDataList = []models.HandcupRespData{} // 要回傳的總資料群

func (server *Server) GetHandcupList(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		fmt.Println("OPTIONS")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Expose-Headers", "*")
		responses.JSON(w, http.StatusOK, "success")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var reqData reqData
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Requset body 內容:", reqData)
	//24.9927061, 121.4491151 24.9888971, 121.4481381
	HistoryRequest := models.HistoryRequest{}
	HistoryRequest.ReqLatitude = reqData.Latitude
	HistoryRequest.ReqLongitude = reqData.Longitude
	HistoryRequest.Distance = reqData.Distance

	hisReqResp, err := HistoryRequest.CheckHistoryReq(server.DB)
	if err != nil {
		panic(err)
	}

	fmt.Println("檢查資料庫歷史紀錄:", hisReqResp)
	if len(hisReqResp) == 0 {
		// 如果 HistoryRequest 內沒有紀錄
		handleMapParms := handleMapParms{w: w, r: r}
		handleMapParms.location.Lat = reqData.Latitude
		handleMapParms.location.Lng = reqData.Longitude
		handleMapParms.distance = 300 // !?暫定500m
		// handleMapParms.distance = reqData.Distance
		server.handleGoogleMap(handleMapParms)
	} else {
		// 如果 HistoryRequest 內有紀錄
		server.handleHistoryReq(saveResultsParms{w: w, r: r, handcupIdResponse: hisReqResp})
	}
}

func (server *Server) handleGoogleMap(parms handleMapParms) {
	key, err := server.loadGoogleKey()
	if err != nil {
		log.Fatalf("Load API key fatal error: %s", err)
	}

	c, err := maps.NewClient(maps.WithAPIKey(key))
	if err != nil {
		log.Fatalf("Connect client fatal error: %s", err)
	}

	location := &maps.LatLng{Lat: parms.location.Lat, Lng: parms.location.Lng}
	distance := parms.distance
	r := &maps.NearbySearchRequest{
		Location: location,
		Radius:   distance,
		Keyword:  "飲料店",
	}
	// fmt.Println(parms.location.Lat, parms.location.Lng, parms.distance)
	if len(parms.nextToken) != 0 {
		r.PageToken = parms.nextToken
	}
	// Call Google map API
	resp, err := c.NearbySearch(context.Background(), r)
	// fmt.Println(resp)
	if err != nil {
		log.Fatalf("Request nearby search fatal error: %s", err)
	}

	handcupInfo := models.HandcupInfo{}
	HistoryRequest := models.HistoryRequest{}
	// respDataList := []models.HandcupRespData{}
	HistoryRequest.ReqLatitude = parms.location.Lat
	HistoryRequest.ReqLongitude = parms.location.Lng
	latestGroupID := HistoryRequest.FindLatestGroupID(server.DB)

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
			Views:        1,
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
		server.handleGoogleMap(handleMapParms{location: *location, distance: distance, nextToken: resp.NextPageToken})
	}

	// parms.w.Header().Set("Access-Control-Allow-Origin", "*")
	responses.JSON(parms.w, http.StatusCreated, respDataList)
}

func (server *Server) handleUpdateGoogleMap(parms handleUpdateMapParms) {
	HistoryRequest := models.HistoryRequest{}
	handcupInfo := models.HandcupInfo{}

	key, err := server.loadGoogleKey()
	if err != nil {
		log.Fatalf("Load API key fatal error: %s", err)
	}

	c, err := maps.NewClient(maps.WithAPIKey(key))
	if err != nil {
		log.Fatalf("Connect client fatal error: %s", err)
	}

	g := HistoryRequest.GetGroupHisReqByGId(server.DB, parms.groupId)
	fmt.Println("重新要一次GOOGLE API! 經度:", g.ReqLatitude)
	fmt.Println("重新要一次GOOGLE API! 緯度:", g.ReqLongitude)
	r := &maps.NearbySearchRequest{
		Location: &maps.LatLng{Lat: g.ReqLatitude, Lng: g.ReqLongitude},
		Radius:   g.Distance,
		Keyword:  "飲料店",
	}
	if len(parms.nextToken) != 0 {
		r.PageToken = parms.nextToken
	}
	resp, err := c.NearbySearch(context.Background(), r)
	if err != nil {
		log.Fatalf("Request nearby search fatal error: %s", err)
	}

	// respDataList := []models.HandcupRespData{}

	for _, s := range resp.Results {
		h, err := handcupInfo.FindHandcupInfoByPlaceID(server.DB, s.PlaceID)
		if err != nil {
			log.Fatal(err)
		}

		handcupInfo = server.handleHandcupInfoData(s) // 將 Google map API 的值塞入 handcupInfo 中

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
		respDataList = append(respDataList, respData) // 把資料塞進 respDataList 中

		// 如果資料庫內有這筆資訊
		if h.ID != 0 {
			handcupInfo.ID = h.ID
			handcupInfoUpdated, err := handcupInfo.UpdateAHandcupInfo(server.DB, h.ID)
			if err != nil {
				formattedError := formaterror.FormatError(err.Error())
				responses.ERROR(parms.w, http.StatusInternalServerError, formattedError)
				return
			}

			fmt.Println("我改飲料店資訊了喔", handcupInfoUpdated)
			HistoryRequest.GroupId = g.GroupId
			HistoryRequest.ID = handcupInfoUpdated.ID
			HistoryRequest.UpdateAHistoryRequest(server.DB)
		} else {
			handcupInfo.ID = handcupInfo.FindLatestID(server.DB) + 1
			// 將 Google API 的資料存入 DB [handcup_infos]
			handcupInfoCreated, err := handcupInfo.SaveHandcupInfo(server.DB)
			if err != nil {
				formattedError := formaterror.FormatError(err.Error())
				responses.ERROR(parms.w, http.StatusInternalServerError, formattedError)
				return
			}

			fmt.Println("欸這一區有新的飲料店，已經新增了喔", handcupInfoCreated)
			latestHisReqID := HistoryRequest.FindLatestHisReqID(server.DB)
			HistoryRequest.InitData(latestHisReqID, g.GroupId, handcupInfoCreated.ID, r.Radius, r.Keyword)
			// 將 Google API 的資料存入 DB [history_requests]
			HistoryRequest.SaveHistoryReq(server.DB)
		}
	}

	// Recall next page with nextToken
	if resp.NextPageToken != "" {
		server.handleUpdateGoogleMap(handleUpdateMapParms{nextToken: resp.NextPageToken})
	}

	// parms.w.Header().Set("Access-Control-Allow-Origin", "*")
	responses.JSON(parms.w, http.StatusCreated, respDataList)
}

func (server *Server) handleHistoryReq(parms saveResultsParms) {
	handcupInfo := models.HandcupInfo{}
	var timeIsExpire bool = false
	_ = timeIsExpire
	// respDataList := []models.HandcupRespData{}

	for _, h := range parms.handcupIdResponse {
		thresholdTime := h.UpdateTime.AddDate(0, 0, 7) // Add 7 days
		// 設定超過七天需要更新資訊
		if time.Now().After(thresholdTime) {
			fmt.Println("這筆資料超過七天啦! ID:", h.HandcupId)
			timeIsExpire = true
		} else {
			fmt.Println("Group ID:", h.GroupId)
			fmt.Println("Handcup ID:", h.HandcupId)
			fmt.Println("Updata time:", h.UpdateTime)
			resp, err := handcupInfo.FindHandcupInfoByID(server.DB, h.HandcupId) // Get handcup infomation by handcup_id
			if err != nil {
				fmt.Println("為甚麼會錯:", err)
			}

			fmt.Println("飲料店資料抓到你啦:", resp)
			respDataList = append(respDataList, resp) // 把資料塞進 respDataList 中
		}
	}

	// 如果資料過期
	if timeIsExpire {
		t := handleUpdateMapParms{r: parms.r, w: parms.w, groupId: parms.handcupIdResponse[0].GroupId}
		server.handleUpdateGoogleMap(t) // Call handleUpdateGoogleMap func
	}

	// parms.w.Header().Set("Access-Control-Allow-Origin", "*")
	responses.JSON(parms.w, http.StatusOK, respDataList)
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
		log.Fatalf("Load API key fatal error: %s", err)
	}

	maxWidth := 400
	requsetBaseURL := "https://maps.googleapis.com/maps/api/place/photo?"
	requestURL := fmt.Sprintf("%smaxwidth=%d&photoreference=%s&key=%s", requsetBaseURL, maxWidth, ref, key)
	resp, err := http.Get(requestURL)
	if err != nil {
		log.Fatalf("http.Get => %v", err.Error())
	}

	// The Request in the Response is the last URL the
	finalURL := resp.Request.URL.String()
	fmt.Printf("The photo url you ended up at is: %v\n", finalURL)

	return finalURL
}

func (server *Server) loadGoogleKey() (string, error) {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, not comming through %v", err)
		return "", errors.New("maps: destination missing")
	}

	return os.Getenv("GOOGLE_MAP_API_KEY"), nil
}
