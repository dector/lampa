package collect

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dector/lampa/internal/report"
	pages "github.com/dector/lampa/internal/templates/html"
	"github.com/dector/lampa/internal/utils"
)

func GenerateHtmlReport(r *report.Report) (string, error) {
	w := &strings.Builder{}
	err := pages.CollectHtml(r).Render(context.Background(), w)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}

func WriteHtmlReportToFile(report *report.Report, args ExecArgs) error {
	err := utils.EnsureParentDirExists(args.HtmlReportFile)
	if err != nil {
		return err
	}

	reportHtml, err := GenerateHtmlReport(report)
	if err != nil {
		return fmt.Errorf("could not generate HTML report: %v", err)
	}
	file, err := os.Create(args.HtmlReportFile)
	if err != nil {
		return fmt.Errorf("could not create HTML report file: %v", err)
	}
	defer file.Close()

	if _, err := file.Write([]byte(reportHtml)); err != nil {
		return fmt.Errorf("could not write HTML report: %v", err)
	}

	return nil
}
