package frigate

//@author Wang Weiwei
//@since 2020/3/25
var taskMap = make(map[string]*Frigate)

// 对进程发起守护
func Protect(f *Frigate) {
	taskMap[f.ProtectTask.Name] = f
	if f.Strategy.AutoStart {
		f.Start()
	}
}
