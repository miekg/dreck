package dreck

import "testing"

func TestIsAutosubmit(t *testing.T) {

	var autosubmit = []struct {
		title        string
		message      string
		expectedBool bool
	}{
		{
			title:        "Correctly autosubmit string",
			message:      "/autosubmit",
			expectedBool: true,
		},
		{
			title:        "Correctly signed off within string",
			message:      "This PR will /autosubmit soon",
			expectedBool: true,
		},
		{
			title:        "Not correct signed full string",
			message:      "autosubmit",
			expectedBool: false,
		},
	}
	for _, test := range autosubmit {
		t.Run(test.title, func(t *testing.T) {

			containsSignoff := isAutosubmit(test.message)

			if containsSignoff != test.expectedBool {
				t.Errorf("Is autosubmit - Testing '%s'  - wanted: %t, found %t", test.message, test.expectedBool, containsSignoff)
			}
		})
	}
}
