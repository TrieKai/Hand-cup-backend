package controllers

import (
	"context"
	"errors"
	"handCup-project-backend/api/models"
	"handCup-project-backend/api/responses"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/kr/pretty"
	"googlemaps.github.io/maps"
)

func (server *Server) GetHandcupList(w http.ResponseWriter, r *http.Request) {
	HistoryRequest := models.HistoryRequest{}

	groupResults, err := HistoryRequest.FindAllHistoryRequests(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, groupResults)
	server.HandleMap()
}

func (server *Server) HandleMap(nextToken ...string) {
	key, err := server.LoadGoogleKey()
	if err != nil {
		log.Fatalf("Load API key fatal error: %s", err)
	}

	c, err := maps.NewClient(maps.WithAPIKey(key))
	if err != nil {
		log.Fatalf("Connect client fatal error: %s", err)
	}

	r := &maps.NearbySearchRequest{
		Location: &maps.LatLng{Lat: 24.988669, Lng: 121.448312},
		Radius:   5,
		Keyword:  "飲料店",
	}
	if len(nextToken) != 0 {
		r.PageToken = nextToken[0]
	}
	resp, err := c.NearbySearch(context.Background(), r)
	if err != nil {
		log.Fatalf("Request nearby search fatal error: %s", err)
	}

	if resp.NextPageToken != "" {
		server.HandleMap(resp.NextPageToken)
	}

	server.SaveResults(resp.Results)
}

func (server *Server) SaveResults(results []maps.PlacesSearchResult) {

	for _, s := range results {
		// handcupInfo := models.HandcupInfo{
		// 	GoogleId:       s.PlaceID,
		// 	Name:           s.Name,
		// 	Latitude:       s.Geometry.Location.Lat,
		// 	Longitude:      s.Geometry.Location.Lng,
		// 	Rating:         s.Rating,
		// 	ImageReference: s.Photos[0].PhotoReference,
		// 	ImageWidth:     s.Photos[0].Width,
		// 	ImageHeigth:    s.Photos[0].Height,
		// 	// ImageUrl:       server.RequestPhoto(s.Photos[0].PhotoReference),
		// }
		server.RequestPhoto(s.Photos[0].PhotoReference)
		pretty.Println(s)
		// handcupInfoCreated, err := handcupInfo.SaveHandcupInfo(server.DB)
	}
}

func (server *Server) RequestPhoto(ref string) {
	key, err := server.LoadGoogleKey()
	if err != nil {
		log.Fatalf("Load API key fatal error: %s", err)
	}

	c, err := maps.NewClient(maps.WithAPIKey(key))
	if err != nil {
		log.Fatalf("Connect client fatal error: %s", err)
	}

	r := &maps.PlacePhotoRequest{
		PhotoReference: ref,
		MaxWidth:       400,
	}

	resp, err := c.PlacePhoto(context.Background(), r)

	if err != nil {
		log.Fatalf("Request photo fatal error: %s", err)
	}

	pretty.Println(resp.ContentType)
	resp.Data.Close()
}

func (server *Server) LoadGoogleKey() (string, error) {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, not comming through %v", err)
		return "", errors.New("maps: destination missing")
	}

	return os.Getenv("GOOGLE_MAP_API_KEY"), nil
}
