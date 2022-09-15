package main

import (
	"fmt"
	"log"
	"strings"
	"time"
	"os"
	"os/signal"
	"io/ioutil"
	"syscall"
	"github.com/tarm/serial"
)

func recoverFile(infile string, outfile string) {
	filename := infile
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	defer f.Close()

	c := make([]byte, 1)
	var buf []byte
	size := 0
	for {
		len, r_err := f.Read(c)
		if len == 0 {
			break
		}
		size += len
		if r_err != nil {
			fmt.Println("err:", err)
			return
		}
		if c[0] != 0 {
			buf = append(buf, c[0])
		}
	}

	err = ioutil.WriteFile(outfile, buf, 0755)
	if err != nil {
		fmt.Println("panic")
		panic(err)
	}
}

func main() {

	recoverFile("/home/zero/Z_Work/sensor/UD-CO2S/ud-co2_result.csv", "/home/zero/Z_Work/sensor/UD-CO2S/ud-co2_result.csv")

	go dataGet()

    quit := make(chan os.Signal)
    signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT)
    <-quit

}
func dataGet() {
	c := &serial.Config{Name: "/dev/ttyACM0", Baud: 115200, ReadTimeout: time.Millisecond * 6000}

	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	s.Flush()

//	fout, err := os.Create("ud-co2_result.csv")
    fout, err := os.OpenFile("ud-co2_result.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	for {
		mySerialWrite(s, "ID?\r\n")

		time.Sleep(time.Millisecond * 300)

		str := mySerialRead(s)
//		fmt.Printf("%s", str)
		if str == "OK ID=UD-CO2S\r\n" {
			break
		}
		time.Sleep(time.Second * 1)
	}

	mySerialWrite(s, "STA\r\n")
	res := mySerialRead(s)

	for {
		res = mySerialRead(s)
		if len(res) > 2 {
			t := time.Now()
//			fmt.Printf("%02d:%02d,%s", t.Hour(), t.Minute(), res)
			res_str := fmt.Sprintf("%04d/%02d/%02d %02d:%02d,%s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), res)
			fout.WriteString(res_str)
		}
		time.Sleep(time.Second * 15)
	}

	mySerialWrite(s, "STP\r\n")

	_ = mySerialRead(s)
//	fmt.Printf("%s", last)
}
func chop(s string) string {
	s = strings.TrimRight(s, "\x00")
	//	s = strings.TrimRight(s, "\n")t
	//	if strings.HasSuffix(s, "\r") {
	//		s = strings.TrimRight(s, "\r")
	//	}

	return s
}
func mySerialRead(port *serial.Port) string {
	buf := make([]byte, 1)
	res := make([]byte, 64)
	var n int
	var err error
	var k int
	k = 0
	for {
		n, err = port.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if n != 0 {
			res[k] = buf[0]
			k = k + 1
		}
		if buf[0] == 0x0a {
			break
		}
	}
	//	fmt.Printf("%s", res)
	str := fmt.Sprintf("%s", res)
	str = chop(str)
	return str
}
func mySerialWrite(port *serial.Port, str string) {
	_, err := port.Write([]byte(str))
	if err != nil {
		log.Fatal(err)
	}
}
