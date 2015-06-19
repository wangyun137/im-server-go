package dispatcher

import (
	"connect-server/model"
	core "connect-server/protocol"
	"connect-server/utils/uuid"
	"fmt"
	"net"
	"strconv"
	"td-server/protocol"
	"time"
)

type Tokener struct {
	IP            string
	Port          int
	TokenListener net.Listener
	Tokens        *model.Tokens
}

func NewTokener(IP string, Port int, tokens *model.Tokens) *Tokener {
	return &Tokener{IP: IP, Port: Port, Tokens: tokens}
}

func (tokener *Tokener) Start() {
	var err error
	tokener.TokenListener, err = net.Listen("tcp", tokener.IP+":"+strconv.Itoa(tokener.Port))
	if err != nil {
		fmt.Println("token-server Listen Error")
		return
	}

	for {
		Conn, err := tokener.TokenListener.Accept()
		if err != nil {
			continue
		}

		go tokener.Handle(&Conn)
	}
}

func (this *Tokener) Handle(Conn *net.Conn) {
	defer (*Conn).Close()

	buffer := make([]byte, 44)
	for {
		n, err := (*Conn).Read(buffer)
		if n == 0 || err != nil {
			fmt.Println("token-server read error")
			return
		}

		related := model.Related{}
		err = related.Decode(buffer)
		if err != nil {
			err = fmt.Errorf("token-server Handle Error:%v", err)
			fmt.Println(err.Error())
			return
		}

		var tokener *model.Tokener
		if t, err := this.Tokens.Query(related.FromUuid, related.DestUuid); t != nil && err == nil && t.TimeOut-time.Now().Unix() > 10 {
			tokener = t
		} else {
			fmt.Println("token-server Handle: Get Token")
			if t != nil {
				this.Tokens.Delete(related.FromUuid, related.DestUuid)
			}

			tokener = &model.Tokener{}
			fromToken := uuid.NewV4()
			tokener.FromToken, err = core.UuidToString(fromToken[:16])
			if err != nil {
				err = fmt.Errorf("token-server Handle Error:%v", err)
				fmt.Println(err.Error())
				return
			}

			destToken := uuid.NewV4()
			tokener.DestToken, err = core.UuidToString(destToken[:16])
			if err != nil {
				err = fmt.Errorf("token-server Handle Error:%v", err)
				fmt.Println(err.Error())
				return
			}

			tokener.TimeOut = time.Now().Unix() + 10
			err = this.Tokens.Add(related.FromUuid, related.DestUuid, tokener)
			if err != nil {
				err = fmt.Errorf("token-server Handle Error:%v", err)
				fmt.Println(err.Error())
				return
			}
		}
		token := model.Token{
			protocol.TsHead{Len: 50, Version: 1, Type: core.CTRL, Number: protocol.USER_RES_SEND_TOKEN},
			related.FromUuid,
			tokener.FromToken,
			related.DestUuid,
			tokener.DestToken,
			tokener.TimeOut,
		}
		tokenBuff, err := token.Encode()
		if err != nil {
			err = fmt.Errorf("token-server Handle Error:%v", err)
			fmt.Println(err.Error())
			return
		}

		n, err = (*Conn).Write(tokenBuff)
		if n == 0 || err != nil {
			fmt.Println("token-server Handle Write data error")
			return
		}
	}
}
