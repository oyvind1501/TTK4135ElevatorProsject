package elev

import (
	"time"

	"./driver/elevio"
)

/*-----------------------------------------------------
Function:	MotorController
Affects:	Motor
Operation:	Sets the motor to either STOP, DOWN OR UP
-----------------------------------------------------*/

func Core_MotorController(motorChannel chan MotorDirection) {
	for {
		select {
		case command := <-motorChannel:
			switch command {
			case DIR_UP:
				elevio.SetMotorDirection(elevio.MD_Up)
			case DIR_DOWN:
				elevio.SetMotorDirection(elevio.MD_Down)
			case DIR_STOP:
				elevio.SetMotorDirection(elevio.MD_Stop)
			}
		}
	}
}

/*-----------------------------------------------------
Function:	LightController
Affects:	Hall/cab lights
Operation:	Sets the Hall/cab lights to on
-----------------------------------------------------*/

func Core_LightController(lightChannel chan Light) {
	for {
		select {
		case command := <-lightChannel:
			switch command.LightType {
			case BUTTON_HALL_UP:
				elevio.SetButtonLamp(elevio.BT_HallUp, command.FloorNumber, command.LightOn)
			case BUTTON_HALL_DOWN:
				elevio.SetButtonLamp(elevio.BT_HallDown, command.FloorNumber, command.LightOn)
			case BUTTON_CAB:
				elevio.SetButtonLamp(elevio.BT_Cab, command.FloorNumber, command.LightOn)
			case FLOOR_INDICATOR:
				elevio.SetFloorIndicator(command.FloorNumber)
			}
		}
	}
}

/*-----------------------------------------------------
Function:	ActionButtonController
Operation:	Controls the hall/cab buttons order
		from the nodes.
-----------------------------------------------------*/

func Core_ActionButtonController(buttonChannel chan elevio.ButtonEvent, lightChannel chan Light, doorChannel chan bool, sendChannel chan ElevatorOrderMessage) {
	for {
		select {
		case buttonEvent := <-buttonChannel:
			switch buttonEvent.Button {
			case (elevio.ButtonType)(BUTTON_HALL_UP):
				sendChannel <- ElevatorOrderMessage{
					Event:     EVENT_NEW_ORDER,
					Direction: DIR_UP,
					Floor:     buttonEvent.Floor,
					Origin:    nodeId,
					Sender:    nodeId,
				}
				openDoorAtFloor = buttonEvent.Floor
				doorOpenedAtFloor = true
			case (elevio.ButtonType)(BUTTON_HALL_DOWN):
				sendChannel <- ElevatorOrderMessage{
					Event:     EVENT_NEW_ORDER,
					Direction: DIR_DOWN,
					Floor:     buttonEvent.Floor,
					Origin:    nodeId,
					Sender:    nodeId,
				}
				openDoorAtFloor = buttonEvent.Floor
				doorOpenedAtFloor = true
			case (elevio.ButtonType)(BUTTON_CAB):
				lightChannel <- Light{
					LightType:   BUTTON_CAB,
					FloorNumber: buttonEvent.Floor,
					LightOn:     true,
				}
				Core_AddCabOrder(buttonEvent.Floor)
			}
		}
	}
}

/*-----------------------------------------------------
Function:	ActionRequestController
Operation:	Controls the flow of requests from the nodes.
-----------------------------------------------------*/
func Core_ActionRequestController(buttonChannel chan elevio.ButtonEvent, lightChannel chan Light, doorChannel chan bool, requestActionChannel chan Action, sendChannel chan ElevatorOrderMessage) {
	for {
		select {
		case requestEvent := <-requestActionChannel:
			switch requestEvent.Command {
			case ACTION_REQUEST_ORDER:
				Core_CheckForOrders(sendChannel)
			case ACTION_REQUEST_SPECIFIC_ORDER:
				sendChannel <- ElevatorOrderMessage{
					Event:     EVENT_ORDER_RESERVE_SPECIFIC,
					Direction: requestEvent.Direction,
					Floor:     requestEvent.Floor,
					Origin:    nodeId,
					Sender:    nodeId,
				}
			case ACTION_ORDER_DONE:
				sendChannel <- ElevatorOrderMessage{
					Event:  EVENT_ORDER_DONE,
					Floor:  requestEvent.Floor,
					Origin: nodeId,
					Sender: nodeId,
				}
			case ACTION_RESET_ALL_LIGHTS:
				core_resetLightsController(lightChannel)
			}
		}
	}
}

func core_resetLightsController(lightChannel chan Light) {
	for i := 0; i < (MAX_FLOOR_NUMBER - 1); i++ {
		lightChannel <- Light{
			LightType:   BUTTON_HALL_UP,
			FloorNumber: i,
			LightOn:     false,
		}
	}
	for i := 0; i < MAX_FLOOR_NUMBER-1; i++ {
		lightChannel <- Light{
			LightType:   BUTTON_HALL_DOWN,
			FloorNumber: i,
			LightOn:     false,
		}
	}
	for i := 0; i < MAX_FLOOR_NUMBER; i++ {
		lightChannel <- Light{
			LightType:   BUTTON_CAB,
			FloorNumber: i,
			LightOn:     false,
		}
	}
	elevio.SetDoorOpenLamp(false)
}

/*-----------------------------------------------------
Function:	DoorController
Affects:	doorlight
Operation:	Sees if its necessary to turn on the doorligt
-----------------------------------------------------*/

func Core_DoorController(doorChannel chan bool) {
	for {
		select {
		case openDoor := <-doorChannel:
			if openDoor {
				Core_OpenDoorActionController()
			} else {
				elevio.SetDoorOpenLamp(false)
			}
		}
	}
}

/*-----------------------------------------------------
Function:	FiniteStateMachineControllers
Operation:	Collects all controllers in one function
-----------------------------------------------------*/
func Core_FiniteStateMachineControllers(buttonChannel chan elevio.ButtonEvent, lightChannel chan Light, doorChannel chan bool, requestActionChannel chan Action, sendChannel chan ElevatorOrderMessage, motorChannel chan MotorDirection) {
	go Core_ActionButtonController(buttonChannel, lightChannel, doorChannel, sendChannel)
	go Core_ActionRequestController(buttonChannel, lightChannel, doorChannel, requestActionChannel, sendChannel)
	go Core_MotorController(motorChannel)
	go Core_LightController(lightChannel)
	go Core_DoorController(doorChannel)
}

/*-----------------------------------------------------
Function:	OpenDoorAction
Affects:	doorlight
Operation:	Turns on the doorlight for 2 seconds
-----------------------------------------------------*/
func Core_OpenDoorActionController() {
	elevio.SetDoorOpenLamp(true)
	time.Sleep(2 * time.Second)
	elevio.SetDoorOpenLamp(false)
}

func Core_ReadFloorSensorController(floorChannel chan int) int {
	select {
	case floor := <-floorChannel:
		return floor
	default:
		return INVALID_FLOOR
	}
}

func Core_UpdateFloorIndicatorController(floorNumber int, prevFloorNumber int, lightChannel chan Light) {
	lightChannel <- Light{LightType: FLOOR_INDICATOR, LightOn: false, FloorNumber: prevFloorNumber}
	lightChannel <- Light{LightType: FLOOR_INDICATOR, LightOn: true, FloorNumber: floorNumber}
}
