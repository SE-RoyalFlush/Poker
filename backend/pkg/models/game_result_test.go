package models_test

import (
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
)

var _ = Describe("GameResult", func() {
	var db *gorm.DB

	BeforeEach(func() {
		tmpDir := GinkgoT().TempDir()
		var err error
		db, err = gorm.Open(sqlite.Open(filepath.Join(tmpDir, "test.db")), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		Expect(db.AutoMigrate(models.AllModels()...)).To(Succeed())
	})

	Describe("GameType constants", func() {
		It("defines TexasHoldEm and Omaha variants", func() {
			Expect(string(models.GameTypeTexasHoldEm)).To(Equal("texas_holdem"))
			Expect(string(models.GameTypeOmaha)).To(Equal("omaha"))
		})
	})

	Describe("persistence", func() {
		var winner models.User

		BeforeEach(func() {
			winner = models.User{Username: "alice", PasswordHash: "hash"}
			Expect(db.Create(&winner).Error).NotTo(HaveOccurred())
		})

		It("creates and retrieves a GameResult", func() {
			result := models.GameResult{
				WinnerID: winner.ID,
				PotSize:  500,
				Date:     time.Now().UTC().Truncate(time.Second),
				GameType: models.GameTypeTexasHoldEm,
			}
			Expect(db.Create(&result).Error).NotTo(HaveOccurred())
			Expect(result.ID).NotTo(BeZero())

			var fetched models.GameResult
			Expect(db.First(&fetched, result.ID).Error).NotTo(HaveOccurred())
			Expect(fetched.WinnerID).To(Equal(winner.ID))
			Expect(fetched.PotSize).To(Equal(500))
			Expect(fetched.GameType).To(Equal(models.GameTypeTexasHoldEm))
		})

		It("stores the winner association via WinnerID", func() {
			result := models.GameResult{
				WinnerID: winner.ID,
				PotSize:  100,
				Date:     time.Now(),
				GameType: models.GameTypeTexasHoldEm,
			}
			Expect(db.Create(&result).Error).NotTo(HaveOccurred())

			var fetched models.GameResult
			Expect(db.Preload("Winner").First(&fetched, result.ID).Error).NotTo(HaveOccurred())
			Expect(fetched.Winner.Username).To(Equal("alice"))
		})

		It("persists Omaha game type", func() {
			result := models.GameResult{
				WinnerID: winner.ID,
				PotSize:  200,
				Date:     time.Now(),
				GameType: models.GameTypeOmaha,
			}
			Expect(db.Create(&result).Error).NotTo(HaveOccurred())

			var fetched models.GameResult
			Expect(db.First(&fetched, result.ID).Error).NotTo(HaveOccurred())
			Expect(fetched.GameType).To(Equal(models.GameTypeOmaha))
		})

		It("records large pot sizes", func() {
			result := models.GameResult{
				WinnerID: winner.ID,
				PotSize:  1_000_000,
				Date:     time.Now(),
				GameType: models.GameTypeTexasHoldEm,
			}
			Expect(db.Create(&result).Error).NotTo(HaveOccurred())

			var fetched models.GameResult
			Expect(db.First(&fetched, result.ID).Error).NotTo(HaveOccurred())
			Expect(fetched.PotSize).To(Equal(1_000_000))
		})

		It("stores a zero pot size", func() {
			result := models.GameResult{
				WinnerID: winner.ID,
				PotSize:  0,
				Date:     time.Now(),
				GameType: models.GameTypeTexasHoldEm,
			}
			Expect(db.Create(&result).Error).NotTo(HaveOccurred())

			var fetched models.GameResult
			Expect(db.First(&fetched, result.ID).Error).NotTo(HaveOccurred())
			Expect(fetched.PotSize).To(Equal(0))
		})
	})
})
