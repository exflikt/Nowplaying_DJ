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

const TIMETABLE_TEMPLATE = `{{with index .Preamble 0}}
<h1 class="dtimer" data-sh="0" data-sm="0" data-eh="{{.Eh}}" data-em="{{.Em}}" data-es="59">Please wait...</h1>
<h2 class="dtimer" data-sh="0" data-sm="0" data-eh="{{.Eh}}" data-em="{{.Em}}" data-es="59">{{.Timeframe}}</h2>
{{end}}
{{with index .Preamble 1}}
<h3 class="dtimer" data-sh="{{.Sh}}" data-sm="{{.Sm}}" data-eh="{{.Eh}}" data-em="{{.Em}}" data-es="59">NOW PLAYING</h3>
{{end}}
{{range .Body}}
<h1 class="dtimer" data-sh="{{.Sh}}" data-sm="{{.Sm}}" data-eh="{{.Eh}}" data-em="{{.Em}}" data-es="59">{{.DjName}}</h1>
<h2 class="dtimer" data-sh="{{.Sh}}" data-sm="{{.Sm}}" data-eh="{{.Eh}}" data-em="{{.Em}}" data-es="59">{{.Timeframe}}</h2>
{{end}}
{{with .Postamble}}
<h1 class="dtimer" data-sh="{{.Sh}}" data-sm="{{.Sm}}" data-eh="23" data-em="59" data-es="59"></h1>
<h2 class="dtimer" data-sh="{{.Sh}}" data-sm="{{.Sm}}" data-eh="23" data-em="59" data-es="59">The event has closed.<br>Thank you for coming!</h2>
{{end}}`

type TimeTable struct {
	Preamble  [2]TableData
	Body     []TableData
	Postamble TableData
}

type TableData struct {
	Sh        uint8
	Sm        uint8
	Eh        uint8
	Em        uint8
	DjName    string
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

func records_to_timetable(records [][]string) TimeTable {
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

	body := make([]TableData, len(records)-1)
	for i := 1; i < len(records); i++ {
		start, end, dj_name := records[i][0], records[i][1], records[i][2]

		start_hh, start_mm := parse_time(split_time(start))
		end_hh, end_mm := parse_time(split_time(end))

		eh, em := one_min_prev(end_hh, end_mm)

		body[i-1] = TableData{
			Sh:        start_hh,
			Sm:        start_mm,
			Eh:        eh,
			Em:        em,
			DjName:    dj_name,
			Timeframe: fmt.Sprintf("%d:%02d - %d:%02d", start_hh, start_mm, end_hh, end_mm),
		}
	}

	event_start := records[1][0]
	event_start_hh, event_start_mm := parse_time(split_time(event_start))
	event_sh, event_sm := one_min_prev(event_start_hh, event_start_mm)
	event_end := records[len(records)-1][1]
	event_end_hh, event_end_mm := parse_time(split_time(event_end))
	event_eh, event_em := one_min_prev(event_end_hh, event_end_mm)
	return TimeTable{
		Preamble: [2]TableData{
			{
				Eh:        event_sh,
				Em:        event_sm,
				Timeframe: fmt.Sprintf("Start at %d:%02d", event_start_hh, event_start_mm),
			},
			{
				Sh: event_start_hh,
				Sm: event_start_mm,
				Eh: event_eh,
				Em: event_em,
			},
		},
		Body: body,
		Postamble: TableData{
			Sh: event_end_hh,
			Sm: event_end_mm,
		},
	}
}

func to_timetable_html(timetable TimeTable) string {
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
