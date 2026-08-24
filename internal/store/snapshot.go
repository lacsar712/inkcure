package store

import "github.com/lacsar712/inkcure/internal/model"

type InkvatSnapshotView struct {
	UnitID   string
	Inkvat     model.InkvatReading
	Alarms   []model.AlarmEvent
	Revision uint64
}

func CloneInkvatSnapshot(s model.PlantSnapshot) InkvatSnapshotView {
	out := InkvatSnapshotView{
		UnitID:   s.UnitID,
		Inkvat:     s.Inkvat,
		Revision: s.Revision,
	}
	out.Alarms = s.Alarms[:len(s.Alarms):len(s.Alarms)]
	return out
}
