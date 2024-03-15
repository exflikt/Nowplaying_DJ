package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
)

const TIMETABLE_TEMPLATE = `{{range $data := .}}<h1 class="dtimer" data-sh="{{$data.Sh}}" data-sm="{{$data.Sm}}" data-eh="{{$data.Eh}}" data-em="{{$data.Em}}" data-es="{{$data.Es}}">{{$data.Dj_name}}</h1>
<h2 class="dtimer" data-sh="{{$data.Sh}}" data-sm="{{$data.Sm}}" data-eh="{{$data.Eh}}" data-em="{{$data.Em}}" data-es="{{$data.Es}}">{{$data.Timeframe}}</h2>
{{end}}`

type TableData struct {
	Sh        uint8
	Sm        uint8
	Eh        uint8
	Em        uint8
	Es        uint8
	Dj_name   string
	Timeframe string
}

func main() {
	records := read_csv_records("timetable.csv")
	timetable := records_to_timetable(records)
	timetable_html := to_timetable_html(timetable)
	generate_html("public/index.html", "index.template", timetable_html)
	generate_html("public/obs.html", "obs.template", timetable_html)
}

func read_csv_records(filepath string) [][]string {
	file, err := os.Open(filepath)
	defer file.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open %s\n", filepath)
		panic(err)
	}

	records, err := csv.NewReader(bufio.NewReader(file)).ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse %s as a CSV file\n", filepath)
		panic(err)
	}
	return records
}

func records_to_timetable(records [][]string) []TableData {
	split_time := func(time string) (string, string) {
		hhmm := strings.SplitN(time, ":", 2)
		if len(hhmm) != 2 {
			panic(fmt.Sprintf("%s is not a valid time string; Start and End field should be of the form HH:MM\n", time))
		}
		return hhmm[0], hhmm[1]
	}

	parse_time := func(hh_str, mm_str string) (uint8, uint8) {
		hh, err := strconv.ParseUint(hh_str, 10, 8)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not convert string %s into uint8\n", hh_str)
			panic(err)
		}
		mm, err := strconv.ParseUint(mm_str, 10, 8)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not convert string %s into uint8\n", mm_str)
			panic(err)
		}
		return uint8(hh), uint8(mm)
	}

	one_min_prev := func(hh, mm uint8) (prev_hh uint8, prev_mm uint8) {
		if mm == 0 {
			if hh == 0 {
				prev_hh = 24
			} else {
				prev_hh = hh - 1
			}
		} else {
			prev_hh = hh
		}

		if mm == 0 {
			prev_mm = 59
		} else {
			prev_mm = mm - 1
		}

		return
	}

	timetable := make([]TableData, len(records)-1)
	for i := 1; i < len(records); i++ {
		start, end, dj_name := records[i][0], records[i][1], records[i][2]

		start_hh, start_mm := parse_time(split_time(start))
		end_hh, end_mm := parse_time(split_time(end))

		eh, em := one_min_prev(end_hh, end_mm)

		timetable[i-1] = TableData{
			Sh:        start_hh,
			Sm:        start_mm,
			Eh:        eh,
			Em:        em,
			Es:        59,
			Dj_name:   dj_name,
			Timeframe: fmt.Sprintf("%d:%02d - %d:%02d", start_hh, start_mm, end_hh, end_mm),
		}
	}
	return timetable
}

func to_timetable_html(timetable []TableData) string {
	var timetable_html strings.Builder
	timetable_tmpl, err := template.New("timetable").Parse(TIMETABLE_TEMPLATE)
	if err != nil {
		panic(err)
	}
	err = timetable_tmpl.Execute(&timetable_html, timetable)
	if err != nil {
		panic(err)
	}
	return timetable_html.String()
}

func generate_html(filepath string, tmpl_filepath string, text string) {
	file, err := os.Create(filepath)
	defer file.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create file %s\n", filepath)
		panic(err)
	}

	html_tmpl, err := template.ParseFiles(tmpl_filepath)
	if err != nil {
		panic(err)
	}
	err = html_tmpl.Execute(file, template.HTML(text))
	if err != nil {
		panic(err)
	}
}
