package messages

import "fmt"

func (b *BroadcastInstance) ToString() string {
	return fmt.Sprintf("{%d;%d}", b.Author, b.SeqNumber)
}

func (m *Started) ToString() string {
	return fmt.Sprintf("Started{Sender: %d}", m.Sender)
}

func (b *BroadcastInstance) Copy() *BroadcastInstance {
	if b == nil {
		return nil
	}
	return &BroadcastInstance{
		Author:    b.Author,
		SeqNumber: b.SeqNumber,
	}
}

func (m *BrachaProtocolMessage) Copy() *BrachaProtocolMessage {
	if m == nil {
		return nil
	}
	return &BrachaProtocolMessage{
		Stage: m.Stage,
		Value: m.Value,
	}
}

func (m *ConsistentProtocolMessage) Copy() *ConsistentProtocolMessage {
	if m == nil {
		return nil
	}
	return &ConsistentProtocolMessage{
		Stage: m.Stage,
		Value: m.Value,
	}
}

func (m *ReliableProtocolMessage) Copy() *ReliableProtocolMessage {
	if m == nil {
		return nil
	}
	return &ReliableProtocolMessage{
		Stage: m.Stage,
		Value: m.Value,
	}
}

func (m *RecoveryProtocolMessage) Copy() *RecoveryProtocolMessage {
	if m == nil {
		return nil
	}
	return &RecoveryProtocolMessage{
		Stage:                   m.Stage,
		ReliableProtocolMessage: m.ReliableProtocolMessage.Copy(),
	}
}

func (m *ScalableProtocolMessage) Copy() *ScalableProtocolMessage {
	if m == nil {
		return nil
	}
	return &ScalableProtocolMessage{
		Stage: m.Stage,
		Value: m.Value,
	}
}
