package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/api"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/gorilla/mux"
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

	Describe("MeHandler", func() {
		var user models.User
		BeforeEach(func() {
			// Create a user in the DB
			user = models.User{Username: "meuser", PasswordHash: ""}
			hash, _ := bcrypt.GenerateFromPassword([]byte("mepass"), bcrypt.DefaultCost)
			user.PasswordHash = string(hash)
			dbInstance, _ := db.GetDB()
			dbInstance.Create(&user)
		})

		It("should return 200 and user info for valid session", func() {
			// Step 1: Login to get session cookie
			loginReqBody := api.LoginRequest{
				Username: "meuser",
				Password: "mepass",
			}
			loginBody, _ := json.Marshal(loginReqBody)
			loginReq := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(loginBody))
			loginRec := httptest.NewRecorder()
			api.LoginHandler(loginRec, loginReq)
			Expect(loginRec.Code).To(Equal(http.StatusNoContent))

			cookie := loginRec.Result().Cookies()[0]

			// Step 2: Use cookie to call MeHandler
			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()
			api.MeHandler(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			var resp models.User
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Username).To(Equal("meuser"))
			Expect(resp.PasswordHash).To(BeEmpty()) // Should not be present in JSON
		})

		It("should return 204 for missing session", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
			rec := httptest.NewRecorder()
			api.MeHandler(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
		})

		It("should return 204 for invalid session cookie", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
			req.AddCookie(&http.Cookie{Name: "session-id", Value: "invalid-value"})
			rec := httptest.NewRecorder()
			api.MeHandler(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
		})
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

	Describe("Admin user management", func() {
		var database = func() *gorm.DB {
			dbClient, err := db.GetDB()
			Expect(err).NotTo(HaveOccurred())
			return dbClient
		}

		BeforeEach(func() {
			Expect(database().Create(&models.User{Username: "alice", PasswordHash: "hash-a"}).Error).NotTo(HaveOccurred())
			Expect(database().Create(&models.User{Username: "bob", PasswordHash: "hash-b"}).Error).NotTo(HaveOccurred())
		})

		It("should reject list requests without admin credentials", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)

			api.AdminListUsersHandler(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should return usernames for authenticated admin", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
			req.SetBasicAuth("admin", "admin")

			api.AdminListUsersHandler(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusOK))

			var response []map[string]any
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(HaveLen(2))
			Expect(response[0]["username"]).To(Equal("alice"))
			Expect(response[1]["username"]).To(Equal("bob"))
		})

		It("should delete existing user for authenticated admin", func() {
			var user models.User
			Expect(database().Where("username = ?", "alice").First(&user).Error).NotTo(HaveOccurred())

			router := mux.NewRouter()
			router.HandleFunc("/api/admin/users/{id:[0-9]+}", api.AdminDeleteUserHandler).Methods(http.MethodDelete)

			req := httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+strconv.FormatUint(uint64(user.ID), 10), nil)
			req.SetBasicAuth("admin", "admin")
			deleteRecorder := httptest.NewRecorder()

			router.ServeHTTP(deleteRecorder, req)

			Expect(deleteRecorder.Code).To(Equal(http.StatusNoContent))

			err := database().First(&models.User{}, user.ID).Error
			Expect(err).To(HaveOccurred())
		})

		It("should return 404 when deleting non-existing user", func() {
			router := mux.NewRouter()
			router.HandleFunc("/api/admin/users/{id:[0-9]+}", api.AdminDeleteUserHandler).Methods(http.MethodDelete)

			req := httptest.NewRequest(http.MethodDelete, "/api/admin/users/99999", nil)
			req.SetBasicAuth("admin", "admin")
			deleteRecorder := httptest.NewRecorder()

			router.ServeHTTP(deleteRecorder, req)

			Expect(deleteRecorder.Code).To(Equal(http.StatusNotFound))
		})
	})
})
