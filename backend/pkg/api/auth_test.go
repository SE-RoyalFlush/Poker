package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm/logger"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/api"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
)

var _ = Describe("Auth API", func() {
	var (
		recorder *httptest.ResponseRecorder
		tempDir  string
	)

	BeforeEach(func() {
		recorder = httptest.NewRecorder()
		db.ResetForTesting()
		tempDir = GinkgoT().TempDir()
		cfg := &db.Config{
			DatabasePath:    tempDir + "/test_auth.db",
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
			LogLevel:        logger.Silent,
		}
		_, err := db.Connect(cfg)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := db.Close()
		if err != nil {
			return
		}
	})

	Describe("RegisterHandler", func() {
		Context("with valid registration data", func() {
			It("should return 201 Created and the user object", func() {
				reqBody := api.RegisterRequest{
					Username: "testuser",
					Password: "password123",
				}
				body, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))

				api.RegisterHandler(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusCreated))

				var user models.User
				err := json.Unmarshal(recorder.Body.Bytes(), &user)
				Expect(err).NotTo(HaveOccurred())

				Expect(user.Username).To(Equal("testuser"))
				Expect(user.ID).NotTo(BeZero())
				Expect(user.PasswordHash).To(BeEmpty())
			})
		})

		Context("with an existing username", func() {
			BeforeEach(func() {
				// Pre-create a user
				database, err := db.GetDB()
				Expect(err).NotTo(HaveOccurred())

				database.Create(&models.User{Username: "existinguser", PasswordHash: "somehash"})
			})

			It("should return 409 Conflict", func() {
				reqBody := api.RegisterRequest{
					Username: "existinguser",
					Password: "password123",
				}
				body, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))

				api.RegisterHandler(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusConflict))

				var errResp api.ErrorResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &errResp)
				Expect(err).NotTo(HaveOccurred())

				Expect(errResp.Error).To(Equal("Conflict"))
			})
		})

		Context("with invalid data", func() {
			It("should return 400 Bad Request for short username", func() {
				reqBody := api.RegisterRequest{
					Username: "tu",
					Password: "password123",
				}
				body, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))

				api.RegisterHandler(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("should return 400 Bad Request for short password", func() {
				reqBody := api.RegisterRequest{
					Username: "testuser",
					Password: "pass",
				}
				body, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))

				api.RegisterHandler(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("should return 400 Bad Request for empty fields", func() {
				reqBody := api.RegisterRequest{
					Username: "",
					Password: "",
				}
				body, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))

				api.RegisterHandler(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})
})
