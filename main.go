package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path"
	"os"
	"io"

	"github.com/go-echarts/go-echarts/charts"
)

const (
	host   = "http://127.0.0.1:8080"
	maxNum = 50
)

func gaugeBase() *charts.Gauge {
	gauge := charts.NewGauge()
	gauge.SetGlobalOptions(charts.TitleOpts{Title: "Gauge-base"})
	m := make(map[string]interface{})
	m["A"] = rand.Intn(50)
	gauge.Add("gauge", m)
	return gauge
}

func gaugeTimer() *charts.Gauge {
	gauge := charts.NewGauge()

	m := make(map[string]interface{})
	m["B"] = rand.Intn(50)
	gauge.Add("gauge1", m)
	gauge.SetGlobalOptions(charts.TitleOpts{Title: "Gauge-timer"})
	fn := fmt.Sprintf(`var xhttp = new XMLHttpRequest();
			var i = 0;
			setInterval(function () {
  			xhttp.open('GET', 'val1', false);
			xhttp.send();
			i = xhttp.responseText;
			option_%s.series[0].data[0].value = (i * 1).toFixed(2) - 0;
			myChart_%s.setOption(option_%s, true);
		}, 2000);`, gauge.ChartID, gauge.ChartID, gauge.ChartID)
	gauge.AddJSFuncs(fn)
	return gauge
}

func gaugeHandler(w http.ResponseWriter, _ *http.Request) {
	page := charts.NewPage(orderRouters("gauge")...)
	page.Add(
		gaugeBase(),
		gaugeTimer(),
	)
	f, err := os.Create(getRenderPath("gauge.html"))
	if err != nil {
		log.Println(err)
	}
	page.Render(w, f)
}

func val1Handler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "42\n")
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

