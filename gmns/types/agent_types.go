package types

type AgentType uint16

const (
	AGENT_UNDEFINED = AgentType(iota)
	AGENT_AUTO
	AGENT_BIKE
	AGENT_WALK
)

func (iotaIdx AgentType) String() string {
	return [...]string{"undefined", "auto", "bike", "walk"}[iotaIdx]
}

var (
	agentTypesAll = map[AgentType]struct{}{
		AGENT_AUTO: {},
		AGENT_BIKE: {},
		AGENT_WALK: {},
	}

	AGENT_TYPES_DEFAULT = []AgentType{AGENT_AUTO}

	agentsAccessIncludeValues = map[AgentType]map[AccessType]map[string]struct{}{
		AGENT_AUTO: {
			ACCESS_MOTOR_VEHICLE: {
				"yes": struct{}{},
			},
			ACCESS_MOTORCAR: {
				"yes": struct{}{},
			},
		},
		AGENT_BIKE: {
			ACCESS_BICYCLE: {
				"yes": struct{}{},
			},
		},
		AGENT_WALK: {
			ACCESS_FOOT: {
				"yes": struct{}{},
			},
		},
	}

	agentsAccessExcludeValues = map[AgentType]map[AccessType]map[string]struct{}{
		AGENT_AUTO: {
			ACCESS_HIGHWAY: {
				"cycleway":      struct{}{},
				"footway":       struct{}{},
				"pedestrian":    struct{}{},
				"steps":         struct{}{},
				"track":         struct{}{},
				"corridor":      struct{}{},
				"elevator":      struct{}{},
				"escalator":     struct{}{},
				"service":       struct{}{},
				"living_street": struct{}{},
			},
			ACCESS_MOTOR_VEHICLE: {
				"no": struct{}{},
			},
			ACCESS_MOTORCAR: {
				"no": struct{}{},
			},
			ACCESS_OSM_ACCESS: {
				"private": struct{}{},
			},
			ACCESS_SERVICE: {
				"parking":          struct{}{},
				"parking_aisle":    struct{}{},
				"driveway":         struct{}{},
				"private":          struct{}{},
				"emergency_access": struct{}{},
			},
		},
		AGENT_BIKE: {
			ACCESS_HIGHWAY: {
				"footway":       struct{}{},
				"steps":         struct{}{},
				"corridor":      struct{}{},
				"elevator":      struct{}{},
				"escalator":     struct{}{},
				"motor":         struct{}{},
				"motorway":      struct{}{},
				"motorway_link": struct{}{},
			},
			ACCESS_BICYCLE: {
				"no": struct{}{},
			},
			ACCESS_SERVICE: {
				"private": struct{}{},
			},
			ACCESS_OSM_ACCESS: {
				"private": struct{}{},
			},
		},
		AGENT_WALK: {
			ACCESS_HIGHWAY: {
				"cycleway":      struct{}{},
				"motor":         struct{}{},
				"motorway":      struct{}{},
				"motorway_link": struct{}{},
			},
			ACCESS_FOOT: {
				"no": struct{}{},
			},
			ACCESS_SERVICE: {
				"private": struct{}{},
			},
			ACCESS_OSM_ACCESS: {
				"private": struct{}{},
			},
		},
	}
)

func agentsIntersects(left []AgentType, right []AgentType) bool {
	for _, l := range left {
		for _, r := range right {
			if l == r {
				return true
			}
		}
	}
	return false
}

func AgentsIntersection(left []AgentType, right []AgentType) map[AgentType]struct{} {
	intersection := make(map[AgentType]struct{})
	for _, l := range left {
		for _, r := range right {
			if l == r {
				intersection[l] = struct{}{}
			}
		}
	}
	return intersection
}

func NewAllowableAgentTypeFrom(motorVehicle, motorcar, bicycle, foot, highway, access, service string) (allowedAgents []AgentType) {
	for agentType := range agentTypesAll {
		included := findIncludedAgent(motorVehicle, motorcar, bicycle, foot, agentType)
		if included {
			allowedAgents = append(allowedAgents, agentType)
			continue
		}
		excluded := findExcludedAgent(motorVehicle, motorcar, bicycle, foot, highway, access, service, agentType)
		if excluded {
			allowedAgents = append(allowedAgents, agentType)
			continue
		}
	}
	return allowedAgents
}

func findIncludedAgent(motorVehicle, motorcar, bicycle, foot string, agentType AgentType) bool {
	accessType, ok := agentsAccessIncludeValues[agentType]
	if !ok {
		return false
	}
	switch agentType {
	case AGENT_AUTO:
		// Check `motor_vehicle`
		if _, ok := accessType[ACCESS_MOTOR_VEHICLE][motorVehicle]; ok {
			return true
		}
		// Check `motorcar`
		if _, ok := accessType[ACCESS_MOTORCAR][motorcar]; ok {
			return true
		}
	case AGENT_BIKE:
		// Check `bicycle`
		if _, ok := accessType[ACCESS_BICYCLE][bicycle]; ok {
			return true
		}
	case AGENT_WALK:
		// Check `foot`
		if _, ok := accessType[ACCESS_FOOT][foot]; ok {
			return true
		}
	default:
		return false
	}
	return false
}

func findExcludedAgent(motorVehicle, motorcar, bicycle, foot, highway, access, service string, agentType AgentType) bool {
	accessType, ok := agentsAccessExcludeValues[agentType]
	if !ok {
		return true
	}
	switch agentType {
	case AGENT_AUTO:
		// Check `highway`
		if _, ok := accessType[ACCESS_HIGHWAY][highway]; ok {
			return false
		}
		// Check `motor_vehicle`
		if _, ok := accessType[ACCESS_MOTOR_VEHICLE][motorVehicle]; ok {
			return false
		}
		// Check `motorcar`
		if _, ok := accessType[ACCESS_MOTORCAR][motorcar]; ok {
			return false
		}
		// Check `access`
		if _, ok := accessType[ACCESS_OSM_ACCESS][access]; ok {
			return false
		}
		// Check `service`
		if _, ok := accessType[ACCESS_SERVICE][service]; ok {
			return false
		}
	case AGENT_BIKE:
		// Check `highway`
		if _, ok := accessType[ACCESS_HIGHWAY][highway]; ok {
			return false
		}
		// Check `bicycle`
		if _, ok := accessType[ACCESS_BICYCLE][bicycle]; ok {
			return false
		}
		// Check `service`
		if _, ok := accessType[ACCESS_SERVICE][service]; ok {
			return false
		}
		// Check `access`
		if _, ok := accessType[ACCESS_OSM_ACCESS][access]; ok {
			return false
		}
	case AGENT_WALK:
		// Check `highway`
		if _, ok := accessType[ACCESS_HIGHWAY][highway]; ok {
			return false
		}
		// Check `foot`
		if _, ok := accessType[ACCESS_FOOT][foot]; ok {
			return false
		}
		// Check `service`
		if _, ok := accessType[ACCESS_SERVICE][service]; ok {
			return false
		}
		// Check `access`
		if _, ok := accessType[ACCESS_OSM_ACCESS][access]; ok {
			return false
		}
	default:
		return true
	}

	return true
}
