package node

import (
	"bytes"
	"fmt"

	"project/internal/logger"
)

const (
	log_init     = "init"
	log_shutdown = "shutdown"
	log_boot     = "boot"
	log_client   = "client"
)

func (this *NODE) Printf(l logger.Level, src, format string, log ...interface{}) {
	if l < this.log_lv {
		return
	}
	buffer := logger.Prefix(l, src)
	if buffer == nil {
		return
	}
	buffer.WriteString(fmt.Sprintf(format, log...))
	this.print_log(buffer)
}

func (this *NODE) Print(l logger.Level, src string, log ...interface{}) {
	if l < this.log_lv {
		return
	}
	buffer := logger.Prefix(l, src)
	if buffer == nil {
		return
	}
	buffer.WriteString(fmt.Sprint(log...))
	this.print_log(buffer)
}

func (this *NODE) Println(l logger.Level, src string, log ...interface{}) {
	if l < this.log_lv {
		return
	}
	buffer := logger.Prefix(l, src)
	if buffer == nil {
		return
	}
	log_str := fmt.Sprintln(log...)
	log_str = log_str[:len(log_str)-1] // delete "\n"
	buffer.WriteString(log_str)
	this.print_log(buffer)
}

func (this *NODE) print_log(b *bytes.Buffer) {
	fmt.Println(b.String())
}
