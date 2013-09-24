package marvin

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/stathat/go"

	"github.com/eikeon/dynamodb"
	"github.com/eikeon/gpio"
	"github.com/eikeon/hue"
	"github.com/eikeon/presence"
	"github.com/eikeon/scheduler"
	"github.com/eikeon/tsl2561"
)

type Message struct {
	Hash string `db:"HASH"`
	When string `db:"RANGE"`
	Who  string
	What string
	Why  string
}

func NewMessage(who, what, why string) Message {
	when := time.Now().Format(time.RFC3339Nano)

	hash := when[0:10]

	return Message{Hash: hash, When: when, What: what, Who: who, Why: why}
}

type activity struct {
	Name string
	Next map[string]bool
}

type messagelisteners struct {
	pendingmessages chan Message
	m               map[*chan Message]bool
	sync.Mutex
}

func (sc *messagelisteners) notifymessagelisteners() {
	for s := range sc.pendingmessages {
		sc.Lock()
		for o := range sc.m {
			select {
			case *o <- s:
			default:
				log.Println("unable to send to channel:", *o)
			}
		}
		sc.Unlock()
	}
}

func (sc *messagelisteners) RegisterMessageListener(c *chan Message) {
	sc.Lock()
	defer sc.Unlock()
	sc.m[c] = true
}

func (sc *messagelisteners) UnregisterMessageListener(c *chan Message) {
	sc.Lock()
	defer sc.Unlock()
	delete(sc.m, c)
	close(*c)
}

type State Marvin

type listeners struct {
	changes chan State
	last    State
	m       map[*chan State]bool
	sync.Mutex
}

func (sc *listeners) notify() {
	for s := range sc.changes {
		sc.Lock()
		sc.last = s
		for o := range sc.m {
			select {
			case *o <- s:
			default:
				log.Println("unable to send to channel:", *o)
			}
		}
		sc.Unlock()
	}
}

func (sc *listeners) Register(c *chan State) {
	sc.Lock()
	defer sc.Unlock()
	sc.m[c] = true
	*c <- sc.last
}

func (sc *listeners) Unregister(c *chan State) {
	sc.Lock()
	defer sc.Unlock()
	delete(sc.m, c)
	close(*c)
}

func (m *Marvin) StateChanged() {
	for len(m.changes) > 0 {
		select {
		case <-m.changes:
			log.Println("ignoring previous change since there's a new one")
		default:
		}
	}
	// encode and decode to make copy
	var buf bytes.Buffer
	ec := json.NewEncoder(&buf)
	if err := ec.Encode(m); err != nil {
		return
	}
	var c Marvin
	dec := json.NewDecoder(&buf)
	if err := dec.Decode(&c); err != nil {
		return
	}

	m.changes <- State(c)
}

type Marvin struct {
	Hue            hue.Hue
	Activities     map[string]*activity
	Activity       string
	Motion         bool
	DayLight       bool
	LastTransition string
	Present        map[string]bool
	Switch         map[string]bool
	Schedule       scheduler.Schedule
	States         map[string]interface{}
	Transitions    map[string]struct {
		Switch   map[string]bool
		Commands []struct {
			Address string
			State   string
		}
	}
	StatHatUserKey string
	StartedOn      time.Time
	MotionOn       time.Time

	do, persist   chan Message
	lightSensor   *tsl2561.TSL2561
	motionChannel <-chan bool
	lightChannel  <-chan int
	path          string
	db            dynamodb.DynamoDB
	*listeners
	*messagelisteners
}

func NewMarvinFromFile(path string) (*Marvin, error) {
	var marvin Marvin
	marvin.path = path
	if j, err := os.OpenFile(marvin.path, os.O_RDONLY, 0666); err == nil {
		dec := json.NewDecoder(j)
		if err = dec.Decode(&marvin); err != nil {
			return nil, err
		}
		j.Close()
	} else {
		return nil, err
	}
	marvin.initDB()
	return &marvin, nil
}

var messageTableName string = "MarvinMessage"

func init() {
	if hostname, err := os.Hostname(); err == nil {
		messageTableName = messageTableName + "-" + hostname
	} else {
		log.Println("error getting hostname:", err)
	}
}

func (m *Marvin) initDB() {
	db := dynamodb.NewDynamoDB()
	if db != nil {
		m.db = db
		t, err := db.Register(messageTableName, (*Message)(nil))
		if err != nil {
			panic(err)
		}
		pt := dynamodb.ProvisionedThroughput{ReadCapacityUnits: 1, WriteCapacityUnits: 1}
		if _, err := db.CreateTable(t.TableName, t.AttributeDefinitions, t.KeySchema, pt, nil); err != nil {
			log.Println("CreateTable:", err)
		}
		for {
			if description, err := db.DescribeTable(messageTableName, nil); err != nil {
				log.Println("DescribeTable err:", err)
			} else {
				log.Println(description.Table.TableStatus)
				if description.Table.TableStatus == "ACTIVE" {
					break
				}
			}
			time.Sleep(time.Second)
		}
		m.persist = make(chan Message, 512)
		go func() {
			for msg := range m.persist {
				db.PutItem(messageTableName, db.ToItem(&msg), nil)
			}
		}()
	}
}

func (m *Marvin) MotionSensor() bool {
	return m.motionChannel != nil
}

func (m *Marvin) LightSensor() bool {
	return m.lightChannel != nil
}

func (m *Marvin) GetActivity(name string) *activity {
	if name != "" {
		a, ok := m.Activities[name]
		if !ok {
			a = &activity{name, map[string]bool{}}
			m.Activities[name] = a
		}
		return a
	} else {
		return nil
	}
}

func (m *Marvin) UpdateActivity(name string) {
	s := m.GetActivity(m.Activity)
	if s != nil {
		s.Next[name] = true
	}
	m.GetActivity(name)
	m.Activity = name
}

func (m *Marvin) Do(who, what, why string) {
	msg := NewMessage(who, what, why)
	m.do <- msg
	if m.persist != nil {
		m.persist <- msg
	}
}

func (m *Marvin) Run() {
	m.listeners = &listeners{changes: make(chan State, 2), m: make(map[*chan State]bool)}
	go m.listeners.notify()
	m.messagelisteners = &messagelisteners{pendingmessages: make(chan Message, 10), m: make(map[*chan Message]bool)}
	go m.messagelisteners.notifymessagelisteners()
	m.StartedOn = time.Now()
	var createUserChan <-chan time.Time
	if err := m.Hue.GetState(); err != nil {
		createUserChan = time.NewTicker(1 * time.Second).C
	} else {
		// TODO:
	}
	m.StateChanged()
	if m.Switch == nil {
		m.Switch = make(map[string]bool)
	}
	if m.Activities == nil {
		m.Activities = make(map[string]*activity)
	}
	if m.Present == nil {
		m.Present = make(map[string]bool)
	}
	m.do = make(chan Message, 100)
	m.Do("Marvin", "chime", "startup")

	var scheduledEventsChannel <-chan scheduler.Event
	if c, err := m.Schedule.Run(); err == nil {
		scheduledEventsChannel = c
	} else {
		log.Println("Warning: Scheduled events off:", err)
	}

	var dayLightTime time.Time
	if t, err := tsl2561.NewTSL2561(1, tsl2561.ADDRESS_FLOAT); err == nil {
		m.lightSensor = t
		m.lightChannel = t.Broadband()
	} else {
		log.Println("Warning: Light sensor off: ", err)
	}

	if c, err := gpio.GPIOInterrupt(7); err == nil {
		m.motionChannel = c
	} else {
		log.Println("Warning: Motion sensor off:", err)
	}
	var motionTimer *time.Timer
	var motionTimeout <-chan time.Time

	presenceChannel := presence.Listen(m.Present)

	notifyChannel := make(chan os.Signal, 1)
	signal.Notify(notifyChannel, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)

	saveChannel := time.NewTicker(3600 * time.Second).C

	for {
		select {
		case <-createUserChan:
			if err := m.Hue.CreateUser(m.Hue.Username, "Marvin"); err == nil {
				createUserChan = nil
			} else {
				m.Do("Marvin", "press hue link button to authenticate", "setup")
			}
		case message := <-m.do:
			m.messagelisteners.pendingmessages <- message
			log.Println("Message:", message)
			what := ""
			const IAM = "I am "
			const SETHUE = "set hue address "
			const DOTRANSITION = "do transition "
			const TURN = "turn "
			if strings.HasPrefix(message.What, IAM) {
				what = message.What[len(IAM):]
				m.UpdateActivity(what)
			} else if strings.HasPrefix(message.What, SETHUE) {
				words := strings.Split(message.What[len(SETHUE):], " ")
				if len(words) == 3 {
					address := words[0]
					state := words[2]
					var s interface{}
					dec := json.NewDecoder(strings.NewReader(state))
					if err := dec.Decode(&s); err != nil {
						log.Println("json decode err:", err)
					} else {
						m.Hue.Set(address, s)
						err := m.Hue.GetState()
						if err != nil {
							log.Println("ERROR:", err)
						}
					}
				} else {
					log.Println("unexpected number of words in:", message)
				}
			} else if strings.HasPrefix(message.What, TURN) {
				words := strings.Split(message.What[len(TURN):], " ")
				if len(words) == 2 {
					var value bool
					if words[0] == "on" {
						value = true
					} else {
						value = false
					}
					if _, ok := m.Switch[words[1]]; ok {
						m.Switch[words[1]] = value
					}
				}
			} else if strings.HasPrefix(message.What, DOTRANSITION) {
				what = message.What[len(DOTRANSITION):]
			} else {
				what = message.What
			}
			t, ok := m.Transitions[what]
			if ok {
				for k, v := range t.Switch {
					m.Switch[k] = v
				}
			}
			m.LastTransition = what
			for _, command := range t.Commands {
				address := command.Address
				if strings.Contains(command.Address, "/light") {
					address += "/state"
				} else {
					address += "/action"
				}
				b, err := json.Marshal(m.States[command.State])
				if err != nil {
					log.Println("ERROR: json.Marshal: " + err.Error())
				} else {
					m.Do("Marvin", "set hue address "+address+" to "+string(b), what)
				}
			}
			m.StateChanged()
		case e := <-scheduledEventsChannel:
			if m.Switch["Schedule"] {
				m.Do("Marvin", e.What, "schedule")
			}
		case light := <-m.lightChannel:
			go m.postStatValue("light broadband", float64(light))
			if time.Since(dayLightTime) > time.Duration(60*time.Second) {
				if light > 5000 && (m.DayLight != true) {
					m.DayLight = true
					dayLightTime = time.Now()
					m.StateChanged()
					m.Do("Marvin", "it is light", "sensors")
					if m.Switch["Daylights"] {
						m.Do("Marvin", "daylight", "it is light")
					}
				} else if light < 4900 && (m.DayLight != false) {
					m.DayLight = false
					dayLightTime = time.Now()
					m.StateChanged()
					m.Do("Marvin", "it is dark", "sensors")
					if m.Switch["Daylights"] {
						m.Do("Marvin", "daylight off", "it is dark")
					}
				}
			}
		case motion := <-m.motionChannel:
			if motion {
				m.Do("Marvin", "motion detected", "sensors")
				m.MotionOn = time.Now()
				if m.Switch["Nightlights"] && m.LastTransition != "all nightlight" {
					m.Do("Marvin", "all nightlight", "motion detected")
				}
				const duration = 60 * time.Second
				if motionTimer == nil {
					m.Motion = true
					m.StateChanged()
					motionTimer = time.NewTimer(duration)
					motionTimeout = motionTimer.C // enable motionTimeout case
				} else {
					motionTimer.Reset(duration)
				}
				go m.postStatCount("motion", 1)
			}
		case <-motionTimeout:
			m.Motion = false
			m.StateChanged()
			motionTimer = nil
			motionTimeout = nil
			if m.Switch["Nightlights"] {
				m.Do("Marvin", "all off", "motion timeout")
			}
		case p := <-presenceChannel:
			if m.Present[p.Name] != p.Status {
				m.Present[p.Name] = p.Status
				m.StateChanged()
				var status string
				if p.Status {
					status = "home"
				} else {
					status = "away"
				}
				m.Do("Marvin", p.Name+" is "+status, "presence")

			}
		case <-saveChannel:
			if err := m.Save(m.path); err == nil {
				log.Println("saved:", m.path)
			} else {
				log.Println("ERROR: saving", err)
			}
		case sig := <-notifyChannel:
			log.Println("handling:", sig)
			goto Done
		}
	}
Done:
	if err := m.Save(m.path); err == nil {
		log.Println("saved:", m.path)
	} else {
		log.Println("ERROR: saving config", err)
	}
}

func (m *Marvin) Save(path string) error {
	if j, err := os.Create(path); err == nil {
		dec := json.NewEncoder(j)
		var c Marvin = *m
		if err = dec.Encode(&c); err != nil {
			return err
		}
		j.Close()
	} else {
		return err
	}
	return nil
}

func (m *Marvin) postStatValue(name string, value float64) {
	if m.StatHatUserKey != "" {
		if err := stathat.PostEZValue(name, m.StatHatUserKey, value); err != nil {
			log.Printf("error posting value %v: %d", err, value)
		}
	}
}

func (m *Marvin) postStatCount(name string, value int) {
	if m.StatHatUserKey != "" {
		if err := stathat.PostEZCount(name, m.StatHatUserKey, value); err != nil {
			log.Printf("error posting value %v: %d", err, value)
		}
	}
}

func (m *Marvin) Log() (messages []*Message) {
	if m.db != nil {
		when := time.Now().Format(time.RFC3339Nano)
		hash := when[0:10]
		conditions := dynamodb.KeyConditions{"Hash": {[]dynamodb.AttributeValue{{"S": hash}}, "EQ"}}
		if sr, err := m.db.Query(messageTableName, &dynamodb.QueryOptions{KeyConditions: conditions}); err == nil {
			for i := 0; i < sr.Count; i++ {
				messages = append(messages, m.db.FromItem(messageTableName, sr.Items[i]).(*Message))
			}
		} else {
			log.Println("scan error:", err)
		}
	}
	return
}
