package image

import (
	"net/http"
	"testing"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

type histogram struct {
	data []uint16
}

func (h *histogram) serve(w http.ResponseWriter, _ *http.Request) {
	line := charts.NewLine()

	items := make([]opts.LineData, 0)
	axis := make([]int, 0)

	for i, val := range h.data {
		axis = append(axis, i)
		items = append(items, opts.LineData{Value: val})
	}

	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros, Width: "100%"}))
	line.SetXAxis(axis)
	line.AddSeries("Histogram", items)
	line.Render(w)
}

func Test(t *testing.T) {
	img, err := ReadFile("./test/salatin_1.png")
	if err != nil {
		t.Fatal(err)
	}

	monochrome := New(img).Monochrome(160, false)
	_histogram := monochrome.Histogram(false)

	h := &histogram{_histogram}
	http.HandleFunc("/", h.serve)
	http.ListenAndServe(":8081", nil)
}
