# Paxos Consensus Implementation
## Distributed Systems Project 4

### Developer
Pruthvi Prakash Navada

## Project Overview
This project implements the Paxos consensus protocol in Go, allowing distributed nodes to reach agreement on proposed values. The implementation supports multiple proposers and acceptors, with configurable network topologies through host files.

## Requirements
- Go 1.21 or higher
- Docker
- Docker Compose

## Building and Running

### Initial Setup
```bash
# Build the Docker image
make docker
```

### Running Test Cases

#### Test Case 1: Single Proposer
This test case demonstrates basic Paxos consensus with one proposer and three acceptors.

**Configuration**
- 1 Proposer (peer1) proposing value 'X'
- 3 Acceptors (peer2, peer3, peer4)
- 1 Learner (peer5)

```bash
# Run test case 1
make run-test1
```

#### Expected Behavior
- peer1 will propose value 'X'
- Consensus should be reached quickly
- Final value 'X' should be chosen
- Messages will be logged showing the protocol progression

#### Test Case 2: Competing Proposers
This test case demonstrates conflict resolution between multiple proposers.

**Configuration**
- 2 Proposers:
  - peer1 (proposing value 'X')
  - peer5 (proposing value 'Y' after 10-second delay)
- 3 Acceptors (peer2, peer3, peer4) participating in both proposer groups

```bash
# Run test case 2
make run-test2
```

#### Expected Behavior
- peer1 will propose value 'X' immediately
- peer5 will propose value 'Y' after 10 seconds (i.e, after peer3 sends accept to peer1)
- The protocol should handle the conflict and reach consensus
- Final value 'X' should be chosen by both proposers

### Cleanup
```bash
# Stop running containers
make stop-test

# Clean Docker images and system
make docker-clean
```

## Implementation Details

### Network Configuration
- TCP communication on port 8080
- Docker network for container communication
- Hostname-based peer discovery

### Key Features
- Thread-safe data structures
- Asynchronous message handling
- Connection pooling
- Message serialization

### Host File Format
```
hostname:role1[,role2,...]
```
Roles can be:
- proposer[N] - Proposer with ID N
- acceptor[N] - Acceptor for proposer group N
- learner[N] - Learner for proposer group N

## Command Line Arguments
- `-h string`: Path to hosts file (required)
- `-v string`: Proposal value for proposer
- `-t int`: Delay in seconds before proposing (optional)

## Monitoring
The implementation logs JSON messages to stderr in the format:
```json
{
    "peer_id": int,
    "action": string,
    "message_type": string,
    "message_value": string,
    "proposal_num": string
}
```