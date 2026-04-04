package auth_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm/logger"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
)

var _ = Describe("Session Helper", func() {
	BeforeEach(func() {
		db.ResetForTesting()
		tempDir := GinkgoT().TempDir()
		cfg := &db.Config{
			DatabasePath:    tempDir + "/test_auth_session.db",
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
			LogLevel:        logger.Silent,
		}

		_, err := db.Connect(cfg)
		Expect(err).NotTo(HaveOccurred())

		database, err := db.GetDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(database.Create(&models.User{Username: "session-user", PasswordHash: "hash"}).Error).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(db.Close()).To(Succeed())
	})

	It("resolves the authenticated user from a valid session cookie", func() {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		Expect(auth.SetSessionCookie(rec, "session-user")).To(Succeed())
		req.AddCookie(rec.Result().Cookies()[0])

		user, err := auth.AuthenticatedUserFromRequest(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(user.Username).To(Equal("session-user"))
	})

	It("returns an error for a missing session cookie", func() {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)

		_, err := auth.AuthenticatedUserFromRequest(req)
		Expect(err).To(MatchError(auth.ErrMissingSession))
	})

	It("returns an error for an invalid session cookie", func() {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "tampered-value"})

		_, err := auth.AuthenticatedUserFromRequest(req)
		Expect(err).To(MatchError(auth.ErrInvalidSession))
	})
})
