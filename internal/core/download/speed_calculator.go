package download

import (
	"sync"
	"time"
)

// SpeedCalculator gère le calcul de la vitesse de téléchargement
type SpeedCalculator struct {
	lastTime     time.Time
	lastBytes    int64
	currentSpeed int64
	mutex        sync.Mutex
}

// NewSpeedCalculator crée une nouvelle instance du calculateur de vitesse
func NewSpeedCalculator() *SpeedCalculator {
	return &SpeedCalculator{
		lastTime:  time.Now(),
		lastBytes: 0,
	}
}

// Update met à jour la vitesse de téléchargement et retourne la vitesse actuelle en octets/seconde
func (sc *SpeedCalculator) Update(totalBytes int64) int64 {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	now := time.Now()
	duration := now.Sub(sc.lastTime).Seconds()

	// Éviter une division par zéro ou des mises à jour trop fréquentes
	if duration < 0.5 {
		return sc.currentSpeed
	}

	bytesIncrement := totalBytes - sc.lastBytes
	speed := int64(float64(bytesIncrement) / duration)

	sc.currentSpeed = speed
	sc.lastTime = now
	sc.lastBytes = totalBytes

	return speed
}
