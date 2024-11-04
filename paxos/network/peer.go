package network

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"paxos/paxos/datastructures"
	"paxos/paxos/types"
	"paxos/paxos/utils"
)

type PeerStore struct {
	MinProposalNumber      *datastructures.SafeValue[int64]
	AcceptedProposalNumber *datastructures.SafeValue[int64]
	AcceptedValue          *datastructures.SafeValue[rune]
	RoundNumber            *datastructures.SafeValue[int]
}

type Peer struct {
	Id            int
	Roles         *datastructures.SafeList[types.Role]
	Acceptors     *datastructures.SafeList[string]
	Peers         *datastructures.SafeList[string]
	Store         *PeerStore
	ProposerId    int
	TCPEgress     *ConnectionPool
	TCPIngress    *ConnectionPool
	ReadChannel   chan types.InboundMessage
	WriteChannel  chan types.OutboundMessage
	ProposalValue *datastructures.SafeValue[rune]
	QuorumSize    *datastructures.SafeValue[int]
	PrepareAck    sync.Map
	AcceptAck     sync.Map // map[string]types.AcceptAckMessage
}

const tcpPort = 8080

func (p *Peer) Start() {
	go p.ListenForTCPConnections()
	// If I am the proposer, send prepare to acceptors
	time.Sleep(1 * time.Second)
	if p.ProposerId != -1 {
		go p.SendPrepare()
	}
}

func (p *Peer) SendPrepare() {
	p.Store.RoundNumber.Set(p.Store.RoundNumber.Get() + 1)
	prepareMessage := types.PrepareMessage{
		ProposalNumber: &types.ProposalNumber{
			RoundNumber: datastructures.NewSafeValue(p.Store.RoundNumber.Get()),
			ServerId:    datastructures.NewSafeValue(p.Id),
		},
		ProposalValue: p.ProposalValue,
	}
	data := types.Serialize(
		int(types.PREPARE),
		prepareMessage.ProposalNumber.RoundNumber.Get(),
		prepareMessage.ProposalNumber.ServerId.Get(),
		int(prepareMessage.ProposalValue.Get()),
	)
	for _, acceptor := range p.Acceptors.GetAll() {
		go p.SendMessageToPeer(acceptor, data)
		p.LogMessage(
			"sent",
			"prepare",
			p.ProposalValue.Get(),
			p.Id,
			fmt.Sprintf(
				"%d.%d",
				p.Store.RoundNumber.Get(),
				p.Id,
			),
		)
	}
}

func (p *Peer) SendPrepareAck(sender string) {
	acceptedRoundNumber, acceptedServerId := utils.SplitN(p.Store.AcceptedProposalNumber.Get())
	prepareAckMessage := types.PrepareAckMessage{
		AcceptedProposalNumber: &types.ProposalNumber{
			RoundNumber: datastructures.NewSafeValue(int(acceptedRoundNumber)),
			ServerId:    datastructures.NewSafeValue(int(acceptedServerId)),
		},
		AcceptedValue: datastructures.NewSafeValue(p.Store.AcceptedValue.Get()),
	}
	data := types.Serialize(
		int(types.PREPARE_ACK),
		prepareAckMessage.AcceptedProposalNumber.RoundNumber.Get(),
		prepareAckMessage.AcceptedProposalNumber.ServerId.Get(),
		int(prepareAckMessage.AcceptedValue.Get()),
	)
	p.SendMessageToPeer(sender, data)
	p.LogMessage(
		"sent",
		"prepare_ack",
		rune(prepareAckMessage.AcceptedValue.Get()),
		p.Id,
		fmt.Sprintf(
			"%d.%d",
			prepareAckMessage.AcceptedProposalNumber.RoundNumber.Get(),
			prepareAckMessage.AcceptedProposalNumber.ServerId.Get(),
		),
	)
}

func (p *Peer) SendAccept() {
	acceptMessage := types.AcceptMessage{
		ProposalNumber: &types.ProposalNumber{
			RoundNumber: datastructures.NewSafeValue(p.Store.RoundNumber.Get()),
			ServerId:    datastructures.NewSafeValue(p.Id),
		},
		ProposalValue: p.ProposalValue,
	}
	data := types.Serialize(
		int(types.ACCEPT),
		acceptMessage.ProposalNumber.RoundNumber.Get(),
		acceptMessage.ProposalNumber.ServerId.Get(),
		int(acceptMessage.ProposalValue.Get()),
	)
	for _, acceptor := range p.Acceptors.GetAll() {
		p.SendMessageToPeer(acceptor, data)
		p.LogMessage(
			"sent",
			"accept",
			rune(acceptMessage.ProposalValue.Get()),
			p.Id,
			fmt.Sprintf(
				"%d.%d",
				acceptMessage.ProposalNumber.RoundNumber.Get(),
				acceptMessage.ProposalNumber.ServerId.Get(),
			),
		)
	}
}

func (p *Peer) SendAcceptAck(sender string) {
	roundNumber, serverId := utils.SplitN(p.Store.MinProposalNumber.Get())
	acceptAckMessage := types.AcceptAckMessage{
		ProposalNumber: &types.ProposalNumber{
			RoundNumber: datastructures.NewSafeValue(int(roundNumber)),
			ServerId:    datastructures.NewSafeValue(int(serverId)),
		},
	}
	data := types.Serialize(
		int(types.ACCEPT_ACK),
		acceptAckMessage.ProposalNumber.RoundNumber.Get(),
		acceptAckMessage.ProposalNumber.ServerId.Get(),
	)
	p.SendMessageToPeer(sender, data)
	p.LogMessage(
		"sent",
		"accept_ack",
		p.Store.AcceptedValue.Get(),
		p.Id,
		fmt.Sprintf(
			"%d.%d",
			acceptAckMessage.ProposalNumber.RoundNumber.Get(),
			acceptAckMessage.ProposalNumber.ServerId.Get(),
		),
	)
}

func (p *Peer) SendMessageToPeer(peer string, data []byte) {
	addr, err := utils.GetAddrFromHostname(peer)
	if err != nil {
		fmt.Printf("error resolving address for host %s: %v\n", peer, err)
		return
	}
	p.WriteChannel <- types.OutboundMessage{
		Data:      data,
		Recipient: addr,
	}
}

func (p *Peer) ListenForTCPConnections() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", tcpPort))
	if err != nil {
		fmt.Println("Error starting TCP listener:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		p.TCPIngress.Add(conn.RemoteAddr(), conn)
		go p.HandleTCPConnection(conn)
	}
}

func (p *Peer) HandleTCPConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from TCP connection:", err)
			}
			break
		}
		p.ReadChannel <- types.InboundMessage{
			Data:   buffer[:n],
			Sender: conn.RemoteAddr(),
		}
	}
}

func NewPeer(hostsFile string, proposalValue rune) (*Peer, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}

	peers, err := utils.GetPeers(hostsFile)
	if err != nil {
		return nil, err
	}

	roles, err := utils.GetRoles(hostname, hostsFile)
	if err != nil {
		return nil, err
	}

	proposerId, err := utils.GetProposerId(hostname, hostsFile)
	if err != nil {
		proposerId = -1
	}

	acceptors, err := utils.GetAcceptorsForProposer(proposerId, hostsFile)
	if err != nil {
		return nil, err
	}

	id, err := utils.GetPeerIdFromName(hostname, peers)
	if err != nil {
		return nil, err
	}

	store := &PeerStore{
		MinProposalNumber:      datastructures.NewSafeValue(utils.GetN(0, int32(id))),
		AcceptedProposalNumber: datastructures.NewSafeValue(utils.GetN(0, int32(id))),
		AcceptedValue:          datastructures.NewSafeValue[rune](0),
		RoundNumber:            datastructures.NewSafeValue(0),
	}

	return &Peer{
		Id:            id,
		Roles:         datastructures.NewSafeList(roles),
		Acceptors:     datastructures.NewSafeList(acceptors),
		Peers:         datastructures.NewSafeList(peers),
		Store:         store,
		ProposerId:    proposerId,
		TCPIngress:    NewTCPConnectionPool(tcpPort, Incoming),
		TCPEgress:     NewTCPConnectionPool(tcpPort, Outgoing),
		ProposalValue: datastructures.NewSafeValue(proposalValue),
		QuorumSize:    datastructures.NewSafeValue(len(acceptors)),
		ReadChannel:   make(chan types.InboundMessage),
		WriteChannel:  make(chan types.OutboundMessage),
	}, nil
}

func (p *Peer) LogMessage(action, messageType string, messageValue rune, peerId int, proposalNumber string) {
	logMessage := fmt.Sprintf(
		`{"peer_id":%d, "action": "%s", "message_type":"%s", "message_value":"%c", "proposal_num":%s}`,
		peerId, action, messageType, messageValue, proposalNumber,
	)
	utils.PrintToStderr(logMessage)
}
