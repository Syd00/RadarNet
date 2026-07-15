package main

import (
	"math"
	"math/rand"
	"time"
)

const (
	RadarMaxRange  = 50.0
	PMax           = 0.89
	EpsilonGeo     = 1.0  // Tolleranza geometrica
	EpsilonRcs     = 0.5  // Tolleranza Radar Cross Section
	EmptyRcsValue  = 0.02 // RCS attesa per lo spazio vuoto
	EmptySnrValue  = 5.0  // SNR di fondo
	TargetRcsValue = 5.0  // RCS nominale di un target solido
	FraudRcsMax    = 0.09 // Valore massimo RCS quando il nodo bara
	FraudSnrMax    = 9.99 // Valore massimo SNR quando il nodo bara
)

// Rappresenta la struttura dati del segnale
type Signal struct {
	Range float64
	Theta float64
}

// Rappresenta la struttura dati del radar
type Radar struct {
	RadarID  int
	MaxRange float64
	POmitted float64
}

// Rappresenta l'output del radar
type RadarScan struct {
	RadarID   int
	Timestamp int64
	Range     float64
	Theta     float64
	X         float64
	Y         float64
	Rcs       float64
	Snr       float64
	TaskID    int
}

// Inizializza il radar
func newRadar(id int, POmitted float64) Radar {
	return Radar{
		RadarID:  id,
		MaxRange: RadarMaxRange,
		POmitted: POmitted,
	}
}

// Aggiunge rumore
func addNoise(val float64, intensity float64) float64 {
	return val + (rand.NormFloat64() * intensity)
}

// Simula il comportamento del Worker (può computare onestamente o barare)
func GenerateRadarScan(hardware Radar, signal Signal) RadarScan {
	var scan RadarScan
	scan.RadarID = hardware.RadarID
	scan.Timestamp = time.Now().Unix()

	// Comportamento opportunistico
	if rand.Float64() <= hardware.POmitted {
		// Il nodo pigro non elabora il segnale reale, ma per non farsi scoprire
		// simula il rumore gaussiano di un Task 1 (Spazio Vuoto).
		scan.Range = 0.0
		scan.Theta = 0.0
		scan.X, scan.Y = 0.0, 0.0

		// Inserimento di rumore realistico
		scan.Rcs = math.Abs(addNoise(EmptyRcsValue, 0.01))
		scan.Snr = addNoise(EmptySnrValue, 0.2)

		scan.TaskID = 1
		return scan
	}

	// Comportamento onesto
	scan.Range = signal.Range
	scan.Theta = signal.Theta

	// Conversione da coordinate polari a cartesiane
	rad := signal.Theta * (math.Pi / 180.0)
	scan.X = math.Round(signal.Range*math.Sin(rad)*100) / 100
	scan.Y = math.Round(signal.Range*math.Cos(rad)*100) / 100

	if signal.Range == 0.0 && signal.Theta == 0.0 {
		// Task 1: Spazio Vuoto Reale
		scan.Rcs = math.Abs(addNoise(EmptyRcsValue, 0.01))
		scan.Snr = addNoise(EmptySnrValue, 0.2)
	} else {
		// Task 2: Ostacolo Presente
		scan.Rcs = addNoise(TargetRcsValue, 0.05)
		scan.Snr = addNoise(1000.0/math.Pow(signal.Range, 2), 0.5)
	}

	// Classificazione logica basata sui dati processati
	scan.TaskID = classifyTask(scan)

	return scan
}

// Analizza una scansione ed estrae la classifica
func classifyTask(r RadarScan) int {
	const Epsilon = 0.5

	if math.Abs(r.Range-0.0) < Epsilon &&
		math.Abs(r.Theta-0.0) < Epsilon &&
		math.Abs(r.X-0.0) < Epsilon &&
		math.Abs(r.Y-0.0) < Epsilon &&
		r.Rcs < 0.1 &&
		r.Snr < 10.0 {
		return 1 // Riconosciuto come Task 1 (Spazio Vuoto)
	}
	return 2 // Riconosciuto come Task 2 (Target Rilevato)
}
