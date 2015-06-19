package protocol

const (
	VERSION = 1
)
const (
	AUDIO = 0x00
	FILE  = 0x01
	VIDEO = 0x02
	PUSH  = 0x03
)

const (
	REQ_TOKEN = 0x00
	RES_TOKEN = 0x10
)

//Text Number
const (
	TEXT_CTOS_REQ               = 0x3101 // client send text request to server
	TEXT_CTOS_RESP_RECEIVE      = 0x3102 // client send receive text response to server
	TEXT_CTOS_RESP_READ         = 0x3103 // client send read text response to server
	TEXT_CTOS_RESP_RECEIVE_FAIL = 0x3103 // client send fail to receive response to server
	TEXT_CTOS_RESP_READ_FAIL    = 0x3104 // client send fail to read response to server

	TEXT_STOC_REQ_RESP_SUCCESS = 0x3110 // server send the success response to client
	TEXT_STOC_REQ_RESP_FAIL    = 0x3111 // server send the fail response to client

	TEXT_STOC_REQ      = 0x3112 // server send text request from clientA to clientB
	TEXT_STOC_REQ_FAIL = 0x3113 // server send to clientA that fail to send request to clientB

	TEXT_STOC_RESP_RECEIVE = 0x3114 // server send the response which clientB receive text to clientA
	TEXT_STOC_RESP_READ    = 0x3115 // server send the response thich clientB read text to clientA

	TEXT_STOC_RESP_RECEIVE_FAIL = 0x3116 // server send the response which clientB fail to receive text to clientA
	TEXT_STOC_RESP_READ_FAIL    = 0x3117 // server send the response which clientB fail to read text to clientA
)

const (
	PHOTO_CTOS_REQ               = 0x4101 // client send photo request to server
	PHOTO_CTOS_RESP_RECEIVE      = 0x4102 // client send receive photo response to server
	PHOTO_CTOS_RESP_READ         = 0x4103 // client send read photo response to server
	PHOTO_CTOS_RESP_RECEIVE_FAIL = 0x4104 // client send fail to receive response to server
	PHOTO_CTOS_RESP_READ_FAIL    = 0x4105 // client send fail to read response to server

	PHOTO_STOC_REQ_RESP_SUCCESS = 0x4110 // server send the success response to client
	PHOTO_STOC_REQ_RESP_FAIL    = 0x4111 // server send the fail response to client

	PHOTO_STOC_REQ      = 0x4112 // server send photo request from clientA to clientB
	PHOTO_STOC_REQ_FAIL = 0x4113 // server send to clientA that fail to send request to clientB

	PHOTO_STOC_RESP_RECEIVE = 0x4114 // server send the response which clientB receive photo to clientB
	PHOTO_STOC_RESP_READ    = 0x4115 // server send the response which clientB read photo to clientB

	PHOTO_STOC_RESP_RECEIVE_FAIL = 0x4116
	PHOTO_STOC_RESP_READ_FAIL    = 0x4117
)

//PUSH CONST
const (
	TEXT           = 0x5100
	AUDIO_FRAGMENT = 0x5101
	PHOTO          = 0x5102
	VIDEO_FRAGMENT = 0x5103
	SHAKE          = 0x5104
)
