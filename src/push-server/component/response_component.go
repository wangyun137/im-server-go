package component

import (
	// "account-server/structure"
	"errors"
	"fmt"
	"push-server/model"
	"push-server/protocol"
)

var responseChan chan *protocol.Message

func GetResponseComponent() *ResponseComponent {
	return &responseComponent
}

func NewResponseComponent() *ResponseComponent {
	return &responseComponent
}

type ResponseComponent struct {
}

var responseComponent ResponseComponent

//provide the init interface
func (this *ResponseComponent) Init() error {
	responseChan = make(chan *protocol.Message, model.QUERY_CHANNEL_COUNT)

	return nil
}

func (this *ResponseComponent) Start() {
	go this.HandleResponseInfo()
}

func (this *ResponseComponent) AcceptResponseInfo(info *protocol.Message) error {
	if info != nil {
		responseChan <- info
	} else {
		return errors.New("ResponseCompoennt AcceptResponseInfo Error:invalid argument")
	}

	return nil
}

func (this *ResponseComponent) HandleResponseInfo() {
	for {
		info := <-responseChan

		if info == nil {
			continue
		}

		if info.MsgType == protocol.MT_UNREGISTER {
			if err := this.HandleRegisterClientCallback(info.Data); err != nil {
				continue
			}
		}

		if info.MsgType == protocol.MT_PUSH && (info.PushType == protocol.PT_SERVICE || info.PushType == protocol.PT_CLIENT) {
			if err := this.HandleRegisterClientCallback(info.Data); err != nil {
				fmt.Println("ResponseComponent HandleResponseInfo Error :" + err.Error())
			}
		}
	}
}

func (this *ResponseComponent) HandleRegisterClientCallback(msgData []byte) error {
	registerhandle := GetRegisterComponent()

	responseMsg, err := this.DecodeMsg(msgData)
	if err != nil {
		fmt.Println("ResponseComponent HandleRegisterClientCallback() Error: ", err)
		return err
	} else {
		if err = registerhandle.RegisterClientCallback(responseMsg); err != nil {
			fmt.Println("ResponseComponent HandleRegisterClientCallback() Error:RegisterClientCallback failed, " + err.Error())
			return err
		}
	}

	return nil
}

// func (this *ResponseComponent) DecodeMsg(data []byte) (*structure.QueryDeviceResponse, error) {
// 	responseMsg := structure.QueryDeviceResponse{}
// 	if err := responseMsg.Decode(data); err != nil {
// 		fmt.Println(err.Error())
// 		return nil, err
// 	}
// 	return &responseMsg, nil
// }
