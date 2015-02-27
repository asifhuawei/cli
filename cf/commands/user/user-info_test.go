package user_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("org-users command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.ReadWriter
		userRepo            *testapi.FakeUserRepository
		spaceRepo           *testapi.FakeSpaceRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		userRepo = &testapi.FakeUserRepository{}
		spaceRepo = &testapi.FakeSpaceRepository{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(ShowUserInfo(ui, configRepo, userRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			Expect(runCommand()).To(BeFalse())
		})

		It("fails when no org or space is targeted", func() {
			requirementsFactory.LoginSuccess = true

			configRepo.SetSpaceFields(models.SpaceFields{})

			configRepo.SetOrganizationFields(models.OrganizationFields{})

			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings([]string{"No org and space targeted"}))
		})

	})

	Context("when logged in and targeted with proper orgs and space", func() {
		BeforeEach(func() {
			org := models.Organization{}
			org.Name = "Org1"
			org.Guid = "org1-guid"
			space := models.Space{}
			space.Name = "Space1"
			space.Guid = "space1-guid"

			user := models.UserFields{}
			user.Username = "user1"

			userRepo.ListUsersByRole = map[string][]models.UserFields{
				models.SPACE_MANAGER:   []models.UserFields{user},
				models.SPACE_DEVELOPER: []models.UserFields{user},
				models.SPACE_AUDITOR:   []models.UserFields{user},
				models.ORG_MANAGER:     []models.UserFields{user},
				models.BILLING_MANAGER: []models.UserFields{user},
				models.ORG_AUDITOR:     []models.UserFields{user},
			}

			requirementsFactory.LoginSuccess = true
			requirementsFactory.UserFields = user
			requirementsFactory.Organization = org
			spaceRepo.FindByNameInOrgSpace = space
			requirementsFactory.Space = space

			configRepo = testconfig.NewRepositoryWithAccessToken(core_config.TokenInfo{Username: "user1"})

			configRepo.SetSpaceFields(models.SpaceFields{
				Name: "Org1",
				Guid: "org1-guid",
			})

			configRepo.SetOrganizationFields(models.OrganizationFields{
				Name: "Space1",
				Guid: "space1-guid",
			})
		})

		It("shows the user-info", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting user information..."},
				[]string{"User", "Org", "Space", "Role"},
				[]string{"user", "Org1", "Space1", "ORG MANAGER", "BILLING MANAGER", "ORG AUDITOR", "SPACE MANAGER", "SPACE DEVELOPER", "SPACE AUDITOR"},
			))
		})

		Context("when cc api verson is >= 2.21.0", func() {
			BeforeEach(func() {
				userRepo.ListUsersInSpaceForRole_CallCount = 0
				userRepo.ListUsersInSpaceForRoleWithNoUAA_CallCount = 0
				userRepo.ListUsersInOrgForRole_CallCount = 0
				userRepo.ListUsersInOrgForRoleWithNoUAA_CallCount = 0
			})

			It("calls User-InfoWithNoUAA()", func() {
				configRepo.SetApiVersion("2.22.0")
				runCommand()

				Expect(userRepo.ListUsersInSpaceForRoleWithNoUAA_CallCount).To(BeNumerically(">=", 1))
				Expect(userRepo.ListUsersInSpaceForRole_CallCount).To(Equal(0))
				Expect(userRepo.ListUsersInOrgForRoleWithNoUAA_CallCount).To(BeNumerically(">=", 1))
				Expect(userRepo.ListUsersInOrgForRole_CallCount).To(Equal(0))
			})
		})

		Context("when cc api verson is < 2.21.0", func() {
			It("calls User-Info()", func() {
				configRepo.SetApiVersion("2.20.0")
				runCommand()

				Expect(userRepo.ListUsersInSpaceForRoleWithNoUAA_CallCount).To(Equal(0))
				Expect(userRepo.ListUsersInSpaceForRole_CallCount).To(BeNumerically(">=", 1))
				Expect(userRepo.ListUsersInOrgForRoleWithNoUAA_CallCount).To(Equal(0))
				Expect(userRepo.ListUsersInOrgForRole_CallCount).To(BeNumerically(">=", 1))
			})
		})

	})
})
