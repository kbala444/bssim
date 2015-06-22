// netvis
package main

import (
	//"fmt"
	//"time"
)

type Event struct{
	time string
	eventType string
	node *Node
	message *Message
}

type Message struct{
	sourceNode *Node
	destinationNode *Node
	departureTime string
	arrivalTime string
	protocol string
	messageType string
	size int
	contents string
}

type Node struct{
	id int
	nodeType string
	addresses []string
	state string
	enterTime string
	exitTime string
	messagesSent int
	messagesReceived int
}
