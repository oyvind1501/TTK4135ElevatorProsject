package elev

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"./network/bcast"
	"./network/localip"
	"./network/peers"
)

/*------------------------------------------------------------------------------
Function:	generateElevatorID
Affects:	Clientid
Operation:	Gets the IP-adress from an incoming client
------------------------------------------------------------------------------*/
func net_generateElevatorID() string {
	localIP, err := localip.LocalIP()
	if err != nil {
		localIP = "DISCONNECTED"
	}
	id := fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	return id
}

/*------------------------------------------------------------------------------
Function:		extractIdentifier
Operation:	Splits the IP-adress string into its last 4+ digits.
------------------------------------------------------------------------------*/

func net_extractIdentifier(nodeElement NetworkNode) string {
	return strings.Split(nodeElement.ClientInfo.Id, "-")[2]
}

/*------------------------------------------------------------------------------
Function:	addNode
Affects:	ClientTable
Operation:
Adds a new client/node,which is connected to
the network, to the ClientTable.
------------------------------------------------------------------------------*/

func net_addNode(peerUpdateMessage peers.PeerUpdate) {
	newNodeId := peerUpdateMessage.New
	if newNodeId == "" {
		return
	}
	shouldInsert := true
	for _, node := range ClientTable {
		nodeId := node.ClientInfo.Id
		if nodeId == newNodeId {
			shouldInsert = false
			break
		}
	}
	if shouldInsert {
		clientRecord := NetClient{
			Id: newNodeId,
		}
		newNetworkNode := NetworkNode{
			ClientInfo:   clientRecord,
			ActivityTime: time.Now(),
		}
		ClientTable = append(ClientTable, newNetworkNode)
	}
}

/*------------------------------------------------------------------------------
Function:	removeNode
Affects:	ClientTable
Operation:	Removes an client/node from the clientTable when the
client/node becomes offline
------------------------------------------------------------------------------*/
func net_removeNode(peerUpdateMessage peers.PeerUpdate) {
	lostNodesId := peerUpdateMessage.Lost

	shouldRemove := false
	var indexToDelete int

	for _, lostId := range lostNodesId {
		for index, node := range ClientTable {
			nodeId := node.ClientInfo.Id
			if nodeId == lostId {
				shouldRemove = true
				indexToDelete = index

				break
			}
		}
		if shouldRemove {
			ClientTable = append(ClientTable[:indexToDelete], ClientTable[indexToDelete+1:]...)
		}
	}
}

/*------------------------------------------------------------------------------
Function:	setMasterId
Affects:	ClientTable
Operation:	Makes the client/node with the smallest 4+ digits in the
IP-information the master
------------------------------------------------------------------------------*/

func net_setMasterId(peerUpdateMessage peers.PeerUpdate) {
	if len(ClientTable) == 0 {
		return
	}

	var smallestIdentifier int
	var masterCandidate NetworkNode

	for index, nodeElement := range ClientTable {
		if nodeElement.ClientInfo.Id == "" {
			return
		}
		currentIdentifier, _ := strconv.Atoi(net_extractIdentifier(nodeElement))
		if index == 0 {
			smallestIdentifier = currentIdentifier
			masterCandidate = nodeElement
		}
		if currentIdentifier < smallestIdentifier {
			smallestIdentifier = currentIdentifier
			masterCandidate = nodeElement
		}
	}
	masterId = masterCandidate.ClientInfo.Id
}

/*------------------------------------------------------------------------------
Function:	setBackupId
Affects:	ClientTable
Operation:
Makes the client/node with the nextsmallest 4+ digits in the
IP-information the backup
------------------------------------------------------------------------------*/
func net_setBackupId(peerUpdateMessage peers.PeerUpdate) {
	if len(ClientTable) == 0 {
		return
	}

	var listOfNodeIds []int

	for _, nodeElement := range ClientTable {
		if nodeElement.ClientInfo.Id == "" {
			return
		}
		currentIdentifier, _ := strconv.Atoi(net_extractIdentifier(nodeElement))
		listOfNodeIds = append(listOfNodeIds, currentIdentifier)
	}

	if len(listOfNodeIds) < 2 {
		backupId = "UNDEFINED"
	} else {
		sort.Sort(sort.IntSlice(listOfNodeIds))
		backupIdentifier := strconv.Itoa(listOfNodeIds[1])
		for _, backupCandidate := range ClientTable {
			currentIdentifier := net_extractIdentifier(backupCandidate)
			if currentIdentifier == backupIdentifier {
				backupId = backupCandidate.ClientInfo.Id
			}
		}
	}
}

/*------------------------------------------------------------------------------
Function:	setBackupId
Affects:	ClientTable
Operation:	Uses the addnode and removenode function to update the ClientTable
------------------------------------------------------------------------------*/
func net_updateClientTable(peerUpdateMessage peers.PeerUpdate) {
	if peerUpdateMessage.New != "" {
		net_addNode(peerUpdateMessage)
	}
	if len(peerUpdateMessage.Lost) != 0 {
		net_removeNode(peerUpdateMessage)
	}
}

/*------------------------------------------------------------------------------
Function:	AddClientInfo
Affects:	ClientTable
Operation:	Puts the clientinformation for each client/node in the clientTable
------------------------------------------------------------------------------*/
func Net_AddClientInfo(message NetClient) {
	for index, node := range ClientTable {
		if message.Id == node.ClientInfo.Id {
			ClientTable[index] = NetworkNode{
				ClientInfo:   message,
				ActivityTime: node.ActivityTime,
			}
		}
	}
}

/*------------------------------------------------------------------------------
Function:	SendClientInfo
Affects:	clientinforamtion to the client network
Operation:	Sets the clientinformation to be sent to the client network
------------------------------------------------------------------------------*/
func net_sendClientInfo(messageChannel chan NetClient) {
	for {
		var clientInformation ClientInfo
		if !clientInfoInitialized {
			clientInformation = ClientInfo{
				Floor: LastFloor,
			}
			clientInfoInitialized = true
		} else {
			clientInformation = ClientInfo{
				Floor: LastFloor,
			}
		}
		message := NetClient{
			Id:   nodeId,
			Info: clientInformation,
		}
		messageChannel <- message
		time.Sleep(time.Millisecond * 100)
	}
}

/*------------------------------------------------------------------------------
Function:		FreeOCCUPIEDOrders
Operation:
Sets the status of orders, that has been occupied for a given threshold time,
to STATUS_AVAILABLE. This do free potentially locked/non-served orders, if any
issues should arise.
------------------------------------------------------------------------------*/
func Net_FreeOCCUPIEDOrders() {
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
					Core_SetOrderStatus(STATUS_AVAILABLE, tableElement.ReserveID, tableElement.Floor)
				}
			}
			time.Sleep(2 * time.Second)
		}
	}
}

/*------------------------------------------------------------------------------
Function:	ClientInfoCommunication
Operation:	Broadcasts client information, master, backup,
and ids to all nodes in the clientNetwork
------------------------------------------------------------------------------*/

func Net_ClientInfoCommunication() {
	var id string
	id = net_generateElevatorID()
	nodeId = id

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(34641, id, peerTxEnable)
	go peers.Receiver(34641, peerUpdateCh)

	outgoingMessageChannel := make(chan NetClient)
	incomingMessageChannel := make(chan NetClient)
	go net_sendClientInfo(outgoingMessageChannel)

	go bcast.Transmitter(15254, outgoingMessageChannel)
	go bcast.Receiver(15254, incomingMessageChannel)

	for {
		select {
		case peerUpdateMessage := <-peerUpdateCh:
			net_updateClientTable(peerUpdateMessage)
			net_setMasterId(peerUpdateMessage)
			net_setBackupId(peerUpdateMessage)
		case message := <-incomingMessageChannel:
			Net_AddClientInfo(message)
		}
	}
}
