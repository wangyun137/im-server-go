package component

import (
	"errors"
	"fmt"
	"io"
	"net"
	"push-server/buffpool"
	"push-server/model"
	"push-server/protocol"
	"push-server/utils"
	snode_component "snode-server/component"
	"time"
	"tree-server/define"
	"tree-server/structure"
	account_structure "account-server/structure"
)

const (
	P_UID           = "00000000000000000000000000000000"
	MESSAGE_MAX_BUF = 1000
	MAX_UINT16      = 0xff
	UUID_LEN        = 16
	DEVNO_LEN       = 15
)

type RegisterComponent struct {
	connQueue chan *net.Conn
	connArray []*net.Conn
	queueLen  int
	arrayLen  int
	IP        uint32
	Port      uint16
	PushID    string
}

var registerComponent RegisterComponent

func NewRegisterComponent(queueLen, arrayLen int, ip uint32, port uint16, serlfId string) *RegisterComponent {
	registerComponent.queueLen = queueLen
	registerComponent.arrayLen = arrayLen
	registerComponent.IP = ip
	registerComponent.Port = port
	registerComponent.PushID = serlfId
}

func GetRegisterComponent() *RegisterComponent {
	return &registerComponent
}

func (this *RegisterComponent) Init() error {
	this.connQueue = make(chan *net.Conn, this, queueLen)
	this.connArray = make([]*net.Conn, this.arrayLen)
	return nil
}

func (this *RegisterComponent) Start() error {
	if this.connQueue == nil {
		err := errors.New("RegisterComponent Start Error:please init RegisterComponent first")
		fmt.Println(err.Error())
		return err
	}
	go this.registerProcess()
	return nil
}

func (this *RegisterComponent) registerProcess() error {
	buffPool, queryComponent, _ := this.GetComponent()

	for {
		conn := <-this.connQueue
		if conn == nil {
			fmt.Println("RegisterComponent Error:get a nil conn")
			continue
		}
		err := setKeepAlive(conn, true, -30)
		if err != nil {
			err = fmt.Errorf("RegisterComponent registerProcess Error: %v", err)
			fmt.Println(err.Error())
			continue
		}
		go this.dispacth(conn, buffPool, queryComponent)
	}
	return nil
}

func (this *RegisterComponent) GetComponent() (*buffpool.BuffPool, *QueryComponent, *DispatcherComponent) {
	var buffPool *buffPool.BuffPool
	for {
		buffPool = buffPool.GetBuffPool()
		if buffPool != nil {
			break
		} else {
			fmt.Println("RegisterComponent GetComponent Error:can not get buffPool")
		}
	}
	var queryComponent *QueryComponent
	for {
		queryComponent = GetQueryComponent()
		if queryComponent != nil {
			break
		} else {
			fmt.Println("RegisterComponent GetComponent Error:can not get QueryComponent")
		}
	}
	var dispatchComponent *DispatcherComponent
	for {
		dispatchComponent = GetDispatcherComponent()
		if dispatchComponent != nil {
			break
		} else {
			fmt.Println("RegisterCompoennt GetComponent Error:can not get DispatcherComponent")
		}
	}
	return buffPool, queryComponent, dispatchComponent
}

func setKeepAlive(conn *net.Conn, keepAlive bool, seconds int) error {
	if val, ok := (*conn).(*net.TCPConn); ok {
		err := val.SetKeepAlive(keepAlive)
		if err != nil {
			err = fmt.Errorf("RegisterComponent setKeepAlive Error: %v", err)
			return err
		}

		if seconds > 0 {
			err = val.SetKeepAlivePeriod(time.Duration(seconds) * time.Second)
			if err != nil {
				err = fmt.Errorf("RegisterComponent setKeepAlive Error : %v", err)
				return err
			}
		}
		return nil
	} else {
		err := errors.New("RegisterCompoennt setKeepAlive Error : the conn is not the TCPConn")
		return err
	}

}

func (this *RegisterComponent) dispatch(conn *net.Conn, buffPool *buffpool.BuffPool, queryComponent *QueryComponent) error {
	if conn == nil || buffPool == nil || queryComponent == nil {
		err := errors.New("RegisterComponent dispatch Error: invalid argument")
		fmt.Println(err.Error())
		return err
	}
	message, err := this.readMessage(conn)
	if err != nil {
		err = errors.New("RegisterCompoennt dispatch Error: %v", err)
		return err
	}
	if message.MsgType != protocol.MT_REGISTER {
		err = fmt.Errorf("RegisterCompoennt dispatch Error : %v", err)
		return err
	}
	if message.PushType == protocol.PT_SERVICE {
		err = this.handleService(buffPool, message, conn)
		if err != nil {
			err = fmt.Errorf("RegisterComponent dispatch Error: %v", err)
			return err
		}
	} else if message.PushType == protocol.PT_CLIENT {
		err = this.handleClient(buffPool, queryComponent, message, conn)
		if err != nil {
			err = fmt.Errorf("RegisterCompoennt dispatch Error : %v", err)
			return err
		}
	}
	return nil
}

func (this *RegisterComponent) handleService(buffPool *buffpool.BuffPool,message *protocol.Message,conn *net.Conn) {
	if buffPool == nil || message == nil || conn == nil {
		err := errors.New("RegisterComponent handleService Error: the argument is nil")
		return err
	}
	err := buffPool.AddService(message.Number,message.UserUuid,conn)
	if err != nil {
		err  = fmt.Errorf("RegisterCompoennt handleService Error: %v", err)
		return err
	}
	responseMessage, err := protocol.NewMessage(protocol.VERSION,protocol.MT_REGISTER,protocol.PT_PUSHSERVER,0,message.Number,0 P_UID,
		message.UserUuid, 0, nil)
	if err != nil {
		err = fmt.Errorf("RegisterComponent handleService Error:%v", err)
		return err
	}
	GetDispatcherOutComponent().NPushBack(responseMessage,false)

	if utils.GetRunWithSnode().Flag {
		onlineInfo := structure.ClientOnlineInfo{
			PushID:   this.PushID,
			ParentID: this.PushID,
		}
		data, err := onlineInfo.Encode()
		if err != nil {
			err = fmt.Errorf("RegisterCompoennt handleService Error:%v", err)
			return err
		}
		snode_component.GetNodeComponent().Forward(stree_define.ONLINE, msg.UserId, 0, data) //added for QueryUser
	}

	go this.listenService(conn,buffPool,message.Number,message.UserUuid)
	return nil
}

func (this *RegisterComponent) listenService(conn *net.Conn,buffPool *buffpool.BuffPool,number uint16,userUuid string) error {
	if buffPool == nil {
		err := errors.New("RegisterComponent listenService Error: buffpool is nil")
		return err
	}
	if conn == nil {
		err := errors.New("RegisterComponent listenService Error: the conn is nil")
		if utils.GetRunWithSnode().Flag {
			this.ServiceOffline(userUuid)
		}

		err = buffPool.DeleteService(number,userUuid)
		if err != nil {
			err = fmt.Errorf("RegisterComponent listenService Error:%v",err)
			return err
		}
		return err
	}
	var dispatcher *DispatcherComponent
	for {
		dispatcher = GetDispatcherComponent()
		if dispatcher != nil {
			break
		} else {
			fmt.Println("RegisterComponent listenService Error: can not get DispatcherCompoennt")
		}
	}
	for {
		message, err := this.readMessage(conn)
		if number == protocol.N_ACCOUNT_SYS {

		}
		if err != nil && err != io.EOF && conn != nil {
			fmt.Println("RegisterComponent listenService Error: read message fail")
			continue
		} else if err != nil && (err == io.EOF | conn == nil) {
			err = errors.New("RegisterComponent listenService Error:the connection is invalid")
			err2 := buffPool.DeleteService(number,userUuid)

			if utils.GetRunWithSnode().Flag {
				this.ServiceOffline(userUuid)
			}

			if err2 != nil {
				err2 = fmt.Errorf("RegisterComponent listenService Error: %v",err)
				fmt.Println(err2.Error())
				return err2
			}
			fmt.Println(err.Error())
			return err
		}

		if message == nil {
			continue
		}
		dispatcher.NPushBack(message,true)
	}
	return nil
}

func (this *RegisterComponent) ServiceOffline(userUuid string) {
	snode_component.GetNodeComponent().Forward(define.OFFLINE,userUuid,0,nil)
}


func (this *RegisterComponent) handleClient(buffPool *buffpool.BuffPool, queryComponent *QueryComponent,
	message *protocol.Message, conn *net.Conn) {
	if buffPool == nil || queryComponent == nil || message == nil || conn == nil || len(msg.Data) != DEVNO_LEN {
		err := errors.New("RegisterComponent handleClient() Error : invaild argument")
		return err
	}
	if !this.checkAccountServer(buffPool,message,conn) {
		err := errors.New("RegisterCompoennt handleClient() Error: the account server is not register")
		return err
	}
	clientConn := model.Client{Conn:conn,Status:false}
	var token uint8
	if message.Token == 0 {
		token = buffPool.AddTmpClient(message.Number,message.UserUuid,string(message.Data),&clientConn)
	} else {
		token = buffPool.AddTmpClient2(message.Number,message.UserUuid,string(message.Data),message.DeviceId,
			message.Token,&clientConn)
	}

	if token == 0 {
		err := errors.New("RegisterCompoennt handleClient() Error: can not add the client to buffpool")
		return err
	}

	queryInfo := protocol.QueryInfo{
		Token: 			token,
		Numebr: 		message.Number,
		UserUuid: 		message.UserUuid,
		DeviceNumber:	string(message.Data),
	}

	err := queryComponent.AcceptInfo(queryInfo)
	if err != nil {
		err = fmt.Errorf("RegisterCompoennt handleClient() Error: %v", err)
		return er
	}
	return nil
}

func (this *RegisterComponent) checkAccountServer(buffPool *buffpool.BuffPool, message *protocol.Message, conn *net.Conn) bool {
	if buffPool == nil || message == nil || conn == nil {
		fmt.Println("RegisterComponent checkAccountServer() Error: invalid argument")
		return false
	}
	connService  := buffPool.GetService(protocol.N_ACCOUNT_SYS)
	if connService == nil {
		fmt.Println("RegisterComponent checkAccountServer() Error: can not get the account_sys server")
		response,_ := protocol.NewMessage(protocol.VERSION, protoocl.MT_ERROR, protocol.PT_PUSHSERVER, message.Token, message.Number, 0, P_UID,
			message.UserUuid, 0, nil)
		err := sendResponse(conn, response)
		if err != nil {
			err = fmt.Errorf("RegisterComponent checkAccountServer() Error:%v", err)
			fmt.Println(err.Error())
		}
		return false
	}
	return true
}

func (this *RegisterComponent) RegisterClientCallback(response *account_structure.QueryDeviceResponse) error {
	if response == nil {
		err := errors.New("RegisterComponent RegisterClientCallback Error: the argument is nil")
		return er
	}

	var buffPool *buffpool.BuffPool
	for {
		buffPool = buffpool.GetBuffPool()
		if buffPool != nil {
			break
		} else {
			fmt.Println("RegisterComponent RegisterClientCallback Error: can not get buffpool")
		}
	}
	clientConn, err := this.checkQueryResponse(buffPool,response)
	if err != nil || clientConn == nil {
		fmt.Println("RegisterComponent RegisterClientCallback Error:can not get the account response")
		return err
	}
	newToken := buffPool.MoveTmpClient(response.Number,response.UserUuid, response.DeviceNumber, 
		resposne.DeviceID, response.Token, clientConn)
	if newToken == 0 {
		err = errors.New("RegisterComponent RegisterClientCallback Error:can not move client")
		fmt.Println(err.Error())
		err2 := this.sendFailedRegisterResponse(clientConn.Conn,response.Number,response.UserUuid)
		if err2 != nil {
			fmt.Println("RegisterComponent RegisterClientCallback Error: " + err2.Error())
		}

		return err
	}

	message,err := protocol.NewMessage(protocol.VERSION,protocol.MT_REGISTER, protocol.PT_PUSHSERVER, newToken, response.Number, 0,
		P_UID, response.UserUuid, response.DeviceID, nil)
	if err != nil {
		err := errors.New("RegisterComponent RegisterClientCallback Error:" + err.Error())
		fmt.Println(err.Error())
		err2 := this.sendFailedRegisterResponse(clientConn.Conn, response.Number, response.UserUuid)
		if err2 != nil {
			fmt.Println("RegisterComponent RegisterClientCallback Error:" + err2.Error())
		}
		err3 := buffPool.RemoveClient(response.Number, response.UserID, response.DeviceID, newToken)
		(*clientConn.Conn).Close()
		if err3 != nil {
			fmt.Println("RegisterComponent RegisterClientCallback Error:" + err3.Error())
		}

		return err
	}
	dispatcherOutComponent := GetDispactherOutComponent()
	dispatcherOutComponent.NPushBack(message,false)

	go listenClient(clientConn.Conn, buffPool, response.Number, response.UserUuid, response.DeviceID, newToken)
	err = this.onlineBroadcast(buffPool, dispatcherOutComponent, response.UserID, response.DeviceID)
	if err != nil {
		err := fmt.Errorf("RegisterComponent onlineBroadcast() Error: can not broadcast to other server, " + err.Error())
		return err
	}
	if !utils.GetRunWithSnode().Flag {
		GetCacheComponent.SendCached(response.UserUuid,response.DeviceID)
	}
	return nil
}


func (this *RegisterComponent) sendFailedRegisterResponse(conn *net.Conn, num uint16, userUuid string) error {
	response, err := protocol.NewMessage(protocol.VERSION, protocol.MT_ERROR, protocol.PT_PUSHSERVER, 0, num, 0,
		P_UID, userUuid, 0, nil)
	if err != nil {
		err = errors.New("RegisterComponent sendFailedRegisterResponse() Error: " + err.Error())
		return err
	}

	err = sendResponse(conn, response)
	if err != nil {
		err = errors.New("RegisterComponent sendFailedRegisterResponse() Error: " + err.Error())
		return err
	}
	return nil
}


func (this *RegisterComponent) onlineBroadcast(buffPool *buffpool.BuffPool, dispatcherOut *DispatcherOutComponent, userUuid string, deviceID uint32) error {

	if utils.GetRunWithSnode().Flag {
		onlineInfo := structure.ClientOnlineInfo{
			PushID:   this.PushID,
			ParentID: this.PushID,
		}
		data, err := onlineInfo.Encode()
		if err != nil {
			err = errors.New("RegisterComponent onlineBroadcast() Error: " + err.Error())
			return err
		}
		snode_component.GetNodeComponent().Forward(define.ONLINE, userUuid, deviceID, data)
		snode_component.GetNodeComponent().AddUsersNumber()
		snode_component.GetNodeComponent().AddConnNumber()
	}

	if buffPool == nil || dispatcherOut == nil {
		err := errors.New("RegisterComponent onlineBroadcast() Error: invalid argument")
		return err
	}

	brMsg := protocol.BRMsg{
		userUuid,
		deviceID,
	}
	brData, err := brMsg.Encode()
	if err != nil {
		err := errors.New("RegisterCompoent onlineBroadcast() Error: " + err.Error())
		return err
	}

	servNoUuidArr := buffPool.GetAllServices()
	for num, uuidArr := range servNoUuidArr {
		for _, uuid := range uuidArr {
			onlineMsg, err := protocol.NewMessage(protocol.VERSION, protocol.MT_BROADCAST_ONLINE, protocol.PT_PUSHSERVER, 0, num, uint16(len(brData)),
				P_UID, uuid, 0, brData)
			if err != nil {
				err = errors.New("RegisterComponent onlineBroadcast() Error: " + err.Error())
				return err
			}
			dispatcherOut.NPushBack(onlineMsg, false) //not need forward
		}
	}
	return nil
}


func (this *RegisterComponent) checkQueryResponse(buffPool *buffpool.BuffPool, response *account_structure.QueryDeviceResponse) (*model.Client, error) {
	if buffPool == nil || response == nil {
		err := errors.New("RegisterComponnet checkQueryResponse() Error: invalid argument")
		return nil, err
	}

	clientConn := buffPool.GetTmpClient(response.Number, response.UserID, response.DeviceNumber, response.Token)
	if clientConn == nil || clientConn.Conn == nil {
		err := errors.New("RegisterComponent checkQueryResponse() Error: can not get client conn from buffPool")
		return nil, err
	}

	if response.Exist == 0 {
		err := errors.New("RegisterCompoennt checkQueryResponse Error: did not pass the authorization")

		resp, _ := protocol.NewMessage(protocol.VERSION, protocol.MT_ERROR, protocol.PT_PUSHSERVER, response.Token, response.Number, 0, P_UID,
			response.UserUuid, response.DeviceID, nil)
		err = sendResponse(clientConn.Conn, resp)
		if err != nil {
			err = errors.New("RegisterCompoennt checkQueryResponse Error: Can not send register response, " + err.Error())
			fmt.Println(err.Error())
		}

		err2 := buffPool.RemoveTmpClient(response.Number, response.UserID, response.DeviceNumber, response.Token)
		(*clientConn.Conn).Close()
		if err2 != nil {
			err2 := errors.New("RegisterCompoennt checkQueryResponse Error: can not remove client conn from tmp buffpool, " + err2.Error())
			return nil, err2
		}
		return nil, err
	}
	return clientConn, nil
}

func (this *RegisterComponent) listenClient(conn *net.Conn, buffPool *buffpool.BuffPool, number uint16, userUuid string, deviceID uint32, token uint8) error {
	if buffPool == nil {
		err := errors.New("RegisterComponent listenClient() Error: invalid argument")
		fmt.Println(err.Error())
		return err
	}
	if conn == nil {
		err := errors.New("RegisterComponent listenClient() Error:invalid argument")
		fmt.Println(err.Error())
		fmt.Println("the invalid client connection will be removed")
		err2 := buffPool.RemoveClient(number, userUuid, deviceID, token)
		if err2 != nil {
			err2 = errors.New("RegisterComponent listenClient() Error:remove the invalid client's connection failed" + err2.Error())
			fmt.Println(err2.Error())
			return err2
		}
		fmt.Println("the invalid client connection has been removed")
		return err
	}
	var dispatcher *DispatcherComponent
	for {
		dispatcher = GetDispatcherComponent()
		if dispatcher != nil {
			break
		} else {
			fmt.Println("RegisterComponent listenClient() Error:can not get DispatcherComponent, will try again")
		}
	}

	for {
		msg, err := this.readMessage(conn)
		if err != nil && err != io.EOF && conn != nil {
			err = errors.New("RegisterComponent listenClient() Error " + err.Error())
			fmt.Println(err.Error())
			continue
		} else if err != nil && (err == io.EOF || conn == nil) {
			err = errors.New("RegisterComponent listenClient() Error the connection is invalid, " + err.Error())
			err2 := buffPool.RemoveClient(number, userUuid, deviceID, token)
			if conn != nil {
				(*conn).Close()
			}
			if err2 != nil {
				err2 = errors.New("RegisterComponent listenClient() Error: can not remove invalid client's connection, " + err2.Error())
				fmt.Println(err2.Error())
				return err2
			}
			clientArr := buffPool.QueryForDeviceId(number, userUuid, deviceID)
			if clientArr == nil || len(clientArr) == 0 {
				//offline broadcast
				unregister := GetUnregisterComponent()
				if unregister != nil {
					err3 := unregister.offlineBroadcast(buffPool, userUuid, deviceID)
					if err3 != nil {
						err3 = errors.New("RegisterComponent listenClient() Error:one client disconnected but offlineBroadcast send failed, " + err3.Error())
						fmt.Println(err3.Error())
					}
				} else {
					fmt.Println("RegisterComponent listenClient() Error:one client disconnected but offlineBroadcast send failed")
				}
			}
			
			return err
		}
		//push to DispatcherQueue
		if msg == nil {
			continue
		}
		dispatcher.NPushBack(msg, true) //need forward
	}
	return nil
}

func (this *RegisterComponent) PushBack(conn *net.Conn) error {
	if conn == nil {
		err := errors.New("RegisterCompoennt PushBack Error: the argument conn is nil")
		return err
	}
	this.connQueue <- conn
	return nil
}

func (this *RegisterComponent) readMessage(conn *net.Conn) (*protocol.Message, error) {
	if conn == nil {
		err := errors.New("RegisterComponent readMessage Error: the conn is nil")
		return nil, err
	}
	buffer := make([]byte, MESSAGE_MAX_BUF)
	err := this.read(buffer[0:44], conn)
	if err != nil {
		err = io.EOF
		return nil, err
	}
	message := &protocol.Message{}
	err = message.Decode(buffer[0:44])
	if err != nil {
		err = fmt.Errorf("RegisterComponent readMessage Error : %v", err)
		return nil, err
	}

	if message.DataLen == 0 {
		return message, nil
	}

	var data []byte
	if (message.DataLen + 44) > MESSAGE_MAX_BUF {
		data = make([]byte, message.DataLen)
	} else {
		data = buffer[44 : 44+message.DataLen]
	}
	err = this.read(data, conn)
	if err != nil {
		err = io.EOF
		return nil, err
	}
	message.Data = data
	return message, nil
}

func (this *RegisterComponent) read(buffer []byte, conn *net.Conn) error {
	if conn == nil {
		err := errors.New("RegisterComponent read Error: the conn is nil")
		return err
	}
	index := 0
	for index < len(buffer) {
		n, err := (*conn).Read(buffer[index:len(buffer)])
		index += n
		if index < len(buffer) && err != nil {
			return err
		}
		if n == 0 {
			return io.EOF
		}
	}
	return nil
}

func sendResponse(conn *net.Conn, message *protocl.Message) error {
	if conn == nil || message == nil {
		err := errors.New("RegisterComponent sendResponse Error : the argument is invalid")
		return err
	}
	buffer, err := message.Encode()
	if err != nil {
		err = fmt.Errorf("RegisterComponent sendResposne Error: %v", err)
		return err
	}
	index := 0
	for index < len(buffer) {
		n, err := (*conn).Write(buffer[index:])
		index += n
		if index < len(buffer) && err != nil {
			return err
		}
	}
	return nil
}
