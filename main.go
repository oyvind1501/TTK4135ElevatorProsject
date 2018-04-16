package main

import (
	"./elev"
	"./elev/driver/elevio"
)

/*******************************************************************************
TTK4135 Elevator Project
The code in this project describes a single elevator node.
It is written in a very general manner, which enables multiple nodes to be
connected together in a network. The nodes communicate with eachother and
selects internally a master node, responsible for handeling the hall orders,
and delegating these orders to each node in the network upon request.
To guarantee that no orders are lost when a node is lost/detached from the
network, the system internally selects a backup node (if there is two or more
nodes connected to the network).
*******************************************************************************/

func main() {
	elevio.Init("localhost:15657", elev.MAX_FLOOR_NUMBER)

	elev.TargetFloor = elev.UNDEFINED_TARGET_FLOOR

	motorChannel := make(chan elev.MotorDirection)
	lightChannel := make(chan elev.Light)
	doorChannel := make(chan bool)
	floorChannel := make(chan int)
	buttonChannel := make(chan elevio.ButtonEvent)
	requestChannel := make(chan elev.Action)
	sendOrderChannel := make(chan elev.ElevatorOrderMessage)
	receiveOrderChannel := make(chan elev.ElevatorOrderMessage)

	go elev.Core_FiniteStateMachine(motorChannel, lightChannel, floorChannel, doorChannel, requestChannel)
	go elev.Core_FiniteStateMachineControllers(buttonChannel, lightChannel, doorChannel, requestChannel, sendOrderChannel, motorChannel)

	go elevio.Net_PollFloorSensor(floorChannel)
	go elevio.Net_PollButtons(buttonChannel)
	go elev.Net_FreeOCCUPIEDOrders()
	go elev.Net_ClientInfoCommunication()
	go elev.Net_ClientOrderCommunication(sendOrderChannel, receiveOrderChannel, lightChannel, doorChannel)

	go elev.PrintElevatorInfo()

	select {}
}
