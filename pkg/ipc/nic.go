package ipc

import (
	"fmt"
)

type NICConfig struct {
	Name string   `json:"name"`
	IPs  []string `json:"ips,omitempty"`
	MTU  int      `json:"mtu,omitempty"`
	Up   bool     `json:"up"`
}

// 结构体方法
func (cfg *NICConfig) Set() error {
	fmt.Println("NIC Set:", cfg)
	return nil
}

func (cfg *NICConfig) UpIf() error {
	fmt.Println("NIC Up:", cfg.Name)
	return nil
}

func (cfg *NICConfig) DownIf() error {
	fmt.Println("NIC Down:", cfg.Name)
	return nil
}

func (cfg *NICConfig) AddIP(ip string) error {
	fmt.Println("NIC AddIP:", cfg.Name, ip)
	return nil
}

// 自动注册 ModuleInterface
func (cfg *NICConfig) Handlers() map[string]Handler {
	return map[string]Handler{
		"nic.set": func(payload any) (any, error) {
			target := &NICConfig{}
			if err := DecodeTypedPayload(payload, target); err != nil {
				return nil, err
			}
			return nil, target.Set()
		},
		"nic.up": func(payload any) (any, error) {
			target := &NICConfig{}
			if err := DecodeTypedPayload(payload, target); err != nil {
				return nil, err
			}
			return nil, target.UpIf()
		},
		"nic.down": func(payload any) (any, error) {
			target := &NICConfig{}
			if err := DecodeTypedPayload(payload, target); err != nil {
				return nil, err
			}
			return nil, target.DownIf()
		},
		"nic.addip": func(payload any) (any, error) {
			target := &NICConfig{}
			if err := DecodeTypedPayload(payload, target); err != nil {
				return nil, err
			}
			if len(target.IPs) == 0 {
				return nil, fmt.Errorf("no IP provided")
			}
			return nil, target.AddIP(target.IPs[0])
		},
	}
}
