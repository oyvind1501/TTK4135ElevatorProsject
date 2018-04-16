package elev

import (
	"./network/bcast"
	"math"
	"time"
)

/*------------------------------------------------------------------------------
Function:	NewOrderEvent
Operation:
Creates a new HallTable - entry, i.e. places a new hall order, at the master
and backup node. The master node will also acknowledge that a new order is placed.
------------------------------------------------------------------------------*/
func Net_NewOrderEvent(message ElevatorOrderMessage, sendChannel chan ElevatorOrderMessage) {
	if nodeId == masterId {
		tableElement := Core_CreateHallTableElement(message)
		if !Core_IsElementInHallTable(tableElement) {
			HallOrderTable = append(HallOrderTable, tableElement)
		}
		sendChannel <- ElevatorOrderMessage{
			Event:     EVENT_ACK_NEW_ORDER,
			Direction: message.Direction,
			Floor:     message.Floor,
			Origin:    message.Origin,
			Sender:    masterId,
		}
	}
	if nodeId == backupId {
		tableElement := Core_CreateHallTableElement(message)
		if !Core_IsElementInHallTable(tableElement) {
			HallOrderTable = append(HallOrderTable, tableElement)
		}
	}
}

/*------------------------------------------------------------------------------
Function:		AckNewOrderEvent
Operation:
The function is activated for all nodes(including master and backup),
upon receiving an New-Order-Event-Acknowledge from the master. By activation it
will turn on the hall button associated with the new order.
------------------------------------------------------------------------------*/

func Net_AckNewOrderEvent(message ElevatorOrderMessage, lightChannel chan Light) {
	var lightButton int
	switch message.Direction {
	case DIR_UP:
		lightButton = BUTTON_HALL_UP
	case DIR_DOWN:
		lightButton = BUTTON_HALL_DOWN
	}
	lightChannel <- Light{
		LightType:   lightButton,
		LightOn:     true,
		FloorNumber: message.Floor,
	}
}

/*-----------------------------------------------------
Function:		OrderReserveEvent
Operation:
Upon s received request the master node and backup node will both reserve
a hall order if appropriate. The master - node will then send a acknowledge to
back to the requesting node with the reservation (if any), or acknowledge with
an unvalid floor, -1.
-----------------------------------------------------*/

func Net_OrderReserveEvent(message ElevatorOrderMessage, sendChannel chan ElevatorOrderMessage) {
	nextFloor := net_setBestParticipantFloor(message)
	if nodeId == masterId {
		sendChannel <- ElevatorOrderMessage{
			Event:      EVENT_ACK_ORDER_RESERVE,
			Floor:      nextFloor,
			AssignedTo: message.Origin,
			Origin:     message.Origin,
			Sender:     masterId,
		}
	}
}

/*------------------------------------------------------------------------------
Function:		AckOrderReserveEvent
Operation:
Upon receiving the response, from the master node, to an order request,
the node will set its target floor.
------------------------------------------------------------------------------*/

func Net_AckOrderReserveEvent(message ElevatorOrderMessage) {
	if message.Origin == nodeId {
		if message.Floor != UNDEFINED {
			TargetFloor = message.Floor
		}
	}
}

/*------------------------------------------------------------------------------
Function:		OrderReserveSpecificEvent
Operation:
Enables all nodes to reserve a specific hall order, by checking if there is any
hall - order at the requested floor and direction. The master node then sends
back a response with the floor, if the request was sucessfull, or master will
respond with an invalid floor, -1.
------------------------------------------------------------------------------*/
func Net_OrderReserveSpecificEvent(message ElevatorOrderMessage, sendChannel chan ElevatorOrderMessage) {
	if nodeId == masterId {
		if Core_IsOrderAt(message.Floor, message.Direction) {
			sendChannel <- ElevatorOrderMessage{
				Event:      EVENT_ACK_ORDER_RESERVE_SPECIFIC,
				Floor:      message.Floor,
				AssignedTo: message.Origin,
				Origin:     message.Origin,
				Sender:     masterId,
			}
		} else {
			sendChannel <- ElevatorOrderMessage{
				Event:      EVENT_ACK_ORDER_RESERVE_SPECIFIC,
				Floor:      UNDEFINED,
				AssignedTo: message.Origin,
				Origin:     message.Origin,
				Sender:     masterId,
			}
		}
	}
}

/*-----------------------------------------------------
Function:		AckOrderReserveSpecificEvent
Operation:
Upon receiving the response to its Reserve-Specific-Order-Request, the node will
flag the floor as an intermediate stop.
-----------------------------------------------------*/

func Net_AckOrderReserveSpecificEvent(message ElevatorOrderMessage) {
	if message.Origin == nodeId {
		if message.Floor != UNDEFINED {
			IsIntermediateStop = true
		} else {
			IsIntermediateStop = false
		}
	}
}

/*------------------------------------------------------------------------------
Function:		OrderDoneEvent
Affects:		HallorderTable
Operation:
Upon receiving a order-done-message, master will responde by removing the
hall - order from the hall order table, and send an acknowledge.
------------------------------------------------------------------------------*/
func Net_OrderDoneEvent(message ElevatorOrderMessage, sendChannel chan ElevatorOrderMessage) {
	if nodeId == masterId {

		Core_RemoveHallOrder(message.Floor)
		sendChannel <- ElevatorOrderMessage{
			Event:      EVENT_ACK_ORDER_DONE,
			Floor:      message.Floor,
			AssignedTo: message.Origin,
			Origin:     message.Origin,
			Sender:     masterId,
		}
	}
	if nodeId == backupId && message.Floor != UNDEFINED {
		Core_RemoveHallOrder(message.Floor)
	}
}

/*------------------------------------------------------------------------------
Function:		AckOrderDoneEvent
Operation:
After receiving an acknowledge from master, that the hall - order
has been removed, all nodes will acknowledge by switching off the light in the
hall button of the floor.
------------------------------------------------------------------------------*/
func Net_AckOrderDoneEvent(message ElevatorOrderMessage, lightChannel chan Light) {
	lightChannel <- Light{
		LightType:   BUTTON_HALL_UP,
		LightOn:     false,
		FloorNumber: message.Floor,
	}
	lightChannel <- Light{
		LightType:   BUTTON_HALL_DOWN,
		LightOn:     false,
		FloorNumber: message.Floor,
	}
}

/*------------------------------------------------------------------------------
Function:	ClientOrderCommunication
Affects:	chan ElevatorOrderMessage, chan ElevatorOrderMessage,
					chan Light, chan bool
Operation:
Selects an appropriate response, based upon the event specified in the message.
------------------------------------------------------------------------------*/
func Net_ClientOrderCommunication(sendOrderChannel chan ElevatorOrderMessage, receiveOrderChannel chan ElevatorOrderMessage, lightChannel chan Light, doorChannel chan bool) {
	go bcast.Transmitter(15100, sendOrderChannel)
	go bcast.Receiver(15100, receiveOrderChannel)

	for {
		select {
		case message := <-receiveOrderChannel:
			switch message.Event {
			case EVENT_NEW_ORDER:
				Net_NewOrderEvent(message, sendOrderChannel)
			case EVENT_ACK_NEW_ORDER:
				Net_AckNewOrderEvent(message, lightChannel)
			case EVENT_ORDER_RESERVE:
				Net_OrderReserveEvent(message, sendOrderChannel)
			case EVENT_ACK_ORDER_RESERVE:
				Net_AckOrderReserveEvent(message)
			case EVENT_ORDER_RESERVE_SPECIFIC:
				Net_OrderReserveSpecificEvent(message, sendOrderChannel)
			case EVENT_ACK_ORDER_RESERVE_SPECIFIC:
				Net_AckOrderReserveSpecificEvent(message)
			case EVENT_ORDER_DONE:
				Net_OrderDoneEvent(message, sendOrderChannel)
			case EVENT_ACK_ORDER_DONE:
				Net_AckOrderDoneEvent(message, lightChannel)
			default:
				// Do nothing
			}
		}
	}
}

/*-----------------------------------------------------
Function:	setBestParticipantFloor
Affects:	Halltable
Operation:	Finds the best suited client in the client network
			for an incoming hallorder
-----------------------------------------------------*/

func net_setBestParticipantFloor(message ElevatorOrderMessage) int {
	if nodeId == masterId || nodeId == backupId {
		nextFloor := UNDEFINED

		var bestParticipantId string
		var participantId string

		participantFloor := -1
		shortestDistance := -1

		time.Sleep(100 * time.Millisecond)

		for index, participant := range ClientTable {
			participantId = participant.ClientInfo.Id
			participantFloor = participant.ClientInfo.Info.Floor
			closestOrder := Core_ClosestFloor(participantFloor)

			if closestOrder == UNDEFINED {
				continue
			}
			if index == 0 {
				bestParticipantId = participantId
				shortestDistance = int(math.Abs(float64(participantFloor - closestOrder)))
				nextFloor = closestOrder
			} else {
				distance := int(math.Abs(float64(participantFloor - closestOrder)))

				if distance < shortestDistance {
					shortestDistance = distance
					bestParticipantId = participantId
					nextFloor = closestOrder
				}
			}

		}
		if bestParticipantId != message.Origin {
			return -1
		}
		isReserved := Core_IsHallOrderReserved(nextFloor)
		if isReserved {
			return -1
		}

		Core_SetOrderStatus(STATUS_OCCUPIED, message.Origin, nextFloor)

		return nextFloor
	}
	return -1
}
