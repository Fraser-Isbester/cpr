package commit_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fraser-isbester/cpr/internal/commit"
)

var _ = Describe("Analyzer", func() {
	Describe("GenerateTitle", func() {
		Context("when adding new functions", func() {
			It("should generate a feat commit type", func() {
				diff := `
diff --git a/main.go b/main.go
index abc123..def456 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,7 @@
 package main

+func NewFeature() string {
+	return "feature"
+}
`
				files := []string{"main.go"}
				analyzer := commit.NewAnalyzer(diff, files)
				title := analyzer.GenerateTitle()
				Expect(title).To(MatchRegexp("^feat.*"))
			})
		})

		Context("when fixing bugs", func() {
			It("should generate a fix commit type", func() {
				diff := `
diff --git a/main.go b/main.go
index abc123..def456 100644
--- a/main.go
+++ b/main.go
@@ -10,7 +10,7 @@ func Process() error {
-	if data == nil {
-		// This causes a panic
-		return data.Error()
+	if data == nil {
+		// Fixed nil pointer dereference
+		return errors.New("data is nil")
 	}
`
				files := []string{"main.go"}
				analyzer := commit.NewAnalyzer(diff, files)
				title := analyzer.GenerateTitle()
				Expect(title).To(MatchRegexp("^fix.*"))
			})
		})

		Context("when updating tests", func() {
			It("should generate a test commit type", func() {
				diff := `
diff --git a/main_test.go b/main_test.go
index abc123..def456 100644
--- a/main_test.go
+++ b/main_test.go
@@ -1,3 +1,7 @@
 package main

+func TestNewFeature(t *testing.T) {
+	// test implementation
+}
`
				files := []string{"main_test.go"}
				analyzer := commit.NewAnalyzer(diff, files)
				title := analyzer.GenerateTitle()
				Expect(title).To(MatchRegexp("^test.*"))
			})
		})

		Context("when updating documentation", func() {
			It("should generate a docs commit type", func() {
				diff := `
diff --git a/README.md b/README.md
index abc123..def456 100644
--- a/README.md
+++ b/README.md
@@ -1,3 +1,5 @@
 # Project
+
+## New Section
`
				files := []string{"README.md"}
				analyzer := commit.NewAnalyzer(diff, files)
				title := analyzer.GenerateTitle()
				Expect(title).To(MatchRegexp("^docs.*"))
			})
		})

		Context("when updating build files", func() {
			It("should generate a build commit type", func() {
				diff := `
diff --git a/go.mod b/go.mod
index abc123..def456 100644
--- a/go.mod
+++ b/go.mod
@@ -3,3 +3,5 @@ module example.com/project
 go 1.21
+
+require github.com/spf13/cobra v1.8.0
`
				files := []string{"go.mod"}
				analyzer := commit.NewAnalyzer(diff, files)
				title := analyzer.GenerateTitle()
				Expect(title).To(MatchRegexp("^build.*"))
			})
		})

		Context("with scope detection", func() {
			It("should include scope when files are in internal directory", func() {
				diff := `
diff --git a/internal/auth/handler.go b/internal/auth/handler.go
index abc123..def456 100644
--- a/internal/auth/handler.go
+++ b/internal/auth/handler.go
@@ -1,3 +1,7 @@
 package auth

+func NewHandler() *Handler {
+	return &Handler{}
+}
`
				files := []string{"internal/auth/handler.go", "internal/auth/middleware.go"}
				analyzer := commit.NewAnalyzer(diff, files)
				title := analyzer.GenerateTitle()
				Expect(title).To(MatchRegexp("^feat\\(auth\\):.*"))
			})
		})
	})

	Describe("GenerateSummary", func() {
		It("should generate a summary with changed files", func() {
			diff := `
diff --git a/main.go b/main.go
index abc123..def456 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,7 @@
 package main

+func NewFeature() string {
+	return "feature"
+}
`
			files := []string{"main.go", "util.go"}
			analyzer := commit.NewAnalyzer(diff, files)
			summary := analyzer.GenerateSummary()

			Expect(summary).To(ContainSubstring("## Summary"))
			Expect(summary).To(ContainSubstring("## Changed Files"))
			Expect(summary).To(ContainSubstring("`main.go`"))
			Expect(summary).To(ContainSubstring("`util.go`"))
		})
	})
})
