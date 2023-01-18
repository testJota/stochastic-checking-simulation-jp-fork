package consistent

import (
	"github.com/asynkron/protoactor-go/actor"
	"stochastic-checking-simulation/messages"
	"stochastic-checking-simulation/utils"
)

type FaultyProcess struct {
	process *CorrectProcess
}

func (p *FaultyProcess) InitFaultyProcess(currPid *actor.PID, pids []*actor.PID) {
	p.process = &CorrectProcess{}
	p.process.InitCorrectProcess(currPid, pids)
}

func (p *FaultyProcess) Receive(context actor.Context) {
	msg, ok := context.Message().(*messages.ProtocolMessage)
	if !ok {
		return
	}
	senderId := context.Sender()

	switch msg.Stage {
	case messages.ProtocolMessage_ECHO:
		p.process.verify(context, utils.PidToString(senderId), msg)
	case messages.ProtocolMessage_VERIFY:
		if msg.Author == utils.PidToString(p.process.currPid) {
			context.RequestWithCustomSender(
				senderId,
				messages.ProtocolMessage{
					Stage: messages.ProtocolMessage_ECHO,
					Author: msg.Author,
					SeqNumber: msg.SeqNumber,
					Value: msg.Value,
				},
				p.process.currPid)
		} else if p.process.verify(context, utils.PidToString(senderId), msg) {
				p.process.broadcast(
					context,
					&messages.ProtocolMessage{
						Stage: messages.ProtocolMessage_ECHO,
						Author: msg.Author,
						SeqNumber: msg.SeqNumber,
						Value: msg.Value,
					})
		}
	}
}

func (p *FaultyProcess) Broadcast(context actor.SenderContext, value1 int32, value2 int32) {
	author := utils.PidToString(p.process.currPid)
	seqNumber := p.process.msgCounter
	p.process.msgCounter++

	msgState := NewMessageState()
	msgState.witnessSet =
		p.process.wSelector.GetWitnessSet(
			author,
			seqNumber,
			p.process.historyHash,
		)
	p.process.messagesLog[author][seqNumber] = msgState

	i := 0
	for witness := range msgState.witnessSet {
		if i == len(msgState.witnessSet) / 2 {
			break
		}
		context.RequestWithCustomSender(
			p.process.pids[witness],
			&messages.ProtocolMessage{
				Stage:     messages.ProtocolMessage_VERIFY,
				Author:    author,
				SeqNumber: seqNumber,
				Value:     value1,
			},
			p.process.currPid)
		i++
	}
	for witness := range msgState.witnessSet {
		context.RequestWithCustomSender(
			p.process.pids[witness],
			&messages.ProtocolMessage{
				Stage:     messages.ProtocolMessage_VERIFY,
				Author:    author,
				SeqNumber: seqNumber,
				Value:     value2,
			},
			p.process.currPid)
	}
}
