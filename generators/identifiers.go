package generators

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
)

// Empty networks use integer allocation; populated networks must have one standard ID kind.
func macroIdentifierType(net *macro.Net) (gmns.IDType, error) {
	kind := gmns.IDTypeInteger
	ids := sortedMacroNodeIDs(net.Nodes)
	if len(ids) > 0 {
		kind = ids[0].Type()
	}
	err := net.ValidateIDType(kind)
	return kind, err
}

func mesoIdentifierType(net *meso.Net) (gmns.IDType, error) {
	kind := gmns.IDTypeInteger
	ids := sortedMesoNodeIDs(net.Nodes)
	if len(ids) > 0 {
		kind = ids[0].Type()
	}
	err := net.ValidateIDType(kind)
	return kind, err
}
