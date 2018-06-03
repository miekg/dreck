package dreck

import (
	"testing"

	"github.com/miekg/dreck/types"
)

func TestEnabledFeature(t *testing.T) {

	var enableFeatureOpts = []struct {
		title            string
		attemptedFeature string
		configFeatures   []string
		expectedVal      bool
	}{
		{
			title:            "dco enabled try dco case sensitive",
			attemptedFeature: "dco",
			configFeatures:   []string{"dco"},
			expectedVal:      true,
		},
		{
			title:            "dco enabled try dco case insensitive",
			attemptedFeature: "dco",
			configFeatures:   []string{"dco"},
			expectedVal:      true,
		},
		{
			title:            "dco enabled try comments",
			attemptedFeature: "comments",
			configFeatures:   []string{"dco"},
			expectedVal:      false,
		},
		{
			title:            "Comments enabled try comments case insensitive",
			attemptedFeature: "Comments",
			configFeatures:   []string{"comments"},
			expectedVal:      true,
		},
		{
			title:            "Comments enabled try comments case sensitive",
			attemptedFeature: "comments",
			configFeatures:   []string{"comments"},
			expectedVal:      true,
		},
		{
			title:            "Comments enabled try dco",
			attemptedFeature: "dco",
			configFeatures:   []string{"comments"},
			expectedVal:      false,
		},
		{
			title:            "Non-existent feature",
			attemptedFeature: "gibberish",
			configFeatures:   []string{"dco", "comments"},
			expectedVal:      false,
		},
		{
			title:            "autosubmit enabled try autosubmit",
			attemptedFeature: "autosubmit",
			configFeatures:   []string{"autosubmit"},
			expectedVal:      true,
		},
		{
			title:            "autosubmit disabled try autosubmit",
			attemptedFeature: "autosubmit",
			configFeatures:   []string{"dco"},
			expectedVal:      false,
		},
	}
	for _, test := range enableFeatureOpts {
		t.Run(test.title, func(t *testing.T) {

			inputConfig := &types.DreckConfig{Features: test.configFeatures}

			featureEnabled := enabledFeature(test.attemptedFeature, inputConfig)

			if featureEnabled != test.expectedVal {
				t.Errorf("Enabled feature - wanted: %t, found %t", test.expectedVal, featureEnabled)
			}
		})
	}
}

func TestPermittedUserFeature(t *testing.T) {

	var permittedUserFeatureOpts = []struct {
		title            string
		attemptedFeature string
		user             string
		config           types.DreckConfig
		expectedVal      bool
	}{
		{
			title:            "Valid feature with valid maintainer",
			attemptedFeature: "comment",
			user:             "Burt",
			config: types.DreckConfig{
				Features:  []string{"comment"},
				Approvers: []string{"Burt", "Tarquin", "Blanche"},
			},
			expectedVal: true,
		},
		{
			title:            "Valid feature with valid maintainer case insensitive",
			attemptedFeature: "comment",
			user:             "burt",
			config: types.DreckConfig{
				Features:  []string{"comment"},
				Approvers: []string{"Burt", "Tarquin", "Blanche"},
			},
			expectedVal: true,
		},
		{
			title:            "Valid feature with invalid maintainer",
			attemptedFeature: "comment",
			user:             "ernie",
			config: types.DreckConfig{
				Features:  []string{"comment"},
				Approvers: []string{"Burt", "Tarquin", "Blanche"},
			},
			expectedVal: false,
		},
		{
			title:            "Valid feature with invalid maintainer case insensitive",
			attemptedFeature: "Comment",
			user:             "ernie",
			config: types.DreckConfig{
				Features:  []string{"comment"},
				Approvers: []string{"Burt", "Tarquin", "Blanche"},
			},
			expectedVal: false,
		},
		{
			title:            "Invalid feature with valid maintainer",
			attemptedFeature: "invalid",
			user:             "Burt",
			config: types.DreckConfig{
				Features:  []string{"comment"},
				Approvers: []string{"Burt", "Tarquin", "Blanche"},
			},
			expectedVal: false,
		},
		{
			title:            "Invalid feature with valid maintainer case insensitive",
			attemptedFeature: "invalid",
			user:             "burt",
			config: types.DreckConfig{
				Features:  []string{"comment"},
				Approvers: []string{"Burt", "Tarquin", "Blanche"},
			},
			expectedVal: false,
		},
	}

	for _, test := range permittedUserFeatureOpts {
		t.Run(test.title, func(t *testing.T) {

			inputConfig := &test.config

			permittedFeature := permittedUserFeature(test.attemptedFeature, inputConfig, test.user)

			if permittedFeature != test.expectedVal {
				t.Errorf("Permitted user feature - wanted: %t, found %t", test.expectedVal, permittedFeature)
			}
		})
	}
}
