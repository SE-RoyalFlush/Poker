package db_test

import (
	"path/filepath"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm/logger"
)

var _ = Describe("Room Model", func() {
	var tempDir string

	BeforeEach(func() {
		db.ResetForTesting()
		tempDir = GinkgoT().TempDir()
	})

	AfterEach(func() {
		db.Close()
	})

	newTestDB := func(filename string) *db.Config {
		return &db.Config{
			DatabasePath:    filepath.Join(tempDir, filename),
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: time.Minute,
			LogLevel:        logger.Silent,
		}
	}

	It("generates a valid room code when one is not provided", func() {
		instance, err := db.Connect(newTestDB("room_create.db"))
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.AutoMigrate(&models.User{}, &models.Room{})).To(Succeed())

		host := models.User{}
		Expect(instance.Create(&host).Error).NotTo(HaveOccurred())

		room := models.Room{
			HostID:     host.ID,
			Status:     models.RoomStatusWaiting,
			MaxPlayers: 6,
		}

		Expect(instance.Create(&room).Error).NotTo(HaveOccurred())
		Expect(room.Code).To(MatchRegexp(`^[A-Z0-9]{6}$`))
	})

	It("rejects invalid room codes", func() {
		instance, err := db.Connect(newTestDB("room_bad_code.db"))
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.AutoMigrate(&models.User{}, &models.Room{})).To(Succeed())

		host := models.User{}
		Expect(instance.Create(&host).Error).NotTo(HaveOccurred())

		room := models.Room{
			Code:       "bad",
			HostID:     host.ID,
			Status:     models.RoomStatusWaiting,
			MaxPlayers: 6,
		}

		Expect(instance.Create(&room).Error).To(HaveOccurred())
	})

	It("rejects invalid room statuses", func() {
		instance, err := db.Connect(newTestDB("room_bad_status.db"))
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.AutoMigrate(&models.User{}, &models.Room{})).To(Succeed())

		host := models.User{}
		Expect(instance.Create(&host).Error).NotTo(HaveOccurred())

		room := models.Room{
			Code:       "ABC123",
			HostID:     host.ID,
			Status:     "OPEN",
			MaxPlayers: 6,
		}

		Expect(instance.Create(&room).Error).To(HaveOccurred())
	})
})
