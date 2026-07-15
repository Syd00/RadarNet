package main

import (
	"fmt"
	"math/rand"
)

func main() {
	numRadars := 8
	totalJobs := 2000
	sentinelRate := 0.15
	numSentinels := int(float64(totalJobs) * sentinelRate)

	// Configurazione controllata dell'omissione (Modello BAR).
	// Specifichiamo quali indici di radar fisici simulano un comportamento pigro.
	// Ad esempio: il radar con indice 2 (il terzo) e indice 3 (il quarto) barano il 20% delle volte.
	radarOmissionRates := map[int]float64{
		2: 0.20,
		3: 0.20,
	}

	net := NewNetwork(numRadars, radarOmissionRates)
	client := NewClient(sentinelRate)

	// Generatore basato sulla distribuzione stocastica di Zipf (size=2, alpha=3.0)
	zipfGen := ZipfGenerator(2, 3.0)
	var exportData []RadarExport

	fmt.Println("==================================================================")
	fmt.Println("       AVVIO SIMULATORE RADAR CONCORRENTE (MODELLO BAR)           ")
	fmt.Println("==================================================================")
	fmt.Printf("Carico complessivo: %d Job totali | Controllo: %d Sentinelle (%.0f%%)\n",
		totalJobs, numSentinels, sentinelRate*100)
	fmt.Printf("Infrastruttura di rete: %d Radar attivi in parallelo (Worker Pool)\n\n", numRadars)

	isSentinelArray := make([]bool, totalJobs)
	for i := 0; i < totalJobs; i++ {
		isSentinelArray[i] = i < numSentinels
	}
	rand.Shuffle(len(isSentinelArray), func(i, j int) {
		isSentinelArray[i], isSentinelArray[j] = isSentinelArray[j], isSentinelArray[i]
	})

	type Job struct {
		ID           int
		RadarID      int
		Signal       Signal
		IsSentinel   bool
		ExpectedScan RadarScan
	}

	// Canali bufferizzati per consentire lo scambio asincrono ed evitare colli di bottiglia
	jobsChan := make(chan Job, totalJobs)
	resultsChan := make(chan RadarExport, totalJobs)

	// worker pool
	for w := 0; w < numRadars; w++ {
		go func(workerRadarID int) {
			for job := range jobsChan {

				// Il segnale attraversa la rete e viene computato dal radar specifico
				scan := net.RouteTask(job.RadarID, job.Signal)

				// Il Client esegue la validazione di conformità matematica solo se è una sentinella
				passed := true
				if job.IsSentinel {
					passed = client.VerifyScan(job.ExpectedScan, scan)
				}

				// Incapsulamento del risultato e invio asincrono al canale di raccolta
				resultsChan <- RadarExport{
					RadarScan:          scan,
					IsSentinel:         job.IsSentinel,
					PassedVerification: passed,
				}
			}
		}(w)
	}

	for i := 0; i < totalJobs; i++ {
		radarId := i % numRadars
		isSentinel := isSentinelArray[i]

		var currentSignal Signal
		var expectedScan RadarScan

		if isSentinel {
			// Strategia Uniforme: alterna casualmente scenari di Spazio Vuoto (1) o Ostacolo (2)
			targetTask := 1
			if rand.Float64() < 0.5 {
				targetTask = 2
			}
			currentSignal = client.GenerateSentinelSignal(targetTask)
			expectedScan = client.PrecomputeExpectedScan(radarId, currentSignal, targetTask)
		} else {
			// Job Genuino: Estratto dal comportamento stocastico dell'ambiente tramite Zipf
			rank := zipfGen.NextInt()
			currentSignal = client.GenerateSentinelSignal(rank)
		}

		// Distribuzione immediata del carico nel canale senza attendere la risposta
		jobsChan <- Job{
			ID:           i,
			RadarID:      radarId,
			Signal:       currentSignal,
			IsSentinel:   isSentinel,
			ExpectedScan: expectedScan,
		}
	}
	close(jobsChan) // Notifichiamo ai worker che la distribuzione dei Job è terminata

	// Raccolta asincrona delle scansioni
	successSentinels := 0
	failedSentinels := 0

	for i := 0; i < totalJobs; i++ {
		result := <-resultsChan
		exportData = append(exportData, result)

		if result.IsSentinel {
			if !result.PassedVerification {
				failedSentinels++
				fmt.Printf("[ALERT INTEGRITÀ] Job %d -> Il Radar %d ha fallito la verifica! Risposta non conforme\n",
					i, result.RadarID)
			} else {
				successSentinels++
			}
		}
	}
	close(resultsChan) // Chiusura formale del canale dei risultati

	// Reporting
	fmt.Println("\n==================================================================")
	fmt.Println("                     STATISTICHE DI SIMULAZIONE                   ")
	fmt.Println("==================================================================")
	fmt.Printf("Sentinelle totali elaborate:   %d\n", successSentinels+failedSentinels)
	fmt.Printf("Verifiche superate con successo: %d\n", successSentinels)
	fmt.Printf("Violazioni/Omissioni rilevate:   %d\n", failedSentinels)
	if failedSentinels > 0 {
		fmt.Printf("Efficacia del monitoraggio:     %.2f%%\n", (float64(failedSentinels)/float64(numSentinels))*100)
	}

	outputFile := "radar_data.csv"
	err := client.SaveLogsToCSV(outputFile, exportData)
	if err != nil {
		fmt.Printf("\n[ERRORE] Impossibile scrivere il file di log: %v\n", err)
	} else {
		fmt.Printf("\nSimulazione completata. Log strutturato esportato in '%s'.\n", outputFile)
	}
}
