package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabels(t *testing.T) {
	testCases := []struct {
		name                string
		have                label
		expectedDescription string
		expectedString      string
	}{
		{
			"ShouldShowCorrectPriorityCriticalValues",
			labelPriorityCritical,
			"Critical",
			"priority/1/critical",
		},
		{
			"ShouldShowCorrectPriorityHighValues",
			labelPriorityHigh,
			"High",
			"priority/2/high",
		},
		{
			"ShouldShowCorrectPriorityMediumValues",
			labelPriorityMedium,
			"Medium",
			"priority/3/medium",
		},
		{
			"ShouldShowCorrectPriorityNormalValues",
			labelPriorityNormal,
			"Normal",
			"priority/4/normal",
		},
		{
			"ShouldShowCorrectPriorityLowValues",
			labelPriorityLow,
			"Low",
			"priority/5/low",
		},
		{
			"ShouldShowCorrectPriorityVeryLowValues",
			labelPriorityVeryLow,
			"Very Low",
			"priority/6/very-low",
		},
		{
			"ShouldShowCorrectStatusNeedsDesignValues",
			labelStatusNeedsDesign,
			"Needs Design",
			"status/needs-design",
		},
		{
			"ShouldShowCorrectStatusNeedsTriageValues",
			labelStatusNeedsTriage,
			"Needs Triage",
			"status/needs-triage",
		},
		{
			"ShouldShowCorrectTypeFeatureValues",
			labelTypeFeature,
			"Feature",
			"type/feature",
		},
		{
			"ShouldShowCorrectTypeBugUnconfirmedValues",
			labelTypeBugUnconfirmed,
			"Bug: Unconfirmed",
			"type/bug/unconfirmed",
		},
		{
			"ShouldShowCorrectTypeBugValues",
			labelTypeBug,
			"Bug",
			"type/bug",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedString, tc.have.String())
			assert.Equal(t, tc.expectedDescription, tc.have.LabelDescription())
		})
	}
}
