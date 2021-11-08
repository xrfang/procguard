package procguard

type (
	guardian struct {
		Slots slots  `json:"slots"` //时段定义
		Procs []proc `json:"procs"` //进程定义
		Check int    `json:"check"` //检查时间间隔
		log   func(int, string, ...interface{})
	}
)
