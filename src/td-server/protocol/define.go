package protocol

const TYPE = 0x00
const VERSION = 1

const (
	//request
	USER_REQ_SEND_TOKEN = 0x0000 // GServer request DServer
	USER_REQ_SEND       = 0x0001 //client request DServer

	//response
	USER_RES_SEND_TOKEN  = 0x0100 //DServer response GServer
	USER_RES_SEND_STATUS = 0x0101 //DServer response client
)

//in ctrl_file_trans.go for response structure ConnMesthod type
const (
	TCP = iota
	UDP
)
