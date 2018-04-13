package elev

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

const (
	INIT       int = 0
	IDLE       int = 1
	UP         int = 2
	FLOOR_UP   int = 3
	DOWN       int = 4
	FLOOR_DOWN int = 5
	//STOP       int = 6
)

const (
	DIR_UP   MotorDirection = 1
	DIR_DOWN MotorDirection = -1
	DIR_STOP MotorDirection = 0
)

/*
const (
	DOOR_CLOSED int = 0
	DOOR_OPEN   int = 1
)
*/

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

const UNDEFINED int = -1

const UNDEFINED_TARGET_FLOOR int = -1

//const UNDEFINED_NOT_AVAILABLE int = -2
const INVALID_FLOOR int = -1
const MAX_FLOOR_NUMBER int = 4

const (
//CloseDoor bool = false
//OpenDoor bool = false
)

var open bool
var doorOpenedAtFloor bool

var doorOpened bool = false
var openDoorAtFloor int = UNDEFINED

var hallTarget int = UNDEFINED
var lastHallTarget int = UNDEFINED

var LastFloor int
var TargetFloor int

var IsIntermediateStop bool
var ElevatorDirection MotorDirection

//var LastDirection MotorDirection

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

/*
func UpdateIndicator(indicator int, lightActive bool, floorNumber int, lightChannel chan Light) {
	lightChannel <- Light{LightType: indicator, LightOn: lightActive, FloorNumber: floorNumber}
}
*/
