package db_test

import (
	"os"
	"path/filepath"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm/logger"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
)

var _ = Describe("Database Client", func() {
	var tempDir string

	BeforeEach(func() {
		db.ResetForTesting()
		tempDir = GinkgoT().TempDir()
	})

	AfterEach(func() {
		db.Close()
	})

	It("should return ErrNotInitialized when Ping is called before Connect", func() {
		err := db.Ping()
		Expect(err).To(Equal(db.ErrNotInitialized))
	})

	It("should return ErrNotInitialized when GetDB is called before Connect", func() {
		instance, err := db.GetDB()
		Expect(err).To(Equal(db.ErrNotInitialized))
		Expect(instance).To(BeNil())
	})

	It("should successfully connect with valid path", func() {
		testDBPath := filepath.Join(tempDir, "test.db")
		cfg := &db.Config{
			DatabasePath:    testDBPath,
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 1 * time.Minute,
			LogLevel:        logger.Silent,
		}

		// Connect
		instance, err := db.Connect(cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance).NotTo(BeNil())

		// Verify file created
		_, err = os.Stat(testDBPath)
		Expect(err).NotTo(HaveOccurred())

		// Verify foreign keys enabled
		var fkEnabled int
		err = instance.Raw("PRAGMA foreign_keys").Scan(&fkEnabled).Error
		Expect(err).NotTo(HaveOccurred())
		Expect(fkEnabled).To(Equal(1))

		// Verify Ping works
		err = db.Ping()
		Expect(err).NotTo(HaveOccurred())

		// Verify schema created for registered models
		Expect(instance.Migrator().HasTable(&models.Room{})).To(BeTrue())

		// Verify GetDB works
		retrievedDB, err := db.GetDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(retrievedDB).To(Equal(instance))
	})

	It("should fail with invalid path and propagate error", func() {
		cfg := &db.Config{
			DatabasePath:    "/invalid/path/test.db",
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 1 * time.Minute,
			LogLevel:        logger.Silent,
		}

		// Connect should fail
		instance, connectErr := db.Connect(cfg)
		Expect(connectErr).To(HaveOccurred())
		Expect(instance).To(BeNil())

		// Ping should return same error
		pingErr := db.Ping()
		Expect(pingErr).To(Equal(connectErr))

		// GetDB should return same error
		retrievedDB, getDBErr := db.GetDB()
		Expect(getDBErr).To(Equal(connectErr))
		Expect(retrievedDB).To(BeNil())
	})

	It("should return same instance on multiple Connect calls", func() {
		testDBPath := filepath.Join(tempDir, "test.db")
		cfg := &db.Config{
			DatabasePath:    testDBPath,
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 1 * time.Minute,
			LogLevel:        logger.Silent,
		}

		instance1, err1 := db.Connect(cfg)
		Expect(err1).NotTo(HaveOccurred())

		instance2, err2 := db.Connect(cfg)
		Expect(err2).NotTo(HaveOccurred())

		Expect(instance1).To(Equal(instance2))
	})

	It("should not error when Close is called before Connect", func() {
		err := db.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should allow retry after failed connection", func() {
		// First attempt with invalid path
		invalidCfg := &db.Config{
			DatabasePath:    "/invalid/path/test.db",
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 1 * time.Minute,
			LogLevel:        logger.Silent,
		}

		instance1, err1 := db.Connect(invalidCfg)
		Expect(err1).To(HaveOccurred())
		Expect(instance1).To(BeNil())

		// Second attempt with valid path should succeed
		testDBPath := filepath.Join(tempDir, "test.db")
		validCfg := &db.Config{
			DatabasePath:    testDBPath,
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 1 * time.Minute,
			LogLevel:        logger.Silent,
		}

		instance2, err2 := db.Connect(validCfg)
		Expect(err2).NotTo(HaveOccurred())
		Expect(instance2).NotTo(BeNil())

		// Verify connection works
		err := db.Ping()
		Expect(err).NotTo(HaveOccurred())
	})
})
