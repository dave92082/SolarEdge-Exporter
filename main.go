package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/goburrow/modbus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)



func main() {
	// Set up configuration
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/solaredge-exporter")
	viper.AddConfigPath("$HOME/.solaredge-exporter")
	viper.ReadInConfig()

	// Open Logger
	f, err := os.OpenFile("SolarEdge-Exporter.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Could not open log file: %s", err.Error())
	}
	defer f.Close()

	multiwriter := io.MultiWriter(zerolog.ConsoleWriter{Out: os.Stdout}, f)

	// Log starting parameters
	zerolog.TimeFieldFormat = time.RFC3339


	log.Logger = log.Output(zerolog.SyncWriter(multiwriter))
	log.Debug().Msg("Starting SolarEdge-Exporter")
	log.Debug().Msgf("Configured Inverter Address: %s", viper.GetString("SolarEdge.InverterAddress"))
	log.Debug().Msgf("Configured Inverter Address: %d", viper.GetInt("SolarEdge.InverterPort"))

	go runCollection()
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func runCollection() {
	interval := viper.GetInt("Exporter.Interval")

	handler := modbus.NewTCPClientHandler(
			fmt.Sprintf("%s:%d",
			viper.GetString("SolarEdge.InverterAddress"),
			viper.GetInt("SolarEdge.InverterPort")))

	handler.Timeout = 10 * time.Second
	handler.SlaveId = 0x01


	err := handler.Connect()
	if err != nil {
		log.Error().Msgf("Error connecting to Inverter: %s", err.Error())
	}

	client := modbus.NewClient(handler)



	for {
		results, err := client.ReadHoldingRegisters(40069, 40)
		if err != nil {
			log.Error().Msgf("Error reading holding registers: %s", err.Error())
			log.Error().Msgf("Attempting to reconnect")
			_ = handler.Close()
			time.Sleep(7 * time.Second)
			_ = handler.Connect()
			continue

		}
		log.Debug().Msg("Data retrieved from inverter")
		if len(results) == 80 {
			Phase.Set(float64(BytesToUInt16(results[0:2])))
			Length.Set(float64(BytesToUInt16(results[2:4])))
			TotalCurrentAmps.Set(float64(BytesToUInt16(results[4:6])))
			PhaseACurrentAmps.Set(float64(BytesToUInt16(results[6:8])))
			PhaseBCurrentAmps.Set(float64(BytesToUInt16(results[8:10])))
			PhaseCCurrentAmps.Set(float64(BytesToUInt16(results[10:12])))
			CurrentScaleFactor.Set(float64(BytesToInt16(results[12:14])))
			VoltagePhaseABVolts.Set(float64(BytesToUInt16(results[14:16])))
			VoltagePhaseBCVolts.Set(float64(BytesToUInt16(results[16:18])))
			VoltagePhaseCAVolts.Set(float64(BytesToUInt16(results[18:20])))
			VoltagePhaseANVolts.Set(float64(BytesToUInt16(results[20:22])))
			VoltagePhaseBNVolts.Set(float64(BytesToUInt16(results[22:24])))
			VoltagePhaseCNVolts.Set(float64(BytesToUInt16(results[24:26])))
			VoltageScaleFactor.Set(float64(BytesToInt16(results[26:28])))
			ACPowerWatts.Set(float64(BytesToUInt16(results[28:30])))
			ACPowerScaleFactor.Set(float64(BytesToInt16(results[30:32])))
			ACFrequencyHertz.Set(float64(BytesToUInt16(results[32:34])))
			ACFrequencyScaleFactor.Set(float64(BytesToInt16(results[34:36])))
			ACApparentPowerVA.Set(float64(BytesToUInt16(results[36:38])))
			ACApparentPowerScaleFactor.Set(float64(BytesToInt16(results[38:40])))
			ACReactivePowerVAR.Set(float64(BytesToUInt16(results[40:42])))
			ACReactivePowerScaleFactor.Set(float64(BytesToInt16(results[42:44])))
			ACPowerFactorPercent.Set(float64(BytesToUInt16(results[44:46])))
			ACPowerFactorScaleFactor.Set(float64(BytesToInt16(results[46:48])))
			ACLifetimeEnergyProductionWH.Set(float64(BytesToInt32(results[48:52])))
			ACLifetimeEnergyProductionScaleFactor.Set(float64(BytesToInt16(results[52:54])))
			DCCurrentAmps.Set(float64(BytesToUInt16(results[54:56])))
			DCCurrentScaleFactor.Set(float64(BytesToInt16(results[56:58])))
			DCVoltage.Set(float64(BytesToUInt16(results[58:60])))
			DCVoltageScaleFactor.Set(float64(BytesToInt16(results[60:62])))
			DCPowerWatts.Set(float64(BytesToUInt16(results[62:64])))
			DCPowerWattsScaleFactor.Set(float64(BytesToInt16(results[64:66])))
			HeatSinkTemperatureC.Set(float64(BytesToUInt16(results[68:70])))
			HeatSinkTemperatureScaleFactor.Set(float64(BytesToInt16(results[74:76])))
			Status.Set(float64(BytesToUInt16(results[76:78])))
		} else {
			log.Error().Msgf("Got bad data. Length: %d", len(results))
		}



		time.Sleep(time.Duration(interval) * time.Second)

	}

}

func BytesToString(i []byte) (string) {
	hexValue := fmt.Sprintf("%x", binary.BigEndian.Uint32(i))
	value, err := hex.DecodeString(hexValue)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s", value)
}


func BytesToInt16(i []byte) int16 {
	return int16(i[1]) | (int16(i[0]) << 8)
}

func BytesToUInt16(i []byte) uint16 {
	return binary.BigEndian.Uint16(i)
}

func BytesToInt32(i []byte) uint32 {
	return binary.BigEndian.Uint32(i)
}

func CleanString(i []byte) string {
	i = bytes.Trim(i, "\x00")
	return strings.TrimSpace(string(i))
}


var (
	TotalCurrentAps = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACTotalCurrent",
		Help: "AC Total Current value in Amps",
	})
	Phase = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Phase",
		Help: "Phase",
	})
	Length = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Length",
		Help: "Length",
	})
	TotalCurrentAmps = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "TotalCurrentAmps",
		Help: "TotalCurrentAmps",
	})
	PhaseACurrentAmps = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "PhaseACurrentAmps",
		Help: "PhaseACurrentAmps",
	})
	PhaseBCurrentAmps = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "PhaseBCurrentAmps",
		Help: "PhaseBCurrentAmps",
	})
	PhaseCCurrentAmps = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "PhaseCCurrentAmps",
		Help: "PhaseCCurrentAmps",
	})
	CurrentScaleFactor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "CurrentScaleFactor",
		Help: "CurrentScaleFactor",
	})
	VoltagePhaseABVolts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "VoltagePhaseABVolts",
		Help: "VoltagePhaseABVolts",
	})
	VoltagePhaseBCVolts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "VoltagePhaseBCVolts",
		Help: "VoltagePhaseBCVolts",
	})
	VoltagePhaseCAVolts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "VoltagePhaseCAVolts",
		Help: "VoltagePhaseCAVolts",
	})
	VoltagePhaseANVolts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "VoltagePhaseANVolts",
		Help: "VoltagePhaseANVolts",
	})
	VoltagePhaseBNVolts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "VoltagePhaseBNVolts",
		Help: "VoltagePhaseBNVolts",
	})
	VoltagePhaseCNVolts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "VoltagePhaseCNVolts",
		Help: "VoltagePhaseCNVolts",
	})
	VoltageScaleFactor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "VoltageScaleFactor",
		Help: "VoltageScaleFactor",
	})
	ACPowerWatts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACPowerWatts",
		Help: "ACPowerWatts",
	})
	ACPowerScaleFactor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACPowerScaleFactor",
		Help: "ACPowerScaleFactor",
	})
	ACFrequencyHertz = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACFrequencyHertz",
		Help: "ACFrequencyHertz",
	})
	ACFrequencyScaleFactor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACFrequencyScaleFactor",
		Help: "ACFrequencyScaleFactor",
	})
	ACApparentPowerVA = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACApparentPowerVA",
		Help: "ACApparentPowerVA",
	})
	ACApparentPowerScaleFactor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACApparentPowerScaleFactor",
		Help: "ACApparentPowerScaleFactor",
	})
	ACReactivePowerVAR = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACReactivePowerVAR",
		Help: "ACReactivePowerVAR",
	})
	ACReactivePowerScaleFactor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACReactivePowerScaleFactor",
		Help: "ACReactivePowerScaleFactor",
	})
	ACPowerFactorPercent = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACPowerFactorPercent",
		Help: "ACPowerFactorPercent",
	})
	ACPowerFactorScaleFactor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACPowerFactorScaleFactor",
		Help: "ACPowerFactorScaleFactor",
	})
	ACLifetimeEnergyProductionWH = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACLifetimeEnergyProductionWH",
		Help: "ACLifetimeEnergyProductionWH",
	})
	ACLifetimeEnergyProductionScaleFactor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ACLifetimeEnergyProductionScaleFactor",
		Help: "ACLifetimeEnergyProductionScaleFactor",
	})
	DCCurrentAmps = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "DCCurrentAmps",
		Help: "DCCurrentAmps",
	})
	DCCurrentScaleFactor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "DCCurrentScaleFactor",
		Help: "DCCurrentScaleFactor",
	})
	DCVoltage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "DCVoltage",
		Help: "DCVoltage",
	})
	DCVoltageScaleFactor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "DCVoltageScaleFactor",
		Help: "DCVoltageScaleFactor",
	})
	DCPowerWatts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "DCPowerWatts",
		Help: "DCPowerWatts",
	})
	DCPowerWattsScaleFactor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "DCPowerWattsScaleFactor",
		Help: "DCPowerWattsScaleFactor",
	})
	HeatSinkTemperatureC = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "HeatSinkTemperatureC",
		Help: "HeatSinkTemperatureC",
	})
	HeatSinkTemperatureScaleFactor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "HeatSinkTemperatureScaleFactor",
		Help: "HeatSinkTemperatureScaleFactor",
	})
	Status = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Status",
		Help: "Status",
	})
	StatusVendor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "StatusVendor",
		Help: "StatusVendor",
	})

)





