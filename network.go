package main

type Network struct {
	Nodes map[int]Radar
}

func NewNetwork(numNodes int, defaultOmissionRates map[int]float64) Network {
	nodes := make(map[int]Radar)

	for i := 0; i < numNodes; i++ {
		rate := 0.0
		if val, exists := defaultOmissionRates[i]; exists {
			rate = val
		}
		nodes[i] = newRadar(i+1, rate)
	}

	return Network{
		Nodes: nodes,
	}
}

// Inoltra il segnale al radar hardware selezionato e ne avvia la computazione
func (n *Network) RouteTask(radarID int, signal Signal) RadarScan {
	targetRadar := n.Nodes[radarID]
	// Invia il lavoro al worker
	return GenerateRadarScan(targetRadar, signal)
}
