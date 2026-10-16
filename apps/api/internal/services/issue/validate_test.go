package issue_service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
)

func TestEstimatesFollowTheTeamScale(t *testing.T) {
	cases := []struct {
		name     string
		team     ent.Team
		accepted []float64
		refused  []float64
	}{
		{"none", ent.Team{EstimateType: schema.EstimateTypeNone}, nil, []float64{0, 1, 2}},
		{"exponential", ent.Team{EstimateType: schema.EstimateTypeExponential}, []float64{1, 2, 4, 8, 16}, []float64{0, 3, 32}},
		{"exponential extended with zero", ent.Team{EstimateType: schema.EstimateTypeExponential, EstimateExtended: true, EstimateAllowZero: true}, []float64{0, 1, 16, 32, 64}, []float64{3, 128}},
		{"fibonacci", ent.Team{EstimateType: schema.EstimateTypeFibonacci}, []float64{1, 2, 3, 5, 8}, []float64{0, 4, 13}},
		{"fibonacci extended", ent.Team{EstimateType: schema.EstimateTypeFibonacci, EstimateExtended: true}, []float64{13, 21}, []float64{34}},
		{"linear", ent.Team{EstimateType: schema.EstimateTypeLinear}, []float64{1, 2, 3, 4, 5}, []float64{0, 6, 1.5}},
		{"linear extended", ent.Team{EstimateType: schema.EstimateTypeLinear, EstimateExtended: true}, []float64{6, 7}, []float64{8}},
		{"tshirt", ent.Team{EstimateType: schema.EstimateTypeTShirt}, []float64{1, 2, 3, 5, 8}, []float64{4, 13}},
		{"tshirt extended with zero", ent.Team{EstimateType: schema.EstimateTypeTShirt, EstimateExtended: true, EstimateAllowZero: true}, []float64{0, 13, 21}, []float64{34}},
	}
	for _, tc := range cases {
		tc.team.Key = "ENG"
		for _, value := range tc.accepted {
			assert.NoError(t, validateEstimate(&tc.team, value), "%s accepts %v", tc.name, value)
		}
		for _, value := range tc.refused {
			assert.ErrorIs(t, validateEstimate(&tc.team, value), errdefs.ErrInvalidInput, "%s refuses %v", tc.name, value)
		}
	}
}
