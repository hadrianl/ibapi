# Interactive Brokers API - GoLang Version
Interactive Brokers API 9.79
pure golang, Unofficial version, smilar to the official python version


## INSTALL
`go get -u github.com/hadrianl/ibapi`

## USAGE
```golang
import (
    . "github.com/hadrianl/ibapi"
    "time"
    log "github.com/sirupsen/logrus"
)

func main(){
    var err error
    ibwrapper := Wrapper{}
    ic := NewIbClient(ibwrapper)
    err = ic.Connect("172.0.0.1", 4002, 0)
    if err != nil {
        log.Panic("Connect failed:", err)
        return
    }

    err = ic.HandShake()
    if err != nil {
        log.Println("HandShake failed:", err)
        return
    }

    ic.ReqCurrentTime()
    ic.ReqAutoOpenOrders(true)
    ic.ReqAccountUpdates(true, "")
    ic.ReqExecutions(ic.GetReqID(), ExecutionFilter{})

    ic.Run()
    time.Sleep(time.Second * 10)
    ic.Disconnect()
}

```

## Reference 
1.[Offical Document](https://interactivebrokers.github.io/tws-api/) 