package core

// BaseProxy contains common fields shared by all proxy types
type BaseProxy struct {
	Type    string `yaml:"type" json:"type"`
	Remark  string `yaml:"remark" json:"remark"`
	Server  string `yaml:"server" json:"server"`
	Port    int    `yaml:"port" json:"port"`
	Group   string `yaml:"group" json:"group"` // group here is not proxy group, but subscription group
	GroupId int    `yaml:"group_id" json:"group_id"`

	// Common flags (tri-state booleans)
	UDP   *bool `yaml:"udp,omitempty" json:"udp,omitempty"`
	TFO   *bool `yaml:"tfo,omitempty" json:"tfo,omitempty"`
	SCV   *bool `yaml:"skip-cert-verify,omitempty" json:"skip-cert-verify,omitempty"`
	TLS13 *bool `yaml:"tls13,omitempty" json:"tls13,omitempty"`

	UnderlyingProxy string `yaml:"underlying-proxy,omitempty" json:"underlying-proxy,omitempty"`
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

func (p *BaseProxy) GetUnderlyingProxy() string {
	return p.UnderlyingProxy
}

func (p *BaseProxy) SetUnderlyingProxy(proxy string) {
	p.UnderlyingProxy = proxy
}
