package main

import (
	"math"
	"math/rand"
	"time"
)

type Zipf struct {
	size       int
	alpha      float64
	normFactor float64
	rnd        *rand.Rand
}

// NewZipfGenerator inizializza il generatore e pre-calcola il fattore di normalizzazione
func ZipfGenerator(size int, alpha float64) *Zipf {
	normFactor := 0.0
	for i := 1; i <= size; i++ {
		normFactor += 1.0 / math.Pow(float64(i), alpha)
	}

	return &Zipf{
		size:       size,
		alpha:      alpha,
		normFactor: normFactor,
		rnd:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NextInt restituisce un numero estratto secondo la distribuzione di Zipf
func (z *Zipf) NextInt() int {
	for {
		// Estrai un rango casuale tra 1 e size
		rank := z.rnd.Intn(z.size) + 1

		// Calcola la probabilità (frequenza) per quel rango
		frequency := (1.0 / math.Pow(float64(rank), z.alpha)) / z.normFactor

		// Accetta il valore se un numero casuale è < probabilità
		if z.rnd.Float64() < frequency {
			return rank
		}
	}
}
