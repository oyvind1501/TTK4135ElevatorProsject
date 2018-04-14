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

func MotorController(motorChannel chan MotorDirection) {
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
Function:	LigthController
Affects:	Hall/cab lights
Operation:	Sets the Hall/cab lights to on
-----------------------------------------------------*/

func LightController(lightChannel chan Light) {
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
Affects:	Kan du skrive på denne Robin? 
Operation:	
-----------------------------------------------------*/

// Denne funksjonen kan deles opp i flere funksjoner
func ActionButtonController(buttonChannel chan elevio.ButtonEvent, lightChannel chan Light, doorChannel chan bool, sendChannel chan ElevatorOrderMessage){
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
				AddCabOrder(buttonEvent.Floor)
			}
		}
	}
}

/*-----------------------------------------------------
Function:	ActionRequestController
Affects:	
Operation:	
-----------------------------------------------------*/

func ActionRequestController(buttonChannel chan elevio.ButtonEvent, lightChannel chan Light, doorChannel chan bool, requestActionChannel chan Action, sendChannel chan ElevatorOrderMessage) {

	for {
		select {
		case requestEvent := <-requestActionChannel:
			switch requestEvent.Command {
			case ACTION_REQUEST_ORDER:
				CheckForOrders(sendChannel)
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
				ResetLights(lightChannel)
			}
		}
	}
}

/*-----------------------------------------------------
Function:	ResetLights
Affects:	Hall/cab lights
Operation:	Turns off hall and cab lights at the corresponding floor
-----------------------------------------------------*/
func ResetLights(lightChannel chan Light) {
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

func DoorController(doorChannel chan bool) {
	for {
		select {
		case openDoor := <-doorChannel:
			if openDoor {
				openDoorAction()
			} else {
				elevio.SetDoorOpenLamp(false)
			}
		}
	}
}

/*-----------------------------------------------------
Function:	OpenDoorAction
Affects:	doorlight
Operation:	Turns on the doorlight for 2 seconds
-----------------------------------------------------*/
func openDoorAction() {
	elevio.SetDoorOpenLamp(true)
	time.Sleep(2 * time.Second)
	elevio.SetDoorOpenLamp(false)
}
