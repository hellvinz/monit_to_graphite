package main

import (
  "http"
  "log"
  "xml"
  "strconv"
  "net"
  "bytes"
  "fmt"
  "flag"
)


var carbonAddress *string = flag.String("c", "localhost:2003", "carbon address")
var forwarderAddress *string = flag.String("l", ":3005", "forwarder listening address")


type Server struct{
  Id string
  Incarnation int
  Version string
  Uptime int
  Poll int
  Localhostname string
}

type Platform struct {
  Name string
  Release string
  Version string
  Machine string
  Cpu int
  Memory int
}

type Memory struct {
  Percent float64
  Percenttotal float64
  Kilobyte int
  Kilobytetotal int
}

type Cpu struct {
  Percent float64
  Percenttotal float64
}

type Load struct {
  Avg01 float64
  Avg05 float64
  Avg15 float64
}

type Cpusys struct {
  User float64
  System float64
  Wait float64
}

type System struct {
  Load Load
  Cpusys Cpusys
  Memory Memory
}

type Service struct {
  Collected_Sec int64
  Type int `xml:"attr"`
  Name string
  Status int
  Monitor int
  MonitorMode int
  Pendingaction int
  Pid int
  Ppid int
  Uptime int
  Children int
  Memory Memory
  Cpu Cpu
  Sytem System
}

type Monit struct {
  XMLName xml.Name `xml:"monit"`
  Server Server
  Platform Platform
  Service []Service
}

type Graphite struct {
  addr string
}

var serviceq chan *Service

func (graphite *Graphite) Setup() {
  log.Println("starting")
  serviceq = make(chan *Service)
  for {
    service := <-serviceq
    if service.Type == 5 {
      continue
    }
    go graphite.Send(service.Name+".status", strconv.Itoa(service.Status), service.Collected_Sec)
    go graphite.Send(service.Name+".monitor", strconv.Itoa(service.Monitor), service.Collected_Sec)
    go graphite.Send(service.Name+".uptime", strconv.Itoa(service.Uptime), service.Collected_Sec)
    go graphite.Send(service.Name+".children", strconv.Itoa(service.Children), service.Collected_Sec)
    go graphite.Send(service.Name+".memory.percent", strconv.FtoaN(service.Memory.Percent,'g', -1, 2), service.Collected_Sec)
    go graphite.Send(service.Name+".memory.percent_total", strconv.FtoaN(service.Memory.Percenttotal,'g', -1, 2), service.Collected_Sec)
    go graphite.Send(service.Name+".memory.kilobyte", strconv.Itoa(service.Memory.Kilobyte), service.Collected_Sec)
    go graphite.Send(service.Name+".memory.kylobytetotal", strconv.Itoa(service.Memory.Kilobytetotal), service.Collected_Sec)
    go graphite.Send(service.Name+".cpu.percent", strconv.FtoaN(service.Cpu.Percent,'g', -1, 2), service.Collected_Sec)
    go graphite.Send(service.Name+".cpu.percenttotal", strconv.FtoaN(service.Cpu.Percenttotal, 'g', -1, 2), service.Collected_Sec)
  }
}

func (graphite *Graphite) Send(metric string, value string, timestamp int64) {
  conn, err := net.Dial("tcp", graphite.addr)
  if err != nil {
    switch t := err.(type){
    default:
      log.Fatal(err)
    case *net.OpError:
      //retry
      conn, _ = net.Dial("tcp", graphite.addr)
    }
  }
  defer conn.Close()
  buffer := bytes.NewBufferString("")
  fmt.Fprintf(buffer, "monit.%s %s %d\n", metric, value, timestamp)
	conn.Write(buffer.Bytes())
}

func MonitServer(w http.ResponseWriter, req *http.Request) {
  defer req.Body.Close()
  var monit Monit
  p := xml.NewParser(req.Body)
  p.CharsetReader = CharsetReader
  err := p.Unmarshal(&monit, nil)
  if err != nil {
    log.Fatal(err)
  }
  log.Println("Got message from", monit.Server.Localhostname)
  for _, service := range monit.Service {
    serviceq <- &service
  }
}

func main(){
  log.Println("Forwarding m/monit to ", *carbonAddress)
  graphite := Graphite{addr: *carbonAddress}
  go graphite.Setup()

	http.HandleFunc("/collector", MonitServer)
  log.Println("Forwarder listening input on: ", *forwarderAddress)
  err := http.ListenAndServe(*forwarderAddress, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.String())
	}
}
