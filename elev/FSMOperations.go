package elev

import (
	"time"
)

func SetState(requestChannel chan Action, motorChannel chan MotorDirection) {
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

func FloorAction(lightChannel chan Light, doorChannel chan bool, requestChannel chan Action) {
	if IsCabFloor(LastFloor) {
		lightChannel <- Light{
			LightType:   BUTTON_CAB,
			LightOn:     false,
			FloorNumber: LastFloor,
		}
		RemoveCabOrder(LastFloor)
		ServeOrder(doorChannel, requestChannel)
	} else if CabOrderAbove(LastFloor) && CabOrderBelow(LastFloor) {
		floorAbove := GetCabOrderAbove(LastFloor)
		floorBelow := GetCabOrderBelow(LastFloor)
		distanceAbove := LastFloor - floorAbove
		distanceBelow := floorBelow - LastFloor

		if distanceBelow <= distanceAbove {
			TargetFloor = floorBelow
		} else {
			TargetFloor = floorAbove
		}
	} else if CabOrderAbove(LastFloor) {
		newTarget := GetCabOrderAbove(LastFloor)
		TargetFloor = newTarget
	} else if CabOrderBelow(LastFloor) {
		newTarget := GetCabOrderBelow(LastFloor)
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

func ServeOrder(doorChannel chan bool, requestChannel chan Action) {
	doorChannel <- true
	time.Sleep(2 * time.Second)
	requestChannel <- Action{
		Command: ACTION_ORDER_DONE,
		Floor:   LastFloor,
	}
}
