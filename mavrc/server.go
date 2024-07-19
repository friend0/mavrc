package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bluenviron/mavp2p/mavp2p"
	"github.com/nats-io/nats.go"
)

var start_port int = 14550

func remoteMavlinkHandler(msg *nats.Msg) {
	fmt.Printf("Received message on remoteHandler: %s\n", string(msg.Data))
	// todo: use current start_port value to create a new mavp2p program
	endpoints := []string{"udpc:192.168.1.129:14550", "udps:0.0.0.0:14550"}
	// create a new mavp2p program
	program, error := mavp2p.NewProgram(endpoints, nil)
	if error != nil {
		fmt.Printf("Error creating new program: %s\n", error)
		return
	}
	go program.Run()

	start_port++
}

func remoteTXHandler(msg *nats.Msg) {
	fmt.Printf("Received message on remoteTXHandler: %s\n", string(msg.Data))
}

func controlMavlinkHandler(msg *nats.Msg) {
	fmt.Printf("Received message on controllerHandler: %s\n", string(msg.Data))
}

// controlTXHandler processes control commands to be sent over Serial to a CyberTX device.
// CyberTX is a serial to PPM converter that can be used to control a drone via standard RC transmitter.
func controlTXHandler(msg *nats.Msg) {
	fmt.Printf("Received message on controllerTXHandler: %s\n", string(msg.Data))
}

func usage() {
	log.Printf("Usage: nats-qsub [-s server] [-creds file] [-nkey file] [-t] <subject> <queue>\n")
	flag.PrintDefaults()
}

func showUsageAndExit(exitcode int) {
	usage()
	os.Exit(exitcode)
}

func printMsg(m *nats.Msg, i int) {
	log.Printf("[#%d] Received on [%s] Queue[%s] Pid[%d]: '%s'", i, m.Subject, m.Sub.Queue, os.Getpid(), string(m.Data))
}

func main() {

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	// Connect Options.
	opts := []nats.Option{nats.Name("NATS RC Queue Subscriber"), nats.Timeout(10 * time.Second)}
	opts = setupConnOptions(opts)

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL, opts...)
	if err != nil {
		log.Fatal(err)
	}

	queue := "mavrc"
	nc.QueueSubscribe("mavrc.remote", queue, remoteMavlinkHandler)
	nc.QueueSubscribe("mavrc.control", queue, controlMavlinkHandler)
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on [%s], queue group [%s]", "mavrc.remote and mavrc.control", queue)

	// Setup the interrupt handler to drain so we don't miss
	// requests when scaling down.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println()
	log.Printf("Draining...")
	nc.Drain()
	log.Fatalf("Exiting")
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Printf("Disconnected due to: %s, will attempt reconnects for %.0fm", err, totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("Exiting: %v", nc.LastError())
	}))
	return opts
}
