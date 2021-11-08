package procguard

type proc struct {
	Cmd  string         `json:"cmd"`  //命令，相对路径为当前程序路径
	Args []string       `json:"args"` //参数
	Plan map[string]int //执行计划，key为slots的key，value为所需状态：0=终止运行；1=启动；2=启动并保活
	pid  int            //子进程ID
}
