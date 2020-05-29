# Interactive Brokers API - GoLang Implement
Interactive Brokers API 9.79
pure golang, Unofficial, smilar to the official python Implement


## INSTALL
`go get -u github.com/hadrianl/ibapi`

## USAGE
### Demo 1
```golang
import (
    . "github.com/hadrianl/ibapi"
    "time"
)

func main(){
    var err error
    log := GetLogger().Sugar()
    defer log.Sync()
    ibwrapper := &Wrapper{}
    ic := NewIbClient(ibwrapper)
    err = ic.Connect("127.0.0.1", 4002, 0)
    if err != nil {
        log.Panic("Connect failed:", err)
        return
    }

    err = ic.HandShake()
    if err != nil {
        log.Info("HandShake failed:", err)
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

### Demo 2 with context 
```golang
import (
    . "github.com/hadrianl/ibapi"
    "time"
    "context"
)

func main(){
    var err error
    log := GetLogger().Sugar()
    defer log.Sync()
    ibwrapper := &Wrapper{}
    ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
    ic := NewIbClient(ibwrapper)
    ic.SetContext(ctx)
    err = ic.Connect("127.0.0.1", 4002, 0)
    if err != nil {
        log.Panic("Connect failed:", err)
        return
    }

    err = ic.HandShake()
    if err != nil {
        log.Error("HandShake failed:", err)
        return
    }

    ic.ReqCurrentTime()
    ic.ReqAutoOpenOrders(true)
    ic.ReqAccountUpdates(true, "")
    ic.ReqExecutions(ic.GetReqID(), ExecutionFilter{})

    ic.Run()
    
    err = ic.LoopUntilDone()  // block until ctx or ic is done, also, 
	log.Info(err)
}

```

## Reference 
1.[Offical Document](https://interactivebrokers.github.io/tws-api/) 