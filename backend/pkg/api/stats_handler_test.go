package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/api"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm/logger"
)

var _ = Describe("UserStatsHandler", func() {
	var (
		router http.Handler
		user   models.User
		other  models.User
	)

	BeforeEach(func() {
		db.ResetForTesting()
		cfg := &db.Config{
			DatabasePath:    GinkgoT().TempDir() + "/stats.db",
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
			LogLevel:        logger.Silent,
		}

		database, err := db.Connect(cfg)
		Expect(err).NotTo(HaveOccurred())

		user = models.User{Username: "stats-user", PasswordHash: "hash"}
		other = models.User{Username: "other-user", PasswordHash: "hash"}
		Expect(database.Create(&user).Error).To(Succeed())
		Expect(database.Create(&other).Error).To(Succeed())

		router = api.NewRouter()
	})

	AfterEach(func() {
		Expect(db.Close()).To(Succeed())
	})

	sessionCookie := func(username string) *http.Cookie {
		rec := httptest.NewRecorder()
		Expect(auth.SetSessionCookie(rec, username)).To(Succeed())
		return rec.Result().Cookies()[0]
	}

	It("returns aggregated stats for a known user", func() {
		database, err := db.GetDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(database.Create(&models.GameResult{WinnerID: user.ID, PotSize: 100, Date: time.Now(), GameType: models.DefaultGameType}).Error).To(Succeed())
		Expect(database.Create(&models.GameResult{WinnerID: user.ID, PotSize: 75, Date: time.Now(), GameType: models.DefaultGameType}).Error).To(Succeed())
		Expect(database.Create(&models.GameResult{WinnerID: other.ID, PotSize: 40, Date: time.Now(), GameType: models.DefaultGameType}).Error).To(Succeed())

		req := httptest.NewRequest(http.MethodGet, "/api/users/"+strconvUint(user.ID)+"/stats", nil)
		req.AddCookie(sessionCookie(user.Username))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusOK))
		var response api.UserStatsResponse
		Expect(json.NewDecoder(rec.Body).Decode(&response)).To(Succeed())
		Expect(response.HandsPlayed).To(Equal(int64(3)))
		Expect(response.Wins).To(Equal(int64(2)))
		Expect(response.Losses).To(Equal(int64(1)))
		Expect(response.WinRate).To(BeNumerically("~", 66.6666667, 0.0001))
		Expect(response.TotalEarnings).To(Equal(int64(135)))
	})

	It("returns zeroed stats when the user has no game history", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/users/"+strconvUint(user.ID)+"/stats", nil)
		req.AddCookie(sessionCookie(user.Username))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusOK))
		var response api.UserStatsResponse
		Expect(json.NewDecoder(rec.Body).Decode(&response)).To(Succeed())
		Expect(response).To(Equal(api.UserStatsResponse{}))
	})

	It("returns 404 for an unknown user", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/users/99999/stats", nil)
		req.AddCookie(sessionCookie(user.Username))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusNotFound))
		var response api.ErrorResponse
		Expect(json.NewDecoder(rec.Body).Decode(&response)).To(Succeed())
		Expect(response.Message).To(Equal("User not found"))
	})

	It("returns 400 for an invalid user ID", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/users/not-a-number/stats", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "not-a-number"})
		rec := httptest.NewRecorder()

		api.UserStatsHandler(rec, req)

		Expect(rec.Code).To(Equal(http.StatusBadRequest))
		var response api.ErrorResponse
		Expect(json.NewDecoder(rec.Body).Decode(&response)).To(Succeed())
		Expect(response.Message).To(Equal("Invalid user ID"))
	})

	It("requires authentication", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/users/"+strconvUint(user.ID)+"/stats", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusUnauthorized))
	})
})

func strconvUint(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}
