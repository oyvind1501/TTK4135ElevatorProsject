package elev

import (
	"math/rand"
	"time"
)

/*-----------------------------------------------------
Function:	IsOrderAt
Affects:	
Operation:	Tests to see if an hallorder is at floor
-----------------------------------------------------*/


func IsOrderAt(floor int, direction MotorDirection) bool {
	for _, element := range HallOrderTable {
		if element.Floor == floor && element.Direction == direction {
			return true
		}
	}
	return false
}

/*-----------------------------------------------------
Function:	SetOrderStatus
Affects:	
Operation:	Gives the client/node the orderinformation from an hallorder
-----------------------------------------------------*/

func SetOrderStatus(status int, id string, floor int) {
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
/*-----------------------------------------------------
Function:	SetOrderStatus
Affects:	
Operation:	When backup is taking over a timer is set 
		to make Hallorders that are OCCUPIED to AVAILABLE
-----------------------------------------------------*/
func FreeLockedOrders() {

	for {
		if nodeId == masterId || nodeId == backupId {
			thresholdTime := 10
			for _, tableElement := range HallOrderTable {
				if tableElement.TimeReserved.Second() == 0 {
					continue
				}
				sinceLastTimestamp := time.Since(tableElement.TimeReserved)
				secondsElapsed := int(sinceLastTimestamp.Seconds())
				if secondsElapsed >= thresholdTime {
					SetOrderStatus(STATUS_AVAILABLE, tableElement.ReserveID, tableElement.Floor)
				}
			}
			time.Sleep(2 * time.Second)
		}
	}
}

func AddCabOrder(floor int) {
	if !isElementInCabTable(floor) {
		cabOrder := CabOrderElement{
			Floor: floor,
		}
		CabOrderTable = append(CabOrderTable, cabOrder)
	}
}

func checkCabOrderAtFloor(floor int) int {
	for _, cabOrder := range CabOrderTable {
		if floor == cabOrder.Floor {
			return floor
		}
	}
	return UNDEFINED
}

func checkCabOrderAbove(floor int, direction MotorDirection) int {
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

func checkCabOrderBelow(floor int, direction MotorDirection) int {
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

func GetCabOrder(floor int, direction MotorDirection) int {
	var nextFloor int
	nextFloor = UNDEFINED
	switch direction {
	case DIR_STOP:
		nextFloor = checkCabOrderAtFloor(floor)
	case DIR_UP:
		nextFloor = checkCabOrderAbove(floor, direction)
	case DIR_DOWN:
		nextFloor = checkCabOrderBelow(floor, direction)
	}
	return nextFloor
}

func isElementInCabTable(floor int) bool {
	for _, element := range CabOrderTable {
		if element.Floor == floor {
			return true
		}
	}
	return false
}

func CabOrderAbove(floor int) bool {
	for _, element := range CabOrderTable {
		if element.Floor > floor {
			return true
		}
	}
	return false
}

func CabOrderBelow(floor int) bool {
	for _, element := range CabOrderTable {
		if element.Floor < floor {
			return true
		}
	}
	return false
}

func IsCabFloor(floor int) bool {
	for _, element := range CabOrderTable {
		if element.Floor == floor {
			return true
		}
	}
	return false
}

func GetCabOrderAbove(floor int) int {
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

func GetCabOrderBelow(floor int) int {
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

func RemoveCabOrder(floor int) {
	for index, element := range CabOrderTable {
		if element.Floor == floor {
			CabOrderTable = append(CabOrderTable[:index], CabOrderTable[index+1:]...)
		}
	}
}

func HallOrderAbove(floor int) bool {
	for _, element := range HallOrderTable {
		if element.Floor > floor {
			return true
		}
	}
	return false
}

func HallOrderBelow(floor int) bool {
	for _, element := range HallOrderTable {
		if element.Floor < floor {
			return true
		}
	}
	return false
}

func GetHallOrderAbove(floor int) int {
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

func GetHallOrderBelow(floor int) int {
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

func RemoveHallOrder(floor int) {
	for index, order := range HallOrderTable {
		if len(HallOrderTable) == 0 || len(HallOrderTable) <= index {
			break
		}
		if order.Floor == floor {
			HallOrderTable = append(HallOrderTable[:index], HallOrderTable[index+1:]...)
		}
	}
}

func IsHallOrderReserved(floor int) bool {
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

func SelectRandom(floorA int, floorB int) int {
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)
	if random.Intn(100) > 50 {
		return floorA
	} else {
		return floorB
	}
}

func ClosestFloor(floor int) int {
	nextFloor := UNDEFINED
	if HallOrderAbove(floor) && HallOrderBelow(floor) {
		floorAbove := GetHallOrderAbove(floor)
		floorBelow := GetHallOrderBelow(floor)
		distanceAbove := floor - floorAbove
		distanceBelow := floorBelow - floor
		if distanceBelow < distanceAbove {
			nextFloor = floorBelow
		} else if distanceBelow > distanceBelow {
			nextFloor = floorAbove
		} else {
			nextFloor = SelectRandom(floorAbove, floorBelow)
		}
	} else if HallOrderAbove(floor) {
		nextFloor = GetHallOrderAbove(floor)
	} else if HallOrderBelow(floor) {
		nextFloor = GetHallOrderBelow(floor)
	} else {
		nextFloor = floor
	}
	return nextFloor
}
