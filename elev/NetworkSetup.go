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

/*-----------------------------------------------------
Function:	generateElevatorID
Affects:	Clientid
Operation:	Gets the IP-adress from an incoming client
-----------------------------------------------------*/

func generateElevatorID() string {
	localIP, err := localip.LocalIP()
	if err != nil {
		localIP = "DISCONNECTED"
	}
	id := fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	return id
}

/*-----------------------------------------------------
Function:	extractIdentifier
Affects:	
Operation:	Splits the IP-adress string into its last 4 digits.
-----------------------------------------------------*/

func extractIdentifier(nodeElement NetworkNode) string {
	return strings.Split(nodeElement.ClientInfo.Id, "-")[2]
}

/*-----------------------------------------------------
Function:	addNode
Affects:	ClientTable
Operation:	Adds a new client/node,which is connected to 
		the network, to the ClientTable.
-----------------------------------------------------*/

func addNode(peerUpdateMessage peers.PeerUpdate) {
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

/*-----------------------------------------------------
Function:	removeNode
Affects:	ClientTable
Operation:	Removes an client/node from the clientTable when the client/node becomes offline
-----------------------------------------------------*/
func removeNode(peerUpdateMessage peers.PeerUpdate) {
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

/*-----------------------------------------------------
Function:	setMasterId
Affects:	ClientTable
Operation:	Makes the client/node with the smallest 4 digits in the IP-information become the master
-----------------------------------------------------*/

func setMasterId(peerUpdateMessage peers.PeerUpdate) {
	if len(ClientTable) == 0 {
		return
	}

	var smallestIdentifier int
	var masterCandidate NetworkNode

	for index, nodeElement := range ClientTable {
		if nodeElement.ClientInfo.Id == "" {
			return
		}
		currentIdentifier, _ := strconv.Atoi(extractIdentifier(nodeElement))
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

/*-----------------------------------------------------
Function:	setBackupId
Affects:	ClientTable
Operation:	Makes the client/node with the nextsmallest 4 digits in the IP-information become the backup
-----------------------------------------------------*/
func setBackupId(peerUpdateMessage peers.PeerUpdate) {
	if len(ClientTable) == 0 {
		return
	}

	var listOfNodeIds []int

	for _, nodeElement := range ClientTable {
		if nodeElement.ClientInfo.Id == "" {
			return
		}
		currentIdentifier, _ := strconv.Atoi(extractIdentifier(nodeElement))
		listOfNodeIds = append(listOfNodeIds, currentIdentifier)
	}

	if len(listOfNodeIds) < 2 {
		backupId = "UNDEFINED"
	} else {
		sort.Sort(sort.IntSlice(listOfNodeIds))
		backupIdentifier := strconv.Itoa(listOfNodeIds[1])
		for _, backupCandidate := range ClientTable {
			currentIdentifier := extractIdentifier(backupCandidate)
			if currentIdentifier == backupIdentifier {
				backupId = backupCandidate.ClientInfo.Id
			}
		}
	}
}

/*-----------------------------------------------------
Function:	setBackupId
Affects:	ClientTable
Operation:	Uses the addnode and removenode function to update the ClientTable
-----------------------------------------------------*/
func updateClientTable(peerUpdateMessage peers.PeerUpdate) {
	if peerUpdateMessage.New != "" {
		addNode(peerUpdateMessage)
	}
	if len(peerUpdateMessage.Lost) != 0 {
		removeNode(peerUpdateMessage)
	}
}
/*-----------------------------------------------------
Function:	AddClientInfo
Affects:	ClientTable
Operation:	Puts the clientinformation for each client/node in the clientTable
-----------------------------------------------------*/
func AddClientInfo(message NetClient) {
	for index, node := range ClientTable {
		if message.Id == node.ClientInfo.Id {
			ClientTable[index] = NetworkNode{
				ClientInfo:   message,
				ActivityTime: node.ActivityTime,
			}
		}
	}
}
/*-----------------------------------------------------
Function:	SendClientInfo
Affects:	clientinforamtion to the client network
Operation:	Sets the clientinformation to be sent to the client network 
-----------------------------------------------------*/
func sendClientInfo(messageChannel chan NetClient) {
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

/*-----------------------------------------------------
Function:	IdCommunication
Affects:	
Operation:	Broadcasts client information, master, backup, and ids to all nodes in the clientNetwork
-----------------------------------------------------*/

func IdCommunication() {
	var id string
	id = generateElevatorID()
	nodeId = id

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(34641, id, peerTxEnable)
	go peers.Receiver(34641, peerUpdateCh)

	outgoingMessageChannel := make(chan NetClient)
	incomingMessageChannel := make(chan NetClient)
	go sendClientInfo(outgoingMessageChannel)

	go bcast.Transmitter(15254, outgoingMessageChannel)
	go bcast.Receiver(15254, incomingMessageChannel)

	for {
		select {
		case peerUpdateMessage := <-peerUpdateCh:
			updateClientTable(peerUpdateMessage)
			setMasterId(peerUpdateMessage)
			setBackupId(peerUpdateMessage)
		case message := <-incomingMessageChannel:
			AddClientInfo(message)
		}
	}
}
