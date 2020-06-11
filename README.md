# Interactive Brokers API - GoLang Implement

<img
style="display: block; margin: 0 auto;"
src="http://interactivebrokers.github.io/tws-api/nav_iblogo.png"
/>

* Interactive Brokers API 9.79
* pure golang Implement
* Unofficial, use at you own risk

## INSTALL

`go get -u github.com/hadrianl/ibapi`

---

## USAGE

Implement `IbWrapper` to handle datas delivered via tws or gateway, `Wrapper` in demo is a default implement that just output data to std.
[Go to IbWrapper](https://github.com/hadrianl/ibapi/blob/83846bf1194bbdc4f039c8c66033f717e015e9fc/wrapper.go#L11)

1. **implement** your own `IbWrapper`
2. **connect** to TWS or Gateway
3. **handshake** with TWS or Gateway
4. **run** the loop
5. do some **request**

### Demo 1

```golang
import (
    . "github.com/hadrianl/ibapi"
    "time"
)

func main(){
    // internal api log is zap log, you could use GetLogger to get the logger
    // besides, you could use SetAPILogger to set you own log option
    // or you can just use the other logger  
    log := GetLogger().Sugar()
    defer log.Sync()
    // implement your own IbWrapper to handle the msg delivered via tws or gateway
    // Wrapper{} below is a default implement which just log the msg 
    ic := NewIbClient(&Wrapper{})

    // tcp connect with tws or gateway
    // fail if tws or gateway had not yet set the trust IP
    if err := ic.Connect("127.0.0.1", 4002, 0);err != nil {
        log.Panic("Connect failed:", err)
    }

    // handshake with tws or gateway, send handshake protocol to tell tws or gateway the version of client and receive the server version and connection time from tws or gateway
    // fail if someone else had already connected to tws or gateway with same clientID
    if err := ic.HandShake();err != nil {
        log.Panic("HandShake failed:", err)
    }

    // make some request, msg would be delivered via wrapper.
    // req will not send to TWS or Gateway until ic.Run()
    // you could just call ic.Run() before these
    ic.ReqCurrentTime()
    ic.ReqAutoOpenOrders(true)
    ic.ReqAccountUpdates(true, "")
    ic.ReqExecutions(ic.GetReqID(), ExecutionFilter{})

    // start to send req and receive msg from tws or gateway after this
    ic.Run()
    <-time.After(time.Second * 60):
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
    }

    err = ic.HandShake()
    if err != nil {
        log.Panic("HandShake failed:", err)
    }

    ic.ReqCurrentTime()
    ic.ReqAutoOpenOrders(true)
    ic.ReqAccountUpdates(true, "")
    ic.ReqExecutions(ic.GetReqID(), ExecutionFilter{})

    ic.Run()
    err = ic.LoopUntilDone()  // block until ctx or ic is done
    log.Info(err)
}

```

---

## Reference

1. [Offical Document](https://interactivebrokers.github.io/tws-api/)
2. [Order Types](https://www.interactivebrokers.com/en/index.php?f=4985)
3. [Product](https://www.interactivebrokers.com/en/index.php?f=4599)
4. [Margin](https://www.interactivebrokers.com/en/index.php?f=24176)
5. [Market Data](https://www.interactivebrokers.com/en/index.php?f=14193)
6. [Commissions](https://www.interactivebrokers.com/en/index.php?f=1590)
