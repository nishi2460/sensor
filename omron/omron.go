package main

import (
	"fmt"
	"log"
	"os"
	"io/ioutil"
	"time"

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

	recoverFile("/home/zero/Z_Work/sensor/omron/omron.csv", "/home/zero/Z_Work/sensor/omron/omron.csv")

	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 115200, ReadTimeout: time.Second * 1}

	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	s.Flush()

	fout, err := os.OpenFile("omron.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	requestRead(s, 0x180A)

	datas := mySerialRead(s)
	//	fmt.Printf("%s\n", datas)

	requestRead(s, 0x5401) // Error status
	datas = mySerialRead(s)

	/*
		requestRead( s, 0x5012 )	// Latest sensing data
		datas = mySerialRead(s)
		fmt.Printf("Sequence number    : %d \n",       datas[7])
		fmt.Printf("Temperature        : %0.2f degC\n",float32(uint16(datas[8])+uint16(datas[9])*uint16(256))*0.01)
		fmt.Printf("Humidity           : %0.2f %%RH\n",float32(uint16(datas[10])+uint16(datas[11])*uint16(256))*0.01)
		fmt.Printf("Ambient Light      : %d lx\n",     int16(uint16(datas[12])+uint16(datas[13])*uint16(256)))
		fmt.Printf("Barometric pressure: %0.3f hPa\n", float32( uint32(datas[14]) + uint32(datas[15])*uint32(256) + uint32(datas[16])*uint32(256*256) + uint32(datas[17])*uint32(256*256*256))*0.001)
		fmt.Printf("Sound noize        : %0.2f dB\n",  float32(uint16(datas[18])+uint16(datas[19])*uint16(256))*0.01)
		fmt.Printf("eTVOC              : %d ppb\n",    int16(uint16(datas[20])+uint16(datas[21])*uint16(256)))
		fmt.Printf("eCO2               : %d ppm\n",    int16(uint16(datas[22])+uint16(datas[23])*uint16(256)))
	*/

	for {
		requestRead(s, 0x5021) // Latest data short
		datas = mySerialRead(s)
		/*
			fmt.Printf("Sequence number    : %d\n",        datas[7])
			fmt.Printf("Temperature        : %0.2f degC\n",float32(uint16(datas[8])+uint16(datas[9])*uint16(256))*0.01)
			fmt.Printf("Humidity           : %0.2f %%RH\n",float32(uint16(datas[10])+uint16(datas[11])*uint16(256))*0.01)
			fmt.Printf("Ambient Light      : %f lx\n",     float32(uint16(datas[12])+uint16(datas[13])*uint16(256)))
			fmt.Printf("Barometric pressure: %0.3f hPa\n", float32( uint32(datas[14]) + uint32(datas[15])*uint32(256) + uint32(datas[16])*uint32(256*256) + uint32(datas[17])*uint32(256*256*256))*0.001)
			fmt.Printf("Sound noize        : %0.2f dB\n",  float32(uint16(datas[18])+uint16(datas[19])*uint16(256))*0.01)
			fmt.Printf("eTVOC              : %f ppb\n",    float32(uint16(datas[20])+uint16(datas[21])*uint16(256)))
			fmt.Printf("eCO2               : %f ppm\n",    float32(uint16(datas[22])+uint16(datas[23])*uint16(256)))
			fmt.Printf("Discomfort index   : %0.2f\n",     float32(uint16(datas[24])+uint16(datas[25])*uint16(256))*0.01)
			fmt.Printf("Heat stroke        : %0.2f degC\n",float32(uint16(datas[26])+uint16(datas[27])*uint16(256))*0.01)
		*/
		//	seq := datas[7]
		tmp := float32(uint16(datas[8])+uint16(datas[9])*uint16(256)) * 0.01
		hum := float32(uint16(datas[10])+uint16(datas[11])*uint16(256)) * 0.01
		led := float32(uint16(datas[12]) + uint16(datas[13])*uint16(256))
		prs := float32(uint32(datas[14])+uint32(datas[15])*uint32(256)+uint32(datas[16])*uint32(256*256)+uint32(datas[17])*uint32(256*256*256)) * 0.001
		noz := float32(uint16(datas[18])+uint16(datas[19])*uint16(256)) * 0.01
		tvo := float32(uint16(datas[20]) + uint16(datas[21])*uint16(256))
		co2 := float32(uint16(datas[22]) + uint16(datas[23])*uint16(256))
		dis := float32(uint16(datas[24])+uint16(datas[25])*uint16(256)) * 0.01
		srk := float32(uint16(datas[26])+uint16(datas[27])*uint16(256)) * 0.01

		t := time.Now()
		res_str := fmt.Sprintf("%02d-%02dT%02d:%02d,%0.2f,%0.2f,%0.0f,%0.3f,%0.2f,%0.0f,%0.0f,%0.2f,%0.2f\n", t.Month(), t.Day(), t.Hour(), t.Minute(), tmp, hum, led, prs, noz, tvo, co2, dis, srk)
		fout.WriteString(res_str)

		time.Sleep(time.Minute * 1)
	}

}

func calc_crc(data []byte) (crc uint16) {
	var generator uint16
	var i int
	var b uint16
	generator = 0xA001
	crc = 0xFFFF

	for i = 0; i < len(data); i++ {
		crc ^= uint16(data[i])
		for b = 0; b < 8; b++ {
			if (crc & 1) != 0 {
				crc >>= 1
				crc ^= generator
			} else {
				crc >>= 1
			}
		}
	}
	return
}

func requestRead(port *serial.Port, addr uint16) {
	snd := make([]byte, 3)

	snd[0] = 0x01 //READ
	snd[1] = uint8(addr & 0x00FF)
	snd[2] = uint8((addr & 0xFF00) >> 8)

	sendDatas(port, snd, 3)
}

func sendDatas(port *serial.Port, dat []byte, len uint16) {

	snd := make([]byte, len+4)

	snd[0] = 'R'
	snd[1] = 'B'

	snd[2] = uint8((len + 2) & 0x00FF)
	snd[3] = uint8(((len + 2) & 0xFF00) >> 8)

	var i uint16
	for i = 0; i < len; i++ {
		snd[i+4] = dat[i]
	}

	crc := calc_crc(snd)
	snd = append(snd, uint8(crc&0x00FF))
	snd = append(snd, uint8((crc&0xFF00)>>8))

	mySerialWrite(port, snd)
}

func mySerialRead(port *serial.Port) (res []byte) {
	buf := make([]byte, 1)
	res = make([]byte, 64)
	var n int
	var err error
	var k uint16
	var len uint16
	k = 0
	len = 0

	for {
		n, err = port.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if n != 0 {
			res[k] = buf[0]
			k = k + 1
		}
		if len == 0 && k > 3 {
			len = uint16(res[2]) + uint16(res[3])*uint16(256)
		} else {
			if len != 0 && len+4 <= k {
				break
			}
		}
	}
	//	fmt.Println(res)
	return
}
func mySerialWrite(port *serial.Port, data []byte) {
	//	fmt.Println(data)
	_, err := port.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}
