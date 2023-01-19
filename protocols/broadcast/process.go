package broadcast

import (
	"fmt"
	"github.com/asynkron/protoactor-go/actor"
	"math"
	"stochastic-checking-simulation/config"
	"stochastic-checking-simulation/messages"
	"stochastic-checking-simulation/utils"
)

type Stage int32
type ValueType int32

var MessagesForEcho = int(math.Ceil(float64(config.ProcessCount + config.FaultyProcesses + 1) / float64(2)))
const MessagesForReady = config.FaultyProcesses + 1
const MessagesForAccept = 2 * config.FaultyProcesses + 1

const (
	Init Stage = iota
	SentEcho
	SentReady
)

type MessageState struct {
	receivedEcho  map[string]bool
	receivedReady map[string]bool
	echoCount     map[ValueType]int
	readyCount    map[ValueType]int
	stage         Stage
}

func NewMessageState() *MessageState {
	ms := new(MessageState)
	ms.receivedEcho = make(map[string]bool)
	ms.receivedReady = make(map[string]bool)
	ms.echoCount = make(map[ValueType]int)
	ms.readyCount = make(map[ValueType]int)
	ms.stage = Init
	return ms
}

type Process struct {
	currPid          *actor.PID
	pids             map[string]*actor.PID
	msgCounter       int32
	acceptedMessages map[string]map[int32]ValueType
	messagesLog      map[string]map[int32]*MessageState
}

func (p *Process) InitProcess(currPid *actor.PID, pids []*actor.PID) {
	p.currPid = currPid
	p.pids = make(map[string]*actor.PID)
	p.msgCounter = 0
	p.acceptedMessages = make(map[string]map[int32]ValueType)
	p.messagesLog = make(map[string]map[int32]*MessageState)
	for _, pid := range pids {
		id := utils.PidToString(pid)
		p.pids[id] = pid
		p.acceptedMessages[id] = make(map[int32]ValueType)
		p.messagesLog[id] = make(map[int32]*MessageState)
	}
}

func (p *Process) broadcast(context actor.SenderContext, message *messages.Message) {
	for _, pid := range p.pids {
		if pid != p.currPid {
			context.RequestWithCustomSender(pid, message, p.currPid)
		}
	}
}

func (p *Process) broadcastEcho(
	context actor.SenderContext,
	initialMessage *messages.Message,
	msgState *MessageState) {
	p.broadcast(
		context,
		&messages.Message{
			Stage:     messages.Message_ECHO,
			Value:     initialMessage.Value,
			SeqNumber: initialMessage.SeqNumber,
			Author:    initialMessage.Author,
		})
	msgState.stage = SentEcho
	msgState.echoCount[ValueType(initialMessage.Value)]++
}

func (p *Process) broadcastReady(
	context actor.SenderContext,
	initialMessage *messages.Message,
	msgState *MessageState) {
	p.broadcast(
		context,
		&messages.Message{
			Stage:     messages.Message_READY,
			Value:     initialMessage.Value,
			SeqNumber: initialMessage.SeqNumber,
			Author:    initialMessage.Author,
		})
	msgState.stage = SentReady
	msgState.readyCount[ValueType(initialMessage.Value)]++
	p.checkForAccept(initialMessage, msgState)
}

func (p *Process) checkForAccept(msg *messages.Message, msgState *MessageState) {
	value := ValueType(msg.Value)
	if msgState.readyCount[value] >= MessagesForAccept {
		p.acceptedMessages[msg.Author][msg.SeqNumber] = value
		delete(p.messagesLog[msg.Author], msg.SeqNumber)

		fmt.Printf(
			"%s accepted value %d with seq number %d\n",
			utils.PidToString(p.currPid), value, msg.SeqNumber)
	}
}

func (p *Process) Receive(context actor.Context) {
	msg, ok := context.Message().(*messages.Message)
	if !ok {
		return
	}
	sender := utils.PidToString(context.Sender())
	value := ValueType(msg.Value)

	acceptedValue, accepted := p.acceptedMessages[msg.Author][msg.SeqNumber]
	if accepted {
		if acceptedValue != value {
			fmt.Printf("%s: Detected a duplicated seq number attack\n", p.currPid.Id)
		}
		return
	}

	msgState := p.messagesLog[msg.Author][msg.SeqNumber]
	if msgState == nil {
		msgState = NewMessageState()
		p.messagesLog[msg.Author][msg.SeqNumber] = msgState
	}

	switch msg.Stage {
	case messages.Message_INITIAL:
		if msgState.stage == Init {
			p.broadcastEcho(context, msg, msgState)
		}
	case messages.Message_ECHO:
		if msgState.stage == SentReady || msgState.receivedEcho[sender] {
			return
		}
		msgState.receivedEcho[sender] = true
		msgState.echoCount[value]++

		if msgState.echoCount[value] >= MessagesForEcho {
			if msgState.stage == Init {
				p.broadcastEcho(context, msg, msgState)
			}
			if msgState.stage == SentEcho {
				p.broadcastReady(context, msg, msgState)
			}
		}
	case messages.Message_READY:
		if msgState.receivedReady[sender] {
			return
		}
		msgState.receivedReady[sender] = true
		msgState.readyCount[value]++

		if msgState.readyCount[value] >= MessagesForReady {
			if msgState.stage == Init {
				p.broadcastEcho(context, msg, msgState)
			}
			if msgState.stage == SentEcho {
				p.broadcastReady(context, msg, msgState)
			}
		}
		p.checkForAccept(msg, msgState)
	}
}

func (p *Process) Broadcast(context actor.SenderContext, value int32) {
	id := utils.PidToString(p.currPid)
	message := &messages.Message{
		Stage:     messages.Message_INITIAL,
		Author:    id,
		SeqNumber: p.msgCounter,
		Value:     value,
	}
	p.broadcast(context, message)

	msgState := NewMessageState()
	p.messagesLog[id][p.msgCounter] = msgState
	p.broadcastEcho(context, message, msgState)

	p.msgCounter++
}
