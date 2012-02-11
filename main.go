package main

import (
  "http"
  "log"
  "xml"
  "strconv"
  "net"
  "bytes"
  "fmt"
  //  "io"
)


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

func SendToGraphite(metric string, value string, timestamp int64) {
  client, err := net.Dial("tcp", "localhost:2003")
  if err != nil {
    log.Fatal(err)
  }
  defer client.Close()
  buffer := bytes.NewBufferString("")
  fmt.Fprintf(buffer, "monit.%s %s %d\n", metric, value, timestamp)
	client.Write(buffer.Bytes())
}

func MonitServer(w http.ResponseWriter, req *http.Request) {
  var monit Monit
  p := xml.NewParser(req.Body)
  p.CharsetReader = CharsetReader
  err := p.Unmarshal(&monit, nil)
  if err != nil {
    log.Fatal(err)
  }
  for _, service := range monit.Service {
    if service.Type == 5 {
      continue
    }
    go SendToGraphite(service.Name+".status", strconv.Itoa(service.Status), service.Collected_Sec)
    go SendToGraphite(service.Name+".monitor", strconv.Itoa(service.Monitor), service.Collected_Sec)
    go SendToGraphite(service.Name+".uptime", strconv.Itoa(service.Uptime), service.Collected_Sec)
    go SendToGraphite(service.Name+".children", strconv.Itoa(service.Children), service.Collected_Sec)
    go SendToGraphite(service.Name+".memory.percent", strconv.FtoaN(service.Memory.Percent,'g', -1, 2), service.Collected_Sec)
    go SendToGraphite(service.Name+".memory.percent_total", strconv.FtoaN(service.Memory.Percenttotal,'g', -1, 2), service.Collected_Sec)
    go SendToGraphite(service.Name+".memory.kilobyte", strconv.Itoa(service.Memory.Kilobyte), service.Collected_Sec)
    go SendToGraphite(service.Name+".memory.kylobytetotal", strconv.Itoa(service.Memory.Kilobytetotal), service.Collected_Sec)
    go SendToGraphite(service.Name+".cpu.percent", strconv.FtoaN(service.Cpu.Percent,'g', -1, 2), service.Collected_Sec)
    go SendToGraphite(service.Name+".cpu.percenttotal", strconv.FtoaN(service.Cpu.Percenttotal, 'g', -1, 2), service.Collected_Sec)
  }
}

func main(){
	http.HandleFunc("/collector", MonitServer)
  err := http.ListenAndServe(":3005", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.String())
	}
}
