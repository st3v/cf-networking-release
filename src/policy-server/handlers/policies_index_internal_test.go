package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"policy-server/handlers"
	"policy-server/handlers/fakes"
	"policy-server/models"

	lfakes "lib/fakes"

	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("PoliciesIndexInternal", func() {
	var (
		handler           *handlers.PoliciesIndexInternal
		resp              *httptest.ResponseRecorder
		fakeStore         *fakes.Store
		fakeErrorResponse *fakes.ErrorResponse
		logger            *lagertest.TestLogger
		marshaler         *lfakes.Marshaler
	)

	BeforeEach(func() {
		allPolicies := []models.Policy{{
			Source: models.Source{ID: "some-app-guid"},
			Destination: models.Destination{
				ID:       "some-other-app-guid",
				Protocol: "tcp",
				Port:     8080,
			},
		}}

		marshaler = &lfakes.Marshaler{}
		marshaler.MarshalStub = json.Marshal
		fakeStore = &fakes.Store{}
		fakeStore.PoliciesWithFilterReturns(allPolicies, nil)
		logger = lagertest.NewTestLogger("test")
		fakeErrorResponse = &fakes.ErrorResponse{}
		handler = &handlers.PoliciesIndexInternal{
			Logger:        logger,
			Store:         fakeStore,
			Marshaler:     marshaler,
			ErrorResponse: fakeErrorResponse,
		}
		resp = httptest.NewRecorder()
	})

	It("it returns the policies returned by PoliciesWithFilter", func() {
		expectedResponseJSON := `{"policies": [
				{
					"source": {
						"id": "some-app-guid"
					},
					"destination": {
						"id": "some-other-app-guid",
						"protocol": "tcp",
						"port": 8080
					}
				}
			]}`
		request, err := http.NewRequest("GET", "/networking/v0/internal/policies?id=some-app-guid", nil)
		Expect(err).NotTo(HaveOccurred())

		request.RemoteAddr = "some-host:some-port"

		handler.ServeHTTP(resp, request)
		Expect(logger).To(gbytes.Say("internal request made to list policies.*RemoteAddr.*some-host:some-port.*URL.*/networking/v0/internal/policies"))

		Expect(fakeStore.PoliciesWithFilterCallCount()).To(Equal(1))
		Expect(fakeStore.PoliciesWithFilterArgsForCall(0)).To(Equal(
			models.PoliciesFilter{
				SourceGuids:      []string{"some-app-guid"},
				DestinationGuids: []string{"some-app-guid"},
			},
		))
		Expect(resp.Code).To(Equal(http.StatusOK))
		Expect(resp.Body).To(MatchJSON(expectedResponseJSON))
	})

	Context("when there are policies and no filter is passed", func() {
		It("it calls PoliciesWithFilter with an empty filter", func() {
			expectedResponseJSON := `{"policies": [
				{
					"source": {
						"id": "some-app-guid"
					},
					"destination": {
						"id": "some-other-app-guid",
						"protocol": "tcp",
						"port": 8080
					}
				}
			]}`
			request, err := http.NewRequest("GET", "/networking/v0/internal/policies", nil)
			Expect(err).NotTo(HaveOccurred())
			handler.ServeHTTP(resp, request)

			Expect(fakeStore.PoliciesWithFilterCallCount()).To(Equal(1))
			Expect(fakeStore.PoliciesWithFilterArgsForCall(0)).To(Equal(models.PoliciesFilter{}))
			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Body).To(MatchJSON(expectedResponseJSON))

		})
	})

	Context("when the store throws an error", func() {
		var request *http.Request

		BeforeEach(func() {
			var err error
			request, err = http.NewRequest("GET", "/networking/v0/internal/policies", nil)
			Expect(err).NotTo(HaveOccurred())
			fakeStore.PoliciesWithFilterReturns(nil, errors.New("banana"))
		})

		It("calls the internal server error handler", func() {
			var err error
			request, err = http.NewRequest("GET", "/networking/v0/internal/policies", nil)
			Expect(err).NotTo(HaveOccurred())
			handler.ServeHTTP(resp, request)

			Expect(fakeErrorResponse.InternalServerErrorCallCount()).To(Equal(1))

			w, err, message, description := fakeErrorResponse.InternalServerErrorArgsForCall(0)
			Expect(w).To(Equal(resp))
			Expect(err).To(MatchError("banana"))
			Expect(message).To(Equal("policies-index-internal"))
			Expect(description).To(Equal("database read failed"))
		})
	})

	Context("when the policy cannot be marshaled", func() {
		var request *http.Request

		BeforeEach(func() {
			marshaler.MarshalStub = func(interface{}) ([]byte, error) {
				return nil, errors.New("grapes")
			}

			var err error
			request, err = http.NewRequest("get", "/networking/v0/internal/policies", nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("calls the internal server error handler", func() {
			handler.ServeHTTP(resp, request)

			Expect(fakeErrorResponse.InternalServerErrorCallCount()).To(Equal(1))

			w, err, message, description := fakeErrorResponse.InternalServerErrorArgsForCall(0)
			Expect(w).To(Equal(resp))
			Expect(err).To(MatchError("grapes"))
			Expect(message).To(Equal("policies-index-internal"))
			Expect(description).To(Equal("database marshaling failed"))
		})
	})
})
