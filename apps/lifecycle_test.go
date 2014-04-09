package apps

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/vito/cmdtest/matchers"

	. "github.com/cloudfoundry/cf-acceptance-tests/helpers"
	. "github.com/pivotal-cf-experimental/cf-test-helpers/cf"
	. "github.com/pivotal-cf-experimental/cf-test-helpers/generator"
)

type AppUsageEvent struct {
	Entity struct {
		AppName       string `json:"app_name"`
		BuildpackName string `json:"buildpack_name"`
		BuildpackGuid string `json:"buildpack_guid"`
	} `json:"entity"`
}

type AppUsageEvents struct {
	Resources []AppUsageEvent `struct:"resources"`
}

var _ = Describe("Application", func() {
	var appName string

	BeforeEach(func() {
		appName = RandomName()

		Expect(Cf("push", appName, "-p", NewAssets().Dora)).To(Say("App started"))
	})

	AfterEach(func() {
		Expect(Cf("delete", appName, "-f")).To(Say("OK"))
	})

	Describe("pushing", func() {
		It("makes the app reachable via its bound route", func() {
			Eventually(Curling(appName, "/", LoadConfig().AppsDomain)).Should(Say("Hi, I'm Dora!"))
		})

		FIt("generates an app usage event", func() {
			var response AppUsageEvents
			AsUser(AdminUserContext, func() {
				ApiRequest("GET", "/v2/app_usage_events?order-direction=desc&page=1", &response)
			})

			println("***************")
			println(appName)

			var matchingEvent AppUsageEvent
			for _, event := range response.Resources {
				fmt.Printf("%#v\n", event)
				if event.Entity.AppName == appName {
					matchingEvent = event
					break
				}
			}

			Expect(matchingEvent).ToNot(BeNil())
			Expect(matchingEvent.Entity.BuildpackName).To(Equal("ruby_buildpack"))
			Expect(matchingEvent.Entity.BuildpackGuid).ToNot(BeNil())
		})
	})

	Describe("stopping", func() {
		BeforeEach(func() {
			Expect(Cf("stop", appName)).To(Say("OK"))
		})

		It("makes the app unreachable", func() {
			Eventually(Curling(appName, "/", LoadConfig().AppsDomain), 5.0).Should(Say("404"))
		})

		Describe("and then starting", func() {
			BeforeEach(func() {
				Expect(Cf("start", appName)).To(Say("App started"))
			})

			It("makes the app reachable again", func() {
				Eventually(Curling(appName, "/", LoadConfig().AppsDomain)).Should(Say("Hi, I'm Dora!"))
			})
		})
	})

	Describe("updating", func() {
		It("is reflected through another push", func() {
			Eventually(Curling(appName, "/", LoadConfig().AppsDomain)).Should(Say("Hi, I'm Dora!"))

			Expect(Cf("push", appName, "-p", NewAssets().HelloWorld)).To(Say("App started"))

			Eventually(Curling(appName, "/", LoadConfig().AppsDomain)).Should(Say("Hello, world!"))
		})
	})

	Describe("deleting", func() {
		BeforeEach(func() {
			Expect(Cf("delete", appName, "-f")).To(Say("OK"))
		})

		It("removes the application", func() {
			Expect(Cf("app", appName)).To(Say("not found"))
		})

		It("makes the app unreachable", func() {
			Eventually(Curling(appName, "/", LoadConfig().AppsDomain)).Should(Say("404"))
		})
	})
})
