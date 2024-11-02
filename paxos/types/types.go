package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"paxos/paxos/datastructures"
)

type Protocol int

const (
	TCP Protocol = iota
	UDP
)

type MessageType int

const (
	PREPARE MessageType = iota
	PREPARE_ACK
	ACCEPT
	ACCEPT_ACK
)

type Role int

const (
	Proposer = iota
	Acceptor
	Learner
)

type InboundMessage struct {
	Data   []byte
	Sender net.Addr
}

type OutboundMessage struct {
	Data      []byte
	Recipient net.Addr
}

type ProposalNumber struct {
	RoundNumber *datastructures.SafeValue[int]
	ServerId    *datastructures.SafeValue[int]
}

type PrepareMessage struct {
	ProposalNumber *ProposalNumber
	ProposalValue  *datastructures.SafeValue[rune]
}

type PrepareAckMessage struct {
	AcceptedProposalNumber *ProposalNumber
	AcceptedValue          *datastructures.SafeValue[rune]
}

type AcceptMessage struct {
	ProposalNumber *ProposalNumber
	ProposalValue  *datastructures.SafeValue[rune]
}

type AcceptAckMessage struct {
	ProposalNumber *ProposalNumber
}

func Serialize(integers ...int) []byte {
	var buffer bytes.Buffer
	for _, integer := range integers {
		err := binary.Write(&buffer, binary.LittleEndian, int32(integer))
		if err != nil {
			fmt.Println("Error encoding integer:", err)
			return nil
		}
	}
	return buffer.Bytes()
}

func Deserialize(data []byte) ([]int, error) {
	var integers []int
	buffer := bytes.NewReader(data)
	for {
		var integer int32
		err := binary.Read(buffer, binary.LittleEndian, &integer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error decoding integer: %v", err)
		}
		integers = append(integers, int(integer))
	}
	return integers, nil
}
