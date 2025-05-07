package goodesl

import (
	"fmt"
	"strings"
)

const (
	Originate = "originate"
	Status    = "status"
)

type CmdBuilder struct {
	baseCmd string
	params  []string
}

// NewCmdBuilder 创建一个新的 CmdBuilder
func NewCmdBuilder(baseCmd string) *CmdBuilder {
	return &CmdBuilder{
		baseCmd: baseCmd,
		params:  []string{},
	}
}

// AddParam 添加命令的参数
func (b *CmdBuilder) AddParam(param string) *CmdBuilder {
	b.params = append(b.params, param)
	return b
}

// Build 返回完整的命令字符串
func (b *CmdBuilder) Build() string {
	return fmt.Sprintf("%s %s", b.baseCmd, strings.Join(b.params, " "))
}
