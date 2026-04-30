package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/api"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/logger"
)

var _ = Describe("Router", func() {
	var (
		router           http.Handler
		originalAppEnv   string
		hadOriginalEnv   bool
		originalAdmin    string
		hadOriginalAdmin bool
		originalAdminPw  string
		hadOriginalPw    bool
	)

	BeforeEach(func() {
		originalAppEnv, hadOriginalEnv = os.LookupEnv("APP_ENV")
		originalAdmin, hadOriginalAdmin = os.LookupEnv("ADMIN_USERNAME")
		originalAdminPw, hadOriginalPw = os.LookupEnv("ADMIN_PASSWORD")

		db.ResetForTesting()
		tempDir := GinkgoT().TempDir()
		cfg := &db.Config{
			DatabasePath:    tempDir + "/test_router.db",
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
			LogLevel:        logger.Silent,
		}

		_, err := db.Connect(cfg)
		Expect(err).NotTo(HaveOccurred())

		hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		Expect(err).NotTo(HaveOccurred())

		database, err := db.GetDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(database.Create(&models.User{
			Username:     "router-user",
			PasswordHash: string(hash),
		}).Error).NotTo(HaveOccurred())

		Expect(os.Setenv("APP_ENV", "development")).To(Succeed())
		router = api.NewRouter()
	})

	AfterEach(func() {
		if hadOriginalEnv {
			Expect(os.Setenv("APP_ENV", originalAppEnv)).To(Succeed())
		} else {
			Expect(os.Unsetenv("APP_ENV")).To(Succeed())
		}
		if hadOriginalAdmin {
			Expect(os.Setenv("ADMIN_USERNAME", originalAdmin)).To(Succeed())
		} else {
			Expect(os.Unsetenv("ADMIN_USERNAME")).To(Succeed())
		}
		if hadOriginalPw {
			Expect(os.Setenv("ADMIN_PASSWORD", originalAdminPw)).To(Succeed())
		} else {
			Expect(os.Unsetenv("ADMIN_PASSWORD")).To(Succeed())
		}
		Expect(db.Close()).To(Succeed())
	})

	sessionCookie := func() *http.Cookie {
		rec := httptest.NewRecorder()
		Expect(auth.SetSessionCookie(rec, "router-user")).To(Succeed())
		return rec.Result().Cookies()[0]
	}

	It("allows authenticated access to protected endpoints", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/rooms?status=open", nil)
		req.AddCookie(sessionCookie())

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusOK))
	})

	It("returns 401 for protected endpoints without auth", func() {
		req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader([]byte(`{}`)))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusUnauthorized))
		var body api.ErrorResponse
		Expect(json.NewDecoder(rec.Body).Decode(&body)).To(Succeed())
		Expect(body.Error).To(Equal("Unauthorized"))
	})

	It("keeps admin endpoints on basic auth without requiring a session", func() {
		Expect(os.Setenv("ADMIN_USERNAME", "admin")).To(Succeed())
		Expect(os.Setenv("ADMIN_PASSWORD", "admin")).To(Succeed())

		req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
		req.SetBasicAuth("admin", "admin")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusOK))
	})

	It("mounts room endpoints and websocket handshake behind auth middleware", func() {
		protectedRequests := []struct {
			method string
			target string
			body   []byte
		}{
			{method: http.MethodGet, target: "/api/rooms?status=open"},
			{method: http.MethodPost, target: "/api/rooms", body: []byte(`{}`)},
			{method: http.MethodPost, target: "/api/rooms/join", body: []byte(`{}`)},
			{method: http.MethodGet, target: "/api/rooms/ABC123"},
			{method: http.MethodGet, target: "/api/users/1/stats"},
			{method: http.MethodGet, target: "/ws"},
		}

		newRequest := func(method, target string, body []byte) *http.Request {
			if body == nil {
				return httptest.NewRequest(method, target, nil)
			}
			req := httptest.NewRequest(method, target, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			return req
		}

		for _, req := range protectedRequests {
			unauthenticatedReq := newRequest(req.method, req.target, req.body)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, unauthenticatedReq)
			Expect(rec.Code).To(Equal(http.StatusUnauthorized), req.target)

			reqWithAuth := newRequest(req.method, req.target, req.body)
			reqWithAuth.AddCookie(sessionCookie())
			recWithAuth := httptest.NewRecorder()
			router.ServeHTTP(recWithAuth, reqWithAuth)
			expectedStatus := http.StatusOK
			switch req.target {
			case "/api/rooms":
				expectedStatus = http.StatusCreated
			case "/api/rooms/join":
				expectedStatus = http.StatusBadRequest
			case "/api/rooms/ABC123":
				expectedStatus = http.StatusNotFound
			case "/ws":
				expectedStatus = http.StatusBadRequest
			}
			Expect(recWithAuth.Code).To(Equal(expectedStatus), req.target)
		}
	})

	It("restricts websocket handshake routing to GET", func() {
		req := httptest.NewRequest(http.MethodPost, "/ws", nil)
		req.AddCookie(sessionCookie())

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusMethodNotAllowed))
	})

	It("keeps public auth and system routes accessible without auth", func() {
		publicRequests := []*http.Request{
			httptest.NewRequest(http.MethodGet, "/api/health", nil),
			httptest.NewRequest(http.MethodGet, "/api/csrf", nil),
			httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader([]byte(`{"username":"router-user","password":"password123"}`))),
			httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader([]byte(`{"username":"freshuser","password":"password123"}`))),
			func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
				req.SetBasicAuth("admin", "admin")
				return req
			}(),
		}

		expectedStatuses := []int{
			http.StatusOK,
			http.StatusOK,
			http.StatusNoContent,
			http.StatusCreated,
			http.StatusOK,
		}

		for i, req := range publicRequests {
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(expectedStatuses[i]), req.URL.Path)
		}
	})
})
