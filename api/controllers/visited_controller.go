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

type visReqData struct {
	PlaceID string `json:"placeId"`
	UserID  string `json:"userId"`
}

func (server *Server) GetVisiteds(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	visited := models.Visited{}
	uid := vars["user_id"]
	visiteds, err := visited.GetVisiteds(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, visiteds)
}

func (server *Server) CreateVisited(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	var reqData visReqData
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		log.Println("JSON Unmarshal:", err)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	visited := models.Visited{}
	log.Println("Requset body 內容:", reqData)
	visited.InitData(reqData.PlaceID, reqData.UserID)
	visCreated, err := visited.SaveVisited(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.RequestURI, visCreated.ID))
	responses.JSON(w, http.StatusCreated, visCreated)
}

func (server *Server) DeleteVisited(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	visited := models.Visited{}
	placeID := vars["place_id"]
	uid := vars["user_id"]
	_, err := visited.DeleteVisited(server.DB, placeID, uid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", uid))
	responses.JSON(w, http.StatusNoContent, "")
}
