package out

import (
	"bytes"
	"crypto_parser/internal/reporting/domain/ports"
	"crypto_parser/internal/reporting/domain/valueobject"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/go-pdf/fpdf"
	chart "github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

const (
	extremeFear  = "Extreme Fear"
	fear         = "Fear"
	neutral      = "Neutral"
	greed        = "Greed"
	extremeGreed = "Extreme Greed"
)

type PdfGenerator struct {
	basePath string
}

func NewPdfGenerator(basePath string) ports.ReportGenerator {
	os.MkdirAll(basePath, os.ModeDir)
	return &PdfGenerator{
		basePath: basePath,
	}
}

// Generate рендерит PDF-отчёт: бар-чарт 24h-изменений + текстовые секции.
func (p PdfGenerator) Generate(data valueobject.ReportData) (string, error) {
	if data.GeneratedAt.IsZero() {
		data.GeneratedAt = time.Now().UTC()
	}

	chartPNG, err := p.renderChart(data)
	if err != nil {
		return "", fmt.Errorf("render chart: %w", err)
	}

	return p.renderPDF(data, chartPNG)
}

// renderChart строит bar-chart 24h-изменений в PNG.
func (p PdfGenerator) renderChart(data valueobject.ReportData) ([]byte, error) {
	colorFor := func(v float64) drawing.Color {
		if v >= 0 {
			return drawing.ColorFromHex("27ae60")
		}
		return drawing.ColorFromHex("e74c3c")
	}

	styleFor := func(v float64) chart.Style {
		c := colorFor(v)
		return chart.Style{FillColor: c, StrokeColor: c}
	}

	v1 := data.MarketCap.Change24hPct
	v2 := data.CMC20.Change24hPct

	// Y range always includes 0; pad 30% of the data spread (min 0.5%)
	yMin := math.Min(math.Min(v1, v2), 0)
	yMax := math.Max(math.Max(v1, v2), 0)
	pad := math.Max(math.Max(math.Abs(yMin), math.Abs(yMax))*0.3, 0.5)
	yMin -= pad
	yMax += pad

	bc := chart.BarChart{
		Title: "24h % Changes",
		TitleStyle: chart.Style{
			FontSize:  13,
			FontColor: drawing.ColorFromHex("222222"),
		},
		Background: chart.Style{
			Padding: chart.Box{Top: 40, Left: 30, Right: 30, Bottom: 50},
		},
		Width:    520,
		Height:   340,
		BarWidth: 90,
		YAxis: chart.YAxis{
			Style: chart.Style{FontSize: 10},
			Range: &chart.ContinuousRange{Min: yMin, Max: yMax},
		},
		Bars: []chart.Value{
			{Label: "Market Cap", Value: v1, Style: styleFor(v1)},
			{Label: "CMC20", Value: v2, Style: styleFor(v2)},
		},
	}

	var buf bytes.Buffer
	if err := bc.Render(chart.PNG, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// renderPDF собирает итоговый PDF из данных и PNG-чарта.
func (p PdfGenerator) renderPDF(data valueobject.ReportData, chartPNG []byte) (string, error) {
	const (
		marginL = 20.0
		pageW   = 170.0 // A4 (210) − 2×20 мм
	)

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetMargins(marginL, 20, marginL)

	// ── Заголовок ────────────────────────────────────────────────
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetTextColor(30, 30, 30)
	pdf.CellFormat(pageW, 10, "Crypto Market Report", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(130, 130, 130)
	pdf.CellFormat(pageW, 6, data.GeneratedAt.Format("02 Jan 2006, 15:04 UTC"), "", 1, "C", false, 0, "")
	pdf.Ln(4)
	hline(pdf, pageW)
	pdf.Ln(5)

	// ── CMC20 & Market Cap ───────────────────────────────────────
	metricRow(pdf, pageW, "CMC20 Index", formatUSD(data.CMC20.Value), data.CMC20.Change24hPct)
	pdf.Ln(3)
	metricRow(pdf, pageW, "Market Cap", formatUSD(data.MarketCap.Value), data.MarketCap.Change24hPct)
	pdf.Ln(5)
	hline(pdf, pageW)
	pdf.Ln(5)

	// ── Fear & Greed ─────────────────────────────────────────────
	r, g, b := fgRGB(data.FearGreed.Label)
	sectionHeader(pdf, pageW, "Fear & Greed Index",
		fmt.Sprintf("%d / 100", data.FearGreed.Value),
		string(data.FearGreed.Label), r, g, b)
	pdf.Ln(1)
	progressBar(pdf, marginL, pageW, data.FearGreed.Value, 100, r, g, b)
	pdf.Ln(6)

	// ── Altcoin Season ────────────────────────────────────────────
	var ar, ag, ab int
	var seasonLabel string
	if data.AltcoinSeason.IsAltSeason {
		ar, ag, ab = 39, 174, 96
		seasonLabel = "Alt Season"
	} else {
		ar, ag, ab = 230, 126, 34
		seasonLabel = "Bitcoin Season"
	}
	sectionHeader(pdf, pageW, "Altcoin Season Index",
		fmt.Sprintf("%.1f%%", data.AltcoinSeason.Index),
		seasonLabel, ar, ag, ab)
	pdf.Ln(1)
	progressBar(pdf, marginL, pageW, int(data.AltcoinSeason.Index), 100, ar, ag, ab)
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(140, 140, 140)
	pdf.CellFormat(pageW, 5,
		fmt.Sprintf("%d / %d coins outperformed BTC (7d)", data.AltcoinSeason.Outperformed, data.AltcoinSeason.Total),
		"", 1, "L", false, 0, "")
	pdf.Ln(6)
	hline(pdf, pageW)
	pdf.Ln(6)

	// ── Chart ─────────────────────────────────────────────────────
	// Чарт: 520×340px → при ширине pageW=170mm высота ≈ 170*(340/520) ≈ 111mm.
	// flow=false не двигает курсор, поэтому сдвигаем Y вручную.
	const chartH = 111.0
	chartY := pdf.GetY()
	pdf.RegisterImageOptionsReader("chart",
		fpdf.ImageOptions{ImageType: "PNG"},
		bytes.NewReader(chartPNG))
	pdf.ImageOptions("chart", marginL, chartY, pageW, chartH, false,
		fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	pdf.SetY(chartY + chartH + 4)

	hline(pdf, pageW)
	pdf.Ln(5)

	// ── Описание параметров ───────────────────────────────────────
	descTitle(pdf, pageW, "About the metrics")

	descItem(pdf, pageW,
		"CMC20 Index",
		"Total market capitalisation of the top-20 coins by CoinMarketCap. "+
			"Reflects the state of the largest crypto assets. "+
			"The 24h change shows the daily dynamics of the index.")

	descItem(pdf, pageW,
		"Market Cap",
		"Total market capitalisation of the entire crypto market in USD. "+
			"Growth signals capital inflow; decline signals outflow. "+
			"The 24h change is calculated relative to the previous day's value.")

	descItem(pdf, pageW,
		"Fear & Greed Index",
		"Sentiment index (0-100): 0-24 Extreme Fear, 25-44 Fear, "+
			"45-55 Neutral, 56-75 Greed, 76-100 Extreme Greed. "+
			"Extreme readings often indicate potential market reversals.")

	descItem(pdf, pageW,
		"Altcoin Season Index",
		"Shows what % of top-150 coins (excl. BTC and stablecoins) BTC outperformed over the last 7 days. "+
			"Low value (<=25%): Alt Season - most altcoins are beating BTC. "+
			"High value: Bitcoin Season - BTC dominates the market.")

	descItem(pdf, pageW,
		"Chart: 24h % Changes",
		"Bar chart showing the 24h percentage change for CMC20 and Market Cap. "+
			"Green bar = growth, red bar = decline. "+
			"Allows quick comparison of the two key index dynamics.")

	stamp := time.Now().UnixNano()

	fpath := fmt.Sprintf("%d.pdf", stamp)

	withBasePath := filepath.Join(p.basePath, fpath)

	w, err := os.Create(withBasePath)
	if err != nil {
		return "", err
	}
	defer w.Close()

	if err := pdf.Output(w); err != nil {
		return "", err
	}
	return w.Name(), nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func hline(pdf *fpdf.Fpdf, w float64) {
	pdf.SetDrawColor(210, 210, 210)
	pdf.Line(20, pdf.GetY(), 20+w, pdf.GetY())
}

func metricRow(pdf *fpdf.Fpdf, pageW float64, label, value string, change float64) {
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(55, 55, 55)
	pdf.CellFormat(75, 7, label, "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(50, 7, value, "", 0, "C", false, 0, "")

	if change >= 0 {
		pdf.SetTextColor(39, 174, 96)
	} else {
		pdf.SetTextColor(231, 76, 60)
	}
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(pageW-125, 7, formatChange(change), "", 1, "R", false, 0, "")
}

func sectionHeader(pdf *fpdf.Fpdf, pageW float64, label, value, badge string, r, g, b int) {
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(55, 55, 55)
	pdf.CellFormat(75, 7, label, "", 0, "L", false, 0, "")
	pdf.CellFormat(40, 7, value, "", 0, "C", false, 0, "")
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(pageW-115, 7, badge, "", 1, "R", false, 0, "")
}

func progressBar(pdf *fpdf.Fpdf, x, w float64, value, max, r, g, b int) {
	const h = 5.0
	y := pdf.GetY()
	pdf.SetFillColor(225, 225, 225)
	pdf.Rect(x, y, w, h, "F")
	if fill := w * float64(value) / float64(max); fill > 0 {
		pdf.SetFillColor(r, g, b)
		pdf.Rect(x, y, fill, h, "F")
	}
	pdf.Ln(h)
}

func formatUSD(v float64) string {
	switch {
	case v >= 1e12:
		return fmt.Sprintf("$%.2fT", v/1e12)
	case v >= 1e9:
		return fmt.Sprintf("$%.2fB", v/1e9)
	case v >= 1e6:
		return fmt.Sprintf("$%.2fM", v/1e6)
	default:
		return fmt.Sprintf("$%.2f", v)
	}
}

func formatChange(v float64) string {
	if v >= 0 {
		return fmt.Sprintf("+%.2f%%", v)
	}
	return fmt.Sprintf("%.2f%%", v)
}

func descTitle(pdf *fpdf.Fpdf, pageW float64, title string) {
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(55, 55, 55)
	pdf.CellFormat(pageW, 6, title, "", 1, "L", false, 0, "")
	pdf.Ln(2)
}

func descItem(pdf *fpdf.Fpdf, pageW float64, label, text string) {
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(pageW, 5, label, "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.MultiCell(pageW, 4, text, "", "L", false)
	pdf.Ln(2)
}

func fgRGB(label string) (r, g, b int) {
	switch label {
	case extremeFear:
		return 192, 57, 43
	case fear:
		return 211, 84, 0
	case neutral:
		return 52, 73, 94
	case greed:
		return 39, 174, 96
	case extremeGreed:
		return 26, 188, 156
	default:
		return 100, 100, 100
	}
}
