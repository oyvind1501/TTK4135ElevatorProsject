package elev

import (
	"time"
)

/*------------------------------------------------------------------------------
Function:	SetState
Affects:	States
Operation:
Enables the state of the Finite State Machine to be set, if needed.
------------------------------------------------------------------------------*/

func Core_SetState(requestChannel chan Action, motorChannel chan MotorDirection) {
	if TargetFloor == LastFloor || TargetFloor == UNDEFINED_TARGET_FLOOR {
		requestChannel <- Action{
			Command: ACTION_ORDER_DONE,
			Floor:   LastFloor,
		}
		state = IDLE
	} else if TargetFloor > LastFloor {
		isOrderServed = false
		state = UP
	} else if TargetFloor < LastFloor {
		isOrderServed = false
		state = DOWN
	} else {
		// Do nothing!!
	}

}

/*------------------------------------------------------------------------------
Function:	TargetFloorAction
Affects:	Cab/hall lights, doorlights and cabfloors
Operation:
Performes general actions at the valid floor, that is turns off cab
light on entry to a cab floor and requests a new order (cab or hall).
------------------------------------------------------------------------------*/

func Core_TargetFloorAction(lightChannel chan Light, doorChannel chan bool, requestChannel chan Action) {
	if Core_IsCabFloor(LastFloor) {
		lightChannel <- Light{
			LightType:   BUTTON_CAB,
			LightOn:     false,
			FloorNumber: LastFloor,
		}
		Core_RemoveCabOrder(LastFloor)
		Core_ServeOrder(doorChannel, requestChannel)
	} else if Core_IsCabOrderAbove(LastFloor) && Core_IsCabOrderBelow(LastFloor) {
		floorAbove := Core_GetCabOrderAbove(LastFloor)
		floorBelow := Core_GetCabOrderBelow(LastFloor)
		distanceAbove := LastFloor - floorAbove
		distanceBelow := floorBelow - LastFloor

		if distanceBelow <= distanceAbove {
			TargetFloor = floorBelow
		} else {
			TargetFloor = floorAbove
		}
	} else if Core_IsCabOrderAbove(LastFloor) {
		newTarget := Core_GetCabOrderAbove(LastFloor)
		TargetFloor = newTarget
	} else if Core_IsCabOrderBelow(LastFloor) {
		newTarget := Core_GetCabOrderBelow(LastFloor)
		TargetFloor = newTarget
	} else {
		requestChannel <- Action{
			Command:   ACTION_REQUEST_ORDER,
			Direction: ElevatorDirection,
			Floor:     LastFloor,
		}
		time.Sleep(500 * time.Millisecond)
	}
}

/*------------------------------------------------------------------------------
Function:	ServeOrder
Affects:	Doorlight and requests
Operation:
Serves an order, by opening the door, waiting and acknowledging that the
order is served.
------------------------------------------------------------------------------*/
func Core_ServeOrder(doorChannel chan bool, requestChannel chan Action) {
	doorChannel <- true
	time.Sleep(2 * time.Second)
	requestChannel <- Action{
		Command: ACTION_ORDER_DONE,
		Floor:   LastFloor,
	}
}
