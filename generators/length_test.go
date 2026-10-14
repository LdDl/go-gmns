package generators

import (
	"math"
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/movement"
	"github.com/stretchr/testify/assert"
)

type requiredLengthCase struct {
	name    string
	value   float64
	present bool
}

func TestGeneratorsRejectUnusableRequiredLengths(t *testing.T) {
	cases := []requiredLengthCase{
		{
			name: "absent",
		},
		{
			name:    "zero",
			present: true,
		},
		{
			name:    "negative",
			value:   -1,
			present: true,
		},
		{
			name:    "NaN",
			value:   math.NaN(),
			present: true,
		},
		{
			name:    "positive infinity",
			value:   math.Inf(1),
			present: true,
		},
		{
			name:    "negative infinity",
			value:   math.Inf(-1),
			present: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			macroNet := macro.NewNet()
			mesoNet := meso.NewNet()
			for _, linkID := range []gmns.LinkID{20, 10} {
				macroLink := macro.NewLinkFrom(linkID, 1, 2)
				mesoLink := meso.NewLinkFrom(linkID, 1, 2, meso.WithLineMacroLinkID(10))
				if testCase.present {
					macro.WithLengthMeters(testCase.value)(macroLink)
					meso.WithLengthMeters(testCase.value)(mesoLink)
				}
				macroNet.Links[linkID] = macroLink
				mesoNet.Links[linkID] = mesoLink
			}
			generatedMeso, err := GenerateMesoscopic(macroNet, movement.NewMovementsStorage())
			assert.Error(t, err)
			assert.Nil(t, generatedMeso)
			if err != nil {
				assert.Contains(t, err.Error(), "macro link 10")
				assert.Contains(t, err.Error(), "length")
			}
			generatedMicro, err := GenerateMicroscopic(macroNet, mesoNet, movement.NewMovementsStorage())
			assert.Error(t, err)
			assert.Nil(t, generatedMicro)
			if err != nil {
				assert.Contains(t, err.Error(), "meso link 10")
				assert.Contains(t, err.Error(), "length")
			}
			macroLength, macroPresent := macroNet.Links[10].LengthMeters()
			mesoLength, mesoPresent := mesoNet.Links[10].LengthMeters()
			assert.Equal(t, testCase.present, macroPresent)
			assert.Equal(t, testCase.present, mesoPresent)
			if math.IsNaN(testCase.value) {
				assert.True(t, math.IsNaN(macroLength))
				assert.True(t, math.IsNaN(mesoLength))
			} else {
				assert.Equal(t, testCase.value, macroLength)
				assert.Equal(t, testCase.value, mesoLength)
			}
			assert.Empty(t, macroNet.Nodes)
			assert.Empty(t, mesoNet.Nodes)
		})
	}
}

func TestMicroUsesPreparedRoadLengths(t *testing.T) {
	macroNet := controlTestNetwork(t, types.CONTROL_TYPE_IS_SIGNAL)
	movements, err := GenerateMovements(macroNet)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	mesoNet, err := GenerateMesoscopic(macroNet, movements)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	for linkID, original := range macroNet.Links {
		lanesNum, hasLanes := original.LanesNum()
		assert.True(t, hasLanes)
		macroNet.Links[linkID] = macro.NewLinkFrom(linkID, original.SourceNode(), original.TargetNode(),
			macro.WithLineGeom(original.Geom()),
			macro.WithLineGeomEuclidean(original.GeomEuclidean()),
			macro.WithLanesNum(lanesNum),
			macro.WithLanesInfo(original.LanesInfo()),
			macro.WithAllowedAgentTypes(original.AllowedAgentTypes()),
		)
	}
	connectorCount := 0
	for linkID, original := range mesoNet.Links {
		if original.MacroNode() != 2 {
			continue
		}
		connectorCount++
		lanesNum, hasLanes := original.LanesNum()
		assert.True(t, hasLanes)
		mesoNet.Links[linkID] = meso.NewLinkFrom(linkID, original.SourceNode(), original.TargetNode(),
			meso.WithLineMacroNodeID(original.MacroNode()),
			meso.WithLineMacroLinkID(-1),
			meso.WithLanesNum(lanesNum),
			meso.WithMovementID(original.Movement()),
			meso.WithMovementMesoLinkIncome(original.MovementMesoLinkIncome()),
			meso.WithMovementMesoLinkOutcome(original.MovementMesoLinkOutcome()),
			meso.WithMovementIncomeLaneStartSeqID(original.MovementIncomeLaneStartSeqID()),
			meso.WithMovementOutcomeLaneStartSeqID(original.MovementOutcomeLaneStartSeqID()),
		)
	}
	assert.Equal(t, 1, connectorCount)
	microNet, err := GenerateMicroscopic(macroNet, mesoNet, movements)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.NotEmpty(t, microNet.Links)
	movementCells := 0
	for _, link := range microNet.Links {
		length, present := link.LengthMeters()
		assert.True(t, present)
		assert.Positive(t, length)
		assert.False(t, math.IsNaN(length) || math.IsInf(length, 0))
		if link.MacroNode() == 2 {
			movementCells++
		}
	}
	assert.Positive(t, movementCells, "connector cell lengths are calculated from their prepared endpoints")
	for _, link := range macroNet.Links {
		_, present := link.LengthMeters()
		assert.False(t, present, "micro generation must not backfill source macro lengths")
	}
	for _, link := range mesoNet.Links {
		if link.MacroNode() == 2 {
			_, present := link.LengthMeters()
			assert.False(t, present, "micro generation must not backfill source connector lengths")
		}
	}
}
