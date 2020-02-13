package controllers

import (
	"context"
	"errors"
	"fmt"
	"handCup-project-backend/api/models"
	"handCup-project-backend/api/responses"
	formaterror "handCup-project-backend/api/utils"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"googlemaps.github.io/maps"
)

type handleMapParms struct {
	nextToken string
	w         http.ResponseWriter
	r         *http.Request
}

type saveResultsParms struct {
	results  []maps.PlacesSearchResult
	w        http.ResponseWriter
	r        *http.Request
	location maps.LatLng
	distance uint
}

func (server *Server) GetHandcupList(w http.ResponseWriter, r *http.Request) {
	HistoryRequest := models.HistoryRequest{}
	// fmt.Print(r)
	groupResults, err := HistoryRequest.FindAllHistoryRequests(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, groupResults)
	server.handleMap(handleMapParms{w: w, r: r})
}

func (server *Server) handleMap(parms handleMapParms) {
	key, err := server.loadGoogleKey()
	if err != nil {
		log.Fatalf("Load API key fatal error: %s", err)
	}

	c, err := maps.NewClient(maps.WithAPIKey(key))
	if err != nil {
		log.Fatalf("Connect client fatal error: %s", err)
	}

	r := &maps.NearbySearchRequest{
		Location: &maps.LatLng{Lat: 24.9888971, Lng: 121.4481381}, //24.9927061, 121.4491151 24.9888971, 121.4481381
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
		server.handleMap(handleMapParms{nextToken: resp.NextPageToken})
	}

	// If google response has data
	if len(resp.Results) != 0 {
		server.saveResults(saveResultsParms{results: resp.Results, w: parms.w, r: parms.r, location: *r.Location, distance: r.Radius})
	} else {
		return
	}
}

func (server *Server) saveResults(parms saveResultsParms) {
	handcupInfo := models.HandcupInfo{}
	HistoryRequest := models.HistoryRequest{}
	HistoryRequest.ReqLatitude = parms.location.Lat
	HistoryRequest.ReqLongitude = parms.location.Lng
	latestGroupID := HistoryRequest.FindLatestGroupID(server.DB)

	hisReqResp, err := HistoryRequest.CheckHistoryReq(server.DB)
	if err != nil {
		panic(err)
	}

	fmt.Println("檢查回傳值", hisReqResp)
	// 如果 HistoryRequest 內沒有紀錄
	if len(hisReqResp) == 0 {
		for _, s := range parms.results {
			handcupInfo.GoogleId = s.ID
			handcupInfo.PlaceId = s.PlaceID
			handcupInfo.Name = s.Name
			handcupInfo.Latitude = s.Geometry.Location.Lat
			handcupInfo.Longitude = s.Geometry.Location.Lng
			handcupInfo.Rating = s.Rating
			handcupInfo.ImageReference = s.Photos[0].PhotoReference
			handcupInfo.ImageWidth = s.Photos[0].Width
			handcupInfo.ImageHeight = s.Photos[0].Height
			handcupInfo.ImageUrl = server.requestPhoto(s.Photos[0].PhotoReference)
			server.requestPhoto(s.Photos[0].PhotoReference)
			// pretty.Println(handcupInfo)

			handcupInfoCreated, err := handcupInfo.SaveHandcupInfo(server.DB)
			if err != nil {
				formattedError := formaterror.FormatError(err.Error())
				responses.ERROR(parms.w, http.StatusInternalServerError, formattedError)
				return
			}

			latestID := handcupInfo.FindLatestID(server.DB)

			latestHisReqID := HistoryRequest.FindLatestHisReqID(server.DB)
			fmt.Print(latestHisReqID, latestGroupID, latestID)
			HistoryRequest.InitData(latestHisReqID, latestGroupID, latestID)

			HistoryRequest.SaveHistoryReq(server.DB)

			parms.w.Header().Set("Location", fmt.Sprintf("%s%s/%s\n", parms.r.Host, parms.r.RequestURI, handcupInfoCreated.GoogleId))
			responses.JSON(parms.w, http.StatusCreated, handcupInfoCreated)
		}
	} else {
		// TODO: Handle has history request
		for _, r := range hisReqResp {
			fmt.Println("Group ID:", r.GroupId)
			fmt.Println("Handcup ID:", r.HandcupId)
			resp, err := handcupInfo.FindHandcupInfoByID(server.DB, r.HandcupId) // Get handcup info by id

			if err != nil {
				fmt.Println("為甚麼會錯:", err)
			}
			fmt.Println("飲料店資料抓到你啦:", resp)
		}
	}
}

func (server *Server) requestPhoto(ref string) string {
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
