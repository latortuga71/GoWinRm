package main

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/latortuga71/GoWinRm/pkg/winapi"
	"golang.org/x/sys/windows"
)

// globals

var hEvent uintptr
var hReceiveEvent uintptr

func readBuffer(start *byte, length uint32) []byte {
	buff := make([]byte, length)
	for x := 0; uint32(x) < length; x++ {
		buff[x] = *start
		start = (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(start)) + 1))
	}
	return buff
}

func recvCallback(operationContext uintptr, flags uint32, err *winapi.WSMAN_ERROR, shell winapi.WSMAN_SHELL_HANDLE, commandHandle winapi.WSMAN_COMMAND_HANDLE, operationHandle winapi.WSMAN_OPERATION_HANDLE, data *winapi.WSMAN_RECEIVE_DATA_RESULT) uintptr {
	var e error
	if err != nil && err.Code != 0 {
		errorStr := windows.UTF16PtrToString(err.ErrorDetail)
		fmt.Printf("Error -> %s\n", errorStr)
		return 0
	}
	if data != nil && data.StreamData.Type == winapi.WSMAN_DATA_TYPE_BINARY && data.StreamData.BinaryData.DataLength > 0 {
		buff := readBuffer(data.StreamData.BinaryData.Data, data.StreamData.BinaryData.DataLength)
		hFile, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
		if err != nil {
			log.Fatal(err)
		}
		var nWrote uint32
		windows.WriteFile(hFile, buff, &nWrote, nil)
		return 0
	}
	if (err != nil && err.Code != 0) || (data != nil) {
		state := windows.UTF16PtrToString(data.CommandState)
		if state == "http://schemas.microsoft.com/wbem/wsman/1/windows/shell/CommandState/Done" {
			e = winapi.SetEvent(hReceiveEvent)
			if e != nil {
				log.Fatal(err)
			}
		}
	}
	return 0
}

func shellCallback(operationContext uintptr, flags uint32, err *winapi.WSMAN_ERROR, shell winapi.WSMAN_SHELL_HANDLE, commandHandle winapi.WSMAN_COMMAND_HANDLE, operationHandle winapi.WSMAN_OPERATION_HANDLE, data *winapi.WSMAN_RECEIVE_DATA_RESULT) uintptr {
	if err.Code != 0 {
		errorStr := windows.UTF16PtrToString(err.ErrorDetail)
		fmt.Printf("Error -> %s\n", errorStr)
	}
	e := winapi.SetEvent(hEvent)
	if e != nil {
		log.Fatal(e)
	}
	return 0
}

func main() {
	var clientHandle uintptr
	err := winapi.WSManInitialize(winapi.WSMAN_FLAG_REQUESTED_API_VERSION_1_0, &clientHandle)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[+] Initialized WSMan")
	authCreds := winapi.WSMAN_AUTHENTICATION_CREDENTIALS{}
	authCreds.UserAccount = winapi.WSMAN_USERNAME_PASSWORD_CREDS{}
	username, err := windows.UTF16PtrFromString("HACKERLAB\\turtleadmin")
	if err != nil {
		log.Fatal(err)
	}
	password, err := windows.UTF16PtrFromString("dawoof7123!!!")
	if err != nil {
		log.Fatal(err)
	}
	authCreds.UserAccount.Username = username
	authCreds.UserAccount.Password = password
	authCreds.AuthenticationMechanism = winapi.WSMAN_FLAG_AUTH_NEGOTIATE
	var session winapi.WSMAN_SESSION_HANDLE
	err = winapi.WSManCreateSession(winapi.WSMAN_API_HANDLE(clientHandle), "https://192.168.56.108:5986", 0, &authCreds, 0, &session)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[+] Created Session")
	// set options
	var optCn winapi.WSManSessionOption = winapi.WSMAN_OPTION_SKIP_CN_CHECK
	var optCnData winapi.WSMAN_DATA
	optCnData.Type = winapi.WSMAN_DATA_TYPE_DWORD
	optCnData.Number = 1
	var optCa winapi.WSManSessionOption = winapi.WSMAN_OPTION_SKIP_CA_CHECK
	var optCaData winapi.WSMAN_DATA
	optCaData.Type = winapi.WSMAN_DATA_TYPE_DWORD
	optCaData.Number = 1
	var optTimeout winapi.WSManSessionOption = winapi.WSMAN_OPTION_DEFAULT_OPERATION_TIMEOUTMS
	var optTimeoutData winapi.WSMAN_DATA
	optTimeoutData.Type = winapi.WSMAN_DATA_TYPE_DWORD
	optTimeoutData.Number = 60000
	err = winapi.WSManSetSessionOption(session, optTimeout, &optTimeoutData)
	if err != nil {
		log.Fatal(err)
	}
	err = winapi.WSManSetSessionOption(session, optCn, &optCnData)
	if err != nil {
		log.Fatal(err)
	}
	err = winapi.WSManSetSessionOption(session, optCa, &optCaData)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[+] Session Options set")
	hEvent, err = winapi.CreateEventW(nil, 0, 0, "")
	if err != nil {
		log.Fatal(err)
	}
	hReceiveEvent, err = winapi.CreateEventW(nil, 0, 0, "")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[+] Events Created")
	shellCall := windows.NewCallback(shellCallback)
	recvCall := windows.NewCallback(recvCallback)
	async := winapi.WSMAN_SHELL_ASYNC{}
	async.CompletionFunction = shellCall
	recvAsync := winapi.WSMAN_SHELL_ASYNC{}
	recvAsync.CompletionFunction = recvCall
	// create shell
	var shell winapi.WSMAN_SHELL_HANDLE
	err = winapi.WSManCreateShell(session, 0, winapi.WSMAN_CMDSHELL_URI, 0, 0, 0, &async, &shell)
	if err != nil {
		log.Fatal(err)
	}
	windows.WaitForSingleObject(windows.Handle(hEvent), windows.INFINITE)
	log.Println("[+] Created Shell")
	var cmdHandle winapi.WSMAN_COMMAND_HANDLE
	err = winapi.WSManRunShellCommand(shell, 0, "hostname", 0, 0, &async, &cmdHandle)
	if err != nil {
		log.Fatal(err)
	}
	windows.WaitForSingleObject(windows.Handle(hEvent), windows.INFINITE)
	log.Println("[+] Running Shell Command")
	var opHandle winapi.WSMAN_OPERATION_HANDLE
	err = winapi.WSManReceiveShellOutput(shell, cmdHandle, 0, 0, &recvAsync, &opHandle)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[+] Receiving Command Output")
	windows.WaitForSingleObject(windows.Handle(hReceiveEvent), windows.INFINITE)
	// cleanup
	log.Println("Future Cleanup")

}
