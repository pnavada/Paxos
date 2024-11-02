package utils

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"paxos/paxos/types"
)

func GetPeers(hostsfile string) ([]string, error) {
	content, err := os.ReadFile(hostsfile)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	var peers []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid format in hostsfile: %s", line)
		}
		peers = append(peers, parts[0])
	}
	return peers, nil
}

func GetProposerId(peer string, hostsfile string) (int, error) {
	content, err := os.ReadFile(hostsfile)
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, peer+":") {
			parts := strings.Split(line, ":")
			if len(parts) < 2 {
				return 0, fmt.Errorf("invalid format in hostsfile: %s", line)
			}
			roles := strings.Split(parts[1], ",")
			for _, role := range roles {
				if strings.HasPrefix(role, "proposer") {
					proposerIdStr := strings.TrimPrefix(role, "proposer")
					proposerId, err := strconv.Atoi(proposerIdStr)
					if err != nil {
						return 0, fmt.Errorf("invalid proposer id: %s", proposerIdStr)
					}
					return proposerId, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("proposer id not found for peer: %s", peer)
}

func GetRoles(peer string, hostsfile string) ([]types.Role, error) {
	content, err := os.ReadFile(hostsfile)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, peer+":") {
			parts := strings.Split(line, ":")
			if len(parts) < 2 {
				return nil, fmt.Errorf("invalid format in hostsfile: %s", line)
			}
			roleStrs := strings.Split(parts[1], ",")
			var roles []types.Role
			for _, roleStr := range roleStrs {
				switch {
				case strings.HasPrefix(roleStr, "proposer"):
					roles = append(roles, types.Proposer)
				case strings.HasPrefix(roleStr, "acceptor"):
					roles = append(roles, types.Acceptor)
				case strings.HasPrefix(roleStr, "learner"):
					roles = append(roles, types.Learner)
				default:
					return nil, fmt.Errorf("unknown role: %s", roleStr)
				}
			}
			return roles, nil
		}
	}
	return nil, fmt.Errorf("roles not found for peer: %s", peer)
}

func GetAcceptorsForProposer(proposerId int, hostsfile string) ([]string, error) {
	content, err := os.ReadFile(hostsfile)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	var acceptors []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid format in hostsfile: %s", line)
		}
		roles := strings.Split(parts[1], ",")
		for _, role := range roles {
			if strings.HasPrefix(role, "acceptor") {
				acceptorProposerIdStr := strings.TrimPrefix(role, "acceptor")
				acceptorProposerId, err := strconv.Atoi(acceptorProposerIdStr)
				if err != nil {
					return nil, fmt.Errorf("invalid acceptor proposer id: %s", acceptorProposerIdStr)
				}
				if acceptorProposerId == proposerId {
					acceptors = append(acceptors, parts[0])
					break
				}
			}
		}
	}
	return acceptors, nil
}

func GetPeerIdFromName(peer string, peers []string) (int, error) {
	for i, p := range peers {
		if p == peer {
			return i + 1, nil
		}
	}
	return 0, fmt.Errorf("peer not found: %s", peer)
}

func GetPeerNameFromId(id int, peers []string) (string, error) {
	if id <= 0 || id > len(peers) {
		return "", fmt.Errorf("invalid peer id: %d", id)
	}
	return peers[id-1], nil
}

func GetAddrFromHostname(hostname string) (net.Addr, error) {
	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("no addresses found for hostname: %s", hostname)
	}
	return &net.TCPAddr{IP: addrs[0]}, nil
}

func GetHostnameFromAddr(addr net.Addr) (string, error) {
	ip, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return "", err
	}

	names, err := net.LookupAddr(ip)
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", fmt.Errorf("no hostnames found for IP: %s", ip)
	}
	return names[0], nil
}

func CleanHostname(hostname string) string {
	if idx := strings.Index(hostname, "."); idx != -1 {
		return hostname[:idx]
	}
	return hostname
}

func RemoveSelf(peers []string, self string) ([]string, error) {
	var peersWithoutSelf []string
	for _, peer := range peers {
		if peer != self {
			peersWithoutSelf = append(peersWithoutSelf, peer)
		}
	}
	return peersWithoutSelf, nil
}

func GetN(high, low int32) int64 {
	result := int64(high) << 32
	result |= int64(uint32(low))
	return result
}

func SplitN(value int64) (high, low int32) {
	high = int32(value >> 32)
	low = int32(uint32(value))
	return high, low
}

func Length(m *sync.Map) int {
	count := 0
	m.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

func PrintToStderr(message string) {
	fmt.Fprintln(os.Stderr, message)
}
