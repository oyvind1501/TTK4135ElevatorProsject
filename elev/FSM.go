package elev

import (
	"time"
)

var isInitialized bool = false
var isOrderServed bool = false
var state int
var previousState int

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

func idle(motorChannel chan MotorDirection, lightChannel chan Light, doorChannel chan bool, requestChannel chan Action) {
	if previousState != IDLE {

		motorChannel <- DIR_STOP
	}
	if open && LastFloor == hallTarget && lastHallTarget != hallTarget {
		doorChannel <- true
		open = false
		doorOpened = true
		lastHallTarget = LastFloor
	}
	if openDoorAtFloor == LastFloor && doorOpenedAtFloor == true {
		doorChannel <- true
		doorOpenedAtFloor = false
		time.Sleep(2 * time.Second)
	}
	FloorAction(lightChannel, doorChannel, requestChannel)
	SetState(requestChannel, motorChannel)

	previousState = IDLE
}

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

func floorUp(motorChannel chan MotorDirection, lightChannel chan Light, floorChannel chan int) {
	var floorPoll int
	for {
		floorPoll = ReadFloorSensor(floorChannel)
		if floorPoll != UNDEFINED_TARGET_FLOOR {
			if floorPoll == TargetFloor {
				motorChannel <- DIR_STOP
				UpdateFloorIndicator(floorPoll, LastFloor, lightChannel)
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

func floorDown(motorChannel chan MotorDirection, lightChannel chan Light, floorChannel chan int) {
	var floorPoll int
	for {
		floorPoll = ReadFloorSensor(floorChannel)
		if floorPoll != UNDEFINED_TARGET_FLOOR {
			if floorPoll == TargetFloor {
				motorChannel <- DIR_STOP
				UpdateFloorIndicator(floorPoll, LastFloor, lightChannel)
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
