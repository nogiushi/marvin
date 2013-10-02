package marvin

import (
	"github.com/eikeon/marvin/activity"
	"github.com/eikeon/marvin/ambientlight"
	"github.com/eikeon/marvin/hue"
	"github.com/eikeon/marvin/motion"
	"github.com/eikeon/marvin/nog"
	"github.com/eikeon/marvin/persist"
	"github.com/eikeon/marvin/presence"
	"github.com/eikeon/marvin/schedule"
	"github.com/eikeon/marvin/switches"
)

type Marvin struct {
	*nog.Nog
	p            *persist.Persist
	Motion       *motion.Motion
	AmbientLight *ambientlight.AmbientLight
}

func NewMarvinFromFile(path string) (*Marvin, error) {
	if n, err := nog.NewNogFromFile(path); err == nil {
		marvin := &Marvin{Nog: n}
		go marvin.Add(&activity.Activity{})
		go marvin.Add(&schedule.Schedule{})
		go marvin.Add(&switches.Switches{})
		go marvin.Add(&hue.Hue{})
		marvin.p = &persist.Persist{}
		go marvin.Add(marvin.p)
		go marvin.Add(&presence.Presence{})
		marvin.Motion = &motion.Motion{}
		go marvin.Add(marvin.Motion)
		marvin.AmbientLight = &ambientlight.AmbientLight{}
		go marvin.Add(marvin.AmbientLight)
		return marvin, err
	} else {
		return nil, err
	}
}

func (m *Marvin) Log() []*nog.Message {
	return m.p.Log()
}
