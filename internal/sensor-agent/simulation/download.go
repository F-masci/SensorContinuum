package simulation

import (
	"SensorContinuum/configs/simulation"
	"SensorContinuum/internal/sensor-agent/environment"
	"SensorContinuum/pkg/logger"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// downloadRandomCSV scarica un file CSV random della data attuale meno 2 giorni e lo salva nella cartella "csv"
func downloadRandomCSV() (string, error) {
	date := time.Now().AddDate(0, 0, -2).Format("2006-01-02")

	files, err := os.ReadDir(simulation.CSV_DIR)
	if err == nil {
		for _, f := range files {
			if filepath.Ext(f.Name()) == ".csv" {
				if f.Name() == date+".csv" {
					filePath := filepath.Join(simulation.CSV_DIR, f.Name())
					logger.Log.Info("CSV for current date already downloaded: ", f.Name())
					return filePath, nil
				} else {
					err := os.Remove(filepath.Join(simulation.CSV_DIR, f.Name()))
					if err != nil {
						logger.Log.Error("Error removing old CSV: ", err)
						return "", err
					}
					logger.Log.Info("Removed old CSV: ", f.Name())
				}
			}
		}
	}

	pageURL := fmt.Sprintf("https://archive.sensor.community/%s/", date)
	logger.Log.Info("Downloading page: ", pageURL)
	resp, err := http.Get(pageURL)
	if err != nil {
		logger.Log.Error("Error downloading page: ", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Log.Error("Page download failed: ", resp.Status)
		return "", fmt.Errorf("page download failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("Error reading page: ", err)
		return "", err
	}

	pattern := ""
	if environment.SensorLocation == "macrozone" {
		pattern = fmt.Sprintf(`href="([^"]*(%s.*indoor|indoor.*%s)[^"]*\.csv)"`, environment.SimulationSensorReference, environment.SimulationSensorReference)
	} else {
		pattern = fmt.Sprintf(`href="([^"]*%s[^"]*\.csv)"`, environment.SimulationSensorReference)
	}
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(string(body), -1)
	if len(matches) == 0 {
		logger.Log.Error("No CSV file found on page")
		return "", fmt.Errorf("no CSV file found on page")
	}

	selected := matches[rand.Intn(len(matches))][1]
	csvURL := fmt.Sprintf("https://archive.sensor.community/%s/%s", date, selected)
	filePath := filepath.Join(simulation.CSV_DIR, date+".csv")

	logger.Log.Info("Downloading CSV file: ", csvURL)

	// Crea la cartella se non esiste
	if _, err := os.Stat(simulation.CSV_DIR); os.IsNotExist(err) {
		logger.Log.Info("Creating folder: ", simulation.CSV_DIR)
		if err := os.MkdirAll(simulation.CSV_DIR, os.ModePerm); err != nil {
			logger.Log.Error("Error creating folder: ", err)
			return "", err
		}
		logger.Log.Info("Created folder: ", simulation.CSV_DIR)
	}

	csvResp, err := http.Get(csvURL)
	if err != nil {
		logger.Log.Error("Error downloading CSV: ", err)
		return "", err
	}
	defer csvResp.Body.Close()

	if csvResp.StatusCode != http.StatusOK {
		logger.Log.Error("CSV download failed: ", csvResp.Status)
		return "", fmt.Errorf("CSV download failed: %s", csvResp.Status)
	}

	out, err := os.Create(filePath)
	if err != nil {
		logger.Log.Error("Error creating file: ", err)
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, csvResp.Body)
	if err != nil {
		logger.Log.Error("Error writing file: ", err)
		return "", err
	}

	logger.Log.Info("Download completed: ", filePath)
	return filePath, nil
}
