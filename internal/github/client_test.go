package github_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fraser-isbester/cpr/internal/github"
)

var _ = Describe("GitHub Client", func() {
	Describe("ParseGitRemoteURL", func() {
		Context("with SSH URLs", func() {
			It("should parse standard SSH URL", func() {
				owner, repo, err := github.ParseGitRemoteURL("git@github.com:octocat/Hello-World.git")
				Expect(err).NotTo(HaveOccurred())
				Expect(owner).To(Equal("octocat"))
				Expect(repo).To(Equal("Hello-World"))
			})

			It("should parse SSH URL without .git extension", func() {
				owner, repo, err := github.ParseGitRemoteURL("git@github.com:octocat/Hello-World")
				Expect(err).NotTo(HaveOccurred())
				Expect(owner).To(Equal("octocat"))
				Expect(repo).To(Equal("Hello-World"))
			})
		})

		Context("with HTTPS URLs", func() {
			It("should parse standard HTTPS URL", func() {
				owner, repo, err := github.ParseGitRemoteURL("https://github.com/octocat/Hello-World.git")
				Expect(err).NotTo(HaveOccurred())
				Expect(owner).To(Equal("octocat"))
				Expect(repo).To(Equal("Hello-World"))
			})

			It("should parse HTTPS URL without .git extension", func() {
				owner, repo, err := github.ParseGitRemoteURL("https://github.com/octocat/Hello-World")
				Expect(err).NotTo(HaveOccurred())
				Expect(owner).To(Equal("octocat"))
				Expect(repo).To(Equal("Hello-World"))
			})

			It("should parse HTTP URL", func() {
				owner, repo, err := github.ParseGitRemoteURL("http://github.com/octocat/Hello-World.git")
				Expect(err).NotTo(HaveOccurred())
				Expect(owner).To(Equal("octocat"))
				Expect(repo).To(Equal("Hello-World"))
			})
		})

		Context("with invalid URLs", func() {
			It("should return error for invalid format", func() {
				_, _, err := github.ParseGitRemoteURL("not-a-valid-url")
				Expect(err).To(HaveOccurred())
			})

			It("should parse non-GitHub URLs", func() {
				owner, repo, err := github.ParseGitRemoteURL("https://gitlab.com/user/repo.git")
				Expect(err).NotTo(HaveOccurred())
				Expect(owner).To(Equal("user"))
				Expect(repo).To(Equal("repo"))
			})
		})
	})

	Describe("GetToken", func() {
		var originalGitHubToken string
		var originalGHToken string

		BeforeEach(func() {
			originalGitHubToken = os.Getenv("GITHUB_TOKEN")
			originalGHToken = os.Getenv("GH_TOKEN")
			os.Unsetenv("GITHUB_TOKEN")
			os.Unsetenv("GH_TOKEN")
		})

		AfterEach(func() {
			if originalGitHubToken != "" {
				os.Setenv("GITHUB_TOKEN", originalGitHubToken)
			}
			if originalGHToken != "" {
				os.Setenv("GH_TOKEN", originalGHToken)
			}
		})

		Context("when GITHUB_TOKEN is set", func() {
			It("should return the token", func() {
				os.Setenv("GITHUB_TOKEN", "test-token")
				token, err := github.GetToken()
				Expect(err).NotTo(HaveOccurred())
				Expect(token).To(Equal("test-token"))
			})
		})

		Context("when GH_TOKEN is set", func() {
			It("should return the token", func() {
				os.Setenv("GH_TOKEN", "gh-test-token")
				token, err := github.GetToken()
				Expect(err).NotTo(HaveOccurred())
				Expect(token).To(Equal("gh-test-token"))
			})
		})

		Context("when no token is set", func() {
			It("should return an error", func() {
				_, err := github.GetToken()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("GitHub token not found"))
			})
		})
	})
})