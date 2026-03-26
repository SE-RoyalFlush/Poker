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

type sqliteIndexInfo struct {
	Name   string
	Unique int
	Origin string
}

type sqliteIndexColumn struct {
	Name string
}

type sqliteForeignKey struct {
	Table string
	From  string
	To    string
}

var _ = Describe("Room Migration", func() {
	var tempDir string

	BeforeEach(func() {
		db.ResetForTesting()
		tempDir = GinkgoT().TempDir()
	})

	AfterEach(func() {
		db.Close()
	})

	It("creates the rooms table with a unique index on code", func() {
		testDBPath := filepath.Join(tempDir, "rooms.db")
		instance, err := db.Connect(&db.Config{
			DatabasePath:    testDBPath,
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: time.Minute,
			LogLevel:        logger.Silent,
		})
		Expect(err).NotTo(HaveOccurred())

		err = instance.AutoMigrate(&models.User{}, &models.Room{})
		Expect(err).NotTo(HaveOccurred())

		Expect(instance.Migrator().HasTable(&models.Room{})).To(BeTrue())

		var indexes []sqliteIndexInfo
		err = instance.Raw("PRAGMA index_list(rooms)").Scan(&indexes).Error
		Expect(err).NotTo(HaveOccurred())

		var codeIndexName string
		for _, index := range indexes {
			if index.Unique != 1 {
				continue
			}

			var columns []sqliteIndexColumn
			err = instance.Raw("PRAGMA index_info(" + "'" + index.Name + "'" + ")").Scan(&columns).Error
			Expect(err).NotTo(HaveOccurred())

			for _, column := range columns {
				if column.Name == "code" {
					codeIndexName = index.Name
					break
				}
			}
		}

		Expect(codeIndexName).NotTo(BeEmpty())
	})

	It("creates a foreign key from rooms.host_id to users.id", func() {
		testDBPath := filepath.Join(tempDir, "rooms_fk.db")
		instance, err := db.Connect(&db.Config{
			DatabasePath:    testDBPath,
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: time.Minute,
			LogLevel:        logger.Silent,
		})
		Expect(err).NotTo(HaveOccurred())

		err = instance.AutoMigrate(&models.User{}, &models.Room{})
		Expect(err).NotTo(HaveOccurred())

		Expect(instance.Migrator().HasTable(&models.User{})).To(BeTrue())

		var foreignKeys []sqliteForeignKey
		err = instance.Raw("PRAGMA foreign_key_list(rooms)").Scan(&foreignKeys).Error
		Expect(err).NotTo(HaveOccurred())

		var hostFK *sqliteForeignKey
		for i := range foreignKeys {
			if foreignKeys[i].Table == "users" && foreignKeys[i].From == "host_id" && foreignKeys[i].To == "id" {
				hostFK = &foreignKeys[i]
				break
			}
		}

		Expect(hostFK).NotTo(BeNil())
	})
})
