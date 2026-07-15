package main

import (
	"encoding/csv"
	"math"
	"math/rand"
	"os"
	"strconv"
)

// Estende la scansione originale aggiungendo i metadati di sicurezza del Client
type RadarExport struct {
	RadarScan
	IsSentinel         bool
	PassedVerification bool
}

// Rappresenta il client che delega e valida i job
type Client struct {
	SentinelRate float64
}

type RadarStats struct {
	TotalJobs   int
	OmittedJobs int // comportamento pigro rilevato
	Sentinels   int
	Failed      int
}

// Inizializza il client impostando la percentuale di sentinelle desiderata
func NewClient(rate float64) *Client {
	return &Client{SentinelRate: rate}
}

// Crea il Ground Truth di una sentinella
func (c *Client) GenerateSentinelSignal(targetTask int) Signal {
	var sig Signal
	if targetTask == 2 {
		// Task 2: Genera un oggetto reale all'interno del campo visivo del radar
		r := rand.Float64()
		sig.Range = 0.5 + r*(RadarMaxRange-0.5)
		sig.Theta = -60.0 + r*(60.0-(-60.0))
	} else {
		// Task 1: Genera uno scenario di spazio vuoto coerente
		sig.Range = 0.0
		sig.Theta = 0.0
	}
	return sig
}

// Calcola deterministicamente la risposta esatta che il worker onesto DOVREBBE inviare
func (c *Client) PrecomputeExpectedScan(radarID int, sig Signal, targetTask int) RadarScan {
	var truthScan RadarScan
	truthScan.RadarID = radarID
	truthScan.Range = sig.Range
	truthScan.Theta = sig.Theta

	if targetTask == 2 {
		rad := sig.Theta * (math.Pi / 180.0)
		truthScan.X = math.Round(sig.Range*math.Sin(rad)*100) / 100
		truthScan.Y = math.Round(sig.Range*math.Cos(rad)*100) / 100
		truthScan.Rcs = 5.0 // Valore nominale teorico atteso
		truthScan.Snr = 1000.0 / math.Pow(sig.Range, 2)
	} else {
		truthScan.X, truthScan.Y = 0.0, 0.0
		truthScan.Rcs = 0.02
		truthScan.Snr = 5.0
	}
	truthScan.TaskID = targetTask
	return truthScan
}

// Confronta l'output inviato dal Worker con la computazione del Client
func (c *Client) VerifyScan(truth, scan RadarScan) bool {
	return truth.TaskID == scan.TaskID
}

// SaveLogsToCSV esporta i dati arricchiti con le metriche di fingerprinting
func (c *Client) SaveLogsToCSV(filename string, data []RadarExport) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("sep=,\n")
	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"RadarID", "Timestamp", "Range", "Theta", "X", "Y",
		"Rcs", "Snr", "TaskID", "IsSentinel", "PassedVerification", "ExecutionTimeMs",
	}
	writer.Write(header)

	for _, r := range data {
		// Logica di formattazione corretta per i dati nulli
		passedStr := "N/A"
		if r.IsSentinel {
			passedStr = strconv.FormatBool(r.PassedVerification)
		}

		row := []string{
			strconv.Itoa(r.RadarID),
			strconv.FormatInt(r.Timestamp, 10),
			strconv.FormatFloat(r.Range, 'f', 4, 64),
			strconv.FormatFloat(r.Theta, 'f', 4, 64),
			strconv.FormatFloat(r.X, 'f', 4, 64),
			strconv.FormatFloat(r.Y, 'f', 4, 64),
			strconv.FormatFloat(r.Rcs, 'f', 4, 64),
			strconv.FormatFloat(r.Snr, 'f', 4, 64),
			strconv.Itoa(r.TaskID),
			strconv.FormatBool(r.IsSentinel),
			passedStr, // Ora stamperà "true", "false" o "N/A"
		}
		writer.Write(row)
	}
	return nil
}
