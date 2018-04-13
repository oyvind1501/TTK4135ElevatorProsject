package elev

import (
	"./network/bcast"
)

func OrderCommunication(sendOrderChannel chan ElevatorOrderMessage, receiveOrderChannel chan ElevatorOrderMessage, lightChannel chan Light, doorChannel chan bool) {
	go bcast.Transmitter(15100, sendOrderChannel)
	go bcast.Receiver(15100, receiveOrderChannel)

	for {
		select {
		case message := <-receiveOrderChannel:
			switch message.Event {
			case EVENT_NEW_ORDER:
				NewOrderEvent(message, sendOrderChannel)
			case EVENT_ACK_NEW_ORDER:
				AckNewOrderEvent(message, lightChannel)
			case EVENT_ORDER_RESERVE:
				OrderReserveEvent(message, sendOrderChannel)
			case EVENT_ACK_ORDER_RESERVE:
				AckOrderReserveEvent(message)
			case EVENT_ORDER_RESERVE_SPECIFIC:
				OrderReserveSpecificEvent(message, sendOrderChannel)
			case EVENT_ACK_ORDER_RESERVE_SPECIFIC:
				AckOrderReserveSpecificEvent(message)
			case EVENT_ORDER_DONE:
				OrderDoneEvent(message, sendOrderChannel)
			case EVENT_ACK_ORDER_DONE:
				AckOrderDoneEvent(message, lightChannel)
			default:
				// Do nothing
			}
		}
	}
}
