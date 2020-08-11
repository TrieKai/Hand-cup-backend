package responses

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Header struct {
		Status int `json:"status"`
	} `json:"header"`
	Body interface{} `json:"body"`
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	fmt.Println("JSON回覆:", data)
	// TODO: Response marshal json
	// respData := Response{}
	// respData.Header.Status = statusCode
	// respData.Body = data
	// jsondata, _ := json.Marshal(respData)
	// fmt.Println(jsondata)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func ERROR(w http.ResponseWriter, statusCode int, err error) {
	if err != nil {
		JSON(w, statusCode, struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
		return
	}
	JSON(w, http.StatusBadRequest, nil)
}
