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
	"archive/zip"
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

var (
	builddate string
)

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
//	fmt.Printf("len:%v\n",dispData)
	if len(dispData.Time) > 6 {
		fmt.Printf(" \n")
		fmt.Printf(" %v\n", dispData.Time[6:16])
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
	}
	aaa, _ := localAddresses()
	fmt.Println(aaa)
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

func getCo2(filename string) (datas []udco2Elem, disp_co2 float64, count int32) {
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

					disp_co2 = c
				}
			}
		}
	}

	return
}
func getETC(filename string) (datas []omronElem, dispdata omronElem, count int32) {
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
		dispdata = omronElem{times, tmp, hum, led, prs, noz, tvo, co2, dis, srk}
		datas = append(datas, data)
		count++
	}

	return
}

func makeHourFile(outfile string){

	_ = os.Remove(outfile)

	co2, _, co2len := getCo2("/home/zero/Z_Work/sensor/UD-CO2S/ud-co2.csv")
	etc, _, etclen := getETC("/home/zero/Z_Work/sensor/omron/omron2.csv")

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
	hostname, _ := os.Hostname()
	zipfilename = fmt.Sprintf("/home/zero/Z_Work/sensor/%s_%02d%02dT%02d%02d.csv.zip", hostname, t.Month(), t.Day(), t.Hour(), t.Minute())
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

func makeZipFile2() (zipfilename string) {
	t := time.Now()
	hostname, _ := os.Hostname()
	zipfilename = fmt.Sprintf("/home/zero/Z_Work/sensor/%s_%02d%02dT%02d%02d.zip", hostname, t.Month(), t.Day(), t.Hour(), t.Minute())
	dist, err := os.Create(zipfilename)
	if err != nil {
		panic(err)
	}
	defer dist.Close()

	w := zip.NewWriter(dist)
	defer w.Close()
	outfilename := fmt.Sprintf("%s_%02d%02dT%02d%02d.csv", hostname, t.Month(), t.Day(), t.Hour(), t.Minute())
	f, _ := w.Create(outfilename)

	src, err := os.Open("/home/zero/Z_Work/sensor/env.csv")
	if err != nil {
		panic(err)
	}
	defer src.Close()

	if _, err := io.Copy(f, src); err != nil {
		panic(err)
	}

	return
}

func UuEncoder(s []byte) []byte {
	
	out := []byte{s[0]>>2 +32, (s[0]<<6)>>2 | s[1]>>4 + 32, (s[1]<<4)>>2 | s[2]>>6 +32,(s[2]<<2)>>2 + 32}
	
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


func send_mail(attached string, mainstring string){

	var subjectstring string

	hostname, _ := os.Hostname()

	if attached != "" {
		data, err := ioutil.ReadFile(attached)
		if err != nil {
			panic(err)
		}
		s_out := Encoder(string(data), attached, false)

		subjectstring = "Mime-Version: 1.0\r\n"
		subjectstring += fmt.Sprintf("Subject: %s %s\r\n",hostname, attached[25:40])
		subjectstring += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"--nishi\"\r\n")
		subjectstring += fmt.Sprintf("\r\n")
		subjectstring += fmt.Sprintf("----nishi\r\n")
		subjectstring += fmt.Sprintf("Content-Type: text/plain; charset=iso-2022-jp\r\n")
		subjectstring += fmt.Sprintf("\r\n\r\n")
		subjectstring += fmt.Sprintf("----nishi\r\n")
		subjectstring += fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n",attached[25:44])
		subjectstring += fmt.Sprintf("Content-Transfer-Encoding: x-uuencode\r\n")
		subjectstring += s_out + "\r\n" 
		subjectstring += "Build: " + builddate + "\r\n"
		subjectstring += "\r\n----nishi--\r\n"
	} else {
		s_out := mainstring
		tinfo := time.Now()

		subjectstring = "Mime-Version: 1.0\r\n"
		subjectstring += fmt.Sprintf("Subject: %s %s %02d%02d\r\n",hostname, "buffer clear", tinfo.Month(), tinfo.Day())
		subjectstring += fmt.Sprintf("Content-Type: text/plain; charset=iso-2022-jp\r\n")
		subjectstring += fmt.Sprintf("\r\n")
		subjectstring += s_out + "\r\n" 
		subjectstring += "Build: " + builddate + "\r\n"
	}

	// Set up authentication information.
	auth := smtp.PlainAuth(
		"",
		"sensor.raspi.9831@gmail.com",
		"lnfetzmhnnjsxgwl",
		"smtp.gmail.com",
	)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	errs := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		"sensor.raspi.9831@gmail.com",
//		[]string{"nishimura.2460.home@gmail.com"},
		[]string{"aict.mem2022@gmail.com","Setestse123123@gmail.com","nishimura.2460.home@gmail.com"},
		[]byte(subjectstring),
	)
	if errs != nil {
		panic(errs)
	}
}


func main() {

	dayinfo := time.Now()
	timeinfo:= time.Now()

//	count := 0

	for {
		_, disp_co2, _ := getCo2("/home/zero/Z_Work/sensor/UD-CO2S/ud-co2.csv")
		setCo2( disp_co2 )

		_, dispdata, _ := getETC("/home/zero/Z_Work/sensor/omron/omron2.csv")
		setETC( dispdata )

		dispAll()

		time.Sleep(time.Minute * 1)

		now := time.Now()
		hostname, _ := os.Hostname()
		hostnum, _  := strconv.ParseInt(string(hostname[4]),16,64)
		if timeinfo.Hour() != now.Hour() && int(hostnum) <= now.Minute() {
//		if count >= 0 {
			makeHourFile("/home/zero/Z_Work/sensor/env.csv")
			zipfile := makeZipFile2()
			send_mail(zipfile, "")
			_ = os.Remove(zipfile)

			timeinfo = time.Now()

//			count++
//			if count > 1 {
//				count = 0
			if dayinfo.Day() != now.Day() {
				///
				f1, err := os.OpenFile("/home/zero/Z_Work/sensor/omron/midnight", os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					panic(err)
				}
				defer f1.Close()
				f1.WriteString("midnight\n")

				f2, err := os.OpenFile("/home/zero/Z_Work/sensor/UD-CO2S/midnight", os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					panic(err)
				}
				defer f2.Close()
				f2.WriteString("midnight\n")

				send_mail("", "notice buffer clear\r\n")
				dayinfo = time.Now()
			}
		}
	}
}

