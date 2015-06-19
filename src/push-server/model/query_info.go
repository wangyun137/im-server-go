package model

const (
	QUERY_CHANNEL_COUNT = 100
)

type QueryInfo struct {
	Token        uint8
	Number       uint16
	UserUuid     string
	DeviceNumber string
}
