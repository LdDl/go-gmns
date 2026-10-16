package generators

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/stretchr/testify/assert"
)

type preparationUsesCase struct {
	name      string
	inbound   string
	outbound  string
	movement  string
	expected  string
	errorText string
}

func preparationUsesNetwork(t *testing.T, testCase *preparationUsesCase) *macro.Net {
	t.Helper()
	net := preparationTestNetwork(t, gmns.IDTypeString)
	for _, name := range []string{"auto", "bike", "walk", "cargo_robot"} {
		err := net.AddUseDefinition(gmns.NewUseDefinition(name, 1, 0.5,
			gmns.WithUseDescription("supplied definition"),
			gmns.WithUseSpecialConditions("opaque supplied condition"),
		))
		assert.NoError(t, err)
	}
	err := net.AddUseGroup(gmns.NewUseGroup("motor", []string{"cargo_robot", "auto"}, gmns.WithUseGroupDescription("supplied group")))
	assert.NoError(t, err)
	macro.WithAllowedUses(testCase.inbound)(net.Links[gmns.LinkIDFromString("13")])
	macro.WithAllowedUses(testCase.outbound)(net.Links[gmns.LinkIDFromString("900719925474099312345")])
	record := net.MovementRecords()[0]
	movement.WithAllowedUses(testCase.movement)(record)
	err = net.UpdateMovement(record)
	assert.NoError(t, err)
	return net
}

func TestPrepareMesoscopicPreservesStandardUsePermissions(t *testing.T) {
	cases := []preparationUsesCase{
		{
			name: "all unspecified",
		},
		{
			name:     "road intersection",
			inbound:  "auto,bike",
			outbound: "bike,walk",
			expected: "bike",
		},
		{
			name:     "movement only",
			movement: "bike",
			expected: "bike",
		},
		{
			name:     "group and movement restrict road",
			inbound:  "motor",
			outbound: "auto",
			movement: "motor",
			expected: "auto",
		},
		{
			name:     "custom use is not a built-in agent",
			inbound:  "motor",
			outbound: "cargo_robot",
			movement: "cargo_robot",
			expected: "cargo_robot",
		},
		{
			name:     "preserve spelling on source roads and sort connector",
			inbound:  "bike, auto, bike",
			outbound: "auto,bike",
			expected: "auto,bike",
		},
		{
			name:      "conflicting roads",
			inbound:   "auto",
			outbound:  "walk",
			errorText: "movement 49 allowed_uses: no common use",
		},
		{
			name:      "conflicting movement",
			inbound:   "motor",
			outbound:  "auto",
			movement:  "bike",
			errorText: "movement 49 allowed_uses: no common use",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			source := preparationUsesNetwork(t, &testCase)
			expectedSource := preparationUsesNetwork(t, &testCase)
			result, err := PrepareMesoscopic(source, WithUniformLaneLayout())
			assert.Equal(t, expectedSource, source)
			if testCase.errorText != "" {
				assertPreparationError(t, err, testCase.errorText)
				assert.Nil(t, result)
				return
			}
			assert.NoError(t, err)
			if err != nil {
				return
			}
			assert.NoError(t, result.Macro().ValidateDataset())
			assert.NoError(t, result.Meso().ValidateDataset())
			assert.Equal(t, source.UseDefinitionRecords(), result.Meso().UseDefinitionRecords())
			assert.Equal(t, source.UseGroupRecords(), result.Meso().UseGroupRecords())
			assert.Equal(t, testCase.movement, result.Macro().MovementRecords()[0].AllowedUses())
			for _, link := range result.Meso().Links {
				if !link.Movement().IsZero() {
					assert.Equal(t, testCase.expected, link.AllowedUses())
				} else {
					assert.Equal(t, source.Links[link.MacroLink()].AllowedUses(), link.AllowedUses())
				}
				assert.Empty(t, link.AllowedAgentTypes())
			}
			second, err := PrepareMesoscopic(source, WithUniformLaneLayout())
			assert.NoError(t, err)
			assert.Equal(t, result, second)
		})
	}
}
