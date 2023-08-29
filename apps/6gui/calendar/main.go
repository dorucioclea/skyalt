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
	Page int64
}

type Translations struct {
	YEAR      string
	MONTH     string
	DAY       string
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

//export FormatDate
func FormatDate(unix_sec int64) {
	str := Format(unix_sec)

	SA_CallSetReturn(str)
}

func Format(unix_sec int64) string {

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
		return y + "-" + fmt.Sprintf("%02d", int(tm.Month())) + "-" + fmt.Sprintf("%02d", tm.Day())

	case 3: //text
		return MonthText(int(tm.Month())) + " " + d + "," + y

	case 4: //2base
		return y + fmt.Sprintf("%02d", int(tm.Month())) + fmt.Sprintf("%02d", tm.Day())
	}

	return ""
}

//export CmpDates
func CmpDates(a int64, b int64) int64 {
	ta := time.Unix(a, 0)
	tb := time.Unix(b, 0)

	if ta.Year() == tb.Year() && ta.Month() == tb.Month() && ta.Day() == tb.Day() {
		return 1
	}
	return 0
}

//export Calendar
func Calendar(value int64) int64 {
	format := SA_InfoFloat("date")

	//init
	if store.Page == 0 {
		store.Page = int64(SA_Time())
	}

	for x := 0; x < 7; x++ {
		SA_ColMax(x, 10)
	}

	//--Today--
	{
		act_tm := int64(SA_Time())
		if SA_Button(Format(act_tm)).Show(0, 0, 7, 1).click {
			value = act_tm
			store.Page = act_tm
		}
	}

	//--Week header navigation--
	{
		tm := time.Unix(store.Page, 0)

		if SA_Button("<<").Show(0, 2, 1, 1).click {
			store.Page = tm.AddDate(-1, 0, 0).Unix()
		}
		if SA_Button("<").Show(1, 2, 1, 1).click {
			store.Page = tm.AddDate(0, -1, 0).Unix()
		}

		SA_Text(MonthText(int(tm.Month()))+" "+strconv.Itoa(tm.Year())).Align(1).Show(2, 2, 3, 1)

		if SA_Button(">").Show(5, 2, 1, 1).click {
			store.Page = tm.AddDate(0, 1, 0).Unix()
		}
		if SA_Button(">>").Show(6, 2, 1, 1).click {
			store.Page = tm.AddDate(1, 0, 0).Unix()
		}
	}

	//fix page(need to start with day 1)
	{
		dtt := time.Unix(store.Page, 0)
		store.Page = dtt.AddDate(0, 0, -(dtt.Day() - 1)).Unix()
	}

	//--Day names(sort)--
	if format == 1 {
		//"us"
		SA_Text(DayTextShort(7)).Show(0, 3, 1, 1)
		for x := 1; x < 7; x++ {
			SA_Text(DayTextShort(x)).Show(x, 3, 1, 1)
		}
	} else {
		for x := 1; x < 8; x++ {
			SA_Text(DayTextShort(x)).Show(x-1, 3, 1, 1)
		}
	}

	//--Week days--
	now := int64(SA_Time())
	orig_dtt := time.Unix(store.Page, 0)
	dtt := orig_dtt
	weekDay := int(dtt.Weekday()) //sun=0, mon=1, etc.
	if format != 1 {
		//not "us"
		weekDay -= 1
		if weekDay < 0 {
			weekDay = 6
		}
	}
	dtt = dtt.AddDate(0, 0, -weekDay)
	for y := 0; y < 6; y++ {
		for x := 0; x < 7; x++ {
			backCd := SA_ThemeCd()
			frontCd := SA_ThemeBlack()

			if CmpDates(dtt.Unix(), value) > 0 { //highlightDate

				//...
				backCd = SA_ThemeBlack()
				frontCd = SA_ThemeWhite()
			}

			if dtt.Month() != orig_dtt.Month() {
				backCd = SA_ThemeCd().Aprox(SA_ThemeWhite(), 0.7)
			}

			if SA_Button(strconv.Itoa(dtt.Day())).FrontCd(frontCd).BackCd(backCd).Border(CmpDates(dtt.Unix(), now) > 0).Show(x, 4+y, 1, 1).click {
				value = dtt.Unix()
				store.Page = value
			}

			dtt = dtt.AddDate(0, 0, 1) //add day
		}
	}

	return value
}

//export CalendarButton
func CalendarButton(date int64, enable uint32) int64 {

	SA_ColMax(0, 100)
	SA_RowMax(0, 100)
	open := false
	if SA_Button(Format(date)).Enable(enable != 0).Show(0, 0, 1, 1).click {
		open = true
	}

	if SA_DialogStart("CalendarButton", 1, open) {
		SA_ColMax(0, 15)
		SA_RowMax(0, 10)
		SA_DivStart(0, 0, 1, 1)
		date = Calendar(date)
		SA_DivEnd()

		SA_DialogEnd()
	}

	return date
}

//export render
func render() uint32 {

	SA_ColMax(0, 100)
	SA_RowMax(0, 100)
	SA_DivStart(0, 0, 1, 1)
	CalendarButton(int64(SA_Time()), 1)
	SA_DivEnd()

	return 0
}

func open(buff []byte) bool {
	return false //default json
}
func save() ([]byte, bool) {
	return nil, false //default json
}
func debug() (int, int, string) {
	return -1, 00, "main"
}
