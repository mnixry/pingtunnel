package main

import (
	"flag"
	"fmt"
	"github.com/mnixry/pingtunnel/src/common"
	"github.com/mnixry/pingtunnel/src/loggo"
	pingtunnel2 "github.com/mnixry/pingtunnel/src/pingtunnel"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"time"
)

func main() {
	defer common.CrashLog()

	t := flag.String("type", "", "client or server")
	listen := flag.String("l", "", "listen addr")
	target := flag.String("t", "", "target addr")
	server := flag.String("s", "", "server addr")
	timeout := flag.Int("timeout", 60, "conn timeout")
	key := flag.Int("key", 0, "key")
	tcpMode := flag.Int("tcp", 0, "tcp mode")
	tcpmodeBuffersize := flag.Int("tcp_bs", 1*1024*1024, "tcp mode buffer size")
	tcpmode_maxwin := flag.Int("tcp_mw", 20000, "tcp mode max win")
	tcpmode_resend_timems := flag.Int("tcp_rst", 400, "tcp mode resend time ms")
	tcpmode_compress := flag.Int("tcp_gz", 0, "tcp data compress")
	tcpmode_stat := flag.Int("tcp_stat", 0, "print tcp stat")
	loglevel := flag.String("loglevel", "info", "log level")
	open_sock5 := flag.Int("sock5", 0, "sock5 mode")
	maxconn := flag.Int("maxconn", 0, "max num of connections")
	max_process_thread := flag.Int("maxprt", 100, "max process thread in server")
	max_process_buffer := flag.Int("maxprb", 1000, "max process thread's buffer in server")
	profile := flag.Int("profile", 0, "open profile")
	conntt := flag.Int("conntt", 1000, "the connect call's timeout")
	flag.Parse()

	if *t != "client" && *t != "server" {
		flag.Usage()
		return
	}
	if *t == "client" {
		if len(*listen) == 0 || len(*server) == 0 {
			flag.Usage()
			return
		}
		if *open_sock5 == 0 && len(*target) == 0 {
			flag.Usage()
			return
		}
		if *open_sock5 != 0 {
			*tcpMode = 1
		}
	}
	if *tcpmode_maxwin*10 > pingtunnel2.FRAME_MAX_ID {
		fmt.Println("set tcp win to big, max = " + strconv.Itoa(pingtunnel2.FRAME_MAX_ID/10))
		return
	}

	level := loggo.LEVEL_INFO
	if loggo.NameToLevel(*loglevel) >= 0 {
		level = loggo.NameToLevel(*loglevel)
	}
	loggo.Ini(loggo.Config{
		Level:     level,
		Prefix:    "pingtunnel",
		MaxDay:    3,
		NoLogFile: true,
		NoPrint:   true,
	})
	loggo.Info("start...")
	loggo.Info("key %d", *key)

	if *t == "server" {
		s, err := pingtunnel2.NewServer(*key, *maxconn, *max_process_thread, *max_process_buffer, *conntt)
		if err != nil {
			loggo.Error("ERROR: %s", err.Error())
			return
		}
		loggo.Info("Server start")
		err = s.Run()
		if err != nil {
			loggo.Error("Run ERROR: %s", err.Error())
			return
		}
	} else if *t == "client" {

		loggo.Info("type %s", *t)
		loggo.Info("listen %s", *listen)
		loggo.Info("server %s", *server)
		loggo.Info("target %s", *target)

		if *tcpMode == 0 {
			*tcpmodeBuffersize = 0
			*tcpmode_maxwin = 0
			*tcpmode_resend_timems = 0
			*tcpmode_compress = 0
			*tcpmode_stat = 0
		}

		filter := func(addr string) bool {
			return true
		}

		c, err := pingtunnel2.NewClient(*listen, *server, *target, *timeout, *key,
			*tcpMode, *tcpmodeBuffersize, *tcpmode_maxwin, *tcpmode_resend_timems, *tcpmode_compress,
			*tcpmode_stat, *open_sock5, *maxconn, &filter)
		if err != nil {
			loggo.Error("ERROR: %s", err.Error())
			return
		}
		loggo.Info("Client Listen %s (%s) Server %s (%s) TargetPort %s:", c.Addr(), c.IPAddr(),
			c.ServerAddr(), c.ServerIPAddr(), c.TargetAddr())
		err = c.Run()
		if err != nil {
			loggo.Error("Run ERROR: %s", err.Error())
			return
		}
	} else {
		return
	}

	if *profile > 0 {
		go func() {
			err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(*profile), nil)
			if err != nil {
				panic(err)
			}
		}()
	}

	for {
		time.Sleep(time.Hour)
	}
}
