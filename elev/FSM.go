package elev

import (
	"time"
)

/*------------------------------------------------------------------------------
Function:	initState
Affects: 	State, motor, hall/cab lights and doorlight
Operation:
 - Initilizes the elevator to move to the closest floor if the floorsensor
	 is UNDEFINED.
 - Resets all lights
 - Sets the state to IDLE
------------------------------------------------------------------------------*/

func core_initState(motorChannel chan MotorDirection, floorChannel chan int, requestChannel chan Action) {
	if isInitialized {
		state = IDLE
	}
	requestChannel <- Action{
		Command: ACTION_RESET_ALL_LIGHTS,
		Floor:   LastFloor,
	}
	var elevatorIsApproaching bool = false
	for {
		floor := Core_ReadFloorSensorController(floorChannel)
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

/*------------------------------------------------------------------------------
Function:	idle
Affects:	State, doorlights, floorlights, motor
Operation:
Idle is activated upon arrival at a valid floor, and performes all nessesary
actions when the elevator is standing fixed at a valid floor. That is:
 -	Idle stops the elevator, opens the door (if required), before setting the
 		next state of the elevator.
 - 	Idle will also send out order requests if required.
------------------------------------------------------------------------------*/
func core_idle(motorChannel chan MotorDirection, lightChannel chan Light, doorChannel chan bool, requestChannel chan Action) {
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
		// Timer to compensate for RTT - delay required to obtain the ACTION_REQUEST_SPECIFIC_ORDER - request
		time.Sleep(1 * time.Second)
		if IsIntermediateStop == true {
			Core_ServeOrder(doorChannel, requestChannel)
		}
		IsIntermediateStop = false
	}
	if openDoorAtFloor == LastFloor && doorOpenedAtFloor == true {
		doorOpenedAtFloor = false
		Core_ServeOrder(doorChannel, requestChannel)
	}
	Core_TargetFloorAction(lightChannel, doorChannel, requestChannel)
	Core_SetState(requestChannel, motorChannel)

	previousState = IDLE
}

/*------------------------------------------------------------------------------
Function:	up
Affects:	Motor, direction, state
Operation:
Sets the elevator direction and state to go 'up', if allowed.
------------------------------------------------------------------------------*/

func core_up(motorChannel chan MotorDirection, floorChannel chan int) {
	floor := Core_ReadFloorSensorController(floorChannel)
	if floor == MAX_FLOOR_NUMBER {
		state = DOWN
		return
	}
	ElevatorDirection = DIR_UP
	motorChannel <- DIR_UP
	state = FLOOR_UP
	previousState = UP
}

/*------------------------------------------------------------------------------
Function:	down
Affects:	Motor, direction, state
Operation:
Sets the motordirection, state and direction to downwards
------------------------------------------------------------------------------*/

func core_down(motorChannel chan MotorDirection, floorChannel chan int) {
	floor := Core_ReadFloorSensorController(floorChannel)
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
Function:	floorUp
Affects:	Motor, hall/cab lights and floor
Operation:
		Moves the elevator to a floor above
		Turns on the cab/hall lights
		Sets the motordirection to stop when the destination is reached
		Sees if the target floor was a floor up
-----------------------------------------------------*/

func core_floorUp(motorChannel chan MotorDirection, lightChannel chan Light, floorChannel chan int) {
	var floorPoll int
	for {
		floorPoll = Core_ReadFloorSensorController(floorChannel)
		if floorPoll != UNDEFINED_TARGET_FLOOR {
			Core_UpdateFloorIndicatorController(floorPoll, LastFloor, lightChannel)
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

/*------------------------------------------------------------------------------
Function:	floorDown
Affects:	Motor, hall/cab lights and floor
Operation:
		Moves the elevator to a floor below
		Turns on the cab/hall lights when the destiniation is reached
		Sets the motordirection to stop when the destination is reached
------------------------------------------------------------------------------*/

func core_floorDown(motorChannel chan MotorDirection, lightChannel chan Light, floorChannel chan int) {
	var floorPoll int
	for {
		floorPoll = Core_ReadFloorSensorController(floorChannel)
		if floorPoll != UNDEFINED_TARGET_FLOOR {
			Core_UpdateFloorIndicatorController(floorPoll, LastFloor, lightChannel)
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

/*------------------------------------------------------------------------------
Function:	FiniteStateMachine
Affects:	state, Motor, hall/cab lights, door lights and requests from server
Operation:
Sees what state the elevator is in and runs the correct action accordingly
------------------------------------------------------------------------------*/
func Core_FiniteStateMachine(motorChannel chan MotorDirection, lightChannel chan Light, floorChannel chan int, doorChannel chan bool, requestChannel chan Action) {
	for {
		if !isInitialized {
			state = INIT
		}
		switch state {
		case INIT:
			core_initState(motorChannel, floorChannel, requestChannel)
		case IDLE:
			core_idle(motorChannel, lightChannel, doorChannel, requestChannel)
		case UP:
			core_up(motorChannel, floorChannel)
		case FLOOR_UP:
			core_floorUp(motorChannel, lightChannel, floorChannel)
		case DOWN:
			core_down(motorChannel, floorChannel)
		case FLOOR_DOWN:
			core_floorDown(motorChannel, lightChannel, floorChannel)
		}
	}
}
