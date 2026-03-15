// SPDX-License-Identifier: Apache-2.0

package evalrunner

import (
	"time"
)

// CheckResult holds the outcome of a single automated check.
type CheckResult struct {
	Check  Check  `json:"check"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail,omitempty"` // e.g., "found as MyModule.Book" or "attribute not found"
}

// PhaseResult holds the results for a single evaluation phase (initial or iteration).
type PhaseResult struct {
	Phase  string        `json:"phase"` // "initial" or "iteration"
	Checks []CheckResult `json:"checks"`
	Passed int           `json:"passed"`
	Total  int           `json:"total"`
	Score  float64       `json:"score"` // 0.0 - 1.0
}

// ComputeScore calculates passed/total/score from the check results.
func (pr *PhaseResult) ComputeScore() {
	pr.Total = len(pr.Checks)
	pr.Passed = 0
	for _, cr := range pr.Checks {
		if cr.Passed {
			pr.Passed++
		}
	}
	if pr.Total > 0 {
		pr.Score = float64(pr.Passed) / float64(pr.Total)
	}
}

// EvalResult holds the complete result of evaluating a single test case.
type EvalResult struct {
	TestID       string        `json:"test_id"`
	Category     string        `json:"category"`
	Title        string        `json:"title"`
	Timestamp    time.Time     `json:"timestamp"`
	Duration     time.Duration `json:"duration"`
	Initial      PhaseResult   `json:"initial"`
	Iteration    *PhaseResult  `json:"iteration,omitempty"`
	OverallScore float64       `json:"overall_score"` // 0.0 - 1.0
	Criteria     []string      `json:"criteria"`
}

// ComputeOverallScore calculates the overall score from phase results.
func (er *EvalResult) ComputeOverallScore() {
	er.Initial.ComputeScore()

	totalPassed := er.Initial.Passed
	totalChecks := er.Initial.Total

	if er.Iteration != nil {
		er.Iteration.ComputeScore()
		totalPassed += er.Iteration.Passed
		totalChecks += er.Iteration.Total
	}

	if totalChecks > 0 {
		er.OverallScore = float64(totalPassed) / float64(totalChecks)
	}
}

// TotalPassed returns the total number of passed checks across all phases.
func (er *EvalResult) TotalPassed() int {
	n := er.Initial.Passed
	if er.Iteration != nil {
		n += er.Iteration.Passed
	}
	return n
}

// TotalChecks returns the total number of checks across all phases.
func (er *EvalResult) TotalChecks() int {
	n := er.Initial.Total
	if er.Iteration != nil {
		n += er.Iteration.Total
	}
	return n
}

// RunSummary holds results for multiple eval tests in a single run.
type RunSummary struct {
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Results   []EvalResult  `json:"results"`
}

// AverageScore returns the average overall score across all results.
func (rs *RunSummary) AverageScore() float64 {
	if len(rs.Results) == 0 {
		return 0
	}
	total := 0.0
	for _, r := range rs.Results {
		total += r.OverallScore
	}
	return total / float64(len(rs.Results))
}
