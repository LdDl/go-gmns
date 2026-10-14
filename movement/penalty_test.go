package movement

import (
	"encoding/json"
	"testing"

	"github.com/paulmach/orb/geojson"
)

type penaltyCase struct {
	name    string
	options []func(*Movement)
	want    float64
	present bool
}

func TestPenaltyGeoJSON(t *testing.T) {
	cases := []penaltyCase{
		{
			name: "absent",
		},
		{
			name: "zero",
			options: []func(*Movement){
				WithPenalty(0),
			},
			present: true,
		},
		{
			name: "fractional",
			options: []func(*Movement){
				WithPenalty(12.5),
			},
			want:    12.5,
			present: true,
		},
		{
			name: "negative is supplied",
			options: []func(*Movement){
				WithPenalty(-1),
			},
			want:    -1,
			present: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			mvmt := NewMovement(1, 2, 3, 4, MOVEMENT_SBT, MOVEMENT_TYPE_THRU, testCase.options...)
			data, err := json.Marshal(mvmt.GeoFeature())
			if err != nil {
				t.Fatal(err)
			}
			feature, err := geojson.UnmarshalFeature(data)
			if err != nil {
				t.Fatal(err)
			}
			value, exists := feature.Properties["penalty"]
			if !exists {
				t.Fatal("penalty property is missing")
			}
			if testCase.present {
				if value != testCase.want {
					t.Fatalf("exported penalty = %v, want %v", value, testCase.want)
				}
			} else if value != nil {
				t.Fatalf("absent penalty = %v, want null", value)
			}
			penalty, present := mvmt.Penalty()
			if penalty != testCase.want || present != testCase.present {
				t.Fatalf("penalty after export = (%v, %t), want (%v, %t)", penalty, present, testCase.want, testCase.present)
			}
		})
	}
}
