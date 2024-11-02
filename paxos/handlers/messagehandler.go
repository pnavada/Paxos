package handlers

import (
	"fmt"
	"net"

	"paxos/paxos/datastructures"
	"paxos/paxos/network"
	"paxos/paxos/types"
	"paxos/paxos/utils"
)

type MessageHandler struct {
	Peer *network.Peer
}

func NewMessageHandler(p *network.Peer) *MessageHandler {
	return &MessageHandler{
		Peer: p,
	}
}

func (mh *MessageHandler) HandleMessages() {
	for {
		select {
		case inboundMessage := <-mh.Peer.ReadChannel:
			mh.processInboundMessage(inboundMessage)
		case outboundMessage := <-mh.Peer.WriteChannel:
			mh.sendMessage(outboundMessage)
		}
	}
}

func (mh *MessageHandler) processInboundMessage(message types.InboundMessage) {
	data, err := types.Deserialize(message.Data)
	if err != nil {
		fmt.Println("Error decoding byte array to integers:", err)
		return
	}

	sender, err := utils.GetHostnameFromAddr(message.Sender)
	if err != nil {
		fmt.Println("Error getting hostname from address:", err)
		return
	}

	sender = utils.CleanHostname(sender)
	mh.handleMessage(types.MessageType(data[0]), data[1:], sender)
}

func (mh *MessageHandler) handleMessage(msgType types.MessageType, data []int, sender string) {
	switch msgType {
	case types.PREPARE:
		mh.handlePrepareMessage(data, sender)
	case types.PREPARE_ACK:
		mh.handlePrepareAckMessage(data, sender)
	case types.ACCEPT:
		mh.handleAcceptMessage(data, sender)
	case types.ACCEPT_ACK:
		mh.handleAcceptAckMessage(data, sender)
	}
}

func (mh *MessageHandler) handlePrepareMessage(data []int, sender string) {
	proposalNumber := &types.ProposalNumber{
		RoundNumber: datastructures.NewSafeValue(data[0]),
		ServerId:    datastructures.NewSafeValue(data[1]),
	}

	n := utils.GetN(int32(proposalNumber.RoundNumber.Get()), int32(proposalNumber.ServerId.Get()))
	if n > mh.Peer.Store.MinProposalNumber.Get() {
		mh.Peer.Store.MinProposalNumber.Set(n)
	}
	if mh.Peer.Store.RoundNumber.Get() < proposalNumber.RoundNumber.Get() {
		mh.Peer.Store.RoundNumber.Set(proposalNumber.RoundNumber.Get())
	}
	go mh.Peer.SendPrepareAck(sender)
}

func (mh *MessageHandler) handlePrepareAckMessage(data []int, sender string) {
	n := utils.GetN(int32(mh.Peer.Store.RoundNumber.Get()), int32(mh.Peer.Id))
	if _, ok := mh.Peer.PrepareAck.Load(n); !ok {
		mh.Peer.PrepareAck.Store(n, datastructures.NewSafeList(make([][]int, 0)))
	}
	var dataList *datastructures.SafeList[[]int]
	value, _ := mh.Peer.PrepareAck.Load(n)
	dataList = value.(*datastructures.SafeList[[]int])
	dataList.Add(data)
	if dataList.Length() == mh.Peer.QuorumSize.Get() {
		highestProposalNumber := utils.GetN(-1, int32(mh.Peer.Id))
		for _, data := range dataList.GetAll() {
			acceptedRoundNumber := int32(data[0])
			acceptedServerId := int32(data[1])
			acceptedN := utils.GetN(acceptedRoundNumber, acceptedServerId)
			acceptedValue := data[2]
			if acceptedValue != 0 && acceptedN > highestProposalNumber {
				highestProposalNumber = acceptedN
				mh.Peer.ProposalValue.Set(rune(acceptedValue))
			}
		}
		go mh.Peer.SendAccept()
	}
}

func (mh *MessageHandler) handleAcceptMessage(data []int, sender string) {
	proposalNumber := types.ProposalNumber{
		RoundNumber: datastructures.NewSafeValue(data[0]),
		ServerId:    datastructures.NewSafeValue(data[1]),
	}
	proposalValue := data[2]
	n := utils.GetN(int32(proposalNumber.RoundNumber.Get()), int32(proposalNumber.ServerId.Get()))
	if n >= mh.Peer.Store.MinProposalNumber.Get() {
		mh.Peer.Store.MinProposalNumber.Set(n)
		mh.Peer.Store.AcceptedProposalNumber.Set(n)
		mh.Peer.Store.AcceptedValue.Set(rune(proposalValue))
	}
	go mh.Peer.SendAcceptAck(sender)
}

func (mh *MessageHandler) handleAcceptAckMessage(data []int, sender string) {
	n := utils.GetN(int32(mh.Peer.Store.RoundNumber.Get()), int32(mh.Peer.Id))
	if _, ok := mh.Peer.AcceptAck.Load(n); !ok {
		mh.Peer.AcceptAck.Store(n, datastructures.NewSafeList(make([][]int, 0)))
	}
	var dataList *datastructures.SafeList[[]int]
	value, _ := mh.Peer.AcceptAck.Load(n)
	dataList = value.(*datastructures.SafeList[[]int])
	dataList.Add(data)
	if dataList.Length() == mh.Peer.QuorumSize.Get() {
		for _, data := range dataList.GetAll() {
			minProposalRoundNumber := int32(data[0])
			minProposalServerId := int32(data[1])
			minProposalN := utils.GetN(minProposalRoundNumber, minProposalServerId)
			if minProposalN < n {
				go mh.Peer.SendPrepare()
				return
			}
		}
	}
	// Value is chosen
}

func (mh *MessageHandler) sendMessage(outboundMessage types.OutboundMessage) {
	conn, err := mh.Peer.TCPEgress.Get(outboundMessage.Recipient)

	if err != nil {
		fmt.Printf("Error getting connection: %v\n", err)
		return
	}

	tcpConn := conn.(net.Conn)
	_, err = tcpConn.Write(outboundMessage.Data)

	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}
