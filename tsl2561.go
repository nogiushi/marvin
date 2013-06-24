package main

import (
	"log"
	"time"
)

const (
	ADDRESS_LOW   = 0x29
	ADDRESS_FLOAT = 0x39
	ADDRESS_HIGH  = 0x49

	COMMAND_BIT      = 0x80
	WORD_BIT         = 0x20
	CONTROL_POWERON  = 0x03
	CONTROL_POWEROFF = 0x00

	REGISTER_CONTROL    = 0x00
	REGISTER_TIMING     = 0x01
	REGISTER_CHAN0_LOW  = 0x0C
	REGISTER_CHAN0_HIGH = 0x0D
	REGISTER_CHAN1_LOW  = 0x0E
	REGISTER_CHAN1_HIGH = 0x0F
	REGISTER_ID         = 0x0A

	GAIN_1X  = 0x00
	GAIN_16X = 0x10

	INTEGRATIONTIME_13MS  = 0x00 // rather 13.7ms
	INTEGRATIONTIME_101MS = 0x01
	INTEGRATIONTIME_402MS = 0x02
)

type TSL2561 struct {
	bus          *i2c
	timing, gain byte
}

func NewTSL2561(bus, address byte) (t *TSL2561, err error) {
	b, err := NewI2C(bus)
	if err != nil {
		return nil, err
	}
	t = &TSL2561{b, INTEGRATIONTIME_402MS, GAIN_16X}
	err = b.SetAddress(address)
	return
}

func (t *TSL2561) On() (err error) {
	err = t.bus.WriteByte(COMMAND_BIT|REGISTER_CONTROL, CONTROL_POWERON)
	if err != nil {
		return err
	}
	err = t.bus.WriteByte(COMMAND_BIT|REGISTER_TIMING, t.timing|t.gain)
	if err != nil {
		return err
	}
	return
}

func (t *TSL2561) Off() (err error) {
	err = t.bus.WriteByte(COMMAND_BIT|REGISTER_CONTROL, CONTROL_POWEROFF)
	if err != nil {
		return err
	}
	return
}

func (t *TSL2561) IntegrationDuration() (duration time.Duration) {
	switch t.timing {
	case INTEGRATIONTIME_13MS:
		duration = time.Duration(13700 * time.Microsecond)
	case INTEGRATIONTIME_101MS:
		duration = time.Duration(102 * time.Millisecond)
	case INTEGRATIONTIME_402MS:
		duration = time.Duration(402 * time.Millisecond)
	}
	duration = duration * 780 / 735 // adjust from nominal to maximum
	return
}

func (t *TSL2561) scale(value int) int {
	switch t.timing {
	case INTEGRATIONTIME_13MS:
		value = value * 1000 / 34
	case INTEGRATIONTIME_101MS:
		value = value * 1000 / 252
	case INTEGRATIONTIME_402MS:
	default:
		panic("unexpected timing")
	}

	switch t.gain {
	case GAIN_1X:
		value *= 16
	case GAIN_16X:
	default:
		panic("unexpected gain")
	}
	return value
}

func (t *TSL2561) GetBroadband() (value int, err error) {
	low, err := t.bus.ReadByte(COMMAND_BIT | WORD_BIT | REGISTER_CHAN0_LOW)
	if err != nil {
		return -1, err
	}
	high, err := t.bus.ReadByte(COMMAND_BIT | WORD_BIT | REGISTER_CHAN0_HIGH)
	if err != nil {
		return -1, err
	}
	value = t.scale(int(high)*256 + int(low))
	return
}

func (t *TSL2561) GetInfrared() (value int, err error) {
	low, err := t.bus.ReadByte(COMMAND_BIT | WORD_BIT | REGISTER_CHAN1_LOW)
	if err != nil {
		return -1, err
	}
	high, err := t.bus.ReadByte(COMMAND_BIT | WORD_BIT | REGISTER_CHAN1_HIGH)
	if err != nil {
		return -1, err
	}
	value = t.scale(int(high)*256 + int(low))
	return
}

func (t *TSL2561) DayLightSingle() (value int, err error) {
	if err := t.On(); err != nil {
		log.Println("could not turn on:", err)
		return 0, err
	}
	time.Sleep(t.IntegrationDuration())

	value, err = t.GetBroadband()
	if err == nil {
		go postStat("light broadband", float64(value))
	} else {
		log.Println("error getting broadband value:", err)
	}

	if e := t.Off(); err != nil {
		log.Println("Could not turn off:", e)
	}
	return value, err
}

func (t *TSL2561) DayLight() chan bool {
	dayLight := make(chan bool, 1)

	go func() {
		var lastDayLight interface{}
		lastDayLightTime := time.Now()
		ticker := time.NewTicker(1 * time.Second)

		for {
			select {
			case <-ticker.C:
				if err := t.On(); err != nil {
					log.Fatal("could not turn on:", err)
				}
				time.Sleep(t.IntegrationDuration())

				if value, err := t.GetBroadband(); err == nil {
					dl := value > 5000
					if lastDayLight == nil {
						lastDayLight = dl
						lastDayLightTime = time.Now()
						dayLight <- dl
					} else if time.Since(lastDayLightTime) > time.Duration(60*time.Second) {
						if value > 5000 && lastDayLight == false {
							lastDayLight = true
							lastDayLightTime = time.Now()
							dayLight <- true
						} else if value < 4900 && lastDayLight == true {
							lastDayLight = false
							lastDayLightTime = time.Now()
							dayLight <- false
						}
					}
					go postStat("light broadband", float64(value))
				} else {
					log.Println("error getting broadband value:", err)
				}
				if value, err := t.GetInfrared(); err == nil {
					go postStat("light infrared", float64(value))
				} else {
					log.Println("error getting infrared value:", err)
				}

				if err := t.Off(); err != nil {
					log.Fatal("Could not turn off:", err)
				}
			}
		}
	}()
	return dayLight
}
