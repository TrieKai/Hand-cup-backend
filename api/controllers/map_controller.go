package controllers

import (
	"net/http"

	"handCup-project-backend/api/models"
	"handCup-project-backend/api/responses"
)

func (server *Server) GetHandcupList(w http.ResponseWriter, r *http.Request) {
	HistoryResult := models.HistoryResult{}

	groupResults, err := HistoryResult.FindAllHistoryResults(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, groupResults)
}
