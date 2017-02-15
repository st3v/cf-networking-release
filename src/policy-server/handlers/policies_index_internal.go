package handlers

import (
	"lib/marshal"
	"net/http"
	"policy-server/models"
	"strings"

	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter -o fakes/store.go --fake-name Store . store
type store interface {
	All() ([]models.Policy, error)
	Create([]models.Policy) error
	Delete([]models.Policy) error
	Tags() ([]models.Tag, error)
	PoliciesWithFilter(models.PoliciesFilter) ([]models.Policy, error)
}

type PoliciesIndexInternal struct {
	Logger        lager.Logger
	Store         store
	Marshaler     marshal.Marshaler
	ErrorResponse errorResponse
}

func (h *PoliciesIndexInternal) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.Logger.Debug("internal request made to list policies", lager.Data{"URL": req.URL, "RemoteAddr": req.RemoteAddr})

	queryValues := req.URL.Query()
	var ids []string
	idList, ok := queryValues["id"]
	if ok {
		ids = strings.Split(idList[0], ",")
	}

	policies, err := h.Store.PoliciesWithFilter(models.PoliciesFilter{
		SourceGuids:      ids,
		DestinationGuids: ids,
	})

	if err != nil {
		h.ErrorResponse.InternalServerError(w, err, "policies-index-internal", "database read failed")
		return
	}

	policyResponse := struct {
		Policies []models.Policy `json:"policies"`
	}{policies}
	bytes, err := h.Marshaler.Marshal(policyResponse)
	if err != nil {
		h.ErrorResponse.InternalServerError(w, err, "policies-index-internal", "database marshaling failed")
		return
	}

	w.Write(bytes)
}
