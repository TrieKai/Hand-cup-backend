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
}

type respData struct {
	placeId   string  `json:"place_id"`
	name      string  `json:"name"`
	latitude  float64 `json:"latitude"`
	longitude float64 `json:"longitude"`
	rating    float32 `json:"rating"`
	imageUrl  string  `json:"image_url"`
}

type respDataList []respData

type fakeCoordinate struct {
	lat float64
	lng float64
}

func (server *Server) GetHandcupList(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println("Requset body內容", body, reqData)
	//24.9927061, 121.4491151 24.9888971, 121.4481381
	var fakeCoordinate fakeCoordinate
	fakeCoordinate.lat = 24.9927061
	fakeCoordinate.lng = 121.4491151

	HistoryRequest := models.HistoryRequest{}
	HistoryRequest.ReqLatitude = fakeCoordinate.lat
	HistoryRequest.ReqLongitude = fakeCoordinate.lng
	// HistoryRequest.ReqLatitude = reqData.Latitude
	// HistoryRequest.ReqLongitude = reqData.Longitude

	hisReqResp, err := HistoryRequest.CheckHistoryReq(server.DB)
	if err != nil {
		panic(err)
	}

	fmt.Println("檢查資料庫歷史紀錄:", hisReqResp)
	if len(hisReqResp) == 0 {
		// 如果 HistoryRequest 內沒有紀錄
		handleMapParms := handleMapParms{w: w, r: r}
		handleMapParms.location.Lat = fakeCoordinate.lat
		handleMapParms.location.Lng = fakeCoordinate.lng
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

	r := &maps.NearbySearchRequest{
		Location: &maps.LatLng{Lat: parms.location.Lat, Lng: parms.location.Lng},
		Radius:   10,
		Keyword:  "飲料店",
	}
	if len(parms.nextToken) != 0 {
		r.PageToken = parms.nextToken
	}
	resp, err := c.NearbySearch(context.Background(), r)
	if err != nil {
		log.Fatalf("Request nearby search fatal error: %s", err)
	}

	// Recall next page with nextToken
	if resp.NextPageToken != "" {
		server.handleGoogleMap(handleMapParms{nextToken: resp.NextPageToken})
	}

	handcupInfo := models.HandcupInfo{}
	HistoryRequest := models.HistoryRequest{}
	HistoryRequest.ReqLatitude = parms.location.Lat
	HistoryRequest.ReqLongitude = parms.location.Lng
	latestGroupID := HistoryRequest.FindLatestGroupID(server.DB)

	for _, s := range resp.Results {
		handcupInfo.ID = handcupInfo.FindLatestID(server.DB) + 1
		handcupInfo.GoogleId = s.ID
		handcupInfo.PlaceId = s.PlaceID
		handcupInfo.Name = s.Name
		handcupInfo.Latitude = s.Geometry.Location.Lat
		handcupInfo.Longitude = s.Geometry.Location.Lng
		handcupInfo.Rating = s.Rating
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

		// 將 Google API 的資料存入 DB [handcup_infos]
		handcupInfoCreated, err := handcupInfo.SaveHandcupInfo(server.DB)
		if err != nil {
			formattedError := formaterror.FormatError(err.Error())
			responses.ERROR(parms.w, http.StatusInternalServerError, formattedError)
			return
		}

		latestHisReqID := HistoryRequest.FindLatestHisReqID(server.DB)
		HistoryRequest.InitData(latestHisReqID, latestGroupID+1, handcupInfoCreated.ID, r.Radius)
		// 將 Google API 的資料存入 DB [history_requests]
		HistoryRequest.SaveHistoryReq(server.DB)

		parms.w.Header().Set("Location", fmt.Sprintf("%s%s/%s\n", parms.r.Host, parms.r.RequestURI, handcupInfoCreated.GoogleId))
		// responses.JSON(parms.w, http.StatusOK, handcupInfoCreated)
		responses.JSON(parms.w, http.StatusCreated, handcupInfoCreated)
	}
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
		Radius:   10,
		Keyword:  "飲料店",
	}
	if len(parms.nextToken) != 0 {
		r.PageToken = parms.nextToken
	}
	resp, err := c.NearbySearch(context.Background(), r)
	if err != nil {
		log.Fatalf("Request nearby search fatal error: %s", err)
	}

	// Recall next page with nextToken
	if resp.NextPageToken != "" {
		server.handleUpdateGoogleMap(handleUpdateMapParms{nextToken: resp.NextPageToken})
	}

	for _, s := range resp.Results {
		h, err := handcupInfo.FindHandcupInfoByPlaceID(server.DB, s.PlaceID)
		if err != nil {
			log.Fatal(err)
		}

		handcupInfo.GoogleId = s.ID
		handcupInfo.Name = s.Name
		handcupInfo.Latitude = s.Geometry.Location.Lat
		handcupInfo.Longitude = s.Geometry.Location.Lng
		handcupInfo.Rating = s.Rating
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
			parms.w.Header().Set("Location", fmt.Sprintf("%s%s/%s\n", parms.r.Host, parms.r.RequestURI, handcupInfoUpdated.GoogleId))
			// responses.JSON(parms.w, http.StatusOK, handcupInfoCreated)
			responses.JSON(parms.w, http.StatusCreated, handcupInfoUpdated)
		} else {
			handcupInfo.ID = handcupInfo.FindLatestID(server.DB) + 1
			// 將 Google API 的資料存入 DB [handcup_infos]
			handcupInfoCreated, err := handcupInfo.SaveHandcupInfo(server.DB)
			if err != nil {
				formattedError := formaterror.FormatError(err.Error())
				responses.ERROR(parms.w, http.StatusInternalServerError, formattedError)
				return
			}
			latestHisReqID := HistoryRequest.FindLatestHisReqID(server.DB)
			HistoryRequest.InitData(latestHisReqID, g.GroupId, handcupInfoCreated.ID, r.Radius)
			// 將 Google API 的資料存入 DB [history_requests]
			HistoryRequest.SaveHistoryReq(server.DB)
			parms.w.Header().Set("Location", fmt.Sprintf("%s%s/%s\n", parms.r.Host, parms.r.RequestURI, handcupInfoCreated.GoogleId))
			// responses.JSON(parms.w, http.StatusOK, handcupInfoCreated)
			responses.JSON(parms.w, http.StatusCreated, handcupInfoCreated)
		}
	}
}

func (server *Server) handleHistoryReq(parms saveResultsParms) {
	handcupInfo := models.HandcupInfo{}
	var timeIsExpire bool = false
	_ = timeIsExpire

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
			parms.w.Header().Set("Location", fmt.Sprintf("%s%s/%s\n", parms.r.Host, parms.r.RequestURI, resp.GoogleId))
			responses.JSON(parms.w, http.StatusOK, resp)
		}
	}

	// 如果資料過期
	if timeIsExpire {
		t := handleUpdateMapParms{r: parms.r, w: parms.w, groupId: parms.handcupIdResponse[0].GroupId}
		server.handleUpdateGoogleMap(t) // Call handleUpdateGoogleMap func
	}
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
