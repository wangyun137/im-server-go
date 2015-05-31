package component

import (
	// "account-server/model"
	// account "account-server/protocol"
	"errors"
	"fmt"
	push_model "push-server/model"
	. "push-server/protocol"
)

const (
	USER_UUID = "00000000000000000000000000000000"
	DEST_UUID = "00000000000000000000000000000000"
	TOKEN     = 0
	DEVICE_ID = 0
)

var queryChan chan *QueryInfo

var queryComponent QueryComponent

type QueryComponent struct {
}

func GetQueryComponent() *QueryComponent {
	return &queryComponent
}

func NewQueryComponent() *QueryComponent {
	return &queryComponent
}

func (this *QueryComponent) Init() error {
	queryChan = make(chan *QueryInfo, push_model.QUERY_CHANNEL_COUNT)
	return nil
}

func (this *QueryComponent) Start() {
	go this.HandleQueryInfo()
}

func (this *QueryComponent) HandleQueryInfo() {
	dispatcherOut := GetDispatcherOutComponent()
	for {
		info := <-data

		Msg, err := this.InitMsg(info)
		if err != nil || Msg == nil {
			continue
		}

		if err = dispatch.NPushBack(Msg, false); err != nil {
			continue
		}
	}
}

func (this *QueryComponent) InitMessage(info *push_model.QueryInfo) (*Message, error) {
	queryRequest := model.QueryDeviceRequest{
	// model.QueryInfo{
	// 	model.Request{account.QUERYDEVICE_REQ},
	// 	info.UserUuid,
	// },
	// info.DeviceNumber,
	// info.Token,
	// info.Number,
	}
	packet, err := queryRequest.Encode()
	if err != nil {
		err = fmt.Errorf("QueryComponent InitMessage Error : %v", err)
		return nil, err
	}

	message, err := protocol.NewMessage(VERSION, MT_PUSH < PY_PUSHSERVER,
		TOKEN, N_ACCOUNT_SYS, uint16(len(packet)), USER_UUID, DEST_UUID, DEVICE_ID, packet)
	if err != nil {
		err = fmt.Errorf("QueryComponent InitMessage Error : %v", err)
		return nil, err
	}

	return message, nil
}

func (this *QueryComponent) AcceptInfo(info *push_model.QueryInfo) error {
	if info != nil {
		data <- info
	} else {
		return errors.New("QueryComponent AcceptInfo Error")
	}

	return nil
}
