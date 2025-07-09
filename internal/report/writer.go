package report

import (
	"fmt"
	"os"

	"github.com/dector/lampa/internal/utils"
)

func (self StatsReport) WriteToFile(file string) error {
	return exportToFile(file, func() ([]byte, error) {
		return self.ToJsonBytes()
	})
}

func (self DiffReport) WriteToFile(file string) error {
	return exportToFile(file, func() ([]byte, error) {
		return self.ToJsonBytes()
	})
}

func exportToFile(file string, toJson func() ([]byte, error)) error {
	err := utils.EnsureParentDirExists(file)
	if err != nil {
		return err
	}

	reportJson, err := toJson()
	if err != nil {
		return fmt.Errorf("could not marshal report: %v", err)
	}

	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("could not create report file: %v", err)
	}
	defer f.Close()

	if _, err := f.Write(reportJson); err != nil {
		return fmt.Errorf("could not write report: %v", err)
	}

	return nil
}
