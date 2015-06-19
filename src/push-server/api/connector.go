package api

import (
	"errors"
	"fmt"
	"net"
	"push-server/protocol"
)

const TAG = "Connector Error: "

const (
	USER_UUID_LEN = 32
	P_SN          = "000000000000000"
	P_SN_LEN      = 15
	P_TOKEN       = uint8(0)
	P_DEST_UUID   = "00000000000000000000000000000000"
)

type Connector struct {
	/*NIL*/
}

func NewConnector() *Connector {
	return &Connector{}
}

const MESSAGE_MAX_BUF = 1000

func (connector *Connector) Register(msg *protocol.Message, tcpConn *net.TCPConn) (token uint8, deviceId uint32, err error) {
	if msg == nil || tcpConn == nil {
		err = errors.New("Connector Register Error : invalid argument")
		fmt.Println(err.Error())
		return
	}
	err = connector.checkRegister(msg)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = connector.Publish(msg, tcpConn)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	response, err := connector.ReadMessage(tcpConn)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if response.MsgType == protocol.MT_ERROR && response.PushType == protocol.PT_PUSHSERVER &&
		response.DestUuid == msg.UserUuid {

		err = errors.New("Connector Register Error: Push-Server refuse the register request")
		return 0, 0, err
	}

	if response.MsgType != protocol.MT_REGISTER || response.PushType != protocol.PT_PUSHSERVER ||
		response.DestUuid != msg.UserUuid {

		err = errors.New("Connector Register Error: receive wrong response from Push-Server")
		return 0, 0, err
	}
	token = response.Token
	deviceId = response.DeviceId
	return token, deviceId, nil

}

func (connector *Connector) checkRegister(msg *protocol.Message) error {
	if msg == nil {
		err := errors.New("Connector checkRegister Error:invalid argument - Message is nil")
		return err
	}
	if msg.DataLen != P_SN_LEN || len(msg.Data) != P_SN_LEN || len(msg.UserUuid) != USER_UUID_LEN ||
		len(msg.DestUuid) != USER_UUID_LEN || msg.MsgType != protocol.MT_REGISTER || msg.PushType == protocol.PT_PUSHSERVER {
		err := errors.New("tConnector checkRegister Error:invalid argument - one of the Message's fields is wrong")
		return err
	}
	return nil
}

func (connector *Connector) Publish(msg *protocol.Message, tcpConn *net.TCPConn) (err error) {
	if msg == nil || tcpConn == nil {
		err = errors.New("Connector Publish Error: invalid argument - Message or TCPConn is nil")
		return err
	}
	err = connector.checkPublish(msg)
	if err != nil {
		return err
	}
	err = connector.SendMessage(msg, tcpConn)
	if err != nil {
		return err
	}
	return nil

}

func (connector *Connector) checkPublish(msg *protocol.Message) error {
	if msg == nil {
		err := errors.New("Connector CheckPublish Error: invalid argument - Message is nil")
		return err
	}
	if len(msg.UserUuid) != USER_UUID_LEN || len(msg.DestUuid) != USER_UUID_LEN || msg.PushType == protocol.PT_PUSHSERVER ||
		msg.DataLen != uint16(len(msg.Data)) {
		err := errors.New("Connector Publish Error: invalid argument - one of the Message's field is wrong")
		return err
	}
	return nil

}

func (connector Connector) SendMessage(msg *protocol.Message, tcpConn *net.TCPConn) error {
	if msg == nil || tcpConn == nil {
		err := errors.New("Connector SendMessage Error: invalid argument - Message or TCPConn is nil")
		return err
	}
	buffer, err := msg.Encode()
	if err != nil {
		err = fmt.Errorf("connector SendMessage Error: %v", err.Error())
		return err
	}
	index := 0
	for index < len(buffer) {
		n, err := tcpConn.Write(buffer[index:])
		index += n
		if index < len(buffer) || err != nil {
			err = errors.New("Connector SendMessage Error: the write byte count is lower than Message's length Or error occur in write")
			return err
		}
	}
	return nil
}

func (connector Connector) ReadMessage(tcpConn *net.TCPConn) (msg *protocol.Message, err error) {
	if tcpConn == nil {
		err = errors.New("Connector ReadMessage Error: invalid argument - TCPConn is nil")
		return err
	}
	buffer := make([]byte, MESSAGE_MAX_BUF)
	err = connector.ReceiveFromTCP(buffer[0:protocol.MESSAGE_LEN], tcpConn)
	if err != nil {
		return err
	}
	msg = &protocol.Message{}
	err = msg.Decode(buffer[0:protocol.MESSAGE_LEN])
	if err != nil {
		err = fmt.Errorf("Connector ReadMessage Error: %v", err.Error())
		return err
	}
	if msg.DataLen == 0 {
		return msg, nil
	}
	var data []byte
	if (msg.DataLen + protocol.MESSAGE_LEN) > MESSAGE_MAX_BUF {
		data = make([]byte, msg.DataLen)
	} else {
		data = buffer[protocol.MESSAGE_LEN : protocol.MESSAGE_LEN+msg.DataLen]
	}
	err = connector.ReceiveFromTCP(data, tcpConn)
	if err != nil {
		return err
	}
	msg.Data = data
	return msg, err
}

func (connector *Connector) ReceiveFromTCP(buffer []byte, tcpConn *net.TCPConn) error {
	if tcpConn == nil {
		err = errors.New("Connector ReceiveFromTCP Error: invalid argument - TCPConn is nil")
		return err
	}
	index := 0
	for index < len(buffer) {
		n, err := tcpConn.Read(buffer[index:len(buffer)])
		index += n
		if index < len(buffer) || err != nil {
			err = errors.New("Connector ReceiveFromTCP Error: the read byte is lower than Message's length Or error occur in read")
			return err
		}
	}
	return nil
}

func (connector *Connector) Unregister(msf *protocol.Message, tcpConn *net.TCPConn) (number uint16, err error) {
	if msg == nil || tcpConn == nil {
		err = errors.New("Connector Unregister Error: invalid argument - Message or TCPConn is nil")
		return err
	}
	err = connector.checkUnRegister(msg)
	if err != nil {
		fmt.Println(err.Error())
		return number, err
	}
	err = connector.SendMessage(msg, tcpConn)
	if err != nil {
		fmt.Println(err.Error())
		return number, err
	}
	response, err := connector.ReadMessage(tcpConn)
	if err != nil {
		fmt.Println(err.Error())
		return number, err
	}
	if response.MsgType == protocol.MT_ERROR && response.PushType == protocol.PT_PUSHSERVER &&
		response.DestUuid == msg.UserUuid {
		err = errors.New("Connector Unregister Error: Push-Server refuse the unregister request")
		return number, err
	}
	if response.MsgType != protocol.MT_UNREGISTER || response.PushType != protocol.PT_PUSHSERVER ||
		response.Number != msg.Number {
		err = errors.New("Connector Unregister Error: receive the wrong response")
		return number, err
	}
	number = response.Number
	return number, err
}

func (connector *Connector) checkUnRegister(msg *protocol.Message) error {
	if msg == nil {
		err := errors.New("Connector checkUnRegister Error: invalid argument - Message is nil")
		return err
	}
	if len(msg.UserUuid) != 32 || len(msg.DestUuid) != 32 || msg.MsgType != protocol.MT_UNREGISTER ||
		msg.PushType == protocol.PT_PUSHSERVER {
		err := errors.New("Connector checkUnRegister Error: one of the Message's fields is wrong")
		return err
	}
	return nil
}
