package protocol

import (
	"time"
)

const (
	VERSION = 0x01
)

const (
	MT_REGISTER           = 0x10
	MT_UNREGISTER         = 0x20
	MT_PUSH               = 0x30 // client,service contact to push server
	MT_PUBLISH            = 0x40 // client,service contact to each other by push server
	MT_BROADCAST_ONLINE   = 0x50 //	push server send message to service if someone is online
	MT_BROADCAST_OFFLINE  = 0x60 //	push server send message to service if someone is offline
	MT_PUSH_WITH_RESPONSE = 0x70 // push server send message to client and wait the response from client
	MT_KEEP_ALIVE         = 0x80

	MT_QUERY_USER          = 0x90
	MT_QUERY_USER_RESPONSE = 0x91

	MT_ALLOC_CACHE_SERVER_ID          = 0xa0
	MT_ALLOC_CACHE_SERVER_ID_RESPONSE = 0xa1

	MT_ADD_CACHE_ENTRY = 0xc0
	MT_DEL_CACHE_ENTRY = 0xc1

	MT_ERROR = 0xff
)

const (
	//5 minutes for keep alive
	KEEP_ALIVE_PERIOD = 900000 * time.Second
)

const (
	PT_PUSHSERVER = 0x10
	PT_SERVICE    = 0x20
	PT_CLIENT     = 0x30
)

const (
	N_PUSHSYS     = 0x0000
	N_ACCOUNT_SYS = 0x0001
	N_QSERVER     = 0x0002
)

const (
	TOKEN_MIN = 0x01
	TOKEN_MAX = 0x02
	TOKEN_ERR = 0x00
)

const (
	NO_FEEDBACK_DISCARD = 0x00
	NO_FEEDBACK_CACHE   = 0x01
	FEEDBACK_DISCARD    = 0x10
	FEEDBACK_DATARETURN = 0x11
)
