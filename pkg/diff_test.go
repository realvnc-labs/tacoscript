package pkg

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCalcDiff(t *testing.T) {
	testCases := []struct{
		name string
		linesBefore string
		linesAfter string
		expectedDiff *Diff
	}{
		{
			name: "no diff, single line",
			linesBefore: "readline-common 7.0-3",
			linesAfter: "readline-common 7.0-3",
			expectedDiff: nil,
		},
		{
			name: "no diff, multi line",
			linesBefore: `readline-common 7.0-3
sed/now 4.4-2
`,
			linesAfter: `readline-common 7.0-3
sed/now 4.4-2

`,
			expectedDiff: nil,
		},
		{
			name: "possibly conflicting names, multi line",
			linesBefore: `readline-common 7.0-3
sed/now 4.4-2
`,
			linesAfter: `readline-common 7.0-3
sed/now 4.4-2
readline 7.0-3
seda 4.4-2
`,
			expectedDiff: &Diff{
				Added:   []string{
					"readline 7.0-3",
					"seda 4.4-2",
				},
				Removed: []string{},
			},
		},
		{
			name: "one added one removed, one changed, multi line",
			linesBefore: `readline-common 7.0-3
sed/now 4.4-2
`,
			linesAfter: `readline-common 7.0-4
util-linux/now 2.31.1-0.4ubuntu3.7

`,
			expectedDiff: &Diff{
				Added:   []string{
					"readline-common 7.0-4",
					"util-linux/now 2.31.1-0.4ubuntu3.7",
				},
				Removed: []string{
					"readline-common 7.0-3",
					"sed/now 4.4-2",
				},
			},
		},
		{
			name: "all removed, multi line",
			linesBefore: `readline-common 7.0-3
sed/now 4.4-2
`,
			linesAfter: "",
			expectedDiff: &Diff{
				Added:   []string{},
				Removed: []string{
					"readline-common 7.0-3",
					"sed/now 4.4-2",
				},
			},
		},
		{
			name: "all added, multi line",
			linesBefore: "",
			linesAfter: `readline-common 7.0-3
sed/now 4.4-2
`,
			expectedDiff: &Diff{
				Added: []string{
					"readline-common 7.0-3",
					"sed/now 4.4-2",
				},
				Removed:   []string{},
			},
		},
		{
			name: "all changed, multi line",
			linesBefore: `readline-common 7.0-3
sed/now 4.4-2
`,
			linesAfter: `readline-common 7.0-4
sed/now 4.4-3
`,
			expectedDiff: &Diff{
				Added: []string{
					"readline-common 7.0-4",
					"sed/now 4.4-3",
				},
				Removed:   []string{
					"readline-common 7.0-3",
					"sed/now 4.4-2",
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			packagesBefore := strings.Split(tc.linesBefore, "\n")
			packagesAfter := strings.Split(tc.linesAfter, "\n")
			actualDiff := CalcDiff(packagesBefore, packagesAfter)
			if tc.expectedDiff == nil {
				assert.Nil(t, actualDiff)
				return
			}

			assert.EqualValues(t, tc.expectedDiff, actualDiff)
		})
	}
}
