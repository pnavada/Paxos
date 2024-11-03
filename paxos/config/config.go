package config

import (
	"flag"
)

type Config struct {
	HostsFile     string
	ProposalValue rune
	ProposalDelay int
}

func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.HostsFile, "h", "", "Path to the hosts file")
	var proposalValue string
	flag.StringVar(&proposalValue, "v", "", "This is the value used if the peer is a proposer")
	flag.IntVar(&cfg.ProposalDelay, "t", 0, "This is the time in seconds the peer will wait before starting its proposal with its value v")

	flag.Parse()

	if len(proposalValue) > 0 {
		cfg.ProposalValue = rune(proposalValue[0])
	} else {
		cfg.ProposalValue = 0
	}

	if cfg.HostsFile == "" {
		flag.Usage()
		return nil
	}

	return cfg
}
