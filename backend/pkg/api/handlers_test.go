package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm/logger"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/api"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
)

var _ = Describe("API Handlers", func() {

	Describe("HealthHandler", func() {
		var (
			req      *http.Request
			recorder *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			req = httptest.NewRequest(http.MethodGet, "/health", nil)
			recorder = httptest.NewRecorder()
		})

		Context("when database is not connected", func() {
			BeforeEach(func() {
				db.ResetForTesting()
			})

			AfterEach(func() {
				db.Close()
			})

			It("should return status alive and database disconnected", func() {
				api.HealthHandler(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))

				var response api.HealthResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Status).To(Equal("alive"))
				Expect(response.Database).To(Equal("disconnected"))
			})
		})

		Context("when database is connected", func() {
			BeforeEach(func() {
				db.ResetForTesting()
				tempDir := GinkgoT().TempDir()
				cfg := &db.Config{
					DatabasePath:    tempDir + "/test.db",
					MaxOpenConns:    10,
					MaxIdleConns:    2,
					ConnMaxLifetime: 5 * time.Minute,
					LogLevel:        logger.Silent,
				}
				_, err := db.Connect(cfg)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				db.Close()
			})

			It("should return status alive and database connected", func() {
				api.HealthHandler(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))

				var response api.HealthResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Status).To(Equal("alive"))
				Expect(response.Database).To(Equal("connected"))
			})
		})
	})

	Describe("NotFoundHandler", func() {
		var (
			req      *http.Request
			recorder *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			req = httptest.NewRequest(http.MethodGet, "/random", nil)
			recorder = httptest.NewRecorder()
		})

		It("should return 404 status code", func() {
			api.NotFoundHandler(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusNotFound))
		})

		It("should set Content-Type to application/json", func() {
			api.NotFoundHandler(recorder, req)

			Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
		})

		It("should return structured error response", func() {
			api.NotFoundHandler(recorder, req)

			var response api.ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Error).To(Equal("Not Found"))
			Expect(response.Message).To(Equal("The requested resource was not found"))
			Expect(response.Status).To(Equal(http.StatusNotFound))
		})
	})
})
