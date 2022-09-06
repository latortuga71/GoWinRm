package winapi

import (
	"errors"
	"fmt"
	"strconv"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	pModWsmSvc               = syscall.NewLazyDLL("WsmSvc.dll")
	pWSManInitialize         = pModWsmSvc.NewProc("WSManInitialize")
	pWSManCreateSession      = pModWsmSvc.NewProc("WSManCreateSession")
	pWSManSetSessionOption   = pModWsmSvc.NewProc("WSManSetSessionOption")
	pWSManCreateShell        = pModWsmSvc.NewProc("WSManCreateShell")
	pWSManRunShellCommand    = pModWsmSvc.NewProc("WSManRunShellCommand")
	pWSManReceiveShellOutput = pModWsmSvc.NewProc("WSManReceiveShellOutput")
	pWSManCloseOperation     = pModWsmSvc.NewProc("WSManCloseOperation")
	pWSManCloseCommand       = pModWsmSvc.NewProc("WSManCloseCommand")
	pWSManCloseShell         = pModWsmSvc.NewProc("WSManCloseShell")
	pWSManCloseSession       = pModWsmSvc.NewProc("WSManCloseSession")
	pWSManDeinitialize       = pModWsmSvc.NewProc("WSManDeinitialize")
)

func WSManInitialize(flag uint32, apiHandle *uintptr) error {
	res, _, err := pWSManInitialize.Call(uintptr(flag), uintptr(unsafe.Pointer(apiHandle)))
	if res != 0 {
		return err
	}
	return nil
}

func WSManCreateSession(apiHandle WSMAN_API_HANDLE, connection string, flags uint32, creds *WSMAN_AUTHENTICATION_CREDENTIALS, proxyInfo uintptr, session *WSMAN_SESSION_HANDLE) error {
	c, err := windows.UTF16PtrFromString(connection)
	if err != nil {
		return err
	}
	res, _, err := pWSManCreateSession.Call(uintptr(apiHandle), uintptr(unsafe.Pointer(c)), uintptr(flags), uintptr(unsafe.Pointer(creds)), proxyInfo, uintptr(unsafe.Pointer(session)))
	if res != 0 {
		return err
	}
	return nil
}

func WSManSetSessionOption(session WSMAN_SESSION_HANDLE, option WSManSessionOption, data *WSMAN_DATA) error {
	res, _, err := pWSManSetSessionOption.Call(uintptr(session), uintptr(option), uintptr(unsafe.Pointer(data)))
	if res != 0 {
		if syserr, ok := err.(syscall.Errno); ok {
			fmt.Printf("%x\n", uint64(syserr))
			return fmt.Errorf("Error Code 0x%s\n", strconv.FormatUint(uint64(syserr), 16))
		}
	}
	return nil
}

func WSManCreateShell(session WSMAN_SESSION_HANDLE, flags uint32, resourceUri string, startupInfo uintptr, options uintptr, createXml uintptr, async *WSMAN_SHELL_ASYNC, shell *WSMAN_SHELL_HANDLE) error {
	r, err := windows.UTF16PtrFromString(resourceUri)
	if err != nil {
		return err
	}
	pWSManCreateShell.Call(uintptr(session), uintptr(flags), uintptr(unsafe.Pointer(r)), 0, 0, 0, uintptr(unsafe.Pointer(async)), uintptr(unsafe.Pointer(shell)))
	if *shell == 0 {
		return errors.New("Failed to write to shell handle")
	}
	return nil
}

func WSManRunShellCommand(shell WSMAN_SHELL_HANDLE, flags uint32, commandLine string, args uintptr, options uintptr, async *WSMAN_SHELL_ASYNC, command *WSMAN_COMMAND_HANDLE) error {
	cmd, err := windows.UTF16PtrFromString(commandLine)
	if err != nil {
		return err
	}
	pWSManRunShellCommand.Call(uintptr(shell), uintptr(flags), uintptr(unsafe.Pointer(cmd)), 0, 0, uintptr(unsafe.Pointer(async)), uintptr(unsafe.Pointer(command)))
	if *command == 0 {
		return errors.New("Failed to write to command handle")
	}
	return nil
}

func WSManReceiveShellOutput(shell WSMAN_SHELL_HANDLE, command WSMAN_COMMAND_HANDLE, flags uint32, desiredStreame uintptr, async *WSMAN_SHELL_ASYNC, recvOperation *WSMAN_OPERATION_HANDLE) error {
	pWSManReceiveShellOutput.Call(uintptr(shell), uintptr(command), uintptr(flags), 0, uintptr(unsafe.Pointer(async)), uintptr(unsafe.Pointer(recvOperation)))
	if *recvOperation == 0 {
		return errors.New("Failed to write to operation handle")
	}
	return nil
}

type WSManSessionOption uint32
type WSManDataType uint32
type WSMAN_SHELL_STARTUP_INFO struct{}

const (
	WSMAN_CMDSHELL_URI = "http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd"
)

const (
	WSMAN_DATA_NONE        = 0
	WSMAN_DATA_TYPE_TEXT   = 1
	WSMAN_DATA_TYPE_BINARY = 2
	WSMAN_DATA_TYPE_DWORD  = 4
)

type WSMAN_DATA_TEXT struct {
	BufferLength uint32
	Buffer       *uint16
}

type WSMAN_DATA_BINARY struct {
	DataLength uint32
	Data       *byte
}

type WSMAN_DATA_BINARY_ENUM struct {
	Type       WSManDataType
	BinaryData WSMAN_DATA_BINARY
}

type WSMAN_DATA_DWORD_ENUM struct {
	Type   WSManDataType
	Number uint32
}

type WSMAN_DATA struct {
	Type       WSManDataType
	Text       WSMAN_DATA_TEXT
	BinaryData WSMAN_DATA_BINARY
	Number     uint32
}

type WSMAN_SHELL_ASYNC struct {
	OperationContext   uintptr
	CompletionFunction uintptr
}
type WSMAN_ERROR struct {
	Code        uint32
	ErrorDetail *uint16
	Language    *uint16
	MachineName *uint16
	PluginName  *uint16
}

type WSMAN_RESPONSE_DATA struct {
	receiveData WSMAN_RECEIVE_DATA_RESULT
	connectData uintptr
	createData  uintptr
}

const (
	//
	//Timeouts
	//
	WSMAN_OPTION_DEFAULT_OPERATION_TIMEOUTMS    = 1  // DWORD - default timeout in ms that applies to all operations on the client side
	WSMAN_OPTION_MAX_RETRY_TIME                 = 11 // DWORD (read only) - maximum time for Robust connection retries
	WSMAN_OPTION_TIMEOUTMS_CREATE_SHELL         = 12 // DWORD - timeout in ms for WSManCreateShell operations
	WSMAN_OPTION_TIMEOUTMS_RUN_SHELL_COMMAND    = 13 // DWORD - timeout in ms for WSManRunShellCommand operations
	WSMAN_OPTION_TIMEOUTMS_RECEIVE_SHELL_OUTPUT = 14 // DWORD - timeout in ms for WSManReceiveShellOutput operations
	WSMAN_OPTION_TIMEOUTMS_SEND_SHELL_INPUT     = 15 // DWORD - timeout in ms for WSManSendShellInput operations
	WSMAN_OPTION_TIMEOUTMS_SIGNAL_SHELL         = 16 // DWORD - timeout in ms for WSManSignalShell and WSManCloseCommand operations
	WSMAN_OPTION_TIMEOUTMS_CLOSE_SHELL          = 17 // DWORD - timeout in ms for WSManCloseShell operations

	//
	// connection options
	//

	WSMAN_OPTION_SKIP_CA_CHECK          = 18 // DWORD  - 1 to not validate the CA on the server certificate; 0 - default
	WSMAN_OPTION_SKIP_CN_CHECK          = 19 // DWORD  - 1 to not validate the CN on the server certificate; 0 - default
	WSMAN_OPTION_UNENCRYPTED_MESSAGES   = 20 // DWORD  - 1 to not encrypt the messages; 0 - default
	WSMAN_OPTION_UTF16                  = 21 // DWORD  - 1 Send all network packets for remote operatons in UTF16; 0 - default is UTF8
	WSMAN_OPTION_ENABLE_SPN_SERVER_PORT = 22 // DWORD  - 1 When using negotiate, include port number in the connection SPN; 0 - default
	// Used when not talking to the main OS on a machine but, for instance, a BMC
	WSMAN_OPTION_MACHINE_ID = 23 // DWORD  - 1 Identify this machine to the server by including the MachineID header; 0 - default

	//
	// other options
	//
	WSMAN_OPTION_LOCALE               = 25 // string - RFC 3066 language code
	WSMAN_OPTION_UI_LANGUAGE          = 26 // string - RFC 3066 language code
	WSMAN_OPTION_MAX_ENVELOPE_SIZE_KB = 28 // DWORD - max SOAP envelope size (kb) - default 150kb from winrm config
	// (see 'winrm help config' for more details); the client SOAP packet size cannot surpass
	//  this value; this value will be also sent to the server in the SOAP request as a
	//  MaxEnvelopeSize header; the server will use min(MaxEnvelopeSizeKb from server configuration,
	//  MaxEnvelopeSize value from SOAP).
	WSMAN_OPTION_SHELL_MAX_DATA_SIZE_PER_MESSAGE_KB = 29 // DWORD (read only) - max data size (kb) provided by the client, guaranteed by
	//  the winrm client implementation to fit into one SOAP packet; this is an
	// approximate value calculated based on the WSMAN_OPTION_MAX_ENVELOPE_SIZE_KB (default 500kb),
	// the maximum possible size of the SOAP headers and the overhead of the base64
	// encoding which is specific to WSManSendShellInput API; this option can be used
	// with WSManGetSessionOptionAsDword API; it cannot be used with WSManSetSessionOption API.
	WSMAN_OPTION_REDIRECT_LOCATION                    = 30 // string - read-only, cannot set
	WSMAN_OPTION_SKIP_REVOCATION_CHECK                = 31 // DWORD  - 1 to not validate the revocation status on the server certificate; 0 - default
	WSMAN_OPTION_ALLOW_NEGOTIATE_IMPLICIT_CREDENTIALS = 32 // DWORD  - 1 to allow default credentials for Negotiate (this is for SSL only); 0 - default
	WSMAN_OPTION_USE_SSL                              = 33 // DWORD - When using just a machine name in the connection string use an SSL connection. 0 means HTTP, 1 means HTTPS.  Default is 0.
	WSMAN_OPTION_USE_INTEARACTIVE_TOKEN               = 34 // DWORD - When creating connection on local machine, use interactive token feature. 1 - default
)

type WSMAN_API_HANDLE uintptr
type WSMAN_SESSION_HANDLE uintptr
type WSMAN_OPERATION_HANDLE uintptr
type WSMAN_SHELL_HANDLE uintptr
type WSMAN_COMMAND_HANDLE uintptr
type WSMAN_RECEIVE_DATA_RESULT struct {
	StreamId     *uint16
	StreamData   WSMAN_DATA_BINARY_ENUM
	CommandState *uint16
	ExitCode     uint32
}

type WSMAN_USERNAME_PASSWORD_CREDS struct {
	Username *uint16
	Password *uint16
}
type WSMAN_AUTHENTICATION_CREDENTIALS struct {
	AuthenticationMechanism uint32
	UserAccount             WSMAN_USERNAME_PASSWORD_CREDS
	CertificateThumbprint   uintptr
}

const (
	WSMAN_FLAG_REQUESTED_API_VERSION_1_0 = 0x0
	WSMAN_FLAG_DEFAULT_AUTHENTICATION    = 0x0 //Use the default authentication
	WSMAN_FLAG_NO_AUTHENTICATION         = 0x1 //Use no authentication for a remote operation
	WSMAN_FLAG_AUTH_DIGEST               = 0x2 //Use digest authentication for a remote operation
	WSMAN_FLAG_AUTH_NEGOTIATE            = 0x4 //Use negotiate authentication for a remote operation (may use kerberos or ntlm)
	WSMAN_FLAG_AUTH_BASIC                = 0x8 //Use basic authentication for a remote operation
	WSMAN_FLAG_AUTH_KERBEROS             = 0x10
)
