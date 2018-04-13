package main

import (
	"./elev"
	"./elev/driver/elevio"
)

func main() {
	// Disse kan legges et annet sted slik at vi kun har go routines her
	elevio.Init("localhost:15657", 4)
	elev.TargetFloor = elev.UNDEFINED_TARGET_FLOOR
	motorChannel := make(chan elev.MotorDirection)
	lightChannel := make(chan elev.Light)
	doorChannel := make(chan bool)
	floorChannel := make(chan int)
	buttonChannel := make(chan elevio.ButtonEvent)
	requestChannel := make(chan elev.Action)
	sendOrderChannel := make(chan elev.ElevatorOrderMessage)
	receiveOrderChannel := make(chan elev.ElevatorOrderMessage)

	go elev.ActionController(buttonChannel, lightChannel, doorChannel, requestChannel, sendOrderChannel)
	go elev.FiniteStateMachine(motorChannel, lightChannel, floorChannel, doorChannel, requestChannel)
	go elev.FreeLockedOrders()

	go elev.MotorController(motorChannel)
	go elev.LightController(lightChannel)
	go elev.DoorController(doorChannel)

	go elevio.PollFloorSensor(floorChannel)
	go elevio.PollButtons(buttonChannel)
	go elev.IdCommunication()
	go elev.OrderCommunication(sendOrderChannel, receiveOrderChannel, lightChannel, doorChannel)

	select {}
}
