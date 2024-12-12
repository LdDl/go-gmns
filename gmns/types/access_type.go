package types

type AccessType uint16

const (
	ACCESS_UNDEFINED = AccessType(iota)
	ACCESS_HIGHWAY
	ACCESS_MOTOR_VEHICLE
	ACCESS_MOTORCAR
	ACCESS_OSM_ACCESS
	ACCESS_SERVICE
	ACCESS_BICYCLE
	ACCESS_FOOT
)

func (iotaIdx AccessType) String() string {
	return [...]string{"undefined", "highway", "motor_vehicle", "motorcar", "access", "service", "bicycle", "foot"}[iotaIdx]
}
