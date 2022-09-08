package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
	"net"
	"compress/gzip"
	"net/smtp"
)

type udco2Elem struct {
	Time   string
	CO2    float64
//	HUM    float64
//	TMP    float64
}

type omronElem struct {
	Time  string
	Temp  float64
	Humi  float64
	Led   float64
	Press float64
	Noize float64
	TVOC  float64
	CO2   float64
	Dis   float64
	Strk  float64
}

type dispElem struct {
	Time  string
	Temp  float64
	Humi  float64
	Led   float64
	Press float64
	Noize float64
	TVOC  float64
	CO2   float64
	Dis   float64
	Strk  float64
}

var dispData dispElem

func setCo2(co2 float64) {
	dispData.CO2 = co2
}
func setETC(etc omronElem) {
	dispData.Time = etc.Time
	
	dispData.Temp = etc.Temp
	dispData.Humi = etc.Humi
	dispData.Led  = etc.Led
	dispData.Press= etc.Press
	dispData.Noize= etc.Noize
	dispData.TVOC = etc.TVOC
//	dispData.CO2  = etc.CO2
	dispData.Dis  = etc.Dis
	dispData.Strk = etc.Strk
}
func dispAll(){
	fmt.Printf(" \n")
	fmt.Printf(" %v\n", dispData.Time)
	fmt.Printf(" \n")
	fmt.Printf(" Temp: %v\n", dispData.Temp)
	fmt.Printf(" Humi: %v\n", dispData.Humi)
	fmt.Printf(" Led : %v\n", dispData.Led)
	fmt.Printf(" Pres: %v\n", dispData.Press)
	fmt.Printf(" Noiz: %v\n", dispData.Noize)
	fmt.Printf(" TVOC: %v\n", dispData.TVOC)
	fmt.Printf(" CO2 : %v\n", dispData.CO2)
	fmt.Printf(" Dis : %v\n", dispData.Dis)
	fmt.Printf(" Strk: %v\n", dispData.Strk)
	fmt.Printf(" \n")

	aaa, _ := localAddresses()
	fmt.Println(aaa)
}

func getCo2() (datas []udco2Elem, disp_co2 float64) {
	fin, err := os.Open("/home/zero/Z_Work/sensor/UD-CO2S/ud-co2_result.csv")
	if err != nil {
		fmt.Println("panic")
		panic(err)
	}
	defer fin.Close()
	reader := csv.NewReader(fin)
	reader.TrimLeadingSpace = true

	var line []string
	for {
		line, err = reader.Read()
		if err != nil {
			break
		}
		times := strings.TrimSpace(line[0])

		okstring := regexp.MustCompile(`OK`)
		if okstring.MatchString(line[1]) {
			continue
		}

		rep := regexp.MustCompile(`CO2=\s*`)
		result := rep.Split(line[1], -1)
		c, _ := strconv.ParseFloat(result[1], 64)

//		rep = regexp.MustCompile(`HUM=\s*`)
//		result = rep.Split(line[2], -1)
//		h, _ := strconv.ParseFloat(result[1], 64)

//		rep = regexp.MustCompile(`TMP=\s*`)
//		result = rep.Split(line[3], -1)
//		t, _ := strconv.ParseFloat(result[1], 64)

		data := udco2Elem{times, c/*, h, t*/}
		datas = append(datas, data)

		disp_co2 = c
	}

	return
}
func getETC() (datas []omronElem, dispdata omronElem) {
	fin, err := os.Open("/home/zero/Z_Work/sensor/omron/omron.csv")
	if err != nil {
		fmt.Println("panic")
		panic(err)
	}
	defer fin.Close()
	reader := csv.NewReader(fin)
	reader.TrimLeadingSpace = true

	var line []string
	for {
		line, err = reader.Read()
		if err != nil {
			break
		}
		times := strings.TrimSpace(line[0])

		tmp, _ := strconv.ParseFloat(strings.TrimSpace(line[1]), 64)
		hum, _ := strconv.ParseFloat(strings.TrimSpace(line[2]), 64)
		led, _ := strconv.ParseFloat(strings.TrimSpace(line[3]), 64)
		prs, _ := strconv.ParseFloat(strings.TrimSpace(line[4]), 64)
		noz, _ := strconv.ParseFloat(strings.TrimSpace(line[5]), 64)
		tvo, _ := strconv.ParseFloat(strings.TrimSpace(line[6]), 64)
		co2, _ := strconv.ParseFloat(strings.TrimSpace(line[7]), 64)
		dis, _ := strconv.ParseFloat(strings.TrimSpace(line[8]), 64)
		srk, _ := strconv.ParseFloat(strings.TrimSpace(line[9]), 64)

		data := omronElem{times, tmp, hum, led, prs, noz, tvo, co2, dis, srk}
		dispdata = omronElem{times, tmp, hum, led, prs, noz, tvo, co2, dis, srk}
		datas = append(datas, data)
	}

	return
}

func localAddresses() (string, error) {
	var list []string

	name, err := os.Hostname()
	if err != nil {
		return "", err
	}

	interfaceList, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	list = append(list, name)
	for _, networkInterface := range interfaceList {
		addressList, err := networkInterface.Addrs()
		if err != nil {
			continue
		}

		for _, netInterfaceAddress := range addressList {
			networkIp, ok := netInterfaceAddress.(*net.IPNet)
			if ok && !networkIp.IP.IsLoopback() && networkIp.IP.To4() != nil && networkInterface.Name == "wlan0" {
				ip := networkIp.IP.String()

				list = append(list, ip)
			}
		}
	}

	return strings.Join(list, "\n"), nil
}

// NumCheck ... Check the argument(string) to determine if it is a number.
func NumCheck(str string) bool {
	for _, r := range str {
		if '0' <= r && r <= '9' {
			return true
		}
	}
	return false
}
func getCo2_hour(filename string) (datas []udco2Elem, count int32) {
	fin, err := os.Open(filename)
	if err != nil {
		fmt.Println("panic")
		panic(err)
	}
	defer fin.Close()
	reader := csv.NewReader(fin)
	reader.TrimLeadingSpace = true

	var line []string
	var times, pretimes string = "", ""
	for {
		line, err = reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			continue
		}

		okstring := regexp.MustCompile(`OK`)
		if okstring.MatchString(line[1]) {
			continue
		}

		times = strings.TrimSpace(line[0])
		if pretimes != times {
			rep := regexp.MustCompile(`CO2=\s*`)
			result := rep.Split(line[1], -1)
			if result[0] == "" && NumCheck(result[1]) {
				c, _ := strconv.ParseFloat(result[1], 64)
				if 300 <= c && c <= 1500 {
					data := udco2Elem{times, c}
					datas = append(datas, data)
					pretimes = times
					count++
				}
			}
		}
	}
	//	datas = append(datas.Label, "a")

	return
}
func getETC_hour(filename string) (datas []omronElem, count int32) {
	fin, err := os.Open(filename)
	if err != nil {
		fmt.Println("panic")
		panic(err)
	}
	defer fin.Close()
	reader := csv.NewReader(fin)
	reader.TrimLeadingSpace = true

	var line []string
	for {
		line, err = reader.Read()
		if err != nil {
			break
		}
		times := strings.TrimSpace(line[0])

		tmp, _ := strconv.ParseFloat(strings.TrimSpace(line[1]), 64)
		hum, _ := strconv.ParseFloat(strings.TrimSpace(line[2]), 64)
		led, _ := strconv.ParseFloat(strings.TrimSpace(line[3]), 64)
		prs, _ := strconv.ParseFloat(strings.TrimSpace(line[4]), 64)
		noz, _ := strconv.ParseFloat(strings.TrimSpace(line[5]), 64)
		tvo, _ := strconv.ParseFloat(strings.TrimSpace(line[6]), 64)
		co2, _ := strconv.ParseFloat(strings.TrimSpace(line[7]), 64)
		dis, _ := strconv.ParseFloat(strings.TrimSpace(line[8]), 64)
		srk, _ := strconv.ParseFloat(strings.TrimSpace(line[9]), 64)

		data := omronElem{times, tmp, hum, led, prs, noz, tvo, co2, dis, srk}
		datas = append(datas, data)
		count++
	}

	return
}

func recoverFile(infile string, outfile string) {
	filename := infile
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

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

func makeHourFile(outfile string){

	_ = os.Remove(outfile)

	recoverFile("/home/zero/Z_Work/sensor/UD-CO2S/ud-co2_result.csv", "/home/zero/Z_Work/sensor/co2.fix.csv")
	recoverFile("/home/zero/Z_Work/sensor/omron/omron.csv", "/home/zero/Z_Work/sensor/etc.fix.csv")
	co2, co2len := getCo2_hour("/home/zero/Z_Work/sensor/co2.fix.csv")
	etc, etclen := getETC_hour("/home/zero/Z_Work/sensor/etc.fix.csv")

	fout, err := os.OpenFile(outfile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	var i,j int32
	for j = 0; j < etclen; j++ {
		for i = 0; i < co2len; i++ {
			if co2[i].Time == etc[j].Time {
				res_str := fmt.Sprintf("%s,%0.2f,%0.2f,%0.0f,%0.3f,%0.2f,%0.0f,%0.0f,%0.2f,%0.2f\n", co2[i].Time, etc[j].Temp, etc[j].Humi, etc[j].Led, etc[j].Press, etc[j].Noize, etc[j].TVOC, co2[i].CO2, etc[j].Dis, etc[j].Strk)
				fout.WriteString(res_str)
			}
		}
	}

}
func makeZipFile() (zipfilename string) {
	t := time.Now()
	zipfilename = fmt.Sprintf("/home/zero/Z_Work/sensor/env.%04d%02d%02dT%02d%02d.csv.zip", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	dist, err := os.Create(zipfilename)
	if err != nil {
		panic(err)
	}
	defer dist.Close()

	gw, err := gzip.NewWriterLevel(dist, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	defer gw.Close()

	src, err := os.Open("/home/zero/Z_Work/sensor/env.csv")
	if err != nil {
		panic(err)
	}
	defer src.Close()

	if _, err := io.Copy(gw, src); err != nil {
		panic(err)
	}

	return
}

func UuEncoder(s []byte) []byte {
	
	out := []byte{s[0]>>2 +32, (s[0]<<6)>>2 | s[1]>>4 + 32, (s[1]<<4)>>2 | s[2]>>6 +32,(s[2]<<2)>>2 + 32}
	
	return out
}

func UuDecoder(s []byte) []byte {
	
	out := []byte{(s[0]-32)<<2 | (s[1]-32)>>4, (s[1]-32)<<4 | (s[2]-32)>>2, (s[2]-32)<<6 | (s[3]-32)}
	
	return out
}

func Encoder(s string, filename string, write bool) string {

	var out []byte
	
	sbytes := []byte(s)
		
	out = append(out,[]byte("begin 644 " + filename)...)
	
	for len(sbytes) >= 45 {
		out = append(out, []byte("\nM")...)
		for i:=0;i<15;i++{
			out = append(out,UuEncoder(sbytes[3*i:3*i+3])...)
		}
		sbytes = sbytes[45:]
	}
	
	if len(sbytes)<45 {

		if len(sbytes)%3 != 0 {
			ext := make([]byte, 3 - len(sbytes)%3)
			sbytes = append(sbytes,ext...)
		}
		out = append(out,'\n',byte(len(sbytes)+32))
		for i:=0;i<len(sbytes)/3 ;i++{
			out = append(out,UuEncoder(sbytes[3*i:3*i+3])...)
		}
		out = append(out,"\n`"...)
		out = append(out,"\nend\n"...)
	}
	
	if write{
		err:= ioutil.WriteFile(filename, out, 0644)
			
		if err != nil {
			panic(err)
		}
	}
	
	return string(out)
	
}

func FileEncoder(path string, write bool) string {
	data, err := ioutil.ReadFile(path)
	
	if err != nil {
		panic(err)
	}
	
	return Encoder(string(data),path, write)
}

func Decoder(s string, write bool) (string,string,string){

	slines := strings.Split(s,"\n")
	
	var perm,filename string
	
	if string(strings.Split(slines[0]," ")[0]) != "begin" {
		fmt.Println("Incorrect File Format")
		panic("Incorrect File Format")
	}
	
	if string(strings.Split(slines[0]," ")[1]) == " " || string(strings.Split(slines[0]," ")[1]) == ""{
		fmt.Println("Incorrect File Permissions")
		panic("Incorrect File Permissions")
	} else {
		perm = strings.Split(slines[0]," ")[1]
	}
	
	if string(strings.Split(slines[0]," ")[2]) == " " || string(strings.Split(slines[0]," ")[2]) == "" {
		fmt.Println("Invalid Filename")
		panic("Invalid Filename")
	} else {
		filename = strings.Split(slines[0], " ")[2]
	}
	
	if string(slines[len(slines)-2]) != "end" {
		fmt.Println("No END Found. Invalid Format")
		panic("No END Found. Invalid Format")
	}
	
	if string(slines[len(slines)-3]) != "`"{
		fmt.Println("Incorrect ending format")
		panic("Incorrect ending format")
	}
	
	
	slines = slines[1:len(slines)-3]
	
	var text []byte
	for _,line := range slines {
		if string(line[0]) == "M" {
			for i:=0; i<15;i++ {
				text = append(text,UuDecoder([]byte(line[4*i+1:4*i+4+1]))...)
			}
		} else {
			ln := int(line[0])
			for i:=0; i<(ln-32)/3;i++ {
				text = append(text, UuDecoder([]byte(line[4*i+1:4*i+4+1]))...)
			}
		}
	}
	
	if write{
		
		if perm == "644" {
			
			err:= ioutil.WriteFile(filename, text, 0644)
			
			if err != nil {
				panic(err)
			}
		}
	}

	return string(text), filename, perm
}

func FileDecoder(path string, write bool) (string,string, string){
	data, err := ioutil.ReadFile(path)
	
	if err != nil {
		panic(err)
	}
	
	return Decoder(string(data), write)
}

func send_mail(attached string){

	data, err := ioutil.ReadFile(attached)
	if err != nil {
		panic(err)
	}
	s_out := Encoder(string(data), attached, false)

	// Set up authentication information.
	auth := smtp.PlainAuth(
		"",
		"sensor.raspi.9831@gmail.com",
		"lnfetzmhnnjsxgwl",
		"smtp.gmail.com",
	)

	hostname, _ := os.Hostname()
	subjectstring := fmt.Sprintf("Subject: %s %s\r\n\r\n",hostname, attached[25:42])

	mainstring := subjectstring + s_out

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	errs := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		"sensor.raspi.9831@gmail.com",
		[]string{"Kaznim21@gmail.com","Setestse123123@gmail.com","nishimura.2460.home@gmail.com"},
	//	[]string{"nishimura.2460.home@gmail.com"},
		[]byte(mainstring),
	)
	if errs != nil {
		panic(err)
	}
}
func main() {

	var disp_co2 float64
	var dispdata omronElem
	var count int32 = 0

	for {
		_, disp_co2 = getCo2()
		setCo2( disp_co2 )

		_, dispdata = getETC()
		setETC( dispdata )

		dispAll()

		time.Sleep(time.Minute * 1)

		count++

		if count>=60 {
			count = 0
			makeHourFile("/home/zero/Z_Work/sensor/env.csv")
			zipfile := makeZipFile()
			send_mail(zipfile)
			_ = os.Remove(zipfile)
		}
	}
}

