package pop3

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
	"time"
)

const (
	CMD_READY = iota
	CMD_USER  = iota
	CMD_PASS  = iota
	CMD_STAT  = iota
	CMD_LIST  = iota
	CMD_RETR  = iota
	CMD_DELE  = iota
	CMD_NOOP  = iota
	CMD_RSET  = iota
	CMD_TOP   = iota
	CMD_UIDL  = iota
	CMD_APOP  = iota
	CMD_QUIT  = iota
)

var stateList = map[int]string{
	CMD_READY: "READY",
	CMD_USER:  "USER",
	CMD_PASS:  "PASS",
	CMD_STAT:  "STAT",
	CMD_LIST:  "LIST",
	CMD_RETR:  "RETR",
	CMD_DELE:  "DELE",
	CMD_NOOP:  "NOOP",
	CMD_RSET:  "RSET",
	CMD_TOP:   "TOP",
	CMD_UIDL:  "UIDL",
	CMD_APOP:  "APOP",
	CMD_QUIT:  "QUIT",
}

const (
	MSG_INIT            = "220.init"
	MSG_OK              = "220"
	MSG_MAIL_OK         = "250"
	MSG_BYE             = "221"
	MSG_BAD_SYNTAX      = "500"
	MSG_LOGIN_INFO      = "502"
	MSG_COMMAND_TM_ERR  = "421"
	MSG_AUTH_LOGIN_USER = "334.user"
	MSG_AUTH_LOGIN_PWD  = "334.passwd"
	MSG_AUTH_OK         = "235"
	MSG_AUTH_FAIL       = "535"
	MSG_DATA            = "354"
	MSG_RETR_DATA       = "%s\r\n."
)

var msgList = map[string]string{
	MSG_INIT:            "+OK Welcome to coremail Mail Pop3 Server (imail)",
	MSG_OK:              "+OK core mail",
	MSG_BYE:             "bye",
	MSG_LOGIN_INFO:      "+OK 2542 message(s) [100298482 byte(s)]",
	MSG_COMMAND_TM_ERR:  "Too many error commands",
	MSG_BAD_SYNTAX:      "Error: bad syntax",
	MSG_AUTH_LOGIN_USER: "dXNlcm5hbWU6",
	MSG_AUTH_LOGIN_PWD:  "UGFzc3dvcmQ6",
	MSG_AUTH_OK:         "Authentication successful",
	MSG_AUTH_FAIL:       "Error: authentication failed",
	MSG_MAIL_OK:         "Mail OK",
	MSG_RETR_DATA:       "%s\r\n.",
}

var GO_EOL = GetGoEol()

func GetGoEol() string {
	if "windows" == runtime.GOOS {
		return "\r\n"
	}
	return "\n"
}

type Pop3Server struct {
	debug     bool
	conn      net.Conn
	state     int
	startTime time.Time
	errCount  int
}

func (this *Pop3Server) setState(state int) {
	this.state = state
}

func (this *Pop3Server) getState() int {
	return this.state
}

func (this *Pop3Server) D(a ...interface{}) (n int, err error) {
	return fmt.Println(a...)
}

func (this *Pop3Server) Debug(d bool) {
	this.debug = d
}

func (this *Pop3Server) W(msg string) {
	_, err := this.conn.Write([]byte(msg))

	if err != nil {
		log.Fatal(err)
	}
}

func (this *Pop3Server) write(code string) {
	info := fmt.Sprintf("%s\r\n", msgList[code])
	this.W(info)
}

func (this *Pop3Server) getString() (string, error) {
	input, err := bufio.NewReader(this.conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	inputTrim := strings.TrimSpace(input)
	return inputTrim, err
}

func (this *Pop3Server) getString0() (string, error) {
	buffer := make([]byte, 2048)

	n, err := this.conn.Read(buffer)
	if err != nil {
		log.Fatal(this.conn.RemoteAddr().String(), " connection error: ", err)
		return "", err
	}

	input := string(buffer[:n])
	inputTrim := strings.TrimSpace(input)
	return inputTrim, err
}

func (this *Pop3Server) close() {
	this.conn.Close()
}

func (this *Pop3Server) cmdCompare(input string, cmd int) bool {
	if strings.EqualFold(input, stateList[cmd]) {
		return true
	}
	return false
}

func (this *Pop3Server) stateCompare(input int, cmd int) bool {
	if input == cmd {
		return true
	}
	return false
}

func (this *Pop3Server) cmdUser(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_USER) {
		if len(inputN) < 2 {
			this.write(MSG_BAD_SYNTAX)
			return false
		}
		this.write(MSG_OK)
		return true
	}
	return false
}

func (this *Pop3Server) cmdPass(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_PASS) {
		if len(inputN) < 2 {
			this.write(MSG_BAD_SYNTAX)
			return false
		}
		this.write(MSG_LOGIN_INFO)
		return true
	}
	return false
}

func (this *Pop3Server) cmdList(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_LIST) {
		if len(inputN) < 2 {
			this.write(MSG_BAD_SYNTAX)
			return false
		}
		this.write(MSG_LOGIN_INFO)
		return true
	}
	return false
}

func (this *Pop3Server) cmdRetr(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_RETR) {
		if len(inputN) < 2 {
			this.write(MSG_BAD_SYNTAX)
			return false
		}
		this.write(MSG_RETR_DATA)
		return true
	}
	return false
}

func (this *Pop3Server) cmdUidl(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_UIDL) {
		if len(inputN) < 2 {
			this.write(MSG_BAD_SYNTAX)
			return false
		}
		this.write(MSG_LOGIN_INFO)
		return true
	}
	return false
}

func (this *Pop3Server) cmdQuit(input string) bool {
	if this.cmdCompare(input, CMD_QUIT) {
		this.write(MSG_BYE)
		this.close()
		return true
	}
	return false
}

func (this *Pop3Server) handle() {
	for {
		state := this.getState()
		input, _ := this.getString()

		// fmt.Printf(input, state)

		if this.cmdQuit(input) {
			break
		}

		if this.stateCompare(state, CMD_READY) {
			if this.cmdUser(input) {
				this.setState(CMD_USER)
			}
		}

		if this.stateCompare(state, CMD_USER) {
			if this.cmdPass(input) {
				this.setState(CMD_PASS)
			}
		}

		if this.stateCompare(state, CMD_PASS) {

			if this.cmdList(input) {

			}

			if this.cmdUidl(input) {

			}

			if this.cmdRetr(input) {

			}
		}

	}
}

func (this *Pop3Server) start(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(time.Minute * 180))
	defer conn.Close()
	this.conn = conn

	this.startTime = time.Now()

	this.write(MSG_INIT)
	this.setState(CMD_READY)

	this.handle()
}

func Start(port int) {
	pop3_port := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", pop3_port)
	if err != nil {
		panic(err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		srv := Pop3Server{}
		go srv.start(conn)
	}
}