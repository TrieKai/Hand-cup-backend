package responses

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Header struct {
		Status string `json:"status"`
	} `json:"header"`
	Body struct {
		Data interface{} `json:"data"`
	} `json:"body"`
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	respData := Response{}
	respData.Header.Status = "success"
	respData.Body.Data = data
	fmt.Println("JSON回覆:", respData)
	err := json.NewEncoder(w).Encode(respData)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func ERROR(w http.ResponseWriter, statusCode int, err error) {
	respData := Response{}
	respData.Header.Status = "error"
	if err != nil {
		w.WriteHeader(statusCode)
		respData.Body.Data = struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		}
		err = json.NewEncoder(w).Encode(respData)
		if err != nil {
			fmt.Fprintf(w, "%s", err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	respData.Body.Data = nil
	err = json.NewEncoder(w).Encode(respData)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}
