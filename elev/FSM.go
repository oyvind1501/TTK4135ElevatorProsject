package elev

import (
	"time"
)

var nowFloor int
var elevDir MotorDirection

/*-----------------------------------------------------
Function:	initstate
Affects:	states, floors, motor and lights
Operation:	Initilizes the elevator to move to the closest floor if the floorsensor is UNDEFINED.
		Resets all lights
		Sets the state to IDLE
-----------------------------------------------------*/

func initState(motorChannel chan MotorDirection, floorChannel chan int, requestChannel chan Action) {
	if isInitialized {
		state = IDLE
	}
	requestChannel <- Action{
		Command: ACTION_RESET_ALL_LIGHTS,
		Floor:   LastFloor,
	}
	var elevatorIsApproaching bool = false
	for {
		floor := ReadFloorSensor(floorChannel)
		if floor == INVALID_FLOOR && elevatorIsApproaching == false {
			motorChannel <- DIR_DOWN
			LastFloor = INVALID_FLOOR
			elevatorIsApproaching = true
		}
		if floor != INVALID_FLOOR {
			motorChannel <- DIR_STOP
			ElevatorDirection = DIR_STOP
			LastFloor = floor
			state = IDLE
			break
		}
	}
	doorOpened = false
	isInitialized = true
	previousState = INIT
}

/*-----------------------------------------------------
Function:	idle
Affects:	State, doorlights, floorlights, motor
Operation:	Kan du skrive hva denne gjør Robin?
-----------------------------------------------------*/

func idle(motorChannel chan MotorDirection, lightChannel chan Light, doorChannel chan bool, requestChannel chan Action) {
	if previousState != IDLE {
		motorChannel <- DIR_STOP

		if nowFloor < LastFloor {
			elevDir = DIR_DOWN
		} else if nowFloor > LastFloor {
			elevDir = DIR_UP
		} else {
			elevDir = DIR_STOP
		}
		requestChannel <- Action{
			Command:   ACTION_REQUEST_SPECIFIC_ORDER,
			Direction: elevDir,
			Floor:     LastFloor,
		}
		nowFloor = LastFloor
		time.Sleep(1 * time.Second) // Timer to compensate for RTT - delay required to obtain the ACTION_REQUEST_SPECIFIC_ORDER - request!
		if IsIntermediateStop == true {
			ServeOrder(doorChannel, requestChannel)
		}
		IsIntermediateStop = false
	}
	if openDoorAtFloor == LastFloor && doorOpenedAtFloor == true {
		doorOpenedAtFloor = false
		ServeOrder(doorChannel, requestChannel)
	}
	FloorAction(lightChannel, doorChannel, requestChannel)
	SetState(requestChannel, motorChannel)

	previousState = IDLE
}

/*-----------------------------------------------------
Function:	up
Affects:	Motor, direction, state
Operation:	Sets the motordirection, state and direction to upwards
-----------------------------------------------------*/

func up(motorChannel chan MotorDirection, floorChannel chan int) {
	floor := ReadFloorSensor(floorChannel)
	if floor == MAX_FLOOR_NUMBER {
		state = DOWN
		return
	}
	ElevatorDirection = DIR_UP
	motorChannel <- DIR_UP
	state = FLOOR_UP
	previousState = UP
}

/*-----------------------------------------------------
Function:	down
Affects:	Motor, direction, state
Operation:	Sets the motordirection, state and direction to downwards
-----------------------------------------------------*/

func down(motorChannel chan MotorDirection, floorChannel chan int) {
	floor := ReadFloorSensor(floorChannel)
	if floor == 0 {
		state = UP
		return
	}
	ElevatorDirection = DIR_DOWN
	motorChannel <- DIR_DOWN
	state = FLOOR_DOWN
	previousState = DOWN
}

/*-----------------------------------------------------
Function:	floorup
Affects:	Motor, hall/cab lights and floor
Operation:	Moves the elevator to a floor above
		Turns on the cab/hall lights
		Sets the motordirection to stop when the destination is reached
		Sees if the target floor was a floor up
-----------------------------------------------------*/

func floorUp(motorChannel chan MotorDirection, lightChannel chan Light, floorChannel chan int) {
	var floorPoll int
	for {
		floorPoll = ReadFloorSensor(floorChannel)
		if floorPoll != UNDEFINED_TARGET_FLOOR {
			UpdateFloorIndicator(floorPoll, LastFloor, lightChannel)
			if floorPoll == TargetFloor {
				motorChannel <- DIR_STOP
				lightChannel <- Light{
					LightType:   BUTTON_CAB,
					LightOn:     false,
					FloorNumber: floorPoll,
				}
			}
			LastFloor = floorPoll
			break
		}
	}
	previousState = FLOOR_UP
	state = IDLE
}

/*-----------------------------------------------------
Function:	floordown
Affects:	Motor, hall/cab lights and floor
Operation:	Moves the elevator to a floor below
		Turns on the cab/hall lights when the destiniation is reached
		Sets the motordirection to stop when the destination is reached
-----------------------------------------------------*/

func floorDown(motorChannel chan MotorDirection, lightChannel chan Light, floorChannel chan int) {
	var floorPoll int
	for {
		floorPoll = ReadFloorSensor(floorChannel)
		if floorPoll != UNDEFINED_TARGET_FLOOR {
			UpdateFloorIndicator(floorPoll, LastFloor, lightChannel)
			if floorPoll == TargetFloor {
				motorChannel <- DIR_STOP
				lightChannel <- Light{
					LightType:   BUTTON_CAB,
					LightOn:     false,
					FloorNumber: floorPoll,
				}
			}
			LastFloor = floorPoll
			break
		}
	}
	previousState = FLOOR_DOWN
	state = IDLE
}

/*-----------------------------------------------------
Function:	FiniteStateMachine
Affects:	state, Motor, hall/cab lights, door lights and requests from server
Operation:	Sees what state the elevator is in and runs the correct action accordingly
-----------------------------------------------------*/
func FiniteStateMachine(motorChannel chan MotorDirection, lightChannel chan Light, floorChannel chan int, doorChannel chan bool, requestChannel chan Action) {
	for {
		if !isInitialized {
			state = INIT
		}
		switch state {
		case INIT:
			initState(motorChannel, floorChannel, requestChannel)
		case IDLE:
			idle(motorChannel, lightChannel, doorChannel, requestChannel)
		case UP:
			up(motorChannel, floorChannel)
		case FLOOR_UP:
			floorUp(motorChannel, lightChannel, floorChannel)
		case DOWN:
			down(motorChannel, floorChannel)
		case FLOOR_DOWN:
			floorDown(motorChannel, lightChannel, floorChannel)
		}
	}
}
