package generators

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

func preparationTestNetwork(t *testing.T, kind gmns.IDType) *macro.Net {
	t.Helper()
	net := macro.NewNet()
	err := net.SetConfig(gmns.NewConfig(kind, gmns.WithDatasetName("baseline")))
	assert.NoError(t, err)
	points := []orb.Point{{37, 55}, {37.002, 55}, {37.004, 55}}
	nodeIDs := make([]gmns.NodeID, 0, len(points))
	for index, text := range []string{"001", "2", "900719925474099312345"} {
		id, parseErr := gmns.ParseNodeID(text, kind)
		assert.NoError(t, parseErr)
		nodeIDs = append(nodeIDs, id)
		net.Nodes[id] = macro.NewNodeFrom(id, macro.WithPointGeom(points[index]), macro.WithAltitudeMeters(0))
	}
	linkIDs := make([]gmns.LinkID, 0, 2)
	for index, text := range []string{"13", "900719925474099312345"} {
		id, parseErr := gmns.ParseLinkID(text, kind)
		assert.NoError(t, parseErr)
		linkIDs = append(linkIDs, id)
		net.Links[id] = macro.NewLinkFrom(id, nodeIDs[index], nodeIDs[index+1],
			macro.WithLineGeom(orb.LineString{points[index], points[index+1]}),
			macro.WithLanesNum(3),
			macro.WithLengthMeters(128),
			macro.WithCapacity(0),
		)
	}
	movementID, err := gmns.ParseMovementID("49", kind)
	assert.NoError(t, err)
	record := movement.NewMovement(movementID, nodeIDs[1], linkIDs[0], linkIDs[1], "", "thru",
		movement.WithIncomeLane(3, 3),
		movement.WithOutcomeLane(2, 2),
		movement.WithPenalty(0),
		movement.WithControlType(types.CONTROL_TYPE_STOP),
	)
	err = net.AddMovement(record)
	assert.NoError(t, err)
	return net
}

func TestPrepareMesoscopicFromStandardRecords(t *testing.T) {
	for _, kind := range []gmns.IDType{gmns.IDTypeInteger, gmns.IDTypeString} {
		t.Run(string(kind), func(t *testing.T) {
			source := preparationTestNetwork(t, kind)
			expectedSource := preparationTestNetwork(t, kind)
			first, err := PrepareMesoscopic(source, WithUniformLaneLayout())
			assert.NoError(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, expectedSource, source)
			assert.NoError(t, first.Meso().ValidateDataset())
			assert.Len(t, first.Meso().Nodes, 4)
			assert.Len(t, first.Meso().Links, 3)
			assert.Empty(t, first.Meso().MovementRecords())
			connectors := 0
			for _, link := range first.Meso().Links {
				assert.Equal(t, kind, link.ID.Type())
				capacity, present := link.Capacity()
				assert.True(t, present)
				assert.Zero(t, capacity)
				_, present = link.FreeSpeed()
				assert.False(t, present)
				if !link.Movement().IsZero() {
					connectors++
					assert.Equal(t, "49", link.Movement().String())
					assert.Equal(t, 2, link.MovementIncomeLaneStartSeqID())
					assert.Equal(t, 1, link.MovementOutcomeLaneStartSeqID())
					lanes, supplied := link.LanesNum()
					assert.True(t, supplied)
					assert.Equal(t, 1, lanes)
				}
			}
			assert.Equal(t, 1, connectors)
			second, err := PrepareMesoscopic(source, WithUniformLaneLayout())
			assert.NoError(t, err)
			assert.Equal(t, first, second)
			// Stale caches and OSM-only attributes must not affect standard-record preparation.
			for _, node := range source.Nodes {
				macro.WithIncomingLinks(gmns.LinkIDFromInt64(-99))(node)
				macro.WithPointGeomEuclidean(orb.Point{1, 2})(node)
			}
			for _, link := range source.Links {
				macro.WithLineGeomEuclidean(orb.LineString{{1, 2}})(link)
				macro.WithLanesInfo(macro.LanesInfo{LanesList: []int{99}})(link)
				macro.WithAllowedAgentTypes([]types.AgentType{types.AGENT_BIKE})(link)
			}
			third, err := PrepareMesoscopic(source, WithUniformLaneLayout())
			assert.NoError(t, err)
			assert.Equal(t, first, third)
			// Mutating generated geometry must not reach the source shape.
			for _, link := range first.Meso().Links {
				link.Geom()[0][0] = 0
			}
			for _, link := range source.Links {
				assert.Equal(t, expectedSource.Links[link.ID].Geom(), link.Geom())
			}
		})
	}
}

func TestPrepareMesoscopicSharedGeometry(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	expected, err := PrepareMesoscopic(source, WithUniformLaneLayout())
	assert.NoError(t, err)
	for _, link := range source.Links {
		shape := link.Geom().Clone()
		shape.Reverse()
		geometryID, parseErr := gmns.ParseGeometryID(link.ID.String(), gmns.IDTypeInteger)
		assert.NoError(t, parseErr)
		err = source.AddGeometry(gmns.NewGeometry(geometryID, shape))
		assert.NoError(t, err)
		macro.WithGeometryID(geometryID)(link)
		macro.WithDirFlag(gmns.DIR_FLAG_REVERSE)(link)
		macro.WithLineGeom(nil)(link)
	}
	actual, err := PrepareMesoscopic(source, WithUniformLaneLayout())
	assert.NoError(t, err)
	assert.Equal(t, expected.Meso(), actual.Meso())
	for _, link := range source.Links {
		assert.Nil(t, link.Geom())
	}
}

type preparationFailureCase struct {
	name    string
	option  func(*macro.Link)
	message string
}

func TestPrepareMesoscopicFailurePreservesInput(t *testing.T) {
	cases := []preparationFailureCase{
		{
			name:    "missing lanes",
			option:  macro.WithoutLanesNum(),
			message: "link 13 lanes",
		},
		{
			name:    "zero lanes",
			option:  macro.WithLanesNum(0),
			message: "got 0",
		},
		{
			name:    "negative lanes",
			option:  macro.WithLanesNum(-1),
			message: "link 13",
		},
		{
			name:    "too many lanes",
			option:  macro.WithLanesNum(100),
			message: "got 100",
		},
		{
			name:    "missing length",
			option:  macro.WithoutLengthMeters(),
			message: "link 13 length",
		},
		{
			name:    "lane drop unresolved",
			option:  macro.WithLanesNum(2),
			message: "movement 49: inbound lane range 3..3 is outside link 13 lanes 1..2",
		},
		{
			name:    "undirected",
			option:  macro.WithDirected(false),
			message: "link 13",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			source := preparationTestNetwork(t, gmns.IDTypeInteger)
			expected := preparationTestNetwork(t, gmns.IDTypeInteger)
			testCase.option(source.Links[gmns.LinkIDFromInt64(13)])
			testCase.option(expected.Links[gmns.LinkIDFromInt64(13)])
			result, err := PrepareMesoscopic(source, WithUniformLaneLayout())
			assert.Nil(t, result)
			assertPreparationError(t, err, testCase.message)
			assert.Equal(t, expected, source)
		})
	}
}

func TestPrepareMesoscopicRequiresExplicitScopes(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	record := source.MovementRecords()[0]
	movement.WithIncomeLane(1, 3)(record)
	err := source.UpdateMovement(record)
	assert.NoError(t, err)
	_, err = PrepareMesoscopic(source, WithUniformLaneLayout())
	assertPreparationError(t, err, "movement 49: baseline preparation requires exact single-lane")
	movement.WithoutIncomeLaneStart()(record)
	err = source.UpdateMovement(record)
	assert.NoError(t, err)
	_, err = PrepareMesoscopic(source, WithUniformLaneLayout())
	assertPreparationError(t, err, "movement 49: supply")
}

func TestPrepareMesoscopicEmptyAndExplicitOptions(t *testing.T) {
	source := macro.NewNet()
	err := source.SetConfig(gmns.NewConfig(gmns.IDTypeString))
	assert.NoError(t, err)
	_, err = PrepareMesoscopic(source)
	assertPreparationError(t, err, "explicit lane layout")
	_, err = PrepareMesoscopic(nil, WithUniformLaneLayout())
	assertPreparationError(t, err, "network is nil")
	prepared, err := PrepareMesoscopic(source, WithUniformLaneLayout())
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Empty(t, prepared.Meso().Nodes)
	assert.Empty(t, prepared.Meso().Links)
	assert.Equal(t, gmns.IDTypeString, prepared.Meso().Config().IDType())
}

func assertPreparationError(t *testing.T, err error, message string) {
	t.Helper()
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), message)
	}
}

func TestPrepareMesoscopicTwoWayAndLegacySettings(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	forward := source.Links[gmns.LinkIDFromInt64(13)]
	reverseID := gmns.LinkIDFromInt64(-1)
	shape := forward.Geom().Clone()
	shape.Reverse()
	source.Links[reverseID] = macro.NewLinkFrom(reverseID, forward.TargetNode(), forward.SourceNode(),
		macro.WithLineGeom(shape), macro.WithLengthMeters(128), macro.WithLanesNum(3),
	)
	first, err := PrepareMesoscopic(source, WithUniformLaneLayout())
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Len(t, first.Meso().Links, 4)
	assert.Len(t, first.Meso().Nodes, 6)
	for _, link := range first.Meso().Links {
		if link.MacroLink() == forward.ID || link.MacroLink() == reverseID {
			assert.NotEqual(t, 55.0, link.Geom()[0][1])
		}
	}
	previous := CUT_LENGTHS
	defer func() {
		CUT_LENGTHS = previous
	}()
	CUT_LENGTHS[3] = 100
	second, err := PrepareMesoscopic(source, WithUniformLaneLayout())
	assert.NoError(t, err)
	assert.Equal(t, first, second)
}
