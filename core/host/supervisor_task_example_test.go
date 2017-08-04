// white-box testing
package host

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/go-siris/siris/configuration"
)

const (
	banner = `
  _________.___ __________ .___   _________
 /   _____/|   |\______   \|   | /   _____/
 \_____  \ |   | |       _/|   | \_____  \
 /        \|   | |    |   \|   | /        \
/_______  /|___| |____|_  /|___|/_______  /
        \/              \/              \/
         the fastest webframework
`

	// Version is the current version number of the Siris Web framework.
	//
	// Look https://github.com/go-siris/siris#where-can-i-find-older-versions for older versions.
	Version = "7.3.4"
)

func ExampleSupervisor_RegisterOnErrorHook() {
	config := configuration.DefaultConfiguration()
	su := New(&http.Server{Addr: ":8273", Handler: http.DefaultServeMux}, &config)

	su.RegisterOnErrorHook(func(err error) {
		fmt.Println(err.Error())
	})

	su.RegisterOnErrorHook(func(err error) {
		fmt.Println(err.Error())
	})

	su.RegisterOnErrorHook(func(err error) {
		fmt.Println(err.Error())
	})

	go su.ListenAndServe()
	time.Sleep(1 * time.Second)
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	su.Shutdown(ctx)
	time.Sleep(1 * time.Second)

	// Output:
	// http: Server closed
	// http: Server closed
	// http: Server closed
}

type myTestTask struct {
	restartEvery time.Duration
	maxRestarts  int
	logger       *log.Logger
}

func (m myTestTask) OnServe(host TaskHost) {
	host.Supervisor.DeferFlow() // don't exit on underline server's Shutdown.

	ticker := time.NewTicker(m.restartEvery)
	defer ticker.Stop()
	rans := 0
	for {
		select {
		case _, ok := <-ticker.C:
			{
				if !ok {
					m.logger.Println("ticker issue, closed channel, exiting from this task...")
					return
				}
				exitAfterXRestarts := m.maxRestarts
				if rans == exitAfterXRestarts {
					m.logger.Println("exit")
					ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
					defer cancel()
					host.Supervisor.Shutdown(ctx) // total shutdown
					host.Supervisor.RestoreFlow() // free to exit (if shutdown)
					return
				}

				rans++

				m.logger.Println(fmt.Sprintf("closed %d times", rans))
				host.Shutdown(context.TODO())

				startDelay := 2 * time.Second
				time.AfterFunc(startDelay, func() {
					m.logger.Println("restart")
					host.Serve() // restart

				})

			}
		}
	}
}

func ExampleSupervisor_RegisterOnServeHook() {
	config := configuration.DefaultConfiguration()
	h := New(&http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		}),
	}, &config)

	logger := log.New(os.Stdout, "Supervisor: ", 0)

	mytask := myTestTask{
		restartEvery: 6 * time.Second,
		maxRestarts:  2,
		logger:       logger,
	}

	h.RegisterOnServeHook(WriteStartupLogOnServe(func() *zap.SugaredLogger { log, _ := zap.NewDevelopment(); return log.Sugar() }(), banner+"V"+Version))

	h.RegisterOnServeHook(mytask.OnServe)

	ln, err := net.Listen("tcp4", ":9394")
	if err != nil {
		panic(err.Error())
	}

	logger.Println("server started...")
	h.Serve(ln)

	// Output:
	// Supervisor: server started...
	// Supervisor: closed 1 times
	// Supervisor: restart
	// Supervisor: closed 2 times
	// Supervisor: restart
	// Supervisor: exit
}
