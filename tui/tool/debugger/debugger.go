package debugger

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

func New() (file *os.File, err error) {
	homedir := "./"
	filename := "debug.log"
	log.SetLevel(log.DebugLevel)

	_, ok := os.LookupEnv("LOG_DEV")
	if !ok {
		filename = ".viscue.error.log"
		log.SetLevel(log.ErrorLevel)
		homedir, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch home directory")
		}
	}

	path := filepath.Join(homedir, filename)
	file, err = tea.LogToFileWith(path, "", log.Default())
	if err != nil {
		return nil, err
	}
	return file, nil
}
