package elev

import (
	"math"
	"time"
)

/*-----------------------------------------------------
Function:	NewOrderEvent
Affects:
Operation:
-----------------------------------------------------*/

func NewOrderEvent(message ElevatorOrderMessage, sendChannel chan ElevatorOrderMessage) {
	if nodeId == masterId {
		tableElement := createHallTableElement(message)
		if !isElementInHallTable(tableElement) {
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
		tableElement := createHallTableElement(message)
		if !isElementInHallTable(tableElement) {
			HallOrderTable = append(HallOrderTable, tableElement)
		}
	}
}

/*-----------------------------------------------------
Function:	AckNewOrderEvent
Affects:
Operation:
-----------------------------------------------------*/

func AckNewOrderEvent(message ElevatorOrderMessage, lightChannel chan Light) {
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
Function:	OrderReserveEvent
Affects:
Operation:
-----------------------------------------------------*/

func OrderReserveEvent(message ElevatorOrderMessage, sendChannel chan ElevatorOrderMessage) {
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
			closestOrder := ClosestFloor(participantFloor)

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
			return
		}
		isReserved := IsHallOrderReserved(nextFloor)
		if isReserved {
			return
		}

		SetOrderStatus(STATUS_OCCUPIED, message.Origin, nextFloor)
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
}

/*-----------------------------------------------------
Function:	AckOrderReserveEvent
Affects:
Operation:
-----------------------------------------------------*/

func AckOrderReserveEvent(message ElevatorOrderMessage) {
	if message.Origin == nodeId {
		if message.Floor != UNDEFINED {
			hallTarget = message.Floor
			doorOpened = false
			TargetFloor = message.Floor
		}
	}
}

/*-----------------------------------------------------
Function:	OrderReserveSpecificEvent
Affects:
Operation:
-----------------------------------------------------*/

func OrderReserveSpecificEvent(message ElevatorOrderMessage, sendChannel chan ElevatorOrderMessage) {
	if nodeId == masterId {
		if IsOrderAt(message.Floor, message.Direction) {
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
Function:	AckOrderReserveSpecificEvent
Affects:
Operation:
-----------------------------------------------------*/

func AckOrderReserveSpecificEvent(message ElevatorOrderMessage) {
	if message.Origin == nodeId {
		if message.Floor != UNDEFINED {
			IsIntermediateStop = true
		} else {
			IsIntermediateStop = false
		}
	}
}

/*-----------------------------------------------------
Function:	OrderDoneEvent
Affects:
Operation:
-----------------------------------------------------*/

func OrderDoneEvent(message ElevatorOrderMessage, sendChannel chan ElevatorOrderMessage) {
	if nodeId == masterId {

		RemoveHallOrder(message.Floor)
		sendChannel <- ElevatorOrderMessage{
			Event:      EVENT_ACK_ORDER_DONE,
			Floor:      message.Floor,
			AssignedTo: message.Origin,
			Origin:     message.Origin,
			Sender:     masterId,
		}
	}
	if nodeId == backupId && message.Floor != UNDEFINED {
		RemoveHallOrder(message.Floor)
	}
}

/*-----------------------------------------------------
Function:	AckOrderDoneEvent
Affects:
Operation:
-----------------------------------------------------*/

func AckOrderDoneEvent(message ElevatorOrderMessage, lightChannel chan Light) {
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
	if message.Origin == nodeId && doorOpened == false {
		open = true
	}
}

func createHallTableElement(message ElevatorOrderMessage) HallOrderElement {
	tableElement := HallOrderElement{
		Command:   message.Event,
		Direction: message.Direction,
		Floor:     message.Floor,
		ReserveID: "RESERVER_UNDEFINED",
		Status:    STATUS_AVAILABLE,
	}
	return tableElement
}

func isElementInHallTable(element HallOrderElement) bool {
	for _, tableElement := range HallOrderTable {
		if isHallTableElementEqual(element, tableElement) {
			return true
		}
	}
	return false
}

func isHallTableElementEqual(element HallOrderElement, tableElement HallOrderElement) bool {
	if element.Command == tableElement.Command && element.Direction == tableElement.Direction && element.Floor == tableElement.Floor {
		return true
	}
	return false
}

func UpdateReservationTable(sendChannel chan ElevatorOrderMessage) {
	if len(CabOrderTable) != 0 {
		for _, cabOrder := range CabOrderTable {
			ReserveTable = append(ReserveTable, ReserveElement{Floor: cabOrder.Floor})
			removeCabOrder(cabOrder)
		}
	}
	sendChannel <- ElevatorOrderMessage{
		Event:     EVENT_ORDER_RESERVE_SPECIFIC,
		Direction: ElevatorDirection,
		Floor:     LastFloor,
		Origin:    nodeId,
		Sender:    nodeId,
	}
}

func CheckForOrders(sendChannel chan ElevatorOrderMessage) {
	if len(CabOrderTable) != 0 {
		for _, cabOrder := range CabOrderTable {
			TargetFloor = GetCabOrder(LastFloor, ElevatorDirection)
			removeCabOrder(cabOrder)
			break
		}
	} else {
		sendChannel <- ElevatorOrderMessage{
			Event:     EVENT_ORDER_RESERVE,
			Direction: ElevatorDirection,
			Floor:     LastFloor,
			Origin:    nodeId,
			Sender:    nodeId,
		}
	}
}

func removeCabOrder(cabOrder CabOrderElement) {
	for index, element := range CabOrderTable {
		if element == cabOrder {
			CabOrderTable = append(CabOrderTable[:index], CabOrderTable[index+1:]...)
		}
	}
}
