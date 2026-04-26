package conf

import "fmt"

type System struct {
	Ip               string `yaml:"ip"`
	Port             int    `yaml:"port"`
	Env              string `yaml:"env"`
	RunMode          string `yaml:"run_mode"`
	Cron             bool   `yaml:"cron"`              //是否开启定时任务,在分布式环境下,只建议一个实例开启
	ScheduledCleanup bool   `yaml:"scheduled_cleanup"` //是否开启清理浏览记录,在分布式环境下,只建议一个实例开启
}

func (this *System) Addr() string {
	return fmt.Sprintf("%s:%d", this.Ip, this.Port)
}
