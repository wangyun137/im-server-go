package server

import (
	core "connect-server/protocol"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"td-server/protocol"
)

type TCPHandler struct {
}
