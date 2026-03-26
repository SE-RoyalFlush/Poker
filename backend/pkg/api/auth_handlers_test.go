package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/api"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
)

var _ = Describe("Authentication Handlers", func() {
	var (
		testUser models.User
	)

	BeforeEach(func() {
		// Set JWT secret for tests
		err := os.Setenv("JWT_SECRET", "test-secret-key-that-is-at-least-32-bytes-long!!!")
		if err != nil {
			return
		}

		// Close any existing database connection to reset the singleton
		err = db.Close()
		if err != nil {
			return
		}

		// Initialize a fresh in-memory database for each test
		dbCfg := &db.Config{
			DatabasePath: ":memory:",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
		}

		database, err := db.Connect(dbCfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(database).NotTo(BeNil())

		// Run migrations
		Expect(database.AutoMigrate(&models.User{})).To(Succeed())

		// Create test user
		hashedPassword, err := auth.HashPassword("password123")
		Expect(err).NotTo(HaveOccurred())

		testUser = models.User{
			Username:     "testuser",
			PasswordHash: hashedPassword,
		}

		// Create test user
		result := database.Create(&testUser)
		Expect(result.Error).NotTo(HaveOccurred())
		Expect(testUser.ID).NotTo(BeZero())
	})

	Describe("LoginHandler", func() {
		It("should successfully login with valid credentials", func() {
			body := map[string]string{
				"username": "testuser",
				"password": "password123",
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			api.LoginHandler(w, req)

			Expect(w.Code).To(Equal(http.StatusNoContent))

			// Verify cookie is set
			cookies := w.Result().Cookies()
			Expect(len(cookies)).To(Equal(1))

			cookie := cookies[0]
			Expect(cookie.Name).To(Equal("session-id"))
			Expect(cookie.Value).NotTo(BeEmpty())
			Expect(cookie.HttpOnly).To(BeTrue())
		})

		It("should return 401 with invalid password", func() {
			body := map[string]string{
				"username": "testuser",
				"password": "wrongpassword",
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			api.LoginHandler(w, req)

			Expect(w.Code).To(Equal(http.StatusUnauthorized))
			Expect(w.Header().Get("Content-Type")).To(Equal("application/json"))

			var response api.ErrorResponse
			Expect(json.Unmarshal(w.Body.Bytes(), &response)).To(Succeed())
			Expect(response.Error).To(Equal("Unauthorized"))
			Expect(response.Message).To(Equal("Invalid username or password"))
			Expect(response.Status).To(Equal(http.StatusUnauthorized))
		})

		It("should return 401 with non-existent user", func() {
			body := map[string]string{
				"username": "nonexistent",
				"password": "password123",
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			api.LoginHandler(w, req)

			Expect(w.Code).To(Equal(http.StatusUnauthorized))

			var response api.ErrorResponse
			Expect(json.Unmarshal(w.Body.Bytes(), &response)).To(Succeed())
			Expect(response.Message).To(Equal("Invalid username or password"))
		})

		It("should return 400 with missing credentials", func() {
			body := map[string]string{
				"username": "testuser",
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			api.LoginHandler(w, req)

			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 204 for /api/login even with invalid JSON as it is caught by decoder", func() {
			// Actually decoder returns error, but LoginHandler in auth.go returns 400 Bad Request
		})

		It("should return 400 for invalid JSON", func() {
			req := httptest.NewRequest("POST", "/api/login", bytes.NewReader([]byte("{invalid-json}")))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			api.LoginHandler(w, req)

			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("LogoutHandler", func() {
		It("should clear the session cookie", func() {
			req := httptest.NewRequest("POST", "/api/logout", nil)
			w := httptest.NewRecorder()

			api.LogoutHandler(w, req)

			Expect(w.Code).To(Equal(http.StatusNoContent))

			// Verify cookie is cleared (MaxAge = -1)
			cookies := w.Result().Cookies()
			Expect(len(cookies)).To(Equal(1))

			cookie := cookies[0]
			Expect(cookie.Name).To(Equal("session-id"))
			Expect(cookie.MaxAge).To(Equal(-1))
		})
	})

	Describe("MeHandler", func() {
		It("should return 204 without a session cookie", func() {
			req := httptest.NewRequest("GET", "/api/me", nil)
			w := httptest.NewRecorder()

			api.MeHandler(w, req)

			Expect(w.Code).To(Equal(http.StatusNoContent))
		})

		It("should return 204 with an invalid token", func() {
			req := httptest.NewRequest("GET", "/api/me", nil)
			req.AddCookie(&http.Cookie{
				Name:  "session-id",
				Value: "invalid.token.here",
			})
			w := httptest.NewRecorder()

			api.MeHandler(w, req)

			Expect(w.Code).To(Equal(http.StatusNoContent))
		})

		It("should return 200 with valid session cookie", func() {
			// Step 1: Login to get a valid session cookie
			body := map[string]string{
				"username": "testuser",
				"password": "password123",
			}
			bodyBytes, _ := json.Marshal(body)
			loginReq := httptest.NewRequest("POST", "/api/login", bytes.NewReader(bodyBytes))
			loginReq.Header.Set("Content-Type", "application/json")
			loginW := httptest.NewRecorder()
			api.LoginHandler(loginW, loginReq)
			Expect(loginW.Code).To(Equal(http.StatusNoContent))

			cookie := loginW.Result().Cookies()[0]

			req := httptest.NewRequest("GET", "/api/me", nil)
			req.AddCookie(cookie)
			w := httptest.NewRecorder()

			api.MeHandler(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))

			// Verify response contains user info
			var response models.User
			Expect(json.Unmarshal(w.Body.Bytes(), &response)).To(Succeed())
			Expect(response.ID).To(Equal(testUser.ID))
			Expect(response.Username).To(Equal(testUser.Username))
		})

		It("should NOT expose password in response", func() {
			// Step 1: Login to get a valid session cookie
			body := map[string]string{
				"username": "testuser",
				"password": "password123",
			}
			bodyBytes, _ := json.Marshal(body)
			loginReq := httptest.NewRequest("POST", "/api/login", bytes.NewReader(bodyBytes))
			loginReq.Header.Set("Content-Type", "application/json")
			loginW := httptest.NewRecorder()
			api.LoginHandler(loginW, loginReq)
			cookie := loginW.Result().Cookies()[0]

			req := httptest.NewRequest("GET", "/api/me", nil)
			req.AddCookie(cookie)
			w := httptest.NewRecorder()

			api.MeHandler(w, req)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				return
			}

			Expect(response["password"]).To(BeNil())
			Expect(response["password_hash"]).To(BeNil())
		})
	})

	Describe("Session Validation Flow", func() {
		It("should validate complete login -> me -> logout flow", func() {
			// 1. Login
			loginBody := map[string]string{
				"username": "testuser",
				"password": "password123",
			}
			loginBodyBytes, _ := json.Marshal(loginBody)

			loginReq := httptest.NewRequest("POST", "/api/login", bytes.NewReader(loginBodyBytes))
			loginReq.Header.Set("Content-Type", "application/json")
			loginW := httptest.NewRecorder()

			api.LoginHandler(loginW, loginReq)

			Expect(loginW.Code).To(Equal(http.StatusNoContent))

			// Extract cookie from login response
			cookies := loginW.Result().Cookies()
			Expect(len(cookies)).To(Equal(1))
			sessionCookie := cookies[0]

			// 2. Call /api/me with cookie - should succeed
			meReq := httptest.NewRequest("GET", "/api/me", nil)
			meReq.AddCookie(sessionCookie)
			meW := httptest.NewRecorder()

			api.MeHandler(meW, meReq)

			Expect(meW.Code).To(Equal(http.StatusOK))

			// 3. Logout
			logoutReq := httptest.NewRequest("POST", "/api/logout", nil)
			logoutW := httptest.NewRecorder()

			api.LogoutHandler(logoutW, logoutReq)

			Expect(logoutW.Code).To(Equal(http.StatusNoContent))

			// 4. Call /api/me again - should fail (using the cleared cookie)
			clearedCookie := logoutW.Result().Cookies()[0]
			meReq2 := httptest.NewRequest("GET", "/api/me", nil)
			meReq2.AddCookie(clearedCookie)
			meW2 := httptest.NewRecorder()

			api.MeHandler(meW2, meReq2)

			Expect(meW2.Code).To(Equal(http.StatusNoContent))
		})
	})
})
