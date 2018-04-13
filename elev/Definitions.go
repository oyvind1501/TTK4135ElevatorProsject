package elev

import (
  "./driver/elevio"
)

type MessageEvent int
type ActionCommand int
type MotorDirection int

type ElevatorOrderMessage struct {
	Event      MessageEvent
	Direction  MotorDirection
	Floor      int
	AssignedTo string
	Origin     string
	Sender     string
}


type HallOrderElement struct {
	Command      MessageEvent
	Direction    MotorDirection
	Floor        int
	Status       int
	ReserveID    string
	TimeReserved time.Time
}

type CabOrderElement struct {
	Floor int
}
type ReserveElement struct {
	Floor int
}

type Action struct {
	Command   ActionCommand
	Direction MotorDirection
	Floor     int
}

type Light struct {
	LightType   int
	LightOn     bool
	FloorNumber int
}

type ClientInfo struct {
	Floor int
}

type NetClient struct {
	Id   string
	Info ClientInfo
}

type NetworkNode struct {
	ClientInfo   NetClient
	ActivityTime time.Time
}

const (
	INIT       int = 0
	IDLE       int = 1
	UP         int = 2
	FLOOR_UP   int = 3
	DOWN       int = 4
	FLOOR_DOWN int = 5
)

const (
	DIR_UP   MotorDirection = 1
	DIR_DOWN MotorDirection = -1
	DIR_STOP MotorDirection = 0
)

const (
	BUTTON_HALL_UP   int = 0
	BUTTON_HALL_DOWN int = 1
	BUTTON_CAB       int = 2
	FLOOR_INDICATOR  int = 3
)

const (
	ACTION_ORDER ActionCommand = iota
	ACTION_REQUEST_ORDER
	ACTION_REQUEST_SPECIFIC_ORDER
	ACTION_ORDER_DONE
	ACTION_RESET_ALL_LIGHTS
)

const (
	EVENT_NEW_ORDER MessageEvent = iota
	EVENT_ACK_NEW_ORDER
	EVENT_ORDER_RESERVE
	EVENT_ACK_ORDER_RESERVE
	EVENT_ORDER_RESERVE_SPECIFIC
	EVENT_ACK_ORDER_RESERVE_SPECIFIC
	EVENT_ORDER_DONE
	EVENT_ACK_ORDER_DONE
)

const (
	STATUS_AVAILABLE int = 0
	STATUS_OCCUPIED  int = 1
)

MotorChannel := make(chan MotorDirection)
LightChannel := make(chan Light)
DoorChannel := make(chan bool)
FloorChannel := make(chan int)
ButtonChannel := make(chan elevio.ButtonEvent)
RequestChannel := make(chan Action)
SendOrderChannel := make(chan ElevatorOrderMessage)
ReceiveOrderChannel := make(chan ElevatorOrderMessage)


const UNDEFINED int = -1
const UNDEFINED_TARGET_FLOOR int = -1
const INVALID_FLOOR int = -1
const MAX_FLOOR_NUMBER int = 4

var HallOrderTable []HallOrderElement
var CabOrderTable []CabOrderElement
var ReserveTable []ReserveElement

var isInitialized bool = false
var isOrderServed bool = false
var state int
var previousState int

var ClientTable []NetworkNode
var masterId string
var backupId string
var nodeId string
var clientInfoInitialized bool = false

var open bool
var doorOpenedAtFloor bool
var doorOpened bool = false
var openDoorAtFloor int = UNDEFINED

var hallTarget int = UNDEFINED
var lastHallTarget int = UNDEFINED
var LastFloor int
var TargetFloor int
// IsIntermediateStop kan tas ut av systemet
var IsIntermediateStop bool

var ElevatorDirection MotorDirection

// Skal dette være i definitions?
func ReadFloorSensor(floorChannel chan int) int {
	select {
	case floor := <-floorChannel:
		return floor
	default:
		return INVALID_FLOOR
	}
}

func UpdateFloorIndicator(floorNumber int, prevFloorNumber int, lightChannel chan Light) {
	lightChannel <- Light{LightType: FLOOR_INDICATOR, LightOn: false, FloorNumber: prevFloorNumber}
	lightChannel <- Light{LightType: FLOOR_INDICATOR, LightOn: true, FloorNumber: floorNumber}
}

