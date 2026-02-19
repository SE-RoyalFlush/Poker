package models_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
)

var _ = Describe("Models", func() {
	Describe("AllModels", func() {
		It("should contain the User model", func() {
			allModels := models.AllModels()
			Expect(allModels).To(ContainElement(&models.User{}))
		})
	})
})
