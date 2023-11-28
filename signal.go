package proxy

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// 监听信号
func (s *Server) goNotifySignal() {
	s.wg.Add(1)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	for {
		si := <-ch
		switch si {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Printf("get a signal %s, stop the process\n", si.String())
			s.Close()
			// wait for program finish processing
			log.Println("waiting for the process to finish 1 minute")
			time.Sleep(time.Minute)
			// notify process exit
			s.wg.Done()
			runtime.Gosched()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
