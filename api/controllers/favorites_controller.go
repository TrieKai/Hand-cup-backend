package controllers

import (
	"encoding/json"
	"fmt"
	"handCup-project-backend/api/models"
	"handCup-project-backend/api/responses"
	formaterror "handCup-project-backend/api/utils"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type favReqData struct {
	PlaceID string `json:"placeId"`
	UserID  string `json:"userId"`
}

func (server *Server) GetFavorites(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	favorite := models.Favorites{}
	uid := vars["user_id"]
	favorites, err := favorite.GetFavorites(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, favorites)
}

func (server *Server) CreateFavorites(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	var reqData favReqData
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		log.Println("JSON Unmarshal:", err)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	favorites := models.Favorites{}
	log.Println("Requset body 內容:", reqData)
	favorites.InitData(reqData.PlaceID, reqData.UserID)
	favCreated, err := favorites.SaveFavorite(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.RequestURI, favCreated.ID))
	responses.JSON(w, http.StatusCreated, favCreated)
}

func (server *Server) DeleteFavorites(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	favorites := models.Favorites{}
	placeID := vars["place_id"]
	uid := vars["user_id"]
	_, err := favorites.DeleteFavorite(server.DB, placeID, uid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", uid))
	responses.JSON(w, http.StatusNoContent, "")
}
