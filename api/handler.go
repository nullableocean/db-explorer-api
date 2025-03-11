package api

import (
	"db_explorer/dbexplorer"
	"db_explorer/pkg/router"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type ExplorerHandler struct {
	explorer dbexplorer.SqlExplorer
}

func NewExplorerHandler(dbexp dbexplorer.SqlExplorer) *ExplorerHandler {
	return &ExplorerHandler{
		explorer: dbexp,
	}
}

func (h *ExplorerHandler) RegisterRoutes(router *router.MuxRouter) {
	router.Route("GET", "/", h.GetTables)
	router.Route("GET", "/{table}/", h.GetRecords)
	router.Route("GET", "/{table}/{id}/", h.GetRecord)
	router.Route("PUT", "/{table}/", h.CreateRecord)
	router.Route("POST", "/{table}/{id}/", h.UpdateRecord)
	router.Route("DELETE", "/{table}/{id}/", h.DeleteRecord)
}

type TablesResponse struct {
	Tables []string `json:"tables"`
}

// GET /
func (h *ExplorerHandler) GetTables(w http.ResponseWriter, r *http.Request) {
	tables, err := h.explorer.GetTables()
	if err != nil {
		h.errorResponse(w, "server error", http.StatusInternalServerError)
		return
	}

	tableRes := &TablesResponse{Tables: tables}
	response := map[string]*TablesResponse{"response": tableRes}

	json.NewEncoder(w).Encode(response)
}

type RecordsResponse struct {
	Records []map[string]interface{} `json:"records"`
}

// GET /$table?limit=5&offset=7
func (h *ExplorerHandler) GetRecords(w http.ResponseWriter, r *http.Request) {
	table := router.PathValue(r, "table")
	if !h.explorer.HasTable(table) {
		h.errorResponse(w, "unknown table", http.StatusNotFound)
		return
	}

	limit := 5
	offset := 0

	vals := r.URL.Query()
	if vals.Has("limit") {
		lim, err := strconv.Atoi(vals.Get("limit"))
		if err == nil {
			limit = lim
		}
	}
	if vals.Has("offset") {
		offs, err := strconv.Atoi(vals.Get("offset"))
		if err == nil {
			offset = offs
		}
	}

	recs, err := h.explorer.GetRecords(table, offset, limit)
	if err != nil {
		fmt.Println(err)
		h.errorResponse(w, "server error", http.StatusInternalServerError)
		return
	}

	rr := &RecordsResponse{Records: recs}
	response := map[string]*RecordsResponse{"response": rr}
	json.NewEncoder(w).Encode(response)
}

type RecordResponse struct {
	Record map[string]interface{} `json:"record"`
}

// GET /$table/$id
func (h *ExplorerHandler) GetRecord(w http.ResponseWriter, r *http.Request) {
	table := router.PathValue(r, "table")
	if !h.explorer.HasTable(table) {
		h.errorResponse(w, "unknown table", http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(router.PathValue(r, "id"))
	if err != nil {
		h.errorResponse(w, "invalid param: id. expect number", http.StatusBadRequest)
		return
	}

	record, err := h.explorer.GetRecord(table, id)
	if err != nil {
		if errors.Is(err, dbexplorer.ErrRecordNotFound) {
			h.errorResponse(w, err.Error(), http.StatusNotFound)
		} else {
			fmt.Println(err)
			h.errorResponse(w, "server error", http.StatusInternalServerError)
		}
		return
	}

	recResponse := &RecordResponse{Record: record}
	response := map[string]*RecordResponse{"response": recResponse}
	json.NewEncoder(w).Encode(response)
}

type CreateRecordResponse struct {
	Id int `json:"id"`
}

// PUT /$table body=formdata
func (h *ExplorerHandler) CreateRecord(w http.ResponseWriter, r *http.Request) {
	table := router.PathValue(r, "table")
	if !h.explorer.HasTable(table) {
		h.errorResponse(w, "unknown table", http.StatusNotFound)
		return
	}

	var body map[string]interface{}
	json.NewDecoder(r.Body).Decode(&body)

	err := h.explorer.ValidateCreateData(table, body)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.explorer.CreateRecord(table, body)
	if err != nil {
		fmt.Println(err)
		h.errorResponse(w, "server error", http.StatusNotFound)
		return
	}

	createResponse := CreateRecordResponse{Id: id}
	response := map[string]*CreateRecordResponse{"response": &createResponse}
	json.NewEncoder(w).Encode(response)
}

type UpdateReponse struct {
	Updated int `json:"updated"`
}

// POST /$table/$id
func (h *ExplorerHandler) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	table := router.PathValue(r, "table")
	if !h.explorer.HasTable(table) {
		h.errorResponse(w, "unknown table", http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(router.PathValue(r, "id"))
	if err != nil {
		h.errorResponse(w, "invalid param: id. expect number", http.StatusBadRequest)
		return
	}

	var body map[string]interface{}
	json.NewDecoder(r.Body).Decode(&body)

	err = h.explorer.ValidateUpdateData(table, body)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := h.explorer.UpdateRecord(table, id, body)
	if err != nil {
		if errors.Is(err, dbexplorer.ErrRecordNotFound) {
			h.errorResponse(w, err.Error(), http.StatusNotFound)
		} else {
			fmt.Println(err)
			h.errorResponse(w, "server error", http.StatusInternalServerError)
		}
		return
	}

	updateResponse := UpdateReponse{Updated: updated}
	response := map[string]*UpdateReponse{"response": &updateResponse}
	json.NewEncoder(w).Encode(response)
}

type DeleteReponse struct {
	Deleted int `json:"deleted"`
}

// DELETE /$table/$id
func (h *ExplorerHandler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	table := router.PathValue(r, "table")
	if !h.explorer.HasTable(table) {
		h.errorResponse(w, "unknown table", http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(router.PathValue(r, "id"))
	if err != nil {
		h.errorResponse(w, "invalid param: id. expect number", http.StatusBadRequest)
		return
	}

	deleted, err := h.explorer.DeleteRecord(table, id)
	if err != nil {
		if errors.Is(err, dbexplorer.ErrRecordNotFound) {
			h.errorResponse(w, err.Error(), http.StatusNotFound)
		} else {
			fmt.Println(err)
			h.errorResponse(w, "server error", http.StatusInternalServerError)
		}
		return
	}

	deletedResponse := DeleteReponse{Deleted: deleted}
	response := map[string]*DeleteReponse{"response": &deletedResponse}
	json.NewEncoder(w).Encode(response)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *ExplorerHandler) errorResponse(w http.ResponseWriter, errorMsg string, code int) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: errorMsg})
}
