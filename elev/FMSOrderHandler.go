package elev

import (
	"math/rand"
	"time"
)

/*------------------------------------------------------------------------------
Function:		IsOrderAt
Operation:
Returns true if there is a hall order with at the floor in the given direction.
------------------------------------------------------------------------------*/
func Core_IsOrderAt(floor int, direction MotorDirection) bool {
	for _, element := range HallOrderTable {
		if element.Floor == floor && element.Direction == direction {
			return true
		}
	}
	return false
}

/*------------------------------------------------------------------------------
Function:		SerOrderStatus
Operation:
Sets the status of the hall - order.
------------------------------------------------------------------------------*/
func Core_SetOrderStatus(status int, id string, floor int) {
	if floor == UNDEFINED {
		return
	}
	for index, tableElement := range HallOrderTable {
		if tableElement.Floor == floor {
			HallOrderTable[index] = HallOrderElement{
				Command:      tableElement.Command,
				Direction:    tableElement.Direction,
				Floor:        tableElement.Floor,
				Status:       status,
				ReserveID:    id,
				TimeReserved: time.Now(),
			}
		}
	}
}

/*------------------------------------------------------------------------------
Function:		ClosestFloor
Operation:
Returns the closest floor to the elevator. If the distance between two
different floors are equal, it will select one of the two floors at random.
------------------------------------------------------------------------------*/
func Core_ClosestFloor(floor int) int {
	nextFloor := UNDEFINED
	if Core_IsHallOrderAbove(floor) && Core_IsHallOrderBelow(floor) {
		floorAbove := Core_GetHallOrderAbove(floor)
		floorBelow := Core_GetHallOrderBelow(floor)
		distanceAbove := floor - floorAbove
		distanceBelow := floorBelow - floor
		if distanceBelow < distanceAbove {
			nextFloor = floorBelow
		} else if distanceBelow > distanceBelow {
			nextFloor = floorAbove
		} else {
			nextFloor = Core_SelectRandom(floorAbove, floorBelow)
		}
	} else if Core_IsHallOrderAbove(floor) {
		nextFloor = Core_GetHallOrderAbove(floor)
	} else if Core_IsHallOrderBelow(floor) {
		nextFloor = Core_GetHallOrderBelow(floor)
	} else {
		nextFloor = floor
	}
	return nextFloor
}

func Core_AddCabOrder(floor int) {
	if !core_isElementInCabTable(floor) {
		cabOrder := CabOrderElement{
			Floor: floor,
		}
		CabOrderTable = append(CabOrderTable, cabOrder)
	}
}

/*------------------------------------------------------------------------------
Function:		SelectRandom
Operation:
Selects randomly between the two floors specified,
and returns the floor selected.
------------------------------------------------------------------------------*/
func Core_SelectRandom(floorA int, floorB int) int {
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)
	if random.Intn(100) > 50 {
		return floorA
	} else {
		return floorB
	}
}

func core_checkCabOrderAtFloor(floor int) int {
	for _, cabOrder := range CabOrderTable {
		if floor == cabOrder.Floor {
			return floor
		}
	}
	return UNDEFINED
}

func core_checkCabOrderAbove(floor int, direction MotorDirection) int {
	var bestFloor int
	var minDistance int
	var distance int

	bestFloor = UNDEFINED
	for _, order := range CabOrderTable {
		if direction == DIR_UP {
			distance = order.Floor - floor
			if minDistance == UNDEFINED {
				minDistance = distance
			}
			if distance < minDistance {
				distance = minDistance
				bestFloor = order.Floor
			}
		}
	}
	return bestFloor
}

func core_checkCabOrderBelow(floor int, direction MotorDirection) int {
	var bestFloor int
	var minDistance int
	var distance int

	bestFloor = UNDEFINED
	for _, order := range CabOrderTable {
		if direction == DIR_DOWN {
			distance = floor - order.Floor
			if minDistance == UNDEFINED {
				minDistance = distance
			}
			if distance < minDistance {
				distance = minDistance
				bestFloor = order.Floor
			}
		}
	}
	return bestFloor
}

/*------------------------------------------------------------------------------
Function:		CheckForOrders
Operation:
Sets the target - floor to a cab order (if any) or will send a request for a
hall - order.
------------------------------------------------------------------------------*/
func Core_CheckForOrders(sendChannel chan ElevatorOrderMessage) {
	if len(CabOrderTable) != 0 {
		for _, cabOrder := range CabOrderTable {
			TargetFloor = Core_GetCabOrder(LastFloor, ElevatorDirection)
			Core_RemoveCabOrderElement(cabOrder)
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

func Core_GetCabOrder(floor int, direction MotorDirection) int {
	var nextFloor int
	nextFloor = UNDEFINED
	switch direction {
	case DIR_STOP:
		nextFloor = core_checkCabOrderAtFloor(floor)
	case DIR_UP:
		nextFloor = core_checkCabOrderAbove(floor, direction)
	case DIR_DOWN:
		nextFloor = core_checkCabOrderBelow(floor, direction)
	}
	return nextFloor
}

func core_isElementInCabTable(floor int) bool {
	for _, element := range CabOrderTable {
		if element.Floor == floor {
			return true
		}
	}
	return false
}

func Core_IsCabOrderAbove(floor int) bool {
	for _, element := range CabOrderTable {
		if element.Floor > floor {
			return true
		}
	}
	return false
}

func Core_IsCabOrderBelow(floor int) bool {
	for _, element := range CabOrderTable {
		if element.Floor < floor {
			return true
		}
	}
	return false
}

func Core_IsCabFloor(floor int) bool {
	for _, element := range CabOrderTable {
		if element.Floor == floor {
			return true
		}
	}
	return false
}

func Core_IsElementInHallTable(element HallOrderElement) bool {
	for _, tableElement := range HallOrderTable {
		if Core_IsHallTableElementEqual(element, tableElement) {
			return true
		}
	}
	return false
}

func Core_IsHallTableElementEqual(element HallOrderElement, tableElement HallOrderElement) bool {
	if element.Command == tableElement.Command && element.Direction == tableElement.Direction && element.Floor == tableElement.Floor {
		return true
	}
	return false
}

func Core_IsHallOrderBelow(floor int) bool {
	for _, element := range HallOrderTable {
		if element.Floor < floor {
			return true
		}
	}
	return false
}

func Core_IsHallOrderReserved(floor int) bool {
	if floor == UNDEFINED {
		return false
	}
	isReserved := false
	for _, order := range HallOrderTable {
		if order.Floor == floor && order.Status == STATUS_OCCUPIED {
			isReserved = true
			break
		}
	}
	return isReserved
}

func Core_IsHallOrderAbove(floor int) bool {
	for _, element := range HallOrderTable {
		if element.Floor > floor {
			return true
		}
	}
	return false
}

func Core_CreateHallTableElement(message ElevatorOrderMessage) HallOrderElement {
	tableElement := HallOrderElement{
		Command:   message.Event,
		Direction: message.Direction,
		Floor:     message.Floor,
		ReserveID: "RESERVER_UNDEFINED",
		Status:    STATUS_AVAILABLE,
	}
	return tableElement
}

func Core_GetCabOrderAbove(floor int) int {
	var bestFloor int
	var minDistance int
	var distance int

	bestFloor = UNDEFINED
	minDistance = UNDEFINED
	for _, order := range CabOrderTable {
		distance = order.Floor - floor
		if minDistance == UNDEFINED {
			minDistance = distance
		}
		if bestFloor == UNDEFINED {
			bestFloor = order.Floor
		}
		if distance < minDistance {
			distance = minDistance
			bestFloor = order.Floor
		}
	}
	return bestFloor
}

func Core_GetCabOrderBelow(floor int) int {
	var bestFloor int
	var minDistance int
	var distance int

	bestFloor = UNDEFINED
	minDistance = UNDEFINED
	for _, order := range CabOrderTable {
		distance = floor - order.Floor
		if minDistance == UNDEFINED {
			minDistance = distance
		}
		if bestFloor == UNDEFINED {
			bestFloor = order.Floor
		}
		if distance < minDistance {
			distance = minDistance
			bestFloor = order.Floor
		}
	}
	return bestFloor
}

func Core_GetHallOrderAbove(floor int) int {
	var bestFloor int
	var minDistance int
	var distance int

	bestFloor = UNDEFINED
	minDistance = UNDEFINED
	for _, order := range HallOrderTable {
		distance = order.Floor - floor
		if minDistance == UNDEFINED {
			minDistance = distance
		}
		if bestFloor == UNDEFINED {
			bestFloor = order.Floor
		}
		if distance < minDistance {
			distance = minDistance
			bestFloor = order.Floor
		}
	}
	return bestFloor
}

func Core_GetHallOrderBelow(floor int) int {
	var bestFloor int
	var minDistance int
	var distance int

	bestFloor = UNDEFINED
	minDistance = UNDEFINED
	for _, order := range HallOrderTable {
		distance = floor - order.Floor
		if minDistance == UNDEFINED {
			minDistance = distance
		}
		if bestFloor == UNDEFINED {
			bestFloor = order.Floor
		}
		if distance < minDistance {
			distance = minDistance
			bestFloor = order.Floor
		}
	}
	return bestFloor
}

func Core_RemoveHallOrder(floor int) {
	for index, order := range HallOrderTable {
		if len(HallOrderTable) == 0 || len(HallOrderTable) <= index {
			break
		}
		if order.Floor == floor {
			HallOrderTable = append(HallOrderTable[:index], HallOrderTable[index+1:]...)
		}
	}
}

func Core_RemoveCabOrderElement(cabOrder CabOrderElement) {
	for index, element := range CabOrderTable {
		if element == cabOrder {
			CabOrderTable = append(CabOrderTable[:index], CabOrderTable[index+1:]...)
		}
	}
}

func Core_RemoveCabOrder(floor int) {
	for index, element := range CabOrderTable {
		if element.Floor == floor {
			CabOrderTable = append(CabOrderTable[:index], CabOrderTable[index+1:]...)
		}
	}
}


