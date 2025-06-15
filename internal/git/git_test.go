package git_test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fraser-isbester/cpr/internal/git"
)

var _ = Describe("Git Repository", func() {
	var (
		tmpDir string
		repo   *git.Repository
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "cpr-test-*")
		Expect(err).NotTo(HaveOccurred())

		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		err = cmd.Run()
		Expect(err).NotTo(HaveOccurred())

		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = tmpDir
		err = cmd.Run()
		Expect(err).NotTo(HaveOccurred())

		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = tmpDir
		err = cmd.Run()
		Expect(err).NotTo(HaveOccurred())

		testFile := filepath.Join(tmpDir, "test.txt")
		err = os.WriteFile(testFile, []byte("initial content"), 0644)
		Expect(err).NotTo(HaveOccurred())

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		err = cmd.Run()
		Expect(err).NotTo(HaveOccurred())

		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = tmpDir
		err = cmd.Run()
		Expect(err).NotTo(HaveOccurred())

		repo = git.NewRepository(tmpDir)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("CurrentBranch", func() {
		Context("when on a branch", func() {
			It("should return the current branch name", func() {
				branch, err := repo.CurrentBranch()
				Expect(err).NotTo(HaveOccurred())
				Expect(branch).To(Or(Equal("main"), Equal("master")))
			})
		})

		Context("when on a new branch", func() {
			BeforeEach(func() {
				cmd := exec.Command("git", "checkout", "-b", "feature-branch")
				cmd.Dir = tmpDir
				err := cmd.Run()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return the new branch name", func() {
				branch, err := repo.CurrentBranch()
				Expect(err).NotTo(HaveOccurred())
				Expect(branch).To(Equal("feature-branch"))
			})
		})
	})

	Describe("DefaultBranch", func() {
		Context("when origin remote exists", func() {
			BeforeEach(func() {
				cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/example/repo.git")
				cmd.Dir = tmpDir
				_ = cmd.Run()
			})

			It("should detect main or master as default branch", func() {
				branch, err := repo.DefaultBranch()
				Expect(err).NotTo(HaveOccurred())
				Expect(branch).To(Or(Equal("main"), Equal("master")))
			})
		})
	})

	Describe("GetChangedFiles", func() {
		Context("when files are changed", func() {
			BeforeEach(func() {
				cmd := exec.Command("git", "checkout", "-b", "feature-branch")
				cmd.Dir = tmpDir
				err := cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				newFile := filepath.Join(tmpDir, "new-file.txt")
				err = os.WriteFile(newFile, []byte("new content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				cmd = exec.Command("git", "add", ".")
				cmd.Dir = tmpDir
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				cmd = exec.Command("git", "commit", "-m", "Add new file")
				cmd.Dir = tmpDir
				err = cmd.Run()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return the list of changed files", func() {
				files, err := repo.GetChangedFiles()
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ContainElement("new-file.txt"))
			})
		})
	})
})