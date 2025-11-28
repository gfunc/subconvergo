package core

// BaseProxy contains common fields shared by all proxy types
type BaseProxy struct {
	Type    string `yaml:"type" json:"type"`
	Remark  string `yaml:"remark" json:"remark"`
	Server  string `yaml:"server" json:"server"`
	Port    int    `yaml:"port" json:"port"`
	Group   string `yaml:"group" json:"group"` // group here is not proxy group, but subscription group
	GroupId int    `yaml:"group_id" json:"group_id"`
}

func (p *BaseProxy) GetType() string {
	return p.Type
}

func (p *BaseProxy) GetRemark() string {
	return p.Remark
}

func (p *BaseProxy) SetRemark(remark string) {
	p.Remark = remark
}

func (p *BaseProxy) GetServer() string {
	return p.Server
}

func (p *BaseProxy) GetPort() int {
	return p.Port
}

func (p *BaseProxy) GetGroup() string {
	return p.Group
}

func (p *BaseProxy) SetGroup(group string) {
	p.Group = group
}

func (p *BaseProxy) GetGroupId() int {
	return p.GroupId
}

func (p *BaseProxy) SetGroupId(groupId int) {
	p.GroupId = groupId
}
