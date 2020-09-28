package controllers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"handCup-project-backend/api/models"
	"handCup-project-backend/api/responses"
)

func (server *Server) SignUp(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	err = user.Validate("")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	err = server.DB.Debug().Model(&models.User{}).Create(&user).Error
	if err != nil {
		log.Println("Cannot create user: %s", err)
	}
}
