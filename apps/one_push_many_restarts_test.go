// This is a defensive test against the CC no longer knowing how to find an
// existing app's bits. This can happen if the scheme of the app's paths in
// the blobstore changes without being backwards-compatible.
//
// If this is not caught before a deploy, all running apps will go down, as
// during evacuation of the DEAs, the CC will not know to look in their old
// path format in the blob store.
//
// This tests pushes the app once (checking if it already exists), and then
// just restarts it on later runs.

package apps

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vito/cmdtest"
	. "github.com/vito/cmdtest/matchers"

	. "github.com/cloudfoundry/cf-acceptance-tests/helpers"
	. "github.com/pivotal-cf-experimental/cf-test-helpers/cf"
)

var _ = Describe("An application that's already been pushed", func() {
	var appName string
	var originalCfHomeDir, currentCfHomeDir string
	var config = LoadConfig()
	var adminUserContext, persistentAppUserContext UserContext

	BeforeEach(func() {
		context := NewPersistentAppContext(config)

		adminUserContext = context.AdminUserContext()
		persistentAppUserContext = context.PersistentAppUserContext()

		context.Setup()

		AsUser(adminUserContext, func() {
			SetUpSpaceWithUserAccess(persistentAppUserContext, persistentAppUserContext.Space)
		})

		originalCfHomeDir, currentCfHomeDir = InitiateUserContext(persistentAppUserContext)
		TargetSpace(persistentAppUserContext)

		appName = config.PersistentAppHost

		Expect(Cf("app", appName)).To(SayBranches(
			cmdtest.ExpectBranch{
				"not found",
				func() {
					Expect(
						Cf("push", appName, "-p", NewAssets().Dora),
					).To(Say("App started"))
				},
			},
			cmdtest.ExpectBranch{"running", func() {}},
		))
	})

	AfterEach(func() {
		RestoreUserContext(persistentAppUserContext, originalCfHomeDir, currentCfHomeDir)

		AsUser(adminUserContext, func() {
			Expect(Cf("delete-user", "-f", persistentAppUserContext.Username)).To(ExitWith(0))
		})
	})

	It("can be restarted and still come up", func() {
		Eventually(Curling(appName, "/", config.AppsDomain)).Should(Say("Hi, I'm Dora!"))

		Expect(Cf("stop", appName)).To(ExitWith(0))

		Eventually(Curling(appName, "/", config.AppsDomain)).Should(Say("404"))

		Expect(Cf("start", appName)).To(Say("App started"))

		Eventually(Curling(appName, "/", config.AppsDomain)).Should(Say("Hi, I'm Dora!"))
	})
})
