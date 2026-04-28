package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

var _ = Describe("Room Handlers", func() {
	var (
		router http.Handler
		user   models.User
	)

	BeforeEach(func() {
		db.ResetForTesting()
		tempDir := GinkgoT().TempDir()
		_, err := db.Connect(&db.Config{
			DatabasePath:    tempDir + "/test_rooms.db",
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
			LogLevel:        logger.Silent,
		})
		Expect(err).NotTo(HaveOccurred())

		hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		Expect(err).NotTo(HaveOccurred())

		database, err := db.GetDB()
		Expect(err).NotTo(HaveOccurred())
		user = models.User{
			Username:     "room-user",
			PasswordHash: string(hash),
		}
		Expect(database.Create(&user).Error).NotTo(HaveOccurred())

		router = api.NewRouter()
	})

	AfterEach(func() {
		Expect(db.Close()).To(Succeed())
	})

	sessionCookie := func() *http.Cookie {
		rec := httptest.NewRecorder()
		Expect(auth.SetSessionCookie(rec, user.Username)).To(Succeed())
		return rec.Result().Cookies()[0]
	}

	It("creates a room for the authenticated user", func() {
		req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader([]byte(`{"maxPlayers":8,"isPrivate":true}`)))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie())

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusCreated))

		var body struct {
			Code           string `json:"code"`
			HostUserID     uint   `json:"hostUserId"`
			Status         string `json:"status"`
			MaxPlayers     int    `json:"maxPlayers"`
			CurrentPlayers int    `json:"currentPlayers"`
			IsPrivate      bool   `json:"isPrivate"`
			Seats          int    `json:"seats"`
			IsFull         bool   `json:"isFull"`
		}
		Expect(json.NewDecoder(rec.Body).Decode(&body)).To(Succeed())
		Expect(body.Code).To(MatchRegexp(`^[A-Z0-9]{6}$`))
		Expect(body.HostUserID).To(Equal(user.ID))
		Expect(body.Status).To(Equal(string(models.RoomStatusOpen)))
		Expect(body.MaxPlayers).To(Equal(8))
		Expect(body.CurrentPlayers).To(Equal(0))
		Expect(body.IsPrivate).To(BeTrue())
		Expect(body.Seats).To(Equal(8))
		Expect(body.IsFull).To(BeFalse())

		database, err := db.GetDB()
		Expect(err).NotTo(HaveOccurred())
		var persisted models.Room
		Expect(database.Where("code = ?", body.Code).First(&persisted).Error).NotTo(HaveOccurred())
		Expect(persisted.HostUserID).To(Equal(user.ID))
	})

	It("returns room details by code when the room is open", func() {
		database, err := db.GetDB()
		Expect(err).NotTo(HaveOccurred())
		room := models.Room{
			Code:       "ZX98QP",
			HostUserID: user.ID,
			Status:     models.RoomStatusOpen,
			MaxPlayers: 6,
			IsPrivate:  false,
		}
		Expect(database.Create(&room).Error).NotTo(HaveOccurred())

		req := httptest.NewRequest(http.MethodGet, "/api/rooms/ZX98QP", nil)
		req.AddCookie(sessionCookie())

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusOK))
		var body struct {
			Code           string `json:"code"`
			HostUsername   string `json:"hostUsername"`
			Status         string `json:"status"`
			CurrentPlayers int    `json:"currentPlayers"`
		}
		Expect(json.NewDecoder(rec.Body).Decode(&body)).To(Succeed())
		Expect(body.Code).To(Equal("ZX98QP"))
		Expect(body.HostUsername).To(Equal(user.Username))
		Expect(body.Status).To(Equal(string(models.RoomStatusOpen)))
		Expect(body.CurrentPlayers).To(Equal(0))
	})

	It("lists rooms filtered by status", func() {
		database, err := db.GetDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(database.Create(&models.Room{
			Code:       "OPEN01",
			HostUserID: user.ID,
			Status:     models.RoomStatusOpen,
			MaxPlayers: 6,
		}).Error).NotTo(HaveOccurred())
		Expect(database.Create(&models.Room{
			Code:       "CLOS01",
			HostUserID: user.ID,
			Status:     models.RoomStatusClosed,
			MaxPlayers: 6,
		}).Error).NotTo(HaveOccurred())

		req := httptest.NewRequest(http.MethodGet, "/api/rooms?status=open", nil)
		req.AddCookie(sessionCookie())

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusOK))
		var body []struct {
			Code           string `json:"code"`
			Status         string `json:"status"`
			CurrentPlayers int    `json:"currentPlayers"`
			Seats          int    `json:"seats"`
		}
		Expect(json.NewDecoder(rec.Body).Decode(&body)).To(Succeed())
		Expect(body).To(HaveLen(1))
		Expect(body[0].Code).To(Equal("OPEN01"))
		Expect(body[0].Status).To(Equal(string(models.RoomStatusOpen)))
		Expect(body[0].CurrentPlayers).To(Equal(0))
		Expect(body[0].Seats).To(Equal(6))
	})

	It("rejects invalid list status filters", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/rooms?status=missing", nil)
		req.AddCookie(sessionCookie())

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusBadRequest))
	})

	It("returns an open room for join validation", func() {
		database, err := db.GetDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(database.Create(&models.Room{
			Code:       "JOIN01",
			HostUserID: user.ID,
			Status:     models.RoomStatusOpen,
			MaxPlayers: 6,
		}).Error).NotTo(HaveOccurred())

		req := httptest.NewRequest(http.MethodPost, "/api/rooms/join", bytes.NewReader([]byte(`{"code":"join01"}`)))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie())

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusOK))
		var body struct {
			Code string `json:"code"`
		}
		Expect(json.NewDecoder(rec.Body).Decode(&body)).To(Succeed())
		Expect(body.Code).To(Equal("JOIN01"))
	})

	It("returns 400 for malformed join codes and 404 for non-open rooms", func() {
		database, err := db.GetDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(database.Create(&models.Room{
			Code:       "NOJOIN",
			HostUserID: user.ID,
			Status:     models.RoomStatusClosed,
			MaxPlayers: 6,
		}).Error).NotTo(HaveOccurred())

		cases := []struct {
			body   string
			status int
		}{
			{body: `{"code":"BAD"}`, status: http.StatusBadRequest},
		{body: `{"code":"NOJOIN"}`, status: http.StatusNotFound},
		}

		for _, tc := range cases {
			req := httptest.NewRequest(http.MethodPost, "/api/rooms/join", bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(sessionCookie())

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(tc.status), tc.body)
		}
	})

	It("returns 400 for malformed room codes and 404 for unknown room codes", func() {
		cases := []struct {
			target string
			status int
		}{
			{target: "/api/rooms/BADCODE", status: http.StatusBadRequest},
			{target: "/api/rooms/ABC123", status: http.StatusNotFound},
		}

		for _, tc := range cases {
			req := httptest.NewRequest(http.MethodGet, tc.target, nil)
			req.AddCookie(sessionCookie())

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(tc.status), tc.target)
		}
	})

	It("returns 404 for rooms that are not open", func() {
		database, err := db.GetDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(database.Create(&models.Room{
			Code:       "CLOSED",
			HostUserID: user.ID,
			Status:     models.RoomStatusClosed,
			MaxPlayers: 6,
		}).Error).NotTo(HaveOccurred())

		req := httptest.NewRequest(http.MethodGet, "/api/rooms/CLOSED", nil)
		req.AddCookie(sessionCookie())

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusNotFound))
	})

	It("rejects invalid maxPlayers on create", func() {
		req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader([]byte(`{"maxPlayers":11}`)))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie())

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusBadRequest))
	})

	It("persists and looks up room metadata without an active socket", func() {
		req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader([]byte(`{"maxPlayers":8,"isPrivate":true}`)))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie())

		createRec := httptest.NewRecorder()
		router.ServeHTTP(createRec, req)

		Expect(createRec.Code).To(Equal(http.StatusCreated))
		var created struct {
			Code           string `json:"code"`
			HostUserID     uint   `json:"hostUserId"`
			MaxPlayers     int    `json:"maxPlayers"`
			CurrentPlayers int    `json:"currentPlayers"`
			IsPrivate      bool   `json:"isPrivate"`
		}
		Expect(json.NewDecoder(createRec.Body).Decode(&created)).To(Succeed())
		Expect(created.Code).NotTo(BeEmpty())
		Expect(created.HostUserID).To(Equal(user.ID))
		Expect(created.MaxPlayers).To(Equal(8))
		Expect(created.CurrentPlayers).To(Equal(0))
		Expect(created.IsPrivate).To(BeTrue())

		lookupReq := httptest.NewRequest(http.MethodGet, "/api/rooms/"+created.Code, nil)
		lookupReq.AddCookie(sessionCookie())
		lookupRec := httptest.NewRecorder()
		router.ServeHTTP(lookupRec, lookupReq)

		Expect(lookupRec.Code).To(Equal(http.StatusOK))
		var found struct {
			Code           string `json:"code"`
			CurrentPlayers int    `json:"currentPlayers"`
		}
		Expect(json.NewDecoder(lookupRec.Body).Decode(&found)).To(Succeed())
		Expect(found.Code).To(Equal(created.Code))
		Expect(found.CurrentPlayers).To(Equal(0))
	})
})
