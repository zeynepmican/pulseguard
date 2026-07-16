package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Config okunurken kritik hata: %v", err)
	}
	// Sadece 'Checker' arayüzüne uyan nesneleri kabul eden bir liste oluşturuyoruz.
	var activeCheckers []Checker

	// YAML dosyasında 'enabled: true' olanları fabrikadan çıkarıp listeye ekliyoruz.
	if config.Checks.CPU.Enabled {
		activeCheckers = append(activeCheckers, NewCPUChecker(config.Checks.CPU, config.Agent.HostID))
	}
	if config.Checks.RAM.Enabled {
		activeCheckers = append(activeCheckers, NewRAMChecker(config.Checks.RAM, config.Agent.HostID))
	}
	if config.Checks.Disk.Enabled {
		activeCheckers = append(activeCheckers, NewDiskChecker(config.Checks.Disk, config.Agent.HostID))
	}
	// no checker enabled, exit
	if len(activeCheckers) == 0 {
		log.Fatal("Hiçbir kontrol aktif değil (enabled: false). Ajan kapatılıyor.")
	}

	// İçinde maksimum 100 tane Event nesnesi barındırabilen bir boru inşa ediyoruz.
	eventQueue := make(chan *Event, 100)

	go eventShipper(eventQueue)

	// Config dosyasından gelen saniyeyi, Go'nun anlayacağı time.Second formatına çeviriyoruz.
	ticker := time.NewTicker(time.Duration(config.Agent.IntervalSeconds) * time.Second)
	defer ticker.Stop()
	// Sonsuz döngü kurduk ancak ticker bir bariyer görevi görerek, vakit gelene kadar bu döngüyü bekletiyor.
	// Böylece CPU'yu boşuna yormuyoruz.
	for {
		<-ticker.C
		runChecks(activeCheckers, eventQueue)
	}

}
func runChecks(checkers []Checker, queue chan<- *Event) {
	fmt.Printf("\n[%s] 🔄 Sistem kontrol ediliyor...\n", time.Now().Format("15:04:05"))
	for _, worker := range checkers {
		event, err := worker.Execute()

		if err != nil {
			log.Printf("❌ OKUMA HATASI: %v\n", err)
			continue
		}
		if event != nil {
			queue <- event
		}
	}
}

// 2nd goroutine: Event'leri alıp, merkez sunucuya gönderecek olan kargo görevlisi.
// Sadece kanaldan veri okuyacağı için (<-chan) şeklinde sadece-okunabilir belirtiyoruz.
func eventShipper(queue <-chan *Event) {
	// range queue: Kanal açık olduğu sürece döngü burada bekler.
	// Kanala yeni bir veri düştüğü anda uyanır, veriyi alır ve işler.
	for event := range queue {
		fmt.Printf("   📦 KARGOCU TESLİM ALDI | %s [%s]: %.2f\n", event.CheckType, event.Level, event.Value)

		// İleride HTTP Batching, Hashleme ve Retry (Yeniden deneme) mekanizmalarını
		// tam olarak bu bloğun içine inşa edeceğiz.
	}
}
