// Package response holds whole logic associated with helper functions
package response

import (
	"encoding/json"
	"net/http"
)

func Send(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if status == http.StatusNoContent || data == nil {
		return
	}
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		panic(err)
	}
}
