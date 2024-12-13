package movement

import (
	"fmt"
	"testing"

	"github.com/LdDl/go-gmns/utils/geomath"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

func TestFindMovementType(t *testing.T) {
	// should return NBL (movement ID - 31, ib - 21, ob - 9)
	givenInboundLine := geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.12687689999999918, 52.97880310000000037}, {36.12666420000000045, 52.97901970000000205}, {36.12655029999999812, 52.97914080000000325}, {36.12633389999999878, 52.97951309999999836}, {36.12644790000000228, 52.97972719999999924}, {36.12617269999999792, 52.98006380000000348}, {36.12591199999999958, 52.98028130000000147}, {36.12566470000000152, 52.98060569999999814}, {36.1256758999999974, 52.98077049999999844}, {36.12550180000000211, 52.98099280000000277}, {36.12535739999999862, 52.98109339999999889}, {36.12520870000000173, 52.98136800000000335}, {36.12476730000000202, 52.9816525000000027}, {36.12474199999999769, 52.98177170000000302}, {36.12476999999999805, 52.98198190000000096}, {36.12384109999999993, 52.98328839999999929}, {36.12374100000000254, 52.98351069999999652}, {36.12317699999999832, 52.98426740000000024}, {36.12142579999999725, 52.98677949999999726}}))
	givenOutboundLine := geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.12142579999999725, 52.98677949999999726}, {36.12084130000000215, 52.98640280000000047}, {36.12050210000000305, 52.98620780000000252}}))
	expectedMovementTextID, expectedMovementType := MOVEMENT_NBL, MOVEMENT_TYPE_LEFT
	ansMovementTextID, ansMovementType := FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return NBT (movement ID - 27, ib - 8, ob - 10)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.12050210000000305, 52.98620780000000252}, {36.12084130000000215, 52.98640280000000047}, {36.12142579999999725, 52.98677949999999726}}))
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.12142579999999725, 52.98677949999999726}, {36.12219069999999732, 52.9873075}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_NBT, MOVEMENT_TYPE_THRU
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return NBR (movement ID - 28, ib - 8, ob - 20)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.1205021, 52.9862078}, {36.1208413, 52.9864028}, {36.1214258, 52.9867795}}))
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.1214258, 52.9867795}, {36.1231770, 52.9842674}, {36.1237410, 52.9835107}, {36.1238411, 52.9832884}, {36.1247700, 52.9819819}, {36.1247420, 52.9817717}, {36.1247673, 52.9816525}, {36.1252087, 52.9813680}, {36.1253574, 52.9810934}, {36.1255018, 52.9809928}, {36.1256759, 52.9807705}, {36.1256647, 52.9806057}, {36.1259120, 52.9802813}, {36.1261727, 52.9800638}, {36.1264479, 52.9797272}, {36.1263339, 52.9795131}, {36.1265503, 52.9791408}, {36.1266642, 52.9790197}, {36.1268769, 52.9788031}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_NBR, MOVEMENT_TYPE_RIGHT
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return NBU (faked geometry)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{37.35504115149979, 55.84981923732954}, {37.354853767911266, 55.85002159103192}}))
	// Just reversed line
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{37.354853767911266, 55.85002159103192}, {37.35504115149979, 55.84981923732954}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_NBU, MOVEMENT_TYPE_U_TURN
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return SBL (movement ID - 30, ib - 11, ob - 20)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.12219069999999732, 52.9873075}, {36.12142579999999725, 52.98677949999999726}}))
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.12142579999999725, 52.98677949999999726}, {36.12317699999999832, 52.98426740000000024}, {36.12374100000000254, 52.98351069999999652}, {36.12384109999999993, 52.98328839999999929}, {36.12476999999999805, 52.98198190000000096}, {36.12474199999999769, 52.98177170000000302}, {36.12476730000000202, 52.9816525000000027}, {36.12520870000000173, 52.98136800000000335}, {36.12535739999999862, 52.98109339999999889}, {36.12550180000000211, 52.98099280000000277}, {36.1256758999999974, 52.98077049999999844}, {36.12566470000000152, 52.98060569999999814}, {36.12591199999999958, 52.98028130000000147}, {36.12617269999999792, 52.98006380000000348}, {36.12644790000000228, 52.97972719999999924}, {36.12633389999999878, 52.97951309999999836}, {36.12655029999999812, 52.97914080000000325}, {36.12666420000000045, 52.97901970000000205}, {36.12687689999999918, 52.97880310000000037}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_SBL, MOVEMENT_TYPE_LEFT
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return SBT (movement ID - 29, ib - 11, ob - 9)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.12219069999999732, 52.9873075}, {36.12142579999999725, 52.98677949999999726}}))
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.12142579999999725, 52.98677949999999726}, {36.12084130000000215, 52.98640280000000047}, {36.12050210000000305, 52.98620780000000252}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_SBT, MOVEMENT_TYPE_THRU
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return SBR (movement ID - 8, ib - 39, ob - 22)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.1251845999999972, 52.99150929999999704}, {36.12489899999999921, 52.99131729999999862}, {36.1228802999999985, 52.98996069999999747}, {36.12229990000000157, 52.98956330000000037}, {36.12044279999999929, 52.98825049999999948}}))
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.12044279999999929, 52.98825049999999948}, {36.12031240000000309, 52.98833659999999668}, {36.11993350000000191, 52.98862239999999701}, {36.11973820000000046, 52.98879199999999656}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_SBR, MOVEMENT_TYPE_RIGHT
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return SBU (faked geometry)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{37.354679202308745, 55.850240985422715}, {37.35474241605007, 55.850147624560464}, {37.3548259484937, 55.85005130633169}}))
	// Just reversed line
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{37.3548259484937, 55.85005130633169}, {37.35474241605007, 55.850147624560464}, {37.354679202308745, 55.850240985422715}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_SBU, MOVEMENT_TYPE_U_TURN
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return EBL (movement ID - 19, ib - 40, ob - 31)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.11166690000000301, 52.98525029999999703}, {36.11274670000000242, 52.98543680000000222}, {36.11434239999999818, 52.98572250000000139}, {36.11547629999999742, 52.98592670000000027}, {36.11673700000000053, 52.98622499999999746}, {36.11700419999999667, 52.98630649999999775}}))
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.11700419999999667, 52.98630649999999775}, {36.11690120000000093, 52.98640089999999958}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_EBL, MOVEMENT_TYPE_LEFT
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return EBT (movement ID - 21, ib - 6, ob - 8)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.11918990000000207, 52.98542189999999863}, {36.12050210000000305, 52.98620780000000252}}))
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.12050210000000305, 52.98620780000000252}, {36.12084130000000215, 52.98640280000000047}, {36.12142579999999725, 52.98677949999999726}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_EBT, MOVEMENT_TYPE_THRU
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return EBR (movement ID - 22, ib - 6, ob - 27)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.11918990000000207, 52.98542189999999863}, {36.12050210000000305, 52.98620780000000252}}))
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.12050210000000305, 52.98620780000000252}, {36.1220641000000029, 52.9839847000000006}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_EBR, MOVEMENT_TYPE_RIGHT
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return EBU (faked geometry)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{37.354533231245625, 55.84995966864818}, {37.35476049969526, 55.84998712788135}, {37.354875639008526, 55.84999642177132}}))
	// Just reversed line
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{37.354875639008526, 55.84999642177132}, {37.35476049969526, 55.84998712788135}, {37.354533231245625, 55.84995966864818}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_EBU, MOVEMENT_TYPE_U_TURN
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return WBL (movement ID - 15, ib - 4, ob - 51)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.11918990000000207, 52.98542189999999863}, {36.11909330000000296, 52.98547599999999846}, {36.11896899999999988, 52.98552140000000321}, {36.11733540000000175, 52.98611069999999756}, {36.1172475999999989, 52.98614649999999671}, {36.11716289999999674, 52.98619159999999795}, {36.11710870000000284, 52.98622540000000214}}))
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.11710870000000284, 52.98622540000000214}, {36.11675999999999931, 52.98601169999999883}, {36.11669100000000299, 52.98596560000000011}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_WBL, MOVEMENT_TYPE_LEFT
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return WBT (movement ID - 14, ib - 4, ob - 5)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.11918990000000207, 52.98542189999999863}, {36.11909330000000296, 52.98547599999999846}, {36.11896899999999988, 52.98552140000000321}, {36.11733540000000175, 52.98611069999999756}, {36.1172475999999989, 52.98614649999999671}, {36.11716289999999674, 52.98619159999999795}, {36.11710870000000284, 52.98622540000000214}}))
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.11710870000000284, 52.98622540000000214}, {36.11705160000000348, 52.9862694999999988}, {36.11700419999999667, 52.98630649999999775}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_WBT, MOVEMENT_TYPE_THRU
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return WBR (movement ID - 13, ib - 7, ob - 4)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.12050210000000305, 52.98620780000000252}, {36.11918990000000207, 52.98542189999999863}}))
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{36.11918990000000207, 52.98542189999999863}, {36.11909330000000296, 52.98547599999999846}, {36.11896899999999988, 52.98552140000000321}, {36.11733540000000175, 52.98611069999999756}, {36.1172475999999989, 52.98614649999999671}, {36.11716289999999674, 52.98619159999999795}, {36.11710870000000284, 52.98622540000000214}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_WBR, MOVEMENT_TYPE_RIGHT
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")

	// should return WBU (faked geometry)
	givenInboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{37.35438801371322, 55.84968963713089}, {37.35426743052332, 55.84968786742411}, {37.354183889229205, 55.84967503704627}}))
	// Just reversed line
	givenOutboundLine = geomath.LineToEuclidean(orb.LineString([]orb.Point{{37.354183889229205, 55.84967503704627}, {37.35426743052332, 55.84968786742411}, {37.35438801371322, 55.84968963713089}}))
	expectedMovementTextID, expectedMovementType = MOVEMENT_WBU, MOVEMENT_TYPE_U_TURN
	ansMovementTextID, ansMovementType = FindMovementType(givenInboundLine, givenOutboundLine)
	assert.Equal(t, expectedMovementTextID, ansMovementTextID, "Wrong movement text ID")
	assert.Equal(t, expectedMovementType, ansMovementType, "Wrong movement type")
}

func TestFindMovementGeom(t *testing.T) {
	precision := 10e-8

	// Check movement with ID 31 (ib - 21, ob - 9)
	givenInboundLine := orb.LineString([]orb.Point{{36.12687689999999918, 52.97880310000000037}, {36.12666420000000045, 52.97901970000000205}, {36.12655029999999812, 52.97914080000000325}, {36.12633389999999878, 52.97951309999999836}, {36.12644790000000228, 52.97972719999999924}, {36.12617269999999792, 52.98006380000000348}, {36.12591199999999958, 52.98028130000000147}, {36.12566470000000152, 52.98060569999999814}, {36.1256758999999974, 52.98077049999999844}, {36.12550180000000211, 52.98099280000000277}, {36.12535739999999862, 52.98109339999999889}, {36.12520870000000173, 52.98136800000000335}, {36.12476730000000202, 52.9816525000000027}, {36.12474199999999769, 52.98177170000000302}, {36.12476999999999805, 52.98198190000000096}, {36.12384109999999993, 52.98328839999999929}, {36.12374100000000254, 52.98351069999999652}, {36.12317699999999832, 52.98426740000000024}, {36.12142579999999725, 52.98677949999999726}})
	givenOutboundLine := orb.LineString([]orb.Point{{36.12142579999999725, 52.98677949999999726}, {36.12084130000000215, 52.98640280000000047}, {36.12050210000000305, 52.98620780000000252}})
	expectedMovementGeom := orb.LineString([]orb.Point{{36.12147199999999714, 52.98671319999999696}, {36.12134429999999696, 52.98672700000000191}})
	ansMovementGeom := FindMovementGeom(givenInboundLine, givenOutboundLine)
	assert.Equal(t, len(expectedMovementGeom), len(ansMovementGeom), "Incorrect number of points in movement geometry")
	for i := range ansMovementGeom {
		pt := ansMovementGeom[i]
		assert.InDelta(t, expectedMovementGeom[i][0], pt[0], precision, fmt.Sprintf("Wrong X (longitude) in EPSG:3857 at pos #%d", i))
		assert.InDelta(t, expectedMovementGeom[i][1], pt[1], precision, fmt.Sprintf("Wrong Y (latitude) in EPSG:3857 at pos #%d", i))
	}

	// Check movement with ID 17 (ib - 5, ob - 31)
	givenInboundLine = orb.LineString([]orb.Point{{36.11710870000000284, 52.98622540000000214}, {36.11705160000000348, 52.9862694999999988}, {36.11700419999999667, 52.98630649999999775}})
	givenOutboundLine = orb.LineString([]orb.Point{{36.11700419999999667, 52.98630649999999775}, {36.11690120000000093, 52.98640089999999958}})
	expectedMovementGeom = orb.LineString([]orb.Point{{36.11707729999999827, 52.98624970000000189}, {36.11693869999999862, 52.98636659999999665}})
	ansMovementGeom = FindMovementGeom(givenInboundLine, givenOutboundLine)
	assert.Equal(t, len(expectedMovementGeom), len(ansMovementGeom), "Incorrect number of points in movement geometry")
	for i := range ansMovementGeom {
		pt := ansMovementGeom[i]
		assert.InDelta(t, expectedMovementGeom[i][0], pt[0], precision, fmt.Sprintf("Wrong X (longitude) in EPSG:3857 at pos #%d", i))
		assert.InDelta(t, expectedMovementGeom[i][1], pt[1], precision, fmt.Sprintf("Wrong Y (latitude) in EPSG:3857 at pos #%d", i))
	}
}
