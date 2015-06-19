package qserver

import (
	"connect-server/model"
	core "connect-server/protocol"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"td-server/protocol"
)

type TCPHandler struct {
	userCache  *model.UserCache
	Conn       *net.Conn
	ReadLength int
	BuffLength int64
	Error      error
	User       *model.User
	TokenIP    string
	TokenPort  int
}

func NewTCPHandler(userCache *model.UserCache, conn *net.Conn, TokenIP string, TokenPort int) *TCPHandler {
	return &TCPHandler{
		userCache: userCache,
		Conn:      conn,
		TokenIP:   TokenIP,
		TokenPort: TokenPort,
	}
}

func (handler *TCPHandler) Handle() {
	defer func() {
		if handler.User != nil {
			handler.userCache.Remove(handler.User.Uuid)
		}
		(*handler.Conn).Close()
	}()

	lengthBuff := make([]byte, 8)
	var packetLen int64
	var number int = 0
	var err error
	var count int32 = 0

	for {
		number, err = (*handler.Conn).Read(lengthBuff)
		if number == 0 || err != nil {
			break
		}
		packetLen = int64(binary.BigEndian.Uint64(lengthBuff))
		if packetLen == 0 {
			requestBuff := make([]byte, 4096)
			number, err = (*handler.Conn).Read(requestBuff)
			if number == 0 || err != nil {
				fmt.Println("Can't read any data")
				return
			}
		}

		if packetLen > (1<<12) || packetLen <= 0 {
			fmt.Println("The packet is too long or error")
			break
		}

		count++
		requestBuff := make([]byte, packetLen-8)
		number, err = (*handler.Conn).Read(requestBuff)
		if number == 0 || err != nil {
			fmt.Println("Can't read any data")
			break
		}

		var n int
		var length int = len(requestBuff)
		offset := length - number
		for offset != 0 {
			buffer := make([]byte, offset)
			n, err = (*handler.Conn).Read(buffer)
			if n == 0 && err != nil {
				return
			}
			copy(requestBuff[number:number+n], buffer[:n])
			offset -= n
		}

		handler.Unpack(lengthBuff, requestBuff, handler.Conn, handler.TokenIP, handler.TokenPort)
	}
}

func (handler *TCPHandler) Unpack(lengthBuff, requestBuff []byte, Conn *net.Conn, TokenIP string, TokenPort int) {
	Type := requestBuff[1]
	buffer := make([]byte, 0)
	buffer = append(buffer, lengthBuff...)
	buffer = append(buffer, requestBuff...)
	Number := binary.BigEndian.Uint16(buffer[10:12])

	if handler.User == nil {
		if Type == core.CTRL && Number == core.CTRL_REQ_LOGIN {
			handler.User = Login(buffer, Conn, handler.User, handler.userCache)
			if handler.User != nil {
				LoginOK(Conn)
				return
			} else {
				LoginError(Conn)
				return
			}
		} else {
			fmt.Println("Not the login order")
			return
		}
	} else {
		switch int8(Type) {
		case core.CTRL:
		case core.AUDIO:
		case core.VIDEO:
		case core.FILE:
			FileHandle(buffer, Number, handler.User, handler.userCache, Conn, TokenIP, TokenPort)
		case core.PUSH:
			PushHandle(buffer, Number, handler.User, handler.userCache, Conn)
		default:
			fmt.Println("Receive unkown order")
		}
	}
}

func Login(buffer []byte, Conn *net.Conn, User *model.User, userCache *model.UserCache) *model.User {
	confirm := core.Confirm{}
	err := confirm.Decode(buffer)
	if err != nil {
		fmt.Printf("TCPHandle Login Error:%v", err.Error)
		return nil
	}
	if confirm.UserUuid != "" {
		remote := (*Conn).RemoteAddr().String()
		ips := strings.Split(remote, ":")
		ip := net.ParseIP(ips[0])
		User = &model.User{IP: ip, Port: int32(confirm.Port), Uuid: confirm.UserUuid, Conn: Conn}
	}
	if User == nil {
		return nil
	}
	userCache.Add(User)
	return User
}

func LoginOK(Conn *net.Conn) {
	response := &core.ConfirmResponse{
		core.QCHeadBase{
			Len:     13,
			Version: 1,
			Type:    core.CTRL,
			Number:  core.CTRL_RES_LOGIN,
		},
		1,
	}
	responseBuff, err := response.Encode()
	if err != nil {
		fmt.Printf("TCPHandle Error: %v", err.Error())
		return
	}
	(*Conn).Write(responseBuff)
}

func LoginError(Conn *net.Conn) {
	response := &core.ConfirmResponse{
		core.QCHeadBase{
			Len:     13,
			Version: 1,
			Type:    core.CTRL,
			Number:  core.CTRL_RES_LOGIN,
		},
		0,
	}
	responseBuff, err := response.Encode()
	if err != nil {
		fmt.Printf("TCPHandle Error: %v", err.Error())
		return
	}
	(*Conn).Write(responseBuff)
}

func FileHandle(buffer []byte, Number uint16, User *model.User, userCache *model.UserCache, Conn *net.Conn, TokenIP string, TokenPort int) {
	switch Number {
	case core.FILE_REQ_SEND_FILE:
		FileRequestHandle(buffer, Conn, userCache, User)
	case core.FILE_REQ_SEND_STATUS:
		FileResponseHandle(buffer, Conn, userCache, User, TokenIP, TokenPort)
	default:

	}
}

func FileRequestHandle(buffer []byte, Conn *net.Conn, userCache *model.UserCache, User *model.User) error {
	fileRequest := core.FileMsgRequest{}
	err := fileRequest.Decode(buffer)
	if err != nil {
		err = fmt.Errorf("FileRequestHandle Error: %v", err)
		fmt.Println(err.Error())
		return err
	}

	if fileRequest.Number != core.FILE_REQ_SEND_FILE {
		err = errors.New("Not the file request number")
		fmt.Println(err.Error())
		return err
	}

	var destUser *model.User
	//这里的DestID应该是请求接收方
	if destUser = userCache.Query(fileRequest.DestUuid); destUser == nil {
		err = fmt.Errorf("Don't have the DestUser : %v", fileRequest.DestUuid)
		fmt.Println(err.Error())
		return err
	}
	//下面这一部分是向请求接收方设置参数，已请求接收方的角度来看，DestID为当前用户的Uuid
	fileRequest.DestUuid = User.Uuid
	//服务器发送发起方的请求给请求接收方
	fileRequest.QCHeadBase.Number = core.FILE_RES_SEND_TRAN
	response, err := fileRequest.Encode()
	if err != nil {
		err = fmt.Errorf("FileRequestHandle Error: %v", err)
		fmt.Println(err.Error())
		return err
	}
	//向目标用户的流中写入数据
	n, err := destUser.Write(response)
	if err != nil || n == 0 {
		err = errors.New("Write file response fail")
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func FileResponseHandle(buffer []byte, Conn *net.Conn, userCache *model.UserCache, User *model.User, TokenIP string, TokenPort int) error {
	fileResponse := core.FileMsgResponse2{}
	err := fileResponse.Decode(buffer)
	if err != nil {
		err = fmt.Errorf("FileResponseHandle Error : %v", err)
		fmt.Println(err.Error())
		return err
	}
	if fileResponse.Number != core.FILE_REQ_SEND_STATUS {
		err = errors.New("Not has file response nubmer ")
		fmt.Println(err.Error())
		return err
	}

	var destUser *model.User
	//这里的DestID应该是之前的文件请求发起方
	if destUser = userCache.Query(fileResponse.DestUuid); destUser == nil {
		err = fmt.Errorf("Don't have the destUser: %v", fileResponse.DestUuid)
		fmt.Println(err.Error())
		return err
	}

	fileResponse.DestUuid = destUser.Uuid
	//服务器发送请求接收方的回应给请求发起方:
	fileResponse.Number = core.FILE_RES_SEND_STATUS
	responseBuff, err := fileResponse.Encode()
	if err != nil {
		err = fmt.Errorf("FileResponseHandle Error: %v", err)
		fmt.Println(err.Error())
		return err
	}
	n, err := (*destUser.Conn).Write(responseBuff)
	if n == 0 || err != nil {
		err = errors.New("Write response to from error")
		fmt.Println(err)
		return err
	}

	if fileResponse.State == 0 {
		return nil
	}

	conn, err := net.Dial("tcp", TokenIP+":"+strconv.Itoa(TokenPort))
	if err != nil {
		err = errors.New("token server is closed")
		fmt.Println(err.Error())
		return err
	}
	//destUuid应该是之前的文件请求发起方，User应该是之前的文件接收方
	related := &model.Related{
		protocol.TsHead{Len: 44, Version: 1, Type: core.CTRL, Number: protocol.USER_REQ_SEND_TOKEN},
		destUser.Uuid,
		User.Uuid,
	}
	tokenRequestBuff, err := related.Encode()
	if err != nil {
		err = fmt.Errorf("FileResponseHandle Error : %v", err)
		fmt.Println(err.Error())
		return err
	}
	n, err = conn.Write(tokenRequestBuff)
	if n == 0 && err != nil {
		err = errors.New("Write the tokenRequestBuff error")
		fmt.Println(err.Error())
		return err
	}
	tokenResponseBuff := make([]byte, 52)
	n, err = conn.Read(tokenResponseBuff)
	if n == 0 && err != nil && n != len(tokenResponseBuff) {
		err = errors.New("Read the tokenResponseBuff error")
		fmt.Println(err.Error())
		return err
	}
	conn.Close()

	token := &model.Token{}
	err = token.Decode(tokenResponseBuff)
	if err != nil {
		err = fmt.Errorf("FileResponseHandle Error : %v", err)
		fmt.Println(err.Error())
		return err
	}

	if token.Number != protocol.USER_RES_SEND_TOKEN {
		err = errors.New("Don't have the USER_RES_SEND_TOKEN type")
		fmt.Println(err.Error())
		return err
	}

	fromResponse := core.FileMsgResponse{}
	destResponse := core.FileMsgResponse{}

	fromResponse.QCHeadBase.Len = 51
	fromResponse.QCHeadBase.Version = 1
	fromResponse.QCHeadBase.Type = core.FILE
	//服务器发送token和TDS的IP地址、对client提供服务的端口以及连接方式给请求发起方
	fromResponse.QCHeadBase.Number = core.FILE_RES_SEND_ACTIVE_STATUS
	//User应该是之前的文件请求接收方
	fromResponse.DestUuid = User.Uuid

	fromResponse.IP = TokenIP
	fromResponse.Port = 9004
	fromResponse.Token = token.FromToken
	fromResponse.ConnMethod = protocol.TCP

	destResponse.QCHeadBase.Len = 51
	destResponse.QCHeadBase.Version = 1
	destResponse.QCHeadBase.Type = core.FILE
	//服务器发送token和TDS的IP地址，对client提供服务的端口以及连接方式给请求接收方:
	destResponse.QCHeadBase.Number = core.FILE_RES_SEND_PASSIVE_STATUS
	destResponse.DestUuid = destUser.Uuid

	destResponse.IP = TokenIP
	destResponse.Port = 9004
	destResponse.Token = token.DestToken
	destResponse.ConnMethod = protocol.TCP

	fromBuff, err := fromResponse.Encode()
	if err != nil {
		err = fmt.Errorf("FileResponseHandle Error : %v", err)
		fmt.Println(err.Error())
		return err
	}

	destBuff, err := destResponse.Encode()
	if err != nil {
		err = fmt.Errorf("FileResponseHandle Error : %v", err)
		fmt.Println(err.Error())
		return err
	}
	//User是文件请求的接收方
	n, err = (*User.Conn).Write(destBuff)
	if n == 0 || err != nil {
		err = errors.New("Send dest buff error")
		fmt.Println(err.Error())
		return err
	}
	//destUser是文件请求的发起方
	n, err = (*destUser.Conn).Write(fromBuff)
	if n == 0 || err != nil {
		err = errors.New("Send from buff error")
		return err
	}
	fmt.Println("send ok")
	return nil
}

func PushHandle(buffer []byte, Number uint16, User *model.User, userCache *model.UserCache, Conn *net.Conn) error {

	switch Number {
	case core.TEXT:
		return TextHandle(buffer, User, userCache, Conn)
	case core.AUDIO_FRAGMENT:
		return AudioFragmentHandle(buffer, User, userCache, Conn)
	case core.PHOTO:
	case core.VIDEO_FRAGMENT:
	case core.SHAKE:
	case core.MUSIC:
	default:
		return errors.New("can't find command")
	}
	return nil
}

func AudioFragmentHandle(buffer []byte, User *model.User, userCache *model.UserCache, Conn *net.Conn) error {
	DestUuid, err := core.UuidToString(buffer[12:28])
	if err != nil {
		err = fmt.Errorf("AudioFragmentHandle Error : %v", err)
		fmt.Println(err.Error())
		return err
	}

	var destUser *model.User
	if destUser = userCache.Query(DestUuid); destUser == nil {
		err = fmt.Errorf("Don't have the destUser : %v", DestUuid)
		fmt.Println(err.Error())
		return err
	}

	DestUuidBuff, err := core.StringToUuid(User.Uuid)
	if err != nil {
		err = fmt.Errorf("AudioFragmentHandle Error : %v", err)
		fmt.Println(err.Error())
		return err
	}

	copy(buffer[12:28], DestUuidBuff[0:16])
	n, err := destUser.Write(buffer)
	if n == 0 || err != nil {
		err = errors.New("Write to " + DestUuid + " error")
		return err
	}
	fmt.Println("send audio ok")
	return nil
}

func TextHandle(buffer []byte, User *model.User, userCache *model.UserCache, Conn *net.Conn) error {

	text := core.TextMsg{}
	err := text.Decode(buffer)
	if err != nil {
		err = fmt.Errorf("TextHandle Error : %v", err)
		fmt.Println(err.Error())
		return err
	}

	var destUser *model.User
	if destUser = userCache.Query(text.DestUuid); destUser == nil {
		err = fmt.Errorf("Don't have the destUser : %v", text.DestUuid)
		fmt.Println(err.Error())
		return err
	}

	text.DestUuid = User.Uuid
	requestBuff, err := text.Encode()
	if err != nil {
		err = fmt.Errorf("TextHandle Error : %v", err)
		fmt.Println(err.Error())
		return err
	}

	n, err := destUser.Write(requestBuff)
	if err != nil || n == 0 {
		err = errors.New("Write to " + text.DestUuid + " error")
		return err
	}
	fmt.Println("send text ok")
	return nil
}
