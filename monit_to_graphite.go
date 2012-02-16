package main

import (
    "bytes"
    "encoding/xml"
    "flag"
    "fmt"
    "log"
    "net"
    "net/http"
    "strconv"
    "os"
    "errors"
)

var carbonAddress *string = flag.String("c", "localhost:2003", "carbon address")
var forwarderAddress *string = flag.String("l", ":3005", "forwarder listening address")

var ErrHelp = errors.New("flag: help requested")
var Usage = func() {
    fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
    flag.PrintDefaults()
}

type Server struct {
    Id            string `xml:"id"`
    Incarnation   int `xml:"incarnation"`
    Version       string `xml:"version"`
    Uptime        int `xml:"uptime"`
    Poll          int `xml:"poll"`
    Localhostname string `xml:"localhostname"`
}

type Platform struct {
    Name    string `xml:"name"`
    Release string `xml:"release"`
    Version string `xml:"version"`
    Machine string `xml:"machine"`
    Cpu     int `xml:"cpu"`
    Memory  int `xml:"memory"`
}

type Memory struct {
    Percent       float64 `xml:"percent"`
    Percenttotal  float64 `xml:"percenttotal"`
    Kilobyte      int `xml:"kylobyte"`
    Kilobytetotal int `xml:"kilobytetotal"`
}

type Cpu struct {
    Percent      float64 `xml:"percent"`
    Percenttotal float64 `xml:"percenttotal"`
}

type Load struct {
    Avg01 float64 `xml:"avg01"`
    Avg05 float64 `xml:"avg05"`
    Avg15 float64 `xml:"avg15"`
}

type Cpusys struct {
    User   float64 `xml:"user"`
    System float64 `xml:"system"`
    Wait   float64 `xml:"wait"`
}

type System struct {
    Load   Load `xml:"load"`
    Cpusys Cpusys `xml:"cpusys"`
    Memory Memory `xml:"memory"`
}

type Service struct {
    Collected_Sec int64 `xml:"collected_sec"`
    Type          int `xml:"attr"`
    Name          string `xml:"name"`
    Status        int `xml:"status"`
    Monitor       int `xml:"monitor"`
    MonitorMode   int `xml:"monitormode"`
    Pendingaction int `xml:"pendingaction"`
    Pid           int `xml:"pid"`
    Ppid          int `xml:"ppid"`
    Uptime        int `xml:"uptime"`
    Children      int `xml:"children"`
    Memory        Memory `xml:"memory"`
    Cpu           Cpu `xml:"cpu"`
    Sytem         System `xml:"system"`
}

type Monit struct {
    XMLName  xml.Name `xml:"monit"`
    Server   Server `xml:"server"`
    Platform Platform `xml:"platform"`
    Service  []Service `xml:"service"`
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
        go graphite.Send(service.Name+".memory.percent", strconv.FormatFloat(service.Memory.Percent, 'g', -1, 64), service.Collected_Sec)
        go graphite.Send(service.Name+".memory.percent_total", strconv.FormatFloat(service.Memory.Percenttotal, 'g', -1, 64), service.Collected_Sec)
        go graphite.Send(service.Name+".memory.kilobyte", strconv.Itoa(service.Memory.Kilobyte), service.Collected_Sec)
        go graphite.Send(service.Name+".memory.kylobytetotal", strconv.Itoa(service.Memory.Kilobytetotal), service.Collected_Sec)
        go graphite.Send(service.Name+".cpu.percent", strconv.FormatFloat(service.Cpu.Percent, 'g', -1, 64), service.Collected_Sec)
        go graphite.Send(service.Name+".cpu.percenttotal", strconv.FormatFloat(service.Cpu.Percenttotal, 'g', -1, 64), service.Collected_Sec)
    }
}

func (graphite *Graphite) Send(metric string, value string, timestamp int64) {
    var conn net.Conn
    var err error
    for i:=0; i<=5; i++ {
        conn, err = net.Dial("tcp", graphite.addr)
        if i==5 {
            log.Fatal(err)
        }
        if conn == nil {
            switch err.(type) {
            default:
                log.Fatal(err)
            case *net.OpError:
                continue
            }
        } else {
            defer conn.Close()
            break
        }
    }
    buffer := bytes.NewBufferString("")
    fmt.Fprintf(buffer, "monit.%s %s %d\n", metric, value, timestamp)
    conn.Write(buffer.Bytes())
}

func MonitServer(w http.ResponseWriter, req *http.Request) {
    defer req.Body.Close()
    var monit Monit
    p := xml.NewDecoder(req.Body)
    //b := new(bytes.Buffer)
    //b.ReadFrom(req.Body)
    //log.Fatal(b.String())
    p.CharsetReader = CharsetReader
    err := p.DecodeElement(&monit, nil)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Got message from", monit.Server.Localhostname)
    for _, service := range monit.Service {
        serviceq <- &service
    }
}

func main() {
    flag.Parse()
    log.Println("Forwarding m/monit to ", *carbonAddress)
    graphite := Graphite{addr: *carbonAddress}
    go graphite.Setup()

    http.HandleFunc("/collector", MonitServer)
    log.Println("Forwarder listening input on: ", *forwarderAddress)
    err := http.ListenAndServe(*forwarderAddress, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
