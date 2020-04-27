package responses

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	header struct {
		status int
	}
	body interface{}
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	fmt.Println("抓到JSON囉:", data)
	respData := Response{}
	respData.header.status = statusCode
	respData.body = data
	// TODO: Response marshal json
	jsondata, _ := json.Marshal(respData)
	fmt.Println(jsondata)
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
