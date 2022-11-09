package responses

import (
	"encoding/json"
	"net/http"
)

type OkResponseStructure struct {
	Message string
	Data    interface{}
}

func JsonResponse(w http.ResponseWriter, v interface{}, statusOptional ...int) {
	status := 200
	if len(statusOptional) > 0 {
		status = statusOptional[0]
	}

	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(js)
}
