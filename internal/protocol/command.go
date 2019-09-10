package protocol

// --------------------------test-----------------------------
const (
	TestMessage uint8 = 0xEF
)

// -----------------------controller--------------------------
const (
	CtrlHeartbeat uint8 = 0x00 + iota
	CtrlReply
	CtrlSyncStart
	CtrlSyncQuery
	CtrlBroadcastToken
	CtrlBroadcast
	CtrlSyncSendToken
	CtrlSyncSend
	CtrlSyncRecvToken
	CtrlSyncRecv
)

// trust node
const (
	CtrlTrustNode uint8 = 0x20 + iota
	CtrlTrustNodeData
)

const (
	CtrlQueryNodeStatus uint8 = 0x30 + iota
	CtrlQueryAllNodes
)

// --------------------------node-----------------------------
const (
	NodeHeartbeat uint8 = 0x00 + iota
	NodeReply
	NodeSyncStart
	NodeSyncQuery
	NodeBroadcastToken
	NodeBroadcast
	NodeSyncSendToken
	NodeSyncSend
	NodeSyncRecvToken
	NodeSyncRecv
)

// node authentication
const (
	NodeQueryCertificate uint8 = 0x20 + iota
)

// query nodes
const (
	NodeQueryGUID uint8 = 0x30 + iota
	NodeQueryAllNodes
)
