/*
MIT License

# Copyright (c) 2019 David Suarez

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package main

import (
	"SolarEdge-Exporter/config"
	"SolarEdge-Exporter/exporter"
	"SolarEdge-Exporter/solaredge"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/goburrow/modbus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Register layout from the SolarEdge SunSpec implementation technical note:
// each meter occupies a 174-register block, the first starting with its
// common block at 40121 and its data block at 40188.
const (
	meterCommonBase  = 40121
	meterDataBase    = 40188
	meterBlockLength = 174
)

func main() {
	// Set up configuration
	config.InitConfig()

	// Open Logger
	logPath := viper.GetString("Log.Path")
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Could not open log file %s: %s", logPath, err.Error())
		return
	}
	defer f.Close()
	m := io.MultiWriter(zerolog.ConsoleWriter{Out: os.Stdout}, f)
	log.Logger = log.Output(zerolog.SyncWriter(m))
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if viper.GetBool("Log.Debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Info().Msg("Starting SolarEdge-Exporter")
	log.Info().Msgf("Configured Inverter Address: %s", viper.GetString("SolarEdge.InverterAddress"))
	log.Info().Msgf("Configured Inverter Port: %d", viper.GetInt("SolarEdge.InverterPort"))
	log.Info().Msgf("Configured Number of Meters: %d", viper.GetInt("SolarEdge.NumMeters"))
	log.Info().Msgf("Configured Listen Address: %s", viper.GetString("Exporter.ListenAddress"))
	log.Info().Msgf("Configured Listen Port: %d", viper.GetInt("Exporter.ListenPort"))
	log.Info().Msgf("Configured Client ID: %x", byte(viper.GetInt("SolarEdge.ClientId")))
	log.Info().Msgf("Configured Log Path: %s", logPath)

	// Register one metric set per configured meter. Meter 1 keeps the
	// historical M_* metric names, meter 2 is exported as M2_*, and so on.
	numMeters := viper.GetInt("SolarEdge.NumMeters")
	meterMetrics := make([]*exporter.MeterMetrics, numMeters)
	for i := range meterMetrics {
		prefix := "M"
		if i > 0 {
			prefix = fmt.Sprintf("M%d", i+1)
		}
		meterMetrics[i] = exporter.NewMeterMetrics(prefix)
	}

	// Start Data Collection for each configured inverter. The context is
	// cancelled on SIGINT/SIGTERM so the collectors close their connections
	// cleanly before the process exits.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var collectors sync.WaitGroup
	for _, target := range inverterTargets() {
		collectors.Add(1)
		go func() {
			defer collectors.Done()
			runCollection(ctx, target, meterMetrics)
		}()
	}

	// Start Prometheus Handler, shutting it down once a stop signal arrives
	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{Addr: viper.GetString("Exporter.ListenAddress") + ":" + strconv.Itoa(viper.GetInt("Exporter.ListenPort"))}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error().Msgf("Could not start the prometheus metric server: %s", err.Error())
	}
	collectors.Wait()
	log.Info().Msg("SolarEdge-Exporter stopped")
}

// inverterTargets returns the list of inverter "host:port" targets from the
// configuration. InverterAddress accepts a comma separated list of addresses,
// each optionally with its own ":port" (defaulting to InverterPort).
func inverterTargets() []string {
	defaultPort := viper.GetInt("SolarEdge.InverterPort")
	var targets []string
	for address := range strings.SplitSeq(viper.GetString("SolarEdge.InverterAddress"), ",") {
		address = strings.TrimSpace(address)
		if address == "" {
			continue
		}
		if !strings.Contains(address, ":") {
			address = fmt.Sprintf("%s:%d", address, defaultPort)
		}
		targets = append(targets, address)
	}
	return targets
}

func runCollection(ctx context.Context, target string, meterMetrics []*exporter.MeterMetrics) {
	// Get Interval from Config
	interval := time.Duration(viper.GetInt("Exporter.Interval")) * time.Second

	handler, client := connectInverter(target)
	defer handler.Close()

	activeMeters := 0
	metersDetected := false

	// Collect logs until the context is cancelled
	for {
		inverterData, err := client.ReadHoldingRegisters(40069, 40)
		if err != nil {
			log.Error().Msgf("[%s] Error reading holding registers: %s", target, err.Error())
			log.Error().Msgf("[%s] Attempting to reconnect", target)
			_ = handler.Close()
			if !sleepOrDone(ctx, 7*time.Second) {
				return
			}
			_ = handler.Connect()
			continue
		}
		id, err := solaredge.NewInverterModel(inverterData)
		if err != nil {
			log.Error().Msgf("[%s] Error parsing data: %s", target, err.Error())
			continue
		}
		setMetrics(id, target)

		if !metersDetected {
			activeMeters, metersDetected = detectMeters(client, target, len(meterMetrics))
		}

		collectMeters(client, target, activeMeters, meterMetrics)

		log.Debug().Msg("-------------------------------------------")
		log.Debug().Msgf("[%s] Data retrieved from inverter", target)
		if !sleepOrDone(ctx, interval) {
			return
		}
	}
}

// connectInverter opens the modbus connection to the inverter and logs its
// common block details. Connection errors are logged, not fatal: the read
// loop re-establishes the connection on its next cycle.
func connectInverter(target string) (*modbus.TCPClientHandler, modbus.Client) {
	handler := modbus.NewTCPClientHandler(target)
	handler.Timeout = 10 * time.Second
	handler.SlaveId = byte(viper.GetInt("SolarEdge.ClientId"))
	err := handler.Connect()
	if err != nil {
		log.Error().Msgf("[%s] Error connecting to Inverter: %s", target, err.Error())
	}
	client := modbus.NewClient(handler)

	infoData, err := client.ReadHoldingRegisters(40000, 70)
	if err != nil {
		log.Error().Msgf("[%s] Error reading inverter common block: %s", target, err.Error())
	}
	cm, err := solaredge.NewCommonModel(infoData)
	if err != nil {
		log.Error().Msgf("[%s] Error parsing inverter common block: %s", target, err.Error())
	}
	log.Info().Msgf("[%s] Inverter Model: %s", target, cm.C_Model)
	log.Info().Msgf("[%s] Inverter Serial: %s", target, cm.C_SerialNumber)
	log.Info().Msgf("[%s] Inverter Version: %s", target, cm.C_Version)
	return handler, client
}

// collectMeters reads and publishes the data block of each active meter.
func collectMeters(client modbus.Client, target string, activeMeters int, meterMetrics []*exporter.MeterMetrics) {
	for i := range activeMeters {
		meterData, err := client.ReadHoldingRegisters(uint16(meterDataBase+i*meterBlockLength), 105)
		if err != nil {
			log.Error().Msgf("[%s] Error reading meter %d registers: %s", target, i+1, err.Error())
			return
		}
		mt, err := solaredge.NewMeterModel(meterData)
		if err != nil {
			log.Error().Msgf("[%s] Error parsing meter %d data: %s", target, i+1, err.Error())
			return
		}
		log.Debug().Msgf("[%s] Meter %d AC Current: %f", target, i+1, float64(mt.M_AC_Current)*math.Pow(10, float64(mt.M_AC_Current_SF)))
		log.Debug().Msgf("[%s] Meter %d VoltageLN: %f", target, i+1, float64(mt.M_AC_VoltageLN)*math.Pow(10, float64(mt.M_AC_Voltage_SF)))
		log.Debug().Msgf("[%s] Meter %d PF: %d", target, i+1, mt.M_AC_PF)
		log.Debug().Msgf("[%s] Meter %d Freq: %f", target, i+1, float64(mt.M_AC_Frequency)*math.Pow(10, float64(mt.M_AC_Frequency_SF)))
		log.Debug().Msgf("[%s] Meter %d AC Power: %f", target, i+1, float64(mt.M_AC_Power)*math.Pow(10.0, float64(mt.M_AC_Power_SF)))
		log.Debug().Msgf("[%s] Meter %d M_AC_VA: %f", target, i+1, float64(mt.M_AC_VA)*math.Pow(10.0, float64(mt.M_AC_VA_SF)))
		log.Debug().Msgf("[%s] Meter %d M_Exported: %f", target, i+1, float64(mt.M_Exported)*math.Pow(10.0, float64(mt.M_Energy_W_SF)))
		log.Debug().Msgf("[%s] Meter %d M_Imported: %f", target, i+1, float64(mt.M_Imported)*math.Pow(10.0, float64(mt.M_Energy_W_SF)))
		setMetricsForMeter(meterMetrics[i], mt, target)
	}
}

// sleepOrDone waits for d and reports whether the wait completed; it returns
// false when ctx is cancelled first.
func sleepOrDone(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// detectMeters probes which of the configured meters are actually present.
// Meters are only readable when enabled, and enabled meters are always
// exposed sequentially starting at the first meter block, so probing stops at
// the first empty block. A meter that is readable but reports no manufacturer
// (e.g. an SE5000H with no meter installed) is treated as absent, keeping its
// metrics off the exporter instead of publishing zeros. Detection is only
// considered complete when no communication error occurred, so a transient
// failure is retried on the next collection cycle.
func detectMeters(client modbus.Client, target string, numMeters int) (int, bool) {
	activeMeters := 0
	for i := range numMeters {
		meterInfoData, err := client.ReadHoldingRegisters(uint16(meterCommonBase+i*meterBlockLength), 65)
		if err != nil {
			if _, isModbusError := err.(*modbus.ModbusError); isModbusError {
				// The inverter answered with an exception: the meter block
				// is not readable, i.e. the meter is not enabled.
				log.Info().Msgf("[%s] Meter %d not enabled (%s), skipping meter collection for it", target, i+1, err.Error())
				break
			}
			log.Warn().Msgf("[%s] Error reading meter %d common block: %s", target, i+1, err.Error())
			return activeMeters, false
		}
		cm, err := solaredge.NewCommonMeter(meterInfoData)
		if err != nil || len(cm.C_Manufacturer) == 0 {
			log.Info().Msgf("[%s] Meter %d not detected, skipping meter collection for it", target, i+1)
			break
		}
		log.Info().Msgf("[%s] Meter %d Manufacturer: %s", target, i+1, cm.C_Manufacturer)
		log.Info().Msgf("[%s] Meter %d Model: %s", target, i+1, cm.C_Model)
		log.Info().Msgf("[%s] Meter %d Serial: %s", target, i+1, cm.C_SerialNumber)
		log.Info().Msgf("[%s] Meter %d Version: %s", target, i+1, cm.C_Version)
		log.Info().Msgf("[%s] Meter %d Option: %s", target, i+1, cm.C_Option)
		activeMeters = i + 1
	}
	return activeMeters, true
}

func setMetrics(i solaredge.InverterModel, target string) {
	exporter.SunSpec_DID.WithLabelValues(target).Set(float64(i.SunSpec_DID))
	exporter.SunSpec_Length.WithLabelValues(target).Set(float64(i.SunSpec_Length))
	exporter.AC_Current.WithLabelValues(target).Set(float64(i.AC_Current))
	exporter.AC_CurrentA.WithLabelValues(target).Set(float64(i.AC_CurrentA))
	exporter.AC_CurrentB.WithLabelValues(target).Set(float64(i.AC_CurrentB))
	exporter.AC_CurrentC.WithLabelValues(target).Set(float64(i.AC_CurrentC))
	exporter.AC_Current_SF.WithLabelValues(target).Set(float64(i.AC_Current_SF))
	exporter.AC_VoltageAB.WithLabelValues(target).Set(float64(i.AC_VoltageAB))
	exporter.AC_VoltageBC.WithLabelValues(target).Set(float64(i.AC_VoltageBC))
	exporter.AC_VoltageCA.WithLabelValues(target).Set(float64(i.AC_VoltageCA))
	exporter.AC_VoltageAN.WithLabelValues(target).Set(float64(i.AC_VoltageAN))
	exporter.AC_VoltageBN.WithLabelValues(target).Set(float64(i.AC_VoltageBN))
	exporter.AC_VoltageCN.WithLabelValues(target).Set(float64(i.AC_VoltageCN))
	exporter.AC_Voltage_SF.WithLabelValues(target).Set(float64(i.AC_Voltage_SF))
	exporter.AC_Power.WithLabelValues(target).Set(float64(i.AC_Power))
	exporter.AC_Power_SF.WithLabelValues(target).Set(float64(i.AC_Power_SF))
	exporter.AC_Frequency.WithLabelValues(target).Set(float64(i.AC_Frequency))
	exporter.AC_Frequency_SF.WithLabelValues(target).Set(float64(i.AC_Frequency_SF))
	exporter.AC_VA.WithLabelValues(target).Set(float64(i.AC_VA))
	exporter.AC_VA_SF.WithLabelValues(target).Set(float64(i.AC_VA_SF))
	exporter.AC_VAR.WithLabelValues(target).Set(float64(i.AC_VAR))
	exporter.AC_VAR_SF.WithLabelValues(target).Set(float64(i.AC_VAR_SF))
	exporter.AC_PF.WithLabelValues(target).Set(float64(i.AC_PF))
	exporter.AC_PF_SF.WithLabelValues(target).Set(float64(i.AC_PF_SF))
	exporter.AC_Energy_WH.WithLabelValues(target).Set(float64(i.AC_Energy_WH))
	exporter.AC_Energy_WH_SF.WithLabelValues(target).Set(float64(i.AC_Energy_WH_SF))
	exporter.DC_Current.WithLabelValues(target).Set(float64(i.DC_Current))
	exporter.DC_Current_SF.WithLabelValues(target).Set(float64(i.DC_Current_SF))
	exporter.DC_Voltage.WithLabelValues(target).Set(float64(i.DC_Voltage))
	exporter.DC_Voltage_SF.WithLabelValues(target).Set(float64(i.DC_Voltage_SF))
	exporter.DC_Power.WithLabelValues(target).Set(float64(i.DC_Power))
	exporter.DC_Power_SF.WithLabelValues(target).Set(float64(i.DC_Power_SF))
	exporter.Temp_Sink.WithLabelValues(target).Set(float64(i.Temp_Sink))
	exporter.Temp_SF.WithLabelValues(target).Set(float64(i.Temp_SF))
	exporter.Status.WithLabelValues(target).Set(float64(i.Status))
	exporter.Status_Vendor.WithLabelValues(target).Set(float64(i.Status_Vendor))
}

func setMetricsForMeter(mm *exporter.MeterMetrics, m solaredge.MeterModel, target string) {
	mm.SunSpec_DID.WithLabelValues(target).Set(float64(m.SunSpec_DID))
	mm.SunSpec_Length.WithLabelValues(target).Set(float64(m.SunSpec_Length))
	mm.AC_Current.WithLabelValues(target).Set(float64(m.M_AC_Current))
	mm.AC_CurrentA.WithLabelValues(target).Set(float64(m.M_AC_CurrentA))
	mm.AC_CurrentB.WithLabelValues(target).Set(float64(m.M_AC_CurrentB))
	mm.AC_CurrentC.WithLabelValues(target).Set(float64(m.M_AC_CurrentC))
	mm.AC_Current_SF.WithLabelValues(target).Set(float64(m.M_AC_Current_SF))
	mm.AC_VoltageLN.WithLabelValues(target).Set(float64(m.M_AC_VoltageLN))
	mm.AC_VoltageAN.WithLabelValues(target).Set(float64(m.M_AC_VoltageAN))
	mm.AC_VoltageBN.WithLabelValues(target).Set(float64(m.M_AC_VoltageBN))
	mm.AC_VoltageCN.WithLabelValues(target).Set(float64(m.M_AC_VoltageCN))
	mm.AC_VoltageLL.WithLabelValues(target).Set(float64(m.M_AC_VoltageLL))
	mm.AC_VoltageAB.WithLabelValues(target).Set(float64(m.M_AC_VoltageAB))
	mm.AC_VoltageBC.WithLabelValues(target).Set(float64(m.M_AC_VoltageBC))
	mm.AC_VoltageCA.WithLabelValues(target).Set(float64(m.M_AC_VoltageCA))
	mm.AC_Voltage_SF.WithLabelValues(target).Set(float64(m.M_AC_Voltage_SF))
	mm.AC_Frequency.WithLabelValues(target).Set(float64(m.M_AC_Frequency))
	mm.AC_Frequency_SF.WithLabelValues(target).Set(float64(m.M_AC_Frequency_SF))
	mm.AC_Power.WithLabelValues(target).Set(float64(m.M_AC_Power))
	mm.AC_Power_A.WithLabelValues(target).Set(float64(m.M_AC_Power_A))
	mm.AC_Power_B.WithLabelValues(target).Set(float64(m.M_AC_Power_B))
	mm.AC_Power_C.WithLabelValues(target).Set(float64(m.M_AC_Power_C))
	mm.AC_Power_SF.WithLabelValues(target).Set(float64(m.M_AC_Power_SF))
	mm.AC_VA.WithLabelValues(target).Set(float64(m.M_AC_VA))
	mm.AC_VA_A.WithLabelValues(target).Set(float64(m.M_AC_VA_A))
	mm.AC_VA_B.WithLabelValues(target).Set(float64(m.M_AC_VA_B))
	mm.AC_VA_C.WithLabelValues(target).Set(float64(m.M_AC_VA_C))
	mm.AC_VA_SF.WithLabelValues(target).Set(float64(m.M_AC_VA_SF))
	mm.AC_VAR.WithLabelValues(target).Set(float64(m.M_AC_VAR))
	mm.AC_VAR_A.WithLabelValues(target).Set(float64(m.M_AC_VAR_A))
	mm.AC_VAR_B.WithLabelValues(target).Set(float64(m.M_AC_VAR_B))
	mm.AC_VAR_C.WithLabelValues(target).Set(float64(m.M_AC_VAR_C))
	mm.AC_VAR_SF.WithLabelValues(target).Set(float64(m.M_AC_VAR_SF))
	mm.AC_PF.WithLabelValues(target).Set(float64(m.M_AC_PF))
	mm.AC_PF_A.WithLabelValues(target).Set(float64(m.M_AC_PF_A))
	mm.AC_PF_B.WithLabelValues(target).Set(float64(m.M_AC_PF_B))
	mm.AC_PF_C.WithLabelValues(target).Set(float64(m.M_AC_PF_C))
	mm.AC_PF_SF.WithLabelValues(target).Set(float64(m.M_AC_PF_SF))
	mm.Exported.WithLabelValues(target).Set(float64(m.M_Exported))
	mm.Exported_A.WithLabelValues(target).Set(float64(m.M_Exported_A))
	mm.Exported_B.WithLabelValues(target).Set(float64(m.M_Exported_B))
	mm.Exported_C.WithLabelValues(target).Set(float64(m.M_Exported_C))
	mm.Imported.WithLabelValues(target).Set(float64(m.M_Imported))
	mm.Imported_A.WithLabelValues(target).Set(float64(m.M_Imported_A))
	mm.Imported_B.WithLabelValues(target).Set(float64(m.M_Imported_B))
	mm.Imported_C.WithLabelValues(target).Set(float64(m.M_Imported_C))
	mm.Energy_W_SF.WithLabelValues(target).Set(float64(m.M_Energy_W_SF))
}
