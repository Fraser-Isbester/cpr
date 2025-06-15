package commit

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type CommitType string

const (
	TypeFeat     CommitType = "feat"
	TypeFix      CommitType = "fix"
	TypeDocs     CommitType = "docs"
	TypeStyle    CommitType = "style"
	TypeRefactor CommitType = "refactor"
	TypePerf     CommitType = "perf"
	TypeTest     CommitType = "test"
	TypeBuild    CommitType = "build"
	TypeCI       CommitType = "ci"
	TypeChore    CommitType = "chore"
)

type Analyzer struct {
	diff          string
	changedFiles  []string
}

func NewAnalyzer(diff string, changedFiles []string) *Analyzer {
	return &Analyzer{
		diff:         diff,
		changedFiles: changedFiles,
	}
}

func (a *Analyzer) GenerateTitle() string {
	commitType := a.detectCommitType()
	scope := a.detectScope()
	description := a.generateDescription()
	
	if scope != "" {
		return fmt.Sprintf("%s(%s): %s", commitType, scope, description)
	}
	return fmt.Sprintf("%s: %s", commitType, description)
}

func (a *Analyzer) GenerateSummary() string {
	var summary strings.Builder
	
	summary.WriteString("## Summary\n\n")
	
	changes := a.analyzeChanges()
	for _, change := range changes {
		summary.WriteString(fmt.Sprintf("- %s\n", change))
	}
	
	if len(a.changedFiles) > 0 {
		summary.WriteString("\n## Changed Files\n\n")
		for _, file := range a.changedFiles {
			summary.WriteString(fmt.Sprintf("- `%s`\n", file))
		}
	}
	
	return summary.String()
}

func (a *Analyzer) detectCommitType() CommitType {
	lowerDiff := strings.ToLower(a.diff)
	
	testFilePattern := regexp.MustCompile(`(?i)(_test\.go|\.test\.|spec\.|test/)`)
	hasTestFiles := false
	for _, file := range a.changedFiles {
		if testFilePattern.MatchString(file) {
			hasTestFiles = true
			break
		}
	}
	if hasTestFiles && !a.hasNonTestChanges() {
		return TypeTest
	}
	
	if a.hasFilePattern(`(?i)(readme|\.md$|docs/)`) {
		return TypeDocs
	}
	
	if a.hasFilePattern(`(?i)(makefile|dockerfile|\.yml$|\.yaml$|go\.mod|go\.sum|package\.json)`) {
		return TypeBuild
	}
	
	if a.hasFilePattern(`(?i)(\.github/|\.circleci/|\.travis|jenkins)`) {
		return TypeCI
	}
	
	bugPatterns := []string{
		`fix\s*\(`,
		`bug\s*fix`,
		`error\s*handling`,
		`nil\s*pointer`,
		`panic`,
		`segfault`,
		`crash`,
		`exception`,
	}
	for _, pattern := range bugPatterns {
		if match, _ := regexp.MatchString(pattern, lowerDiff); match {
			return TypeFix
		}
	}
	
	perfPatterns := []string{
		`performance`,
		`optimize`,
		`speed\s*up`,
		`reduce\s*memory`,
		`cache`,
	}
	for _, pattern := range perfPatterns {
		if match, _ := regexp.MatchString(pattern, lowerDiff); match {
			return TypePerf
		}
	}
	
	refactorPatterns := []string{
		`refactor`,
		`rename`,
		`move\s*to`,
		`extract`,
		`simplify`,
		`clean\s*up`,
	}
	for _, pattern := range refactorPatterns {
		if match, _ := regexp.MatchString(pattern, lowerDiff); match {
			return TypeRefactor
		}
	}
	
	if strings.Contains(lowerDiff, "+func") || strings.Contains(lowerDiff, "+type") || 
	   strings.Contains(lowerDiff, "+struct") || strings.Contains(lowerDiff, "+interface") {
		return TypeFeat
	}
	
	return TypeFeat
}

func (a *Analyzer) detectScope() string {
	if len(a.changedFiles) == 0 {
		return ""
	}
	
	commonPrefixes := make(map[string]int)
	for _, file := range a.changedFiles {
		dir := filepath.Dir(file)
		parts := strings.Split(dir, "/")
		
		for i, part := range parts {
			if part == "internal" || part == "pkg" || part == "cmd" {
				if i+1 < len(parts) {
					scope := parts[i+1]
					commonPrefixes[scope]++
				}
				break
			}
		}
		
		if strings.HasPrefix(file, "cmd/") {
			commonPrefixes["cli"]++
		}
	}
	
	var maxScope string
	maxCount := 0
	for scope, count := range commonPrefixes {
		if count > maxCount {
			maxScope = scope
			maxCount = count
		}
	}
	
	return maxScope
}

func (a *Analyzer) generateDescription() string {
	commitType := a.detectCommitType()
	
	primaryChanges := a.analyzePrimaryChanges()
	if len(primaryChanges) > 0 {
		return primaryChanges[0]
	}
	
	switch commitType {
	case TypeFeat:
		return "add new functionality"
	case TypeFix:
		return "resolve issues"
	case TypeDocs:
		return "update documentation"
	case TypeTest:
		return "add tests"
	case TypeBuild:
		return "update build configuration"
	case TypeCI:
		return "update CI configuration"
	case TypeRefactor:
		return "improve code structure"
	case TypePerf:
		return "improve performance"
	default:
		return "update codebase"
	}
}

func (a *Analyzer) analyzePrimaryChanges() []string {
	var changes []string
	
	funcPattern := regexp.MustCompile(`^\+func\s+(\w+)`)
	typePattern := regexp.MustCompile(`^\+type\s+(\w+)`)
	
	lines := strings.Split(a.diff, "\n")
	for _, line := range lines {
		if matches := funcPattern.FindStringSubmatch(line); len(matches) > 1 {
			changes = append(changes, fmt.Sprintf("add %s function", matches[1]))
		} else if matches := typePattern.FindStringSubmatch(line); len(matches) > 1 {
			changes = append(changes, fmt.Sprintf("add %s type", matches[1]))
		}
	}
	
	if len(changes) > 3 {
		return changes[:3]
	}
	
	return changes
}

func (a *Analyzer) analyzeChanges() []string {
	changes := a.analyzePrimaryChanges()
	
	if a.hasFilePattern(`go\.mod`) {
		changes = append(changes, "Update dependencies")
	}
	
	if a.hasFilePattern(`(?i)test`) {
		changes = append(changes, "Add or update tests")
	}
	
	if len(changes) == 0 {
		changes = append(changes, "Various improvements and updates")
	}
	
	return changes
}

func (a *Analyzer) hasFilePattern(pattern string) bool {
	re := regexp.MustCompile(pattern)
	for _, file := range a.changedFiles {
		if re.MatchString(file) {
			return true
		}
	}
	return false
}

func (a *Analyzer) hasNonTestChanges() bool {
	testPattern := regexp.MustCompile(`(?i)(_test\.go|\.test\.|spec\.|test/)`)
	for _, file := range a.changedFiles {
		if !testPattern.MatchString(file) {
			return true
		}
	}
	return false
}