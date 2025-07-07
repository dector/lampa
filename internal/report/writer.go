package report

import (
	"fmt"
	"os"

	"github.com/dector/lampa/internal/utils"
)

func (self Report) WriteToFile(file string) error {
	err := utils.EnsureParentDirExists(file)
	if err != nil {
		return err
	}

	reportJson, err := self.ToJsonBytes()
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
