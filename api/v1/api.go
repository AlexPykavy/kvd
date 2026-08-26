package v1

import (
	"encoding/json"
	"fmt"
	"kvd/api"
	"kvd/internal/store"

	"net/http"
)

type StoreHandler struct {
	s store.Store
}

func NewStoreHandler(s store.Store) *StoreHandler {
	return &StoreHandler{
		s: s,
	}
}

// Get retrieves an entry by its key.
//
//	@Summary		Get an entry
//	@Description	Retrieves a stored value by its key.
//	@Tags			keys
//	@Produce		json
//	@Param			key	path		string	true	"Entry key"
//	@Success		200	{object}	EntryResponse
//	@Failure		400	{object}	api.ErrorResponse	"Key is required"
//	@Failure		404	{object}	api.ErrorResponse	"Entry not found"
//	@Router			/keys/{key} [get]
func (h *StoreHandler) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if len(key) == 0 {
		api.WriteError(w, http.StatusBadRequest, "Key is required")
		return
	}

	value, err := h.s.Get(key)
	if err != nil {
		api.WriteError(w, http.StatusNotFound, fmt.Sprintf("An entry with the key %s not found", key))
		return
	}

	dto := &EntryResponse{
		Key:   key,
		Value: value,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(dto)
}

// Put creates or updates an entry.
//
//	@Summary		Create or update an entry
//	@Description	Stores a value under the specified key. If the key already exists, its value is replaced.
//	@Tags			keys
//	@Accept			json
//	@Produce		json
//	@Param			key		path	string			true	"Entry key"
//	@Param			entry	body	PutEntryRequest	true	"Entry to store"
//	@Success		200
//	@Failure		400	{object}	api.ErrorResponse	"Invalid request"
//	@Router			/keys/{key} [put]
func (h *StoreHandler) Put(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if len(key) == 0 {
		api.WriteError(w, http.StatusBadRequest, "Key is required")
		return
	}

	var dto PutEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		api.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Error decoding request body: %v", err))
		return
	}

	if len(dto.Value) == 0 {
		api.WriteError(w, http.StatusBadRequest, "Value is required")
		return
	}

	h.s.Put(key, dto.Value)

	w.WriteHeader(http.StatusOK)
}

// Delete removes an entry by its key.
//
//	@Summary		Delete an entry
//	@Description	Deletes the entry associated with the specified key.
//	@Tags			keys
//	@Param			key	path	string	true	"Entry key"
//	@Success		204
//	@Failure		400	{object}	api.ErrorResponse	"Key is required"
//	@Failure		404	{object}	api.ErrorResponse	"Entry not found"
//	@Router			/keys/{key} [delete]
func (h *StoreHandler) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if len(key) == 0 {
		api.WriteError(w, http.StatusBadRequest, "Key is required")
		return
	}

	if err := h.s.Delete(key); err != nil {
		api.WriteError(w, http.StatusNotFound, fmt.Sprintf("An entry with the key %s not found", key))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Count returns the number of stored entries.
//
//	@Summary		Count entries
//	@Description	Returns the total number of stored entries.
//	@Tags			keys
//	@Produce		json
//	@Success		200	{object}	CountResponse
//	@Router			/count [get]
func (h *StoreHandler) Count(w http.ResponseWriter, r *http.Request) {
	dto := &CountResponse{
		Size: h.s.Len(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(dto)
}
