package handlers

import (
	"download-server/cmd/utils"
	"download-server/db"
	"net/http"
)

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.CreateToast(w, "error", "Missing task ID")
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()

	_, err := queries.DeleteTaskByID(r.Context(), id)
	if err != nil {
		utils.CreateToast(w, "error", "Failed to delete task: "+err.Error())
		http.Error(w, "Failed to delete task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.CreateToast(w, "success", "Successfully deleted task")
	w.WriteHeader(http.StatusOK)
}
