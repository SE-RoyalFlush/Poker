package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm/logger"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/middleware"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
)

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

var _ = Describe("AuthMiddleware", func() {
	BeforeEach(func() {
		db.ResetForTesting()
		tempDir := GinkgoT().TempDir()
		cfg := &db.Config{
			DatabasePath:    tempDir + "/test_auth_middleware.db",
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
			LogLevel:        logger.Silent,
		}

		_, err := db.Connect(cfg)
		Expect(err).NotTo(HaveOccurred())

		database, err := db.GetDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(database.Create(&models.User{Username: "middleware-user", PasswordHash: "hash"}).Error).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(db.Close()).To(Succeed())
	})

	It("allows requests with a valid session cookie", func() {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()
		Expect(auth.SetSessionCookie(rec, "middleware-user")).To(Succeed())
		req.AddCookie(rec.Result().Cookies()[0])

		protected := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		resp := httptest.NewRecorder()
		protected.ServeHTTP(resp, req)
		Expect(resp.Code).To(Equal(http.StatusOK))
	})

	It("rejects requests without a session cookie", func() {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		protected := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		resp := httptest.NewRecorder()
		protected.ServeHTTP(resp, req)
		Expect(resp.Code).To(Equal(http.StatusUnauthorized))

		var body errorResponse
		Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())
		Expect(body).To(Equal(errorResponse{
			Error:   "Unauthorized",
			Message: "Authentication required",
			Status:  http.StatusUnauthorized,
		}))
	})

	It("rejects requests with an invalid session cookie", func() {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "tampered-value"})

		protected := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		resp := httptest.NewRecorder()
		protected.ServeHTTP(resp, req)
		Expect(resp.Code).To(Equal(http.StatusUnauthorized))

		var body errorResponse
		Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())
		Expect(body).To(Equal(errorResponse{
			Error:   "Unauthorized",
			Message: "Authentication required",
			Status:  http.StatusUnauthorized,
		}))
	})
})
