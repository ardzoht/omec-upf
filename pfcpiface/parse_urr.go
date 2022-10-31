package pfcpiface

import (
	"fmt"

	"github.com/wmnsk/go-pfcp/ie"
)

type Urr struct {
	UrrID          uint32
	CtrID          uint32
	PdrID          uint32
	FseidIP        uint32
	MeasureMethod  uint8
	ReportOpen     bool
	Trigger        ReportTrigger
	LocalThreshold uint64
	VolThreshold   VolumeData
	VolQuota       VolumeData
	// local session ID
	FseID uint64
}

type ReportTrigger struct {
	Flags uint16
}

type VolumeData struct {
	Flags       uint8
	TotalVol    uint64
	UplinkVol   uint64
	DownlinkVol uint64
}

func (r *ReportTrigger) isVOLTHSet() bool {
	u8 := uint8(r.Flags >> 8)
	return has2ndBit(u8)
}
func (r *ReportTrigger) isVOLQUSet() bool {
	u8 := uint8(r.Flags)
	return has1stBit(u8)
}

func (u *Urr) String() string {
	return fmt.Sprintf("URR(id=%v, ctrID=%v, pdrID=%v, F-SEID IPv4=%v, measureMethod=%v, "+
		"reportOpen=%v, trigger=%v, localThreshold=%v, volThreshold=%v, volQuota=%v)",
		u.UrrID, u.CtrID, u.PdrID, u.FseidIP, u.MeasureMethod, u.ReportOpen, u.Trigger,
		u.LocalThreshold, u.VolThreshold, u.VolQuota)
}

func (u *Urr) parseURR(ie1 *ie.IE, seid uint64) error {
	log.Info("Parse Create URR")
	volumeThresh := VolumeData{}
	volumeQuota := VolumeData{}

	urrID, err := ie1.URRID()
	if err != nil {
		log.Error("Could not read urrID!")
		return err
	}

	measureMethod, err := ie1.MeasurementMethod()
	if err != nil {
		log.Error("Could not read Measurement method!")
		return err
	}

	trigger, err := ie1.ReportingTriggers()
	if err != nil {
		log.Error("Could not read Reporting triggers!")
		return err
	}

	reportTrigger := ReportTrigger{Flags: trigger}
	volThreshField, err := ie1.VolumeThreshold()
	if err == nil {
		volumeThresh.Flags = volThreshField.Flags
		volumeThresh.TotalVol = volThreshField.TotalVolume
		volumeThresh.UplinkVol = volThreshField.UplinkVolume
		volumeThresh.DownlinkVol = volThreshField.DownlinkVolume
	}

	volQuotaField, err := ie1.VolumeQuota()
	if err == nil {
		volumeQuota.Flags = volQuotaField.Flags
		volumeQuota.TotalVol = volQuotaField.TotalVolume
		volumeQuota.UplinkVol = volQuotaField.UplinkVolume
		volumeQuota.DownlinkVol = volQuotaField.DownlinkVolume
	}

	u.UrrID = uint32(urrID)
	u.MeasureMethod = measureMethod
	u.Trigger = reportTrigger
	u.ReportOpen = true
	u.VolThreshold = volumeThresh
	u.LocalThreshold = volumeThresh.TotalVol
	u.VolQuota = volumeQuota
	u.FseID = seid

	return nil
}
