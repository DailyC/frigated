package frigate

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/pkg/reexec"
)

//@author Wang Weiwei
//@since 2020/3/24
// 历史注册过的golang函数
var registeredInitializers = make(map[string]func())

const (
	CANCEL_PROCESS string = "cancel"
	COMPLETE_PROCESS string = "complete"
)
// 受保护任务定义
type ProtectTask struct {
	// the task run in child process
	Cmd       *exec.Cmd
	Name      string
	Process   *os.Process
	StartTime time.Time
	// 任务相关信号管道
	signalChan chan error
	// 信号管道是否被正常关闭
	isCloseSC bool
}



// 注册golang 任务函数，如果不注册golang函数，接下来
// 在执行golang任务函数之前需要先对任务函数进行注册
// 因为golang使用了pthread，所以不能正常使用fork()函数。故而强烈的推荐您，在项目
// go没有类似C中的fork调用可以达到在fork之后根据返回的pid然后进入不同的函数的方案。原因主要是：
//
//fork 早出现在只有进程，没有线程的年代
//C中是自行控制线程，这样fork之后才不会发生紊乱。一般都是单线程fork之后，才会开始多线程执行。
//Go中多线程是runtime自行决定的，所以Go中没有提供单纯的fork，而是fork之后立即就exec执行新的二进制文件
//
// 下面使用一个实例程序演示一下 RegisterGolangTask 函数的使用规范:
// 强烈推荐在init（）函数里使用 RegisterGolangTask
//
//func init()  {
//	frigate.RegisterGolangTask("childTask", child)
//}
//func child()  {
//	println("child pid := " + strconv.Itoa(syscall.Getpid()))
//}
//func main()  {
//	// 使用刚刚注册的函数名声明一个 frigate
//	f := frigate.Create("childTask")
//	f.Protect()
//	println("parent pid = " + strconv.Itoa(syscall.Getpid()))
//	// 等待所有被保护的子进程执行完成后自身退出
//	frigate.Done()
//}
func RegisterGolangTask(name string, task func()) {
	registeredInitializers[name] = task
	reexec.Register(name, func() {
		task()
		os.Exit(0)
	})
	if reexec.Init() {
		os.Exit(0)
	}
}

// 创建受守护的进程执行对象
// name : 如果是golang 的函数，则代表已经注册的函数名
// 如果是外部可执行程序，则代表程序绝对路径或相对路径
func NewProtectTask(path string) *ProtectTask {
	if _, ok := registeredInitializers[path]; ok {
		return newGolangTask(path)
	}
	return newExecTask(path)
}

//基于可执行文件构件任务子进程
//path 代表可执行文件的绝对路径
func newExecTask(path string) *ProtectTask {
	paths := strings.Split(path, string(os.PathSeparator))
	return &ProtectTask{
		Cmd: func() *exec.Cmd {
			if runtime.GOOS == "windows" {
				return exec.Command(path)
			} else {
				return exec.Command(path)
			}
		}(),
		Name:      paths[len(path)-1],
		Process:   nil,
		StartTime: time.Now(),
		signalChan: make(chan error, 1),
		isCloseSC: false,
	}
}

// 基于golang的函数，构造新的任务子进程
// name 代表 frigate 守护进程的标识，frigate 将使用这个名字创建子进程
// task 是要作为子进程启动的函数。注意不要忘记对函数引用的变量做初始化。
// 注意函数执行完成要使子进程自动退出
func newGolangTask(name string) *ProtectTask {
	return &ProtectTask{
		Cmd:       reexec.Command(name),
		Name:      name,
		Process:   nil,
		StartTime: time.Now(),
		signalChan: make(chan error, 1),
		isCloseSC: false,
	}
}

/**
 * 获取进程信号通道
*/
func (t *ProtectTask) Done() <-chan error{
	return t.signalChan
}


/**
 * close the channel od signal
*/
func (t *ProtectTask) closeSC(f func ()) {
	if !t.isCloseSC {
		f()
		close(t.signalChan)
		t.isCloseSC = true
	}
}

// Start
/**
 *  启动进程
 * 启动进程后，需要主动wait，等待子进程结束，接收SINGCHILD信号，否则子进程可能变成僵尸进程
*/
func (t *ProtectTask) Start() (err error) {
	t.StartTime = time.Now()
	err = t.Cmd.Start()
	if err != nil {
		return err
	}
	
	t.Process = t.Cmd.Process
	go func() {
		err = t.Cmd.Wait()
		t.closeSC(func () {
			if err == nil {
				t.signalChan <- errors.New(COMPLETE_PROCESS)
			}else {
				t.signalChan <- err
			}
		})
	}()
	return nil
}


// Stop 用户主动关闭进程
//
// 优先使用 SIGTERM 信号量优雅关闭进程
// 如果超时后还未能正常关闭进程，则使用 kill 信号强行关闭进程
// d 关闭进程最大等大时间
func (t *ProtectTask) Stop(d time.Duration) (err error) {
	t.closeSC(func () {
		t.signalChan <- errors.New(CANCEL_PROCESS)
	})
	err = t.Process.Signal(syscall.SIGTERM)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(d))
	go func() {
		states, err1 := t.Process.Wait()
		if err1 != nil {
			err = err1
			cancel()
			return
		}
		if states.Exited() {
			cancel()
			
		}
	}()
	select {
	case <- ctx.Done() :{
		// 超时 或 关闭出错，都尝试kill
		if ctx.Err() != nil || err != nil {
			err = t.Process.Kill()
			return err	
		}
	}
	}
	return err
}