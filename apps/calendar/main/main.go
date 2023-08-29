/*
Copyright 2023 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"strconv"
	"time"
)

type Storage struct {
	ShowSide bool
	Mode     string

	Small_date int64
	Small_page int64
}

type Translations struct {
	YEAR      string
	MONTH     string
	DAY       string
	WEEK      string
	JANUARY   string
	FEBRUARY  string
	MARCH     string
	APRIL     string
	MAY       string
	JUNE      string
	JULY      string
	AUGUST    string
	SEPTEMBER string
	OCTOBER   string
	NOVEMBER  string
	DECEMBER  string
	MON       string
	TUE       string
	WED       string
	THU       string
	FRI       string
	SAT       string
	SUN       string

	MONDAY    string
	TUESDAY   string
	WEDNESDAY string
	THURSDAY  string
	FRIDAY    string
	SATURDAY  string
	SUNDAY    string

	NEW_ITEM string
	OK       string
	TODAY    string
}

func MonthText(month int) string {
	switch month {
	case 1:
		return trns.JANUARY
	case 2:
		return trns.FEBRUARY
	case 3:
		return trns.MARCH
	case 4:
		return trns.APRIL
	case 5:
		return trns.MAY
	case 6:
		return trns.JUNE
	case 7:
		return trns.JULY
	case 8:
		return trns.AUGUST
	case 9:
		return trns.SEPTEMBER
	case 10:
		return trns.OCTOBER
	case 11:
		return trns.NOVEMBER
	case 12:
		return trns.DECEMBER
	}
	return ""
}

func DayTextFull(day int) string {

	switch day {
	case 1:
		return trns.MONDAY
	case 2:
		return trns.TUESDAY
	case 3:
		return trns.WEDNESDAY
	case 4:
		return trns.THURSDAY
	case 5:
		return trns.FRIDAY
	case 6:
		return trns.SATURDAY
	case 7:
		return trns.SUNDAY
	}
	return ""
}

func DayTextShort(day int) string {

	switch day {
	case 1:
		return trns.MON
	case 2:
		return trns.TUE
	case 3:
		return trns.WED
	case 4:
		return trns.THU
	case 5:
		return trns.FRI
	case 6:
		return trns.SAT
	case 7:
		return trns.SUN
	}
	return ""
}

func GetWeekDayPure(date time.Time) int {
	week := int(date.Weekday()) //sun=0, mon=1, etc.
	if week == 0 {
		week = 7
	}
	return week
}

func GetWeekDay(date time.Time, format float64) int {
	week := int(date.Weekday()) //sun=0, mon=1, etc.
	if format != 1 {
		//not "us"
		week -= 1
		if week < 0 {
			week = 6
		}
	}
	return week
}

func GetStartWeekDay(date time.Time, format float64) time.Time {
	weekDay := GetWeekDay(date, format)
	return date.AddDate(0, 0, -weekDay)
}

func GetFullDay(unix_sec int64) string {

	tm := time.Unix(unix_sec, 0)

	d := strconv.Itoa(tm.Day())
	m := strconv.Itoa(int(tm.Month()))
	y := strconv.Itoa(tm.Year())

	switch SA_InfoFloat("date") {
	case 0: //eu
		return d + "/" + m + "/" + y

	case 1: //us
		return m + "/" + d + "/" + y

	case 2: //iso
		return y + "-" + fmt.Sprintf("%02d", m) + "-" + fmt.Sprintf("%02d", d)

	case 3: //text
		//return month + " " + d + "," + y	//...

	case 4: //2base
		return y + fmt.Sprintf("%02d", m) + fmt.Sprintf("%02d", d)
	}

	return ""
}

func GetMonthYear(unix_sec int64) string {
	tm := time.Unix(unix_sec, 0)
	return MonthText(int(tm.Month())) + " " + strconv.Itoa(tm.Year())
}

func GetYear(unix_sec int64) string {
	tm := time.Unix(unix_sec, 0)
	return strconv.Itoa(tm.Year())
}

func CmpDates(a int64, b int64) int64 {
	ta := time.Unix(a, 0)
	tb := time.Unix(b, 0)

	if ta.Year() == tb.Year() && ta.Month() == tb.Month() && ta.Day() == tb.Day() {
		return 1
	}
	return 0
}

func Calendar(value *int64, page *int64) {
	format := SA_InfoFloat("date")

	for x := 0; x < 7; x++ {
		SA_ColMax(x, 0.9)
	}
	for y := 0; y < 6; y++ {
		SA_RowMax(y, 0.9)
	}

	//fix page(need to start with day 1)
	{
		dtt := time.Unix(*page, 0)
		*page = dtt.AddDate(0, 0, -(dtt.Day() - 1)).Unix()
	}

	//--Day names(short)--
	if format == 1 {
		//"us"
		SA_Text(DayTextShort(7)).Align(1).Show(0, 0, 1, 1)
		for x := 1; x < 7; x++ {
			SA_Text(DayTextShort(x)).Align(1).Show(x, 0, 1, 1)
		}
	} else {
		for x := 1; x < 8; x++ {
			SA_Text(DayTextShort(x)).Align(1).Show(x-1, 0, 1, 1)
		}
	}

	//--Week days--
	now := int64(SA_Time())
	orig_dtt := time.Unix(*page, 0)
	dtt := GetStartWeekDay(orig_dtt, format)

	for y := 0; y < 6; y++ {
		for x := 0; x < 7; x++ {
			backCd := SA_ThemeCd()
			frontCd := SA_ThemeBlack()

			if CmpDates(dtt.Unix(), *value) > 0 { //selected day
				backCd = SA_ThemeBlack()
				frontCd = SA_ThemeWhite()
			}

			if dtt.Month() != orig_dtt.Month() { //is day in current month
				backCd = SA_ThemeCd().Aprox(SA_ThemeWhite(), 0.7)
			}

			if SA_Button(strconv.Itoa(dtt.Day())).Alpha(1).FrontCd(frontCd).BackCd(backCd).Border(CmpDates(dtt.Unix(), now) > 0).Show(x, 1+y, 1, 1).click {
				*value = dtt.Unix()
				*page = *value
			}

			dtt = dtt.AddDate(0, 0, 1) //add day
		}
	}
}

func Side() {

	SA_ColMax(0, 100)

	SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeGrey(0.97), 0) //paintRect(color: themeGrey(0.97))

	if store.ShowSide {
		SA_Row(1, 0.3)
		SA_Row(2, 1.2)
		SA_Row(3, 6.5)
		SA_Row(4, 0.3)
		SA_RowMax(5, 100)

		if SA_Button(trns.NEW_ITEM).Show(0, 0, 1, 1).click {
			//dialog + save into db ...
		}

		SA_RowSpacer(0, 1, 1, 1)

		SA_DivStart(0, 2, 1, 1)
		{
			SA_ColMax(0, 100)
			SA_Text(GetMonthYear(store.Small_page)).RatioH(0.5).Show(0, 0, 1, 1)

			if SA_Button("<").Alpha(0.5).Show(1, 0, 1, 1).click {
				tm := time.Unix(store.Small_page, 0)
				store.Small_page = tm.AddDate(0, -1, 0).Unix()
			}
			if SA_Button(">").Alpha(0.5).Show(2, 0, 1, 1).click {
				tm := time.Unix(store.Small_page, 0)
				store.Small_page = tm.AddDate(0, 1, 0).Unix()
			}
		}
		SA_DivEnd()

		SA_DivStart(0, 3, 1, 1)
		Calendar(&store.Small_date, &store.Small_page)
		SA_DivEnd()

		SA_RowSpacer(0, 4, 1, 1)

		SA_DivStart(0, 6, 1, 1)
		{
			SA_ColMax(0, 100)
			if SA_Button("<<").Show(1, 0, 1, 1).click {
				store.ShowSide = false
			}
		}
		SA_DivEnd()
	} else {
		SA_RowMax(0, 100)
		if SA_Button(">>").Show(0, 1, 1, 1).click {
			store.ShowSide = true
		}
	}
}

func ModeYear() {

	w := int(SA_DivInfo("screenWidth") / 7) //7cells is needed for one calendar
	if w == 0 {
		return
	}
	h := 12 / w
	if 12%w > 0 {
		h++
	}

	for x := 0; x < w*2; x += 2 {
		SA_Col(x, 7)

		SA_Col(x+1, 0.1)
		SA_ColMax(x+1, 100)
	}
	for y := 0; y < h; y++ {
		SA_Row(y, 8)
	}

	year := time.Unix(store.Small_date, 0).Year()
	i := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if i < 12 {
				if SA_DivStart(x*2, y, 1, 1) {

					page := time.Date(year, time.Month(1+i), 1, 0, 0, 0, 0, time.Now().Location()).Unix()

					SA_ColMax(0, 100)
					SA_RowMax(1, 100)

					if SA_Button(MonthText(1+i)).Alpha(1).RatioH(0.45).Align(0).AlphaNoBack(true).Show(0, 0, 1, 1).click {

						//change month = i+1
						t := time.Unix(store.Small_date, 0)
						store.Small_date = time.Date(t.Year(), time.Month(i+1), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
						store.Mode = "month"
					}

					if SA_DivStart(0, 1, 1, 1) {
						Calendar(&store.Small_date, &page)
					}
					SA_DivEnd()

				}
				SA_DivEnd()
				i += 1
			}
		}
	}
}

func ModeMonth() {

	format := SA_InfoFloat("date")

	for x := 0; x < 7; x++ {
		SA_ColMax(x, 100)
	}
	for y := 0; y < 6; y++ {
		SA_RowMax(1+y, 100)
	}

	//--Day names(short)--
	if format == 1 {
		//"us"
		SA_Text(DayTextShort(7)).Align(1).Show(0, 0, 1, 1)
		for x := 1; x < 7; x++ {
			SA_Text(DayTextShort(x)).Align(1).Show(x, 0, 1, 1)
		}
	} else {
		for x := 1; x < 8; x++ {
			SA_Text(DayTextShort(x)).Align(1).Show(x-1, 0, 1, 1)
		}
	}

	//grid
	for y := 0; y < 6; y++ {
		SA_DivStart(0, 1+y, 7, 1)
		SAPaint_Line(0, 0, 1, 0, SA_ThemeGrey(0.75), 0.03)
		//paintLine(sy: 0, ey: 0, width: 0.03, color: themeGrey(0.75))
		SA_DivEnd()
	}

	for x := 0; x < 6; x++ {
		SA_DivStart(1+x, 1, 1, 6)
		SAPaint_Line(0, 0, 0, 1, SA_ThemeGrey(0.75), 0.03)
		//paintLine(sx: 0, ex: 0, width: 0.03, color: themeGrey(0.75))
		SA_DivEnd()
	}

	{
		//fix page(need to start with day 1)
		orig_dtt := time.Unix(store.Small_page, 0)
		//dtt := orig_dtt
		//store.Small_page = dtt.AddDate(0, 0, -(dtt.Day() - 1)).Unix()
		dtt := GetStartWeekDay(orig_dtt, format)

		for y := 0; y < 6; y++ {
			for x := 0; x < 7; x++ {
				frontCd := SA_ThemeBlack()
				if dtt.Month() != orig_dtt.Month() { //is day in current month
					frontCd = SA_ThemeGrey(0.75)
				}

				SA_DivStart(x, 1+y, 1, 1)
				{
					isToday := CmpDates(dtt.Unix(), int64(SA_Time())) != 0
					if isToday {
						SAPaint_Rect(0, 0, 1, 1, 0.03, SA_ThemeWhite().Aprox(SA_ThemeCd(), 0.3), 0)
					}

					SA_ColMax(0, 100)
					if SA_Button(strconv.Itoa(dtt.Day())+".").Alpha(1).FrontCd(frontCd).RatioH(0.4).Align(0).AlphaNoBack(true).Show(0, 0, 1, 1).click {
						store.Small_date = dtt.Unix()
						store.Mode = "day"
					}

					//db ...
					//paintRect(borderWidth:0.03, margin: 0.1, color: themeGrey())

				}
				SA_DivEnd()

				dtt = dtt.AddDate(0, 0, 1) //add day
			}
		}
	}
}

func ModeWeek() {

	format := SA_InfoFloat("date")

	SA_ColMax(0, 100)
	SA_RowMax(1, 100)

	//header
	dtt := GetStartWeekDay(time.Unix(store.Small_date, 0), format)

	SA_DivStart(0, 0, 1, 1)
	{
		SA_Col(0, 1.5) //time
		for i := 1; i < 8; i++ {
			SA_ColMax(i, 100)
		}

		changeDay := -1
		if format == 1 {
			//"us"
			if SA_Button(strconv.Itoa(dtt.Day())+". "+DayTextShort(7)).Alpha(1).RatioH(0.4).AlphaNoBack(true).Align(0).Show(1, 0, 1, 1).click {
				changeDay = dtt.Day()
			}

			dtt = dtt.AddDate(0, 0, 1) //add day

			for x := 1; x < 7; x++ {
				if SA_Button(strconv.Itoa(dtt.Day())+". "+DayTextShort(x)).Alpha(1).RatioH(0.4).AlphaNoBack(true).Align(0).Show(1+x, 0, 1, 1).click {
					changeDay = dtt.Day()
				}
				dtt = dtt.AddDate(0, 0, 1) //add day
			}
		} else {
			for x := 1; x < 8; x++ {
				if SA_Button(strconv.Itoa(dtt.Day())+". "+DayTextShort(x)).Alpha(1).RatioH(0.4).AlphaNoBack(true).Align(0).Show(x, 0, 1, 1).click {
					changeDay = dtt.Day()
				}
				dtt = dtt.AddDate(0, 0, 1) //add day
			}
		}

		if changeDay >= 0 {
			//change day = changeDay
			t := time.Unix(store.Small_date, 0)
			store.Small_date = time.Date(t.Year(), t.Month(), changeDay, 0, 0, 0, 0, t.Location()).Unix()

			store.Mode = "day"
		}

	}
	SA_DivEnd()

	//days
	SA_DivStart(0, 1, 1, 1)
	{
		SA_Col(0, 1.5) //time
		for i := 1; i < 8; i++ {
			SA_ColMax(i, 100)
		}
		for i := 0; i < 25; i++ {
			SA_Row(i*2, 0.3)
			SA_Row(i*2+1, 1.5)
		}

		//time
		for y := 0; y < 25; y++ {
			SA_Text(strconv.Itoa(y)+":00").RatioH(0.3).Align(1).AlignV(0).Show(0, y*2, 1, 1)
		}

		//grid
		for y := 0; y < 25; y++ {
			SA_DivStart(1, y*2, 7, 1)
			SAPaint_Line(0, 0.5, 1, 0.5, SA_ThemeGrey(0.75), 0.03)
			SA_DivEnd()
		}

		for x := 1; x < 7; x++ {
			SA_DivStart(1+x, 0, 1, 24*2+1)
			SAPaint_Line(0, 0, 0, 1, SA_ThemeGrey(0.75), 0.03)
			SA_DivEnd()
		}

		for y := 0; y < 24; y++ {
			for x := 0; x < 7; x++ {
				SA_DivStart(1+x, y*2+1, 1, 1)
				//db ...
				//paintRect(borderWidth:0.03, margin: 0.1)
				SA_DivEnd()
			}
		}

		//time-line
		w1 := GetStartWeekDay(time.Now(), format).Unix()
		w2 := GetStartWeekDay(time.Unix(store.Small_date, 0), format).Unix()
		if CmpDates(w1, w2) > 0 { //today is in current week

			dt := time.Now()
			h := (float64(dt.Hour()) + (float64(dt.Minute()) / 60)) / 24
			week := GetWeekDay(dt, format)

			SA_DivStart(1, 0, 7, 24*2+1)
			{
				SAPaint_Line(0, h, 1, h, SA_ThemeEdit(), 0.03)             //... marginY: 0.3/2)
				SAPaint_Circle(float64(week)/7, h, 0.1, SA_ThemeEdit(), 0) //... marginX: 0.1, marginY: 0.3/2)
			}
			SA_DivEnd()
		}

	}
	SA_DivEnd()

}

func ModeDay() {
	SA_ColMax(0, 100)
	SA_RowMax(1, 100)

	//header
	SA_DivStart(0, 0, 1, 1)
	{
		SA_Col(0, 1.5) //time
		SA_ColMax(1, 100)

		dtt := time.Unix(store.Small_date, 0)
		SA_Text(strconv.Itoa(dtt.Day())+". "+DayTextFull(GetWeekDayPure(dtt))).RatioH(0.4).Show(1, 0, 1, 1)
	}
	SA_DivEnd()

	//days
	SA_DivStart(0, 1, 1, 1)
	{
		SA_Col(0, 1.5) //time
		SA_ColMax(1, 100)

		for y := 0; y < 25; y++ {
			SA_Row(y*2, 0.3)
			SA_Row(y*2+1, 1.5)
		}

		//time
		for y := 0; y < 25; y++ {
			SA_Text(strconv.Itoa(y)+":00").RatioH(0.3).Align(1).AlignV(0).Show(0, y*2, 1, 1)
		}

		//grid
		for y := 0; y < 25; y++ {
			SA_DivStart(1, y*2, 1, 1)
			SAPaint_Line(0, 0.5, 1, 0.5, SA_ThemeGrey(0.75), 0.03)
			SA_DivEnd()
		}

		for y := 0; y < 24; y++ {
			SA_DivStart(1, y*2+1, 1, 1)
			//db ...
			//paintRect(borderWidth:0.03, margin: 0.1)
			SA_DivEnd()
		}

		//time-line
		if CmpDates(time.Now().Unix(), store.Small_date) > 0 { //today == day

			dt := time.Now()
			h := (float64(dt.Hour()) + (float64(dt.Minute()) / 60)) / 24
			SA_DivStart(1, 0, 1, 24*2+1)
			{
				SAPaint_Line(0, h, 1, h, SA_ThemeEdit(), 0.03) //... marginY: 0.3/2)
				SAPaint_Circle(0, h, 0.1, SA_ThemeEdit(), 0)   //... marginX: 0.1, marginY: 0.3/2)
			}
			SA_DivEnd()
		}

	}
	SA_DivEnd()
}

func ModePanel() {
	SA_ColMax(0, 100)
	SA_RowMax(1, 100)

	var title string
	SA_DivStart(0, 1, 1, 1)
	{
		if store.Mode == "year" {
			title = GetYear(store.Small_date)
			ModeYear()
		} else if store.Mode == "month" {
			title = GetMonthYear(store.Small_date)
			ModeMonth()
		} else if store.Mode == "week" {
			title = GetMonthYear(store.Small_date)
			ModeWeek()
		} else if store.Mode == "day" {
			title = GetFullDay(store.Small_date)
			ModeDay()
		}
	}
	SA_DivEnd()

	SA_DivStart(0, 0, 1, 1)
	{
		SA_ColMax(0, 2)
		SA_ColMax(3, 100)
		SA_ColMax(4, 8)

		//today
		if SA_Button(trns.TODAY).Title(GetFullDay(int64(SA_Time()))).Alpha(0.5).Show(0, 0, 1, 1).click {
			store.Small_date = int64(SA_Time())
			store.Small_page = int64(SA_Time())
		}

		//arrows
		if SA_Button("<").Alpha(0.5).Show(1, 0, 1, 1).click {
			tm := time.Unix(store.Small_date, 0)
			if store.Mode == "year" {
				store.Small_date = tm.AddDate(-1, 0, 0).Unix()
			} else if store.Mode == "month" {
				store.Small_date = tm.AddDate(0, -1, 0).Unix()
			} else if store.Mode == "week" {
				store.Small_date = tm.AddDate(0, 0, -7).Unix()
			} else if store.Mode == "day" {
				store.Small_date = tm.AddDate(0, 0, -1).Unix()
			}
		}
		if SA_Button(">").Alpha(0.5).Show(2, 0, 1, 1).click {
			tm := time.Unix(store.Small_date, 0)
			if store.Mode == "year" {
				store.Small_date = tm.AddDate(1, 0, 0).Unix()
			} else if store.Mode == "month" {
				store.Small_date = tm.AddDate(0, 1, 0).Unix()
			} else if store.Mode == "week" {
				store.Small_date = tm.AddDate(0, 0, 7).Unix()
			} else if store.Mode == "day" {
				store.Small_date = tm.AddDate(0, 0, 1).Unix()
			}
		}

		//title
		SA_Text(title).RatioH(0.5).Align(1).Show(3, 0, 1, 1)

		//Modes
		SA_DivStart(4, 0, 1, 1)
		{
			SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeGrey(0.9), 0)

			for i := 0; i < 4; i++ {
				SA_ColMax(i*2, 2)
				SA_Col(i*2+1, 0.2)
			}
			for i := 0; i < 4; i++ {
				if i < 3 {
					SA_ColSpacer(i*2+1, 0, 1, 1)
				}
			}

			if SA_Button(trns.DAY).Alpha(1).Highlight(store.Mode == "day").Show(0, 0, 1, 1).click {
				store.Mode = "day"
			}
			if SA_Button(trns.WEEK).Alpha(1).Highlight(store.Mode == "week").Show(2, 0, 1, 1).click {
				store.Mode = "week"
			}
			if SA_Button(trns.MONTH).Alpha(1).Highlight(store.Mode == "month").Show(4, 0, 1, 1).click {
				store.Mode = "month"
			}
			if SA_Button(trns.YEAR).Alpha(1).Highlight(store.Mode == "year").Show(6, 0, 1, 1).click {
				store.Mode = "year"
			}

			//paintRect(color: themeGrey(0.3), borderWidth: 0.03)
		}
		SA_DivEnd()

	}
	SA_DivEnd()

}

//export render
func render() uint32 {

	if store.ShowSide {
		SA_Col(0, 6.3)
	}
	SA_ColMax(1, 100)
	SA_RowMax(0, 100)

	SA_DivStart(0, 0, 1, 1)
	Side()
	SA_DivEnd()

	SA_DivStart(1, 0, 1, 1)
	ModePanel()
	SA_DivEnd()

	return 0
}

func open(buff []byte) bool {
	//init
	store.ShowSide = true
	store.Small_date = int64(SA_Time())
	store.Small_page = int64(SA_Time())

	return false //default json
}
func save() ([]byte, bool) {
	return nil, false //default json
}
func debug() (int, int, string) {
	return -1, 00, "main"
}
