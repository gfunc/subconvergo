package core

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
	SingleConvertableMixin
	ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error)
}

type ClashConvertableMixin interface {
	ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error)
}

type SingleConvertableMixin interface {
	ToSingleConfig(ext *config.ProxySetting) (string, error)
}

type LoonConvertableMixin interface {
	ToLoonConfig(ext *config.ProxySetting) (string, error)
}

type SurgeConvertableMixin interface {
	ToSurgeConfig(ext *config.ProxySetting) (string, error)
}

type QuantumultXConvertableMixin interface {
	ToQuantumultXConfig(ext *config.ProxySetting) (string, error)
}

type SingboxConvertableMixin interface {
	ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error)
}
