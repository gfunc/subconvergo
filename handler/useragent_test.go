package handler

import (
	"testing"
)

func TestMatchUserAgent(t *testing.T) {
	tests := []struct {
		ua             string
		expectedTarget string
		expectedCNN    *bool // ClashNewName
		expectedSV     int   // SurgeVer
	}{
		{
			ua:             "ClashForAndroid/2.5.12",
			expectedTarget: "clash",
			expectedCNN:    &trueVal,
			expectedSV:     -1,
		},
		{
			ua:             "ClashForAndroid/1.0.0",
			expectedTarget: "clash",
			expectedCNN:    &falseVal,
			expectedSV:     -1,
		},
		{
			ua:             "ClashForAndroid/2.5.12R",
			expectedTarget: "clash", // Matches first rule
			expectedCNN:    &trueVal,
			expectedSV:     -1,
		},
		{
			ua:             "ClashforWindows/0.19.11",
			expectedTarget: "clash",
			expectedCNN:    &trueVal,
			expectedSV:     -1,
		},
		{
			ua:             "Surge/2000 (iPhone; iOS 14.0; Scale/3.00)",
			expectedTarget: "surge",
			expectedCNN:    &falseVal,
			expectedSV:     4,
		},
		{
			ua:             "Surge/1000 (iPhone; iOS 13.0; Scale/3.00)",
			expectedTarget: "surge",
			expectedCNN:    &falseVal,
			expectedSV:     3,
		},
		{
			ua:             "Quantumult%20X/1.0.0",
			expectedTarget: "quanx",
			expectedCNN:    nil,
			expectedSV:     -1,
		},
		{
			ua:             "Unknown/1.0",
			expectedTarget: "",
			expectedCNN:    nil,
			expectedSV:     -1,
		},
	}

	for _, tt := range tests {
		target, cnn, sv := matchUserAgent(tt.ua)
		if target != tt.expectedTarget {
			t.Errorf("matchUserAgent(%q) target = %v, want %v", tt.ua, target, tt.expectedTarget)
		}
		if cnn != nil && tt.expectedCNN != nil {
			if *cnn != *tt.expectedCNN {
				t.Errorf("matchUserAgent(%q) cnn = %v, want %v", tt.ua, *cnn, *tt.expectedCNN)
			}
		} else if cnn != tt.expectedCNN {
			t.Errorf("matchUserAgent(%q) cnn = %v, want %v", tt.ua, cnn, tt.expectedCNN)
		}
		if sv != tt.expectedSV {
			t.Errorf("matchUserAgent(%q) sv = %v, want %v", tt.ua, sv, tt.expectedSV)
		}
	}
}
