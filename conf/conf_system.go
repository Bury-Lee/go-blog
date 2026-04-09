package conf

import "fmt"

type System struct {
	Ip      string `yaml:"ip"`
	Port    int    `yaml:"port"`
	Env     string `yaml:"env"`
	RunMode string `yaml:"run_mode"`
}

func (this *System) Addr() string {
	return fmt.Sprintf("%s:%d", this.Ip, this.Port)
}
