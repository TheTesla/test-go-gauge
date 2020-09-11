package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"os"
	"io"
	"io/ioutil"
	"strconv"


	"github.com/go-echarts/go-echarts/charts"
)

const (
	host   = "http://127.0.0.1:8080"
	maxNum = 50
)

func gaugeTimer() *charts.Gauge {
	gauge := charts.NewGauge()

	m := make(map[string]interface{})
	m["Wh"] = 0
	gauge.Add("BAT0", m)
	gauge.SetGlobalOptions(charts.TitleOpts{Title: "Battery-Gauge"})
	fn := fmt.Sprintf(`var xhttp = new XMLHttpRequest();
			var i = 0;
			xhttp.onreadystatechange = function() {
				i = xhttp.responseText;
				option_%s.series[0].data[0].value = (i * 1).toFixed(3) - 0;
				myChart_%s.setOption(option_%s, true);
			};
			setInterval(function () {
  				xhttp.open('GET', 'val1', false);
				xhttp.send();
			}, 200);
		`, gauge.ChartID, gauge.ChartID, gauge.ChartID)
	gauge.AddJSFuncs(fn)
	return gauge
}

func gaugeHandler(w http.ResponseWriter, _ *http.Request) {
	page := charts.NewPage(orderRouters("gauge")...)
	page.Add(
		gaugeTimer(),
	)
	f, err := os.Create(getRenderPath("gauge.html"))
	if err != nil {
		log.Println(err)
	}
	page.Render(w, f)
}


func check(e error) {
    if e != nil {
        panic(e)
    }
}


func val1Handler(w http.ResponseWriter, _ *http.Request) {
	sbat, err := ioutil.ReadFile("/sys/devices/LNXSYSTM:00/LNXSYBUS:00/PNP0A08:00/device:08/PNP0C09:00/PNP0C0A:00/power_supply/BAT0/energy_now")
	check(err)
	fbat, err := strconv.ParseFloat(string(sbat)[:len(sbat)-1], 64)
	check(err)
	sbatmax, err := ioutil.ReadFile("/sys/devices/LNXSYSTM:00/LNXSYBUS:00/PNP0A08:00/device:08/PNP0C09:00/PNP0C0A:00/power_supply/BAT0/energy_full")
	check(err)
	fbatmax, err := strconv.ParseFloat(string(sbatmax)[:len(sbatmax)-1], 64)
	check(err)


	w.WriteHeader(http.StatusOK)
	io.WriteString(w, strconv.FormatFloat(fbat/fbatmax*100, 'f', 3, 64))
}

type router struct {
	name string
	charts.RouterOpts
}

var (
	routers = []router{
		{"gauge", charts.RouterOpts{URL: host + "/gauge", Text: "Gauge"}},
	}
)

func orderRouters(chartType string) []charts.RouterOpts {
	for i := 0; i < len(routers); i++ {
		if routers[i].name == chartType {
			routers[i], routers[0] = routers[0], routers[i]
			break
		}
	}

	rs := make([]charts.RouterOpts, 0)
	for i := 0; i < len(routers); i++ {
		rs = append(rs, routers[i].RouterOpts)
	}
	return rs
}

func getRenderPath(f string) string {
	return path.Join("html", f)
}
func logTracing(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Tracing request for %s\n", r.RequestURI)
		next.ServeHTTP(w, r)
	}
}

func main() {
	// Avoid "404 page not found".
	http.HandleFunc("/", logTracing(gaugeHandler))

	http.HandleFunc("/gauge", logTracing(gaugeHandler))
	http.HandleFunc("/val1", logTracing(val1Handler))

	log.Println("Run server at " + host)
	http.ListenAndServe(":8080", nil)
}

