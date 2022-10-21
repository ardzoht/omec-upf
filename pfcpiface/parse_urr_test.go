// SPDX-License-Identifier: Apache-2.0
// Copyright 2022-present Open Networking Foundation

package pfcpiface

import (
	"testing"

	pfcpsimLib "github.com/omec-project/pfcpsim/pkg/pfcpsim/session"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wmnsk/go-pfcp/ie"
)

type urrTestCase struct {
	input       *ie.IE
	expected    *Urr
	description string
}

func TestParseURR(t *testing.T) {
	FSEID := uint64(100)

	for _, scenario := range []urrTestCase{
		{
			input: pfcpsimLib.NewURRBuilder().
				WithID(999).
				WithMethod(pfcpsimLib.IEMethod(create)).
				WithMeasurementMethodVolume(1).
				WithVolThresholdFlags(7).
				WithVolThresholdTotalVol(1000).
				WithVolThresholdUplinkVol(200).
				WithVolThresholdDownlinkVol(800).
				WithVolQuotaFlags(3).
				WithVolQuotaTotalVol(700).
				WithVolQuotaUplinkVol(300).
				WithVolQuotaDownlinkVol(400).
				WithTriggers(2).
				Build(),
			expected: &Urr{
				UrrID: 999,
				MeasureMethod: 2,
				ReportOpen: true,
				Trigger: ReportTrigger{
					Flags: 2,
				},
				LocalThreshold: 1000,
				VolThreshold: VolumeData{
					Flags: 7,
					TotalVol: 1000,
					UplinkVol: 200,
					DownlinkVol: 800,
				},
				VolQuota: VolumeData{
					// only TotalVol and UplinkVol are set
					Flags: 3,
					TotalVol: 700,
					UplinkVol: 300,
					DownlinkVol: 0,
				},
				FseID: FSEID,
			},
			description: "Valid Create URR input",
		},
		{
			input: pfcpsimLib.NewURRBuilder().
				WithID(999).
				WithMethod(pfcpsimLib.IEMethod(update)).
				WithMeasurementMethodVolume(1).
				WithVolThresholdFlags(7).
				WithVolThresholdTotalVol(1000).
				WithVolThresholdUplinkVol(200).
				WithVolThresholdDownlinkVol(800).
				WithVolQuotaFlags(3).
				WithVolQuotaTotalVol(700).
				WithVolQuotaUplinkVol(300).
				WithVolQuotaDownlinkVol(400).
				WithTriggers(2).
				Build(),
			expected: &Urr{
				UrrID: 999,
				MeasureMethod: 2,
				ReportOpen: true,
				Trigger: ReportTrigger{
					Flags: 2,
				},
				LocalThreshold: 1000,
				VolThreshold: VolumeData{
					Flags: 7,
					TotalVol: 1000,
					UplinkVol: 200,
					DownlinkVol: 800,
				},
				VolQuota: VolumeData{
					// only TotalVol and UplinkVol are set
					Flags: 3,
					TotalVol: 700,
					UplinkVol: 300,
					DownlinkVol: 0,
				},
				FseID: FSEID,
			},
			description: "Valid Update URR input",
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			mockURR := &Urr{}

			err := mockURR.parseURR(scenario.input, FSEID)
			require.NoError(t, err)

			assert.Equal(t, scenario.expected, mockURR)
		})
	}
}

func TestParseURRShouldError(t *testing.T) {
	FSEID := uint64(100)

	for _, scenario := range []urrTestCase{
		{
			input: ie.NewCreateURR(
				ie.NewReportingTriggers(2),
			),
			expected:    &Urr{},
			description: "Invalid URR input: no URR ID provided",
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			mockURR := &Urr{}

			err := mockURR.parseURR(scenario.input, FSEID)
			require.Error(t, err)

			assert.Equal(t, scenario.expected, mockURR)
		})
	}
}
