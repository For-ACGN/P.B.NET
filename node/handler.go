package node

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/davecgh/go-spew/spew"

	"project/internal/convert"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/messages"
	"project/internal/patch/msgpack"
	"project/internal/protocol"
	"project/internal/xpanic"
)

type handler struct {
	ctx *Node

	context context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func newHandler(ctx *Node) *handler {
	h := handler{
		ctx: ctx,
	}
	h.context, h.cancel = context.WithCancel(context.Background())
	return &h
}

func (h *handler) Cancel() {
	h.cancel()
}

func (h *handler) Close() {
	h.wg.Wait()
	h.ctx = nil
}

func (h *handler) logf(lv logger.Level, format string, log ...interface{}) {
	h.ctx.logger.Printf(lv, "handler", format, log...)
}

func (h *handler) log(lv logger.Level, log ...interface{}) {
	h.ctx.logger.Println(lv, "handler", log...)
}

// logPanic must use like defer h.logPanic("title")
func (h *handler) logPanic(title string) {
	if r := recover(); r != nil {
		h.log(logger.Fatal, xpanic.Print(r, title))
	}
}

// logfWithInfo will print log with role GUID and message
// [2020-01-30 15:13:07] [info] <handler> foo logf
// spew output...
//
// first log interface must be *protocol.Send or *protocol.Broadcast

// func (h *handler) logfWithInfo(lv logger.Level, format string, log ...interface{}) {
// 	buf := new(bytes.Buffer)
// 	_, _ = fmt.Fprintf(buf, format, log[1:]...)
// 	buf.WriteString("\n")
// 	spew.Fdump(buf, log[0])
// 	h.ctx.logger.Print(lv, "handler", buf)
// }

// logWithInfo will print log with role GUID and message
// [2020-01-30 15:13:07] [info] <handler> foo log
// spew output...
//
// first log interface must be *protocol.Send or *protocol.Broadcast
func (h *handler) logWithInfo(lv logger.Level, log ...interface{}) {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintln(buf, log[1:]...)
	spew.Fdump(buf, log[0])
	h.ctx.logger.Print(lv, "handler", buf)
}

// ------------------------------------------send--------------------------------------------------

func (h *handler) OnSend(send *protocol.Send) {
	defer h.logPanic("handler.OnSend")
	if len(send.Message) < messages.HeaderSize {
		h.logWithInfo(logger.Exploit, send, "send with invalid size")
		return
	}
	typ := send.Message[messages.RandomDataSize:messages.HeaderSize]
	msgType := convert.BEBytesToUint32(typ)
	send.Message = send.Message[messages.HeaderSize:]
	switch msgType {
	case messages.CMDCtrlAnswerNodeKey:
		h.handleAnswerNodeKey(send)
	case messages.CMDCtrlAnswerBeaconKey:
		h.handleAnswerBeaconKey(send)
	case messages.CMDCtrlNodeRegisterResponse:
		h.handleNodeRegisterResponse(send)
	case messages.CMDCtrlBeaconRegisterResponse:
		h.handleBeaconRegisterResponse(send)
	case messages.CMDCtrlNodeNop:
		h.handleNopCommand()
	case messages.CMDTest:
		h.handleSendTestMessage(send)
	case messages.CMDRTTestRequest:
		h.handleSendTestRequest(send)
	case messages.CMDRTTestResponse:
		h.handleSendTestResponse(send)
	default:
		const format = "send unknown message\ntype: 0x%08X\n%s"
		h.logf(logger.Exploit, format, msgType, spew.Sdump(send))
	}
}

func (h *handler) handleAnswerNodeKey(send *protocol.Send) {
	defer h.logPanic("handler.handleAnswerNodeKey")
	ank := messages.AnswerNodeKey{}
	err := msgpack.Unmarshal(send.Message, &ank)
	if err != nil {
		const log = "send invalid answer node key data"
		h.logWithInfo(logger.Exploit, send, log)
		return
	}
	err = ank.Validate()
	if err != nil {
		const log = "send invalid answer node key"
		h.logWithInfo(logger.Exploit, &ank, log)
		return
	}
	h.ctx.messageMgr.HandleReply(&ank.ID, &ank)
}

func (h *handler) handleAnswerBeaconKey(send *protocol.Send) {
	defer h.logPanic("handler.handleAnswerBeaconKey")
	abk := messages.AnswerBeaconKey{}
	err := msgpack.Unmarshal(send.Message, &abk)
	if err != nil {
		const log = "send invalid answer beacon key data\nerror:"
		h.logWithInfo(logger.Exploit, send, log, err)
		return
	}
	err = abk.Validate()
	if err != nil {
		const log = "send invalid answer beacon key\nerror:"
		h.logWithInfo(logger.Exploit, &abk, log, err)
		return
	}
	h.ctx.messageMgr.HandleReply(&abk.ID, &abk)
}

func (h *handler) handleNodeRegisterResponse(send *protocol.Send) {
	defer h.logPanic("handler.handleNodeRegisterResponse")
	nrr := messages.NodeRegisterResponse{}
	err := msgpack.Unmarshal(send.Message, &nrr)
	if err != nil {
		const log = "send invalid node register response data\nerror:"
		h.logWithInfo(logger.Exploit, send, log, err)
		return
	}
	err = nrr.Validate()
	if err != nil {
		const log = "send invalid node register response\nerror:"
		h.logWithInfo(logger.Exploit, &nrr, log, err)
		return
	}
	h.ctx.storage.AddNodeKey(&nrr.GUID, &protocol.NodeKey{
		PublicKey:    nrr.PublicKey,
		KexPublicKey: nrr.KexPublicKey,
		ReplyTime:    nrr.ReplyTime,
	})
	h.ctx.messageMgr.HandleReply(&nrr.ID, &nrr)
}

func (h *handler) handleBeaconRegisterResponse(send *protocol.Send) {
	defer h.logPanic("handler.handleBeaconRegisterResponse")
	brr := messages.BeaconRegisterResponse{}
	err := msgpack.Unmarshal(send.Message, &brr)
	if err != nil {
		const log = "send invalid beacon register response data"
		h.logWithInfo(logger.Exploit, send, log)
		return
	}
	err = brr.Validate()
	if err != nil {
		const log = "send invalid beacon register response"
		h.logWithInfo(logger.Exploit, &brr, log)
		return
	}
	h.ctx.storage.AddBeaconKey(&brr.GUID, &protocol.BeaconKey{
		PublicKey:    brr.PublicKey,
		KexPublicKey: brr.KexPublicKey,
		ReplyTime:    brr.ReplyTime,
	})
	h.ctx.messageMgr.HandleReply(&brr.ID, &brr)
}

// check execute number for prevent attack.
func (h *handler) handleNopCommand() {

}

// -----------------------------------------send test----------------------------------------------

func (h *handler) handleSendTestMessage(send *protocol.Send) {
	defer h.logPanic("handler.handleSendTestMessage")
	err := h.ctx.Test.AddSendMessage(h.context, send.Message)
	if err != nil {
		const log = "failed to add send test message\nerror:"
		h.logWithInfo(logger.Fatal, send, log, err)
	}
}

func (h *handler) handleSendTestRequest(send *protocol.Send) {
	defer h.logPanic("handler.handleSendTestRequest")
	request := messages.TestRequest{}
	err := msgpack.Unmarshal(send.Message, &request)
	if err != nil {
		const log = "invalid test request data\nerror:"
		h.logWithInfo(logger.Exploit, send, log, err)
		return
	}
	// send response
	response := messages.TestResponse{
		ID:       request.ID,
		Response: request.Request,
	}
	err = h.ctx.sender.Send(h.context, messages.CMDBRTTestResponse, &response, true)
	if err != nil {
		const log = "failed to send test response\nerror:"
		h.logWithInfo(logger.Exploit, send, log, err)
	}
}

func (h *handler) handleSendTestResponse(send *protocol.Send) {
	defer h.logPanic("handler.handleSendTestResponse")
	response := messages.TestResponse{}
	err := msgpack.Unmarshal(send.Message, &response)
	if err != nil {
		const log = "invalid test response data\nerror:"
		h.logWithInfo(logger.Exploit, send, log, err)
		return
	}
	h.ctx.messageMgr.HandleReply(&response.ID, &response)
}

// ----------------------------------------broadcast-------------------------------------------------

func (h *handler) OnBroadcast(broadcast *protocol.Broadcast) {
	defer h.logPanic("handler.OnBroadcast")
	if len(broadcast.Message) < messages.HeaderSize {
		h.logWithInfo(logger.Exploit, broadcast, "broadcast with invalid size")
		return
	}
	typ := broadcast.Message[messages.RandomDataSize:messages.HeaderSize]
	msgType := convert.BEBytesToUint32(typ)
	broadcast.Message = broadcast.Message[messages.HeaderSize:]
	switch msgType {
	case messages.CMDCtrlDeleteNode:
		h.handleDeleteNode(broadcast)
	case messages.CMDCtrlDeleteBeacon:
		h.handleDeleteBeacon(broadcast)
	case messages.CMDTest:
		h.handleBroadcastTestMessage(broadcast)
	default:
		const format = "broadcast unknown message\ntype: 0x%08X\n%s"
		h.logf(logger.Exploit, format, msgType, spew.Sdump(broadcast))
	}
}

func (h *handler) handleDeleteNode(broadcast *protocol.Broadcast) {
	defer h.logPanic("handler.handleDeleteNode")
	if len(broadcast.Message) != guid.Size {
		h.logWithInfo(logger.Exploit, broadcast, "invalid guid size about delete node key")
		return
	}
	// delete Node's key
	nodeGUID := guid.GUID{}
	_ = nodeGUID.Write(broadcast.Message)
	h.ctx.storage.DeleteNodeKey(&nodeGUID)
	// close connection
	_ = h.ctx.CloseNodeConnByGUID(&nodeGUID)
}

func (h *handler) handleDeleteBeacon(broadcast *protocol.Broadcast) {
	defer h.logPanic("handler.handleDeleteBeacon")
	if len(broadcast.Message) != guid.Size {
		h.logWithInfo(logger.Exploit, broadcast, "invalid guid size about delete beacon key")
		return
	}
	// delete Beacon's key
	beaconGUID := guid.GUID{}
	_ = beaconGUID.Write(broadcast.Message)
	h.ctx.storage.DeleteBeaconKey(&beaconGUID)
	// close connection
	_ = h.ctx.CloseBeaconConnByGUID(&beaconGUID)
}

// ---------------------------------------broadcast test-------------------------------------------

func (h *handler) handleBroadcastTestMessage(broadcast *protocol.Broadcast) {
	defer h.logPanic("handler.handleBroadcastTestMessage")
	err := h.ctx.Test.AddBroadcastMessage(h.context, broadcast.Message)
	if err != nil {
		const log = "failed to add broadcast test message\nerror:"
		h.logWithInfo(logger.Fatal, broadcast, log, err)
	}
}
