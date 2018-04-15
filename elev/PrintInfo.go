package elev

import (
	"fmt"
	"strconv"
	"time"
)

func printHallTable() {
	fmt.Println("-----------------Hall Order Table:----------------------")
	if len(HallOrderTable) == 0 {
		fmt.Println("	No hall")
	} else {
		for _, tableElement := range HallOrderTable {
			switch tableElement.Direction {
			case DIR_UP:
				if tableElement.Status == STATUS_AVAILABLE {
					fmt.Println(string(tableElement.Command) + " " + "UP" + " " + strconv.Itoa(tableElement.Floor) + " " + strconv.Itoa(tableElement.TimeReserved.Minute()) + ":" + strconv.Itoa(tableElement.TimeReserved.Second()) + " " + tableElement.ReserveID + " " + "AVAILABLE")
				}
				if tableElement.Status == STATUS_OCCUPIED {
					fmt.Println(string(tableElement.Command) + " " + "UP" + " " + strconv.Itoa(tableElement.Floor) + " " + strconv.Itoa(tableElement.TimeReserved.Minute()) + ":" + strconv.Itoa(tableElement.TimeReserved.Second()) + " " + tableElement.ReserveID + " " + "OCCUPIED")
				}

			case DIR_DOWN:
				if tableElement.Status == STATUS_AVAILABLE {
					fmt.Println(string(tableElement.Command) + " " + "DOWN" + " " + strconv.Itoa(tableElement.Floor) + " " + strconv.Itoa(tableElement.TimeReserved.Minute()) + ":" + strconv.Itoa(tableElement.TimeReserved.Second()) + " " + tableElement.ReserveID + " " + "AVAILABLE")
				}
				if tableElement.Status == STATUS_OCCUPIED {
					fmt.Println(string(tableElement.Command) + " " + "DOWN" + " " + strconv.Itoa(tableElement.Floor) + " " + strconv.Itoa(tableElement.TimeReserved.Minute()) + ":" + strconv.Itoa(tableElement.TimeReserved.Second()) + " " + tableElement.ReserveID + " " + "OCCUPIED")
				}
			}
		}
	}
	fmt.Println("--------------------------------------------------------")
}

func printCabTable() {
	fmt.Println("-----------------Cab Order Table------------------------")
	if len(CabOrderTable) == 0 {
		fmt.Println("	No cab")
	} else {
		for _, tableElement := range CabOrderTable {
			fmt.Println("	Order at Floor:\t\t" + strconv.Itoa(tableElement.Floor))
		}
	}
	fmt.Println("--------------------------------------------------------")
}

func printCommunicationTable() {
	fmt.Println("----------------Node Network Information----------------")
	for _, clients := range ClientTable {
		if clients.ClientInfo.Id == masterId {
			fmt.Println("Master:  ", masterId)
		}
		if clients.ClientInfo.Id == backupId {
			fmt.Println("Backup:  ", backupId)
		}
		if clients.ClientInfo.Id != masterId && clients.ClientInfo.Id != backupId {
			fmt.Println("Node:   ", clients.ClientInfo.Id)
		}
	}
	fmt.Println("--------------------------------------------------------")
}

/*------------------------------------------------------------------------------
Function:		PrintStateInfo
Operation:
Prints the targetfloor, lastfloor and the direction of the current node
------------------------------------------------------------------------------*/
func printStateInfo() {
	fmt.Println("-----------------Elevator Info--------------------------")
	fmt.Println("	Target floor:\t\t" + strconv.Itoa(TargetFloor))
	fmt.Println("	Last floor: \t\t" + strconv.Itoa(LastFloor))
	switch ElevatorDirection {
	case DIR_STOP:
		fmt.Println("	Direction:\t\tDIR_STOP")
	case DIR_UP:
		fmt.Println("	Direction:\t\tDIR_UP")
	case DIR_DOWN:
		fmt.Println("	Direction:\t\tDIR_DOWN")
	}
	fmt.Println("--------------------------------------------------------")
}

func PrintElevatorInfo() {
	for {
		printCommunicationTable()
		fmt.Println()
		printStateInfo()
		fmt.Println()
		printCabTable()
		fmt.Println()
		printHallTable()
		fmt.Println()

		time.Sleep(500 * time.Millisecond)
	}
}
