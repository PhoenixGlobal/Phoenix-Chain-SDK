package network

import (
	ctpyes "github.com/PhoenixGlobal/Phoenix-Chain-SDK/consensus/pbft/types"
)

// SetSendQueueHook
func (h *EngineManager) SetSendQueueHook(f func(*ctpyes.MsgPackage)) {
	h.sendQueueHook = f
}
