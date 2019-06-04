package ibapi

type IbError struct {
	code int64
	msg  string
}

func (ie IbError) Error() string {
	return ie.msg
}

var (
	ALREADY_CONNECTED   IbError
	CONNECT_FAIL        IbError
	UPDATE_TWS          IbError
	NOT_CONNECTED       IbError
	UNKNOWN_ID          IbError
	UNSUPPORTED_VERSION IbError
	BAD_LENGTH          IbError
	BAD_MESSAGE         IbError
	SOCKET_EXCEPTION    IbError
	FAIL_CREATE_SOCK    IbError
	SSL_FAIL            IbError
)

func init() {
	ALREADY_CONNECTED = IbError{501, "Already connected."}
	CONNECT_FAIL = IbError{502, `Couldn't connect to TWS. Confirm that "Enable ActiveX and Socket EClients" 
	is enabled and connection port is the same as "Socket Port" on the 
	TWS "Edit->Global Configuration...->API->Settings" menu. Live Trading ports: 
	TWS: 7496; IB Gateway: 4001. Simulated Trading ports for new installations 
	of version 954.1 or newer:  TWS: 7497; IB Gateway: 4002`}
	UPDATE_TWS = IbError{503, "The TWS is out of date and must be upgraded."}
	NOT_CONNECTED = IbError{504, "Not connected"}
	UNKNOWN_ID = IbError{505, "Fatal Error: Unknown message id."}
	UNSUPPORTED_VERSION = IbError{506, "Unsupported version"}
	BAD_LENGTH = IbError{507, "Bad message length"}
	BAD_MESSAGE = IbError{508, "Bad message"}
	SOCKET_EXCEPTION = IbError{509, "Exception caught while reading socket - "}
	FAIL_CREATE_SOCK = IbError{520, "Failed to create socket"}
	SSL_FAIL = IbError{530, "SSL specific error: "}

}
