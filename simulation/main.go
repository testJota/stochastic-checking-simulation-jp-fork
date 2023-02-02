package main

import (
	"flag"
	"fmt"
	console "github.com/asynkron/goconsole"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/remote"
	"net"
	"stochastic-checking-simulation/config"
	"stochastic-checking-simulation/impl/messages"
	"stochastic-checking-simulation/impl/protocols"
	"stochastic-checking-simulation/impl/protocols/accountability/consistent"
	"stochastic-checking-simulation/impl/protocols/accountability/reliable"
	"stochastic-checking-simulation/impl/protocols/bracha"
	"stochastic-checking-simulation/impl/utils"
	"strconv"
	"strings"
)

var (
	nodesStr = flag.String(
		"nodes", "",
		"a string representing bindings host:port separated by comma, e.g. 127.0.0.1:8081,127.0.0.1:8082")
	address        = flag.String("address", "", "current node's address, e.g. 127.0.0.1:8081")
	mainServerAddr = flag.String("mainserver", "", "address of the main server, e.g. 127.0.0.1:8080")
	protocol       = flag.String("protocol", "reliable_accountability",
		"A protocol to run, one of: reliable_accountability, consistent_accountability, bracha")
	faultyProcesses = flag.Int("f", 0, "max number of faulty processes in the system")
	ownWitnessSetSize = flag.Int("w", 0, "size of the own witness set W")
	potWitnessSetSize = flag.Int("v", 0, "size of the pot witness set V")
	witnessThreshold = flag.Int("u", 0, "witnesses threshold to accept a transaction")
	nodeIdSize = flag.Int("node_id_size", 256, "node id size, default is 256")
	numberOfBins = flag.Int("number_of_bins", 32, "number of bins in history hash, default is 32")
)

func main() {
	flag.Parse()

	nodes := strings.Split(*nodesStr, ",")
	host, portStr, e := net.SplitHostPort(*address)
	if e != nil {
		fmt.Printf("Could not split %s into host and port\n", *address)
		return
	}

	port, e := strconv.Atoi(portStr)
	if e != nil {
		fmt.Printf("Could not convert port string representation into int: %s\n", e)
		return
	}

	config.ProcessCount = len(nodes)
	config.FaultyProcesses = *faultyProcesses
	config.OwnWitnessSetSize = *ownWitnessSetSize
	config.PotWitnessSetSize = *potWitnessSetSize
	config.WitnessThreshold = *witnessThreshold
	config.NodeIdSize = *nodeIdSize
	config.NumberOfBins = *numberOfBins

	pids := make([]*actor.PID, len(nodes))
	for i := 0; i < len(nodes); i++ {
		pids[i] = actor.NewPID(nodes[i], "pid")
	}

	mainServer := actor.NewPID(*mainServerAddr,"mainserver")

	system := actor.NewActorSystem()
	remoteConfig := remote.Configure(host, port)
	remoter := remote.NewRemote(system, remoteConfig)
	remoter.Start()

	var process protocols.Process

	if *protocol == "reliable_accountability" {
		process = &reliable.Process{}
	} else if *protocol == "consistent_accountability" {
		process = &consistent.CorrectProcess{}
	} else {
		process = &bracha.Process{}
	}

	currPid, e :=
		system.Root.SpawnNamed(
			actor.PropsFromProducer(
				func() actor.Actor {
					return process
				}),
			"pid",
		)
	if e != nil {
		fmt.Printf("Error while generating pid happened: %s\n", e)
		return
	}
	process.InitProcess(currPid, pids)
	fmt.Printf("%s: started\n", utils.PidToString(currPid))
	system.Root.RequestWithCustomSender(mainServer, &messages.Started{}, currPid)

	_, _ = console.ReadLine()
}
