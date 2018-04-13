package main

import (
	"./elev"
	"./elev/driver/elevio"
)

func main() {
	// Disse kan legges et annet sted slik at vi kun har go routines her
	elevio.Init("localhost:15657", 4)
	elev.TargetFloor = elev.UNDEFINED_TARGET_FLOOR
	
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
