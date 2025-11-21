package proxy

import (
	"github.com/gfunc/subconvergo/config"
)

type ProxyInterface interface {
	GetType() string
	GetRemark() string
	GetServer() string
	GetPort() int
	SetRemark(remark string)
	GetGroup() string
	SetGroup(group string)
	GetGroupId() int
	SetGroupId(groupId int)
}

type SubconverterProxy interface {
	ProxyInterface
	ToShareLink(ext *config.ProxySetting) (string, error)
	ToClashConfig(ext *config.ProxySetting) map[string]interface{}
}
