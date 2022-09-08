package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type udco2Elem struct {
	Time   string
	CO2    float64
	HUM    float64
    TMP    float64
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

func mainHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("index.html")
	if err != nil {
		panic(err.Error())
	}
	if err := t.Execute(w, nil); err != nil {
		panic(err.Error())
	}
}

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

	dispAll()
}
func dispAll(){
	fmt.Printf(" \n")
	fmt.Printf(" Time: %v\n", dispData.Time)
	fmt.Printf(" \n")
	fmt.Printf(" Temp: %v\n", dispData.Temp)
	fmt.Printf(" Humi: %v\n", dispData.Humi)
	fmt.Printf(" \n")
	fmt.Printf(" Led : %v\n", dispData.Led)
	fmt.Printf(" Pres: %v\n", dispData.Press)
	fmt.Printf(" Noiz: %v\n", dispData.Noize)
	fmt.Printf(" TVOC: %v\n", dispData.TVOC)
	fmt.Printf(" CO2 : %v\n", dispData.CO2)
	fmt.Printf(" Dis : %v\n", dispData.Dis)
	fmt.Printf(" Strk: %v\n", dispData.Strk)
	fmt.Printf(" \n")
	fmt.Printf(" \n")
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

		rep := regexp.MustCompile(`CO2=\s*`)
		result := rep.Split(line[1], -1)
		c, _ := strconv.ParseFloat(result[1], 64)

		rep = regexp.MustCompile(`HUM=\s*`)
		result = rep.Split(line[2], -1)
		h, _ := strconv.ParseFloat(result[1], 64)

		rep = regexp.MustCompile(`TMP=\s*`)
		result = rep.Split(line[3], -1)
		t, _ := strconv.ParseFloat(result[1], 64)

		data := udco2Elem{times, c, h, t}
		datas = append(datas, data)

		disp_co2 = c
	}
	//	datas = append(datas.Label, "a")

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
	//	datas = append(datas.Label, "a")

	return
}
func g3(w http.ResponseWriter, r *http.Request) {

	var datas []udco2Elem
	var disp_co2 float64

	datas, disp_co2 = getCo2()
	
	setCo2( disp_co2 )

	js, err := json.Marshal(datas)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Write(js)

}

func g1x(w http.ResponseWriter, r *http.Request) {

	var datas []omronElem
	var dispdata omronElem

	datas, dispdata = getETC()

	setETC( dispdata )

	js, err := json.Marshal(datas)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Write(js)

}


func main() {

	var disp_co2 float64
	_, disp_co2 = getCo2()
	setCo2( disp_co2 )

	var dispdata omronElem
	_, dispdata = getETC()
	setETC( dispdata )


	dir, _ := os.Getwd()

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/api/g3/", g3)
	http.HandleFunc("/api/g1x/", g1x)

	http.Handle("/UD-CO2S/", http.StripPrefix("/UD-CO2S/", http.FileServer(http.Dir(dir+"/UD-CO2S/"))))
	http.Handle("/omron/", http.StripPrefix("/omron/", http.FileServer(http.Dir(dir+"/omron/"))))

	http.ListenAndServe(":8000", nil)
}

