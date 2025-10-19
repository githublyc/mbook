package channel

import (
	"testing"
	"time"
)

func TestChannel(t *testing.T) {
	// 声明
	//var ch chan struct{}
	// 声明并创建
	// 没有指定缓冲大小，默认容量是 0
	//ch1 := make(chan int)
	// 这种是带 buffer 的
	ch2 := make(chan int, 3)
	// 把 123 发送到 ch2 里面
	ch2 <- 123

	data := <-ch2
	t.Log(data)
	// 这个是关闭 channel
	close(ch2)
}
func TestChannelBlocking(t *testing.T) {
	ch := make(chan int)
	b1 := BigStruct{}
	go func() {
		var b BigStruct
		// 这个就是goroutine 泄露
		ch <- 123
		t.Log(b, b1)
	}()
}

type BigStruct struct {
	//
}

func TestChannelSelect(t *testing.T) {
	ch1 := make(chan int, 1)
	ch2 := make(chan int, 2)
	go func() {
		time.Sleep(time.Second)
		ch1 <- 123
	}()

	go func() {
		time.Sleep(time.Second)
		ch2 <- 123
	}()
	select {
	case val := <-ch1:
		t.Log("进来了 ch1 这里", val)
	case val := <-ch2:
		t.Log("进来了 ch2 这里", val)
	}
	//
}
