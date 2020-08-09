/*

MIT License

Copyright (c) 2019 David Suarez

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
	"fmt"
	"github.com/goburrow/modbus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	// Set up configuration
	config.InitConfig()

	// Open Logger
	f, err := os.OpenFile("SolarEdge-Exporter.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Could not open log file: %s", err.Error())
		return
	}
	defer f.Close()
	m := io.MultiWriter(zerolog.ConsoleWriter{Out: os.Stdout}, f)
	log.Logger = log.Output(zerolog.SyncWriter(m))
	log.Info().Msg("Starting SolarEdge-Exporter")
	log.Info().Msgf("Configured Inverter Address: %s", viper.GetString("SolarEdge.InverterAddress"))
	log.Info().Msgf("Configured Inverter Port: %d", viper.GetInt("SolarEdge.InverterPort"))
	log.Info().Msgf("Configured Number of Meters: %d", viper.GetInt("SolarEdge.NumMeters"))
	log.Info().Msgf("Configured Listen Address: %s", viper.GetString("Exporter.ListenAddress"))
	log.Info().Msgf("Configured Listen Port: %d", viper.GetInt("Exporter.ListenPort"))

	// Start Data Collection
	// TODO: Add a cancellation context on SIGINT to cleanly close the connection
	go runCollection()

	// Start Prometheus Handler
	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(viper.GetString("Exporter.ListenAddress")+":"+strconv.Itoa(viper.GetInt("Exporter.ListenPort")), nil)
	if err != nil {
		log.Error().Msgf("Could not start the prometheus metric server: %s", err.Error())
	}
}

func runCollection() {
	// Get Interval from Config
	interval := viper.GetInt("Exporter.Interval")

	// Configure Modbus Connection and Handler/Client
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
	defer handler.Close()

	// Collect and log common inverter data
	infoData, err := client.ReadHoldingRegisters(40000, 70)
	cm, err := solaredge.NewCommonModel(infoData)
	log.Info().Msgf("Inverter Model: %s", cm.C_Model)
	log.Info().Msgf("Inverter Serial: %s", cm.C_SerialNumber)
	log.Info().Msgf("Inverter Version: %s", cm.C_Version)

	nummeters := viper.GetInt("SolarEdge.NumMeters")
    	// Meter 1 - common data
	if nummeters > 0 {
		infoData2, err := client.ReadHoldingRegisters(40121, 65)
		if err != nil {
			log.Error().Msgf("Error reading holding registers: %s", err.Error())
		}
		cm2, err := solaredge.NewCommonMeter(infoData2)
		log.Info().Msgf("Meter 1 Manufacturer: %s", cm2.C_Manufacturer)
		log.Info().Msgf("Meter 1 Model: %s", cm2.C_Model)
		log.Info().Msgf("Meter 1 Serial: %s", cm2.C_SerialNumber)
		log.Info().Msgf("Meter 1 Version: %s", cm2.C_Version)
		log.Info().Msgf("Meter 1 Option: %s", cm2.C_Option)
	}
	// Meter 2 - common data
	if nummeters > 1 {
		infoData3, err := client.ReadHoldingRegisters(40295, 65)
		if err != nil {
			log.Error().Msgf("Error reading holding registers: %s", err.Error())
		}		
		cm3, err := solaredge.NewCommonMeter(infoData3)
		log.Info().Msgf("Meter 2 Manufacturer: %s", cm3.C_Manufacturer)
		log.Info().Msgf("Meter 2 Model: %s", cm3.C_Model)
		log.Info().Msgf("Meter 2 Serial: %s", cm3.C_SerialNumber)
		log.Info().Msgf("Meter 2 Version: %s", cm3.C_Version)
		log.Info().Msgf("Meter 2 Option: %s", cm3.C_Option)
	}
	// Collect logs forever
	for {
		inverterData, err := client.ReadHoldingRegisters(40069, 40)
		if err != nil {
			log.Error().Msgf("Error reading holding registers: %s", err.Error())
			log.Error().Msgf("Attempting to reconnect")
			_ = handler.Close()
			time.Sleep(7 * time.Second)
			_ = handler.Connect()
			continue
		}
		id, err := solaredge.NewInverterModel(inverterData)
		if err != nil {
			log.Error().Msgf("Error parsing data: %s", err.Error())
		}

        	// Meter 1 - meter data
		if nummeters > 0 {		
			infoData4, err := client.ReadHoldingRegisters(40188, 105)
			if err != nil {
				log.Error().Msgf("Error reading holding registers: %s", err.Error())
			}			
			mt1, err := solaredge.NewMeterModel(infoData4)
			log.Info().Msgf("Meter 1 AC Current: %f", float64(mt1.M_AC_Current)*math.Pow(10, float64(mt1.M_AC_Current_SF)))
			log.Info().Msgf("Meter 1 VoltageLN: %f", float64(mt1.M_AC_VoltageLN)*math.Pow(10, float64(mt1.M_AC_Voltage_SF)))
			log.Info().Msgf("Meter 1 PF: %d", mt1.M_AC_PF)
			log.Info().Msgf("Meter 1 Freq: %f", float64(mt1.M_AC_Frequency)*math.Pow(10, float64(mt1.M_AC_Frequency_SF)))
			log.Info().Msgf("Meter 1 AC Power: %f", float64(mt1.M_AC_Power)*math.Pow(10.0, float64(mt1.M_AC_Power_SF)))
			log.Info().Msgf("Meter 1 M_AC_VA: %f", float64(mt1.M_AC_VA)*math.Pow(10.0, float64(mt1.M_AC_VA_SF)))
			log.Info().Msgf("Meter 1 M_Exported: %f", float64(mt1.M_Exported)*math.Pow(10.0, float64(mt1.M_Energy_W_SF)))
			log.Info().Msgf("Meter 1 M_Imported: %f", float64(mt1.M_Imported)*math.Pow(10.0, float64(mt1.M_Energy_W_SF)))
			setMetricsForMeter1(mt1)
		}
	        // Meter 2 - meter data
		if nummeters > 1 {
			infoData5, err := client.ReadHoldingRegisters(40362, 105)
			if err != nil {
				log.Error().Msgf("Error reading holding registers: %s", err.Error())
			}
			mt2, err := solaredge.NewMeterModel(infoData5)
			log.Info().Msgf("Meter 2 AC Current: %f", float64(mt2.M_AC_Current)*math.Pow(10, float64(mt2.M_AC_Current_SF)))
			log.Info().Msgf("Meter 2 VoltageLN: %f", float64(mt2.M_AC_VoltageLN)*math.Pow(10, float64(mt2.M_AC_Voltage_SF)))
			log.Info().Msgf("Meter 2 PF: %d", mt2.M_AC_PF)
			log.Info().Msgf("Meter 2 Freq: %f", float64(mt2.M_AC_Frequency)*math.Pow(10, float64(mt2.M_AC_Frequency_SF)))
			log.Info().Msgf("Meter 2 AC Power: %f", float64(mt2.M_AC_Power)*math.Pow(10.0, float64(mt2.M_AC_Power_SF)))
			log.Info().Msgf("Meter 2 M_AC_VA: %f", float64(mt2.M_AC_VA)*math.Pow(10.0, float64(mt2.M_AC_VA_SF)))
			log.Info().Msgf("Meter 2 M_Exported: %f", float64(mt2.M_Exported)*math.Pow(10.0, float64(mt2.M_Energy_W_SF)))
			log.Info().Msgf("Meter 2 M_Imported: %f", float64(mt2.M_Imported)*math.Pow(10.0, float64(mt2.M_Energy_W_SF)))
			setMetricsForMeter2(mt2)
		}
		log.Debug().Msg("-------------------------------------------")
		log.Debug().Msg("Data retrieved from inverter")
		setMetrics(id)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func setMetrics(i solaredge.InverterModel) {
	exporter.SunSpec_DID.Set(float64(i.SunSpec_DID))
	exporter.SunSpec_Length.Set(float64(i.SunSpec_Length))
	// *math.Pow(10, float64(mt1.M_AC_Frequency_SF))
	exporter.AC_Current.Set(float64(i.AC_Current)*math.Pow(10, float64(i.AC_Current_SF)))
	exporter.AC_CurrentA.Set(float64(i.AC_CurrentA)*math.Pow(10, float64(i.AC_Current_SF)))
	exporter.AC_CurrentB.Set(float64(i.AC_CurrentB)*math.Pow(10, float64(i.AC_Current_SF)))
	exporter.AC_CurrentC.Set(float64(i.AC_CurrentC)*math.Pow(10, float64(i.AC_Current_SF)))
	exporter.AC_Current_SF.Set(float64(0))
	exporter.AC_VoltageAB.Set(float64(i.AC_VoltageAB)*math.Pow(10, float64(i.AC_Voltage_SF)))
	exporter.AC_VoltageBC.Set(float64(i.AC_VoltageBC)*math.Pow(10, float64(i.AC_Voltage_SF)))
	exporter.AC_VoltageCA.Set(float64(i.AC_VoltageCA)*math.Pow(10, float64(i.AC_Voltage_SF)))
	exporter.AC_VoltageAN.Set(float64(i.AC_VoltageAN)*math.Pow(10, float64(i.AC_Voltage_SF)))
	exporter.AC_VoltageBN.Set(float64(i.AC_VoltageBN)*math.Pow(10, float64(i.AC_Voltage_SF)))
	exporter.AC_VoltageCN.Set(float64(i.AC_VoltageCN)*math.Pow(10, float64(i.AC_Voltage_SF)))
	exporter.AC_Voltage_SF.Set(float64(0))
	exporter.AC_Power.Set(float64(i.AC_Power)*math.Pow(10, float64(i.AC_Power_SF)))
	exporter.AC_Power_SF.Set(float64(0))
	exporter.AC_Frequency.Set(float64(i.AC_Frequency)*math.Pow(10, float64(i.AC_Frequency_SF)))
	exporter.AC_Frequency_SF.Set(float64(0))
	exporter.AC_VA.Set(float64(i.AC_VA)*math.Pow(10, float64(i.AC_VA_SF)))
	exporter.AC_VA_SF.Set(float64(0))
	exporter.AC_VAR.Set(float64(i.AC_VAR)*math.Pow(10, float64(i.AC_VAR_SF)))
	exporter.AC_VAR_SF.Set(float64(0))
	exporter.AC_PF.Set(float64(i.AC_PF)*math.Pow(10, float64(i.AC_PF_SF)))
	exporter.AC_PF_SF.Set(float64(0))
	exporter.AC_Energy_WH.Set(float64(i.AC_Energy_WH)*math.Pow(10, float64(i.AC_Energy_WH_SF)))
	exporter.AC_Energy_WH_SF.Set(float64(0))
	exporter.DC_Current.Set(float64(i.DC_Current)*math.Pow(10, float64(i.DC_Current_SF)))
	exporter.DC_Current_SF.Set(float64(0))
	exporter.DC_Voltage.Set(float64(i.DC_Voltage)*math.Pow(10, float64(i.DC_Voltage_SF)))
	exporter.DC_Voltage_SF.Set(float64(0))
	exporter.DC_Power.Set(float64(i.DC_Power)*math.Pow(10, float64(i.DC_Power_SF)))
	exporter.DC_Power_SF.Set(float64(0))
	exporter.Temp_Sink.Set(float64(i.Temp_Sink)*math.Pow(10, float64(i.Temp_SF)))
	exporter.Temp_SF.Set(float64(0))
	exporter.Status.Set(float64(i.Status))
	exporter.Status_Vendor.Set(float64(i.Status_Vendor))
}

func setMetricsForMeter1(m solaredge.MeterModel) {
	exporter.M_SunSpec_DID.Set(float64(m.SunSpec_DID))
	exporter.M_SunSpec_Length.Set(float64(m.SunSpec_Length))
	exporter.M_AC_Current.Set(float64(m.M_AC_Current)*math.Pow(10, float64(m.M_AC_Current_SF)))
	exporter.M_AC_CurrentA.Set(float64(m.M_AC_CurrentA)*math.Pow(10, float64(m.M_AC_Current_SF)))
	exporter.M_AC_CurrentB.Set(float64(m.M_AC_CurrentB)*math.Pow(10, float64(m.M_AC_Current_SF)))
	exporter.M_AC_CurrentC.Set(float64(m.M_AC_CurrentC)*math.Pow(10, float64(m.M_AC_Current_SF)))
	exporter.M_AC_Current_SF.Set(float64(0))
	exporter.M_AC_VoltageLN.Set(float64(m.M_AC_VoltageLN)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M_AC_VoltageAN.Set(float64(m.M_AC_VoltageAN)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M_AC_VoltageBN.Set(float64(m.M_AC_VoltageBN)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M_AC_VoltageCN.Set(float64(m.M_AC_VoltageCN)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M_AC_VoltageLL.Set(float64(m.M_AC_VoltageLL)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M_AC_VoltageAB.Set(float64(m.M_AC_VoltageAB)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M_AC_VoltageBC.Set(float64(m.M_AC_VoltageBC)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M_AC_VoltageCA.Set(float64(m.M_AC_VoltageCA)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M_AC_Voltage_SF.Set(float64(0))
	exporter.M_AC_Frequency.Set(float64(m.M_AC_Frequency)*math.Pow(10, float64(m.M_AC_Frequency_SF)))
	exporter.M_AC_Frequency_SF.Set(float64(0))
	exporter.M_AC_Power.Set(float64(m.M_AC_Power)*math.Pow(10, float64(m.M_AC_Power_SF)))
	exporter.M_AC_Power_A.Set(float64(m.M_AC_Power_A)*math.Pow(10, float64(m.M_AC_Power_SF)))
	exporter.M_AC_Power_B.Set(float64(m.M_AC_Power_B)*math.Pow(10, float64(m.M_AC_Power_SF)))
	exporter.M_AC_Power_C.Set(float64(m.M_AC_Power_C)*math.Pow(10, float64(m.M_AC_Power_SF)))
	exporter.M_AC_Power_SF.Set(float64(0))
	exporter.M_AC_VA.Set(float64(m.M_AC_VA)*math.Pow(10, float64(m.M_AC_VA_SF)))
	exporter.M_AC_VA_A.Set(float64(m.M_AC_VA_A)*math.Pow(10, float64(m.M_AC_VA_SF)))
	exporter.M_AC_VA_B.Set(float64(m.M_AC_VA_B)*math.Pow(10, float64(m.M_AC_VA_SF)))
	exporter.M_AC_VA_C.Set(float64(m.M_AC_VA_C)*math.Pow(10, float64(m.M_AC_VA_SF)))
	exporter.M_AC_VA_SF.Set(float64(0))
	exporter.M_AC_VAR.Set(float64(m.M_AC_VAR)*math.Pow(10, float64(m.M_AC_VAR_SF)))
	exporter.M_AC_VAR_A.Set(float64(m.M_AC_VAR_A)*math.Pow(10, float64(m.M_AC_VAR_SF)))
	exporter.M_AC_VAR_B.Set(float64(m.M_AC_VAR_B)*math.Pow(10, float64(m.M_AC_VAR_SF)))
	exporter.M_AC_VAR_C.Set(float64(m.M_AC_VAR_C)*math.Pow(10, float64(m.M_AC_VAR_SF)))
	exporter.M_AC_VAR_SF.Set(float64(0))
	exporter.M_AC_PF.Set(float64(m.M_AC_PF)*math.Pow(10, float64(m.M_AC_PF_SF)))
	exporter.M_AC_PF_A.Set(float64(m.M_AC_PF_A)*math.Pow(10, float64(m.M_AC_PF_SF)))
	exporter.M_AC_PF_B.Set(float64(m.M_AC_PF_B)*math.Pow(10, float64(m.M_AC_PF_SF)))
	exporter.M_AC_PF_C.Set(float64(m.M_AC_PF_C)*math.Pow(10, float64(m.M_AC_PF_SF)))
	exporter.M_AC_PF_SF.Set(float64(0))
	exporter.M_Exported.Set(float64(m.M_Exported)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M_Exported_A.Set(float64(m.M_Exported_A)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M_Exported_B.Set(float64(m.M_Exported_B)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M_Exported_C.Set(float64(m.M_Exported_C)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M_Imported.Set(float64(m.M_Imported)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M_Imported_A.Set(float64(m.M_Imported_A)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M_Imported_B.Set(float64(m.M_Imported_B)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M_Imported_C.Set(float64(m.M_Imported_C)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M_Energy_W_SF.Set(float64(0))
}

func setMetricsForMeter2(m solaredge.MeterModel) {
	exporter.M2_SunSpec_DID.Set(float64(m.SunSpec_DID))
	exporter.M2_SunSpec_Length.Set(float64(m.SunSpec_Length))
	exporter.M2_AC_Current.Set(float64(m.M_AC_Current)*math.Pow(10, float64(m.M_AC_Current_SF)))
	exporter.M2_AC_CurrentA.Set(float64(m.M_AC_CurrentA)*math.Pow(10, float64(m.M_AC_Current_SF)))
	exporter.M2_AC_CurrentB.Set(float64(m.M_AC_CurrentB)*math.Pow(10, float64(m.M_AC_Current_SF)))
	exporter.M2_AC_CurrentC.Set(float64(m.M_AC_CurrentC)*math.Pow(10, float64(m.M_AC_Current_SF)))
	exporter.M2_AC_Current_SF.Set(float64(0))
	exporter.M2_AC_VoltageLN.Set(float64(m.M_AC_VoltageLN)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M2_AC_VoltageAN.Set(float64(m.M_AC_VoltageAN)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M2_AC_VoltageBN.Set(float64(m.M_AC_VoltageBN)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M2_AC_VoltageCN.Set(float64(m.M_AC_VoltageCN)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M2_AC_VoltageLL.Set(float64(m.M_AC_VoltageLL)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M2_AC_VoltageAB.Set(float64(m.M_AC_VoltageAB)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M2_AC_VoltageBC.Set(float64(m.M_AC_VoltageBC)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M2_AC_VoltageCA.Set(float64(m.M_AC_VoltageCA)*math.Pow(10, float64(m.M_AC_Voltage_SF)))
	exporter.M2_AC_Voltage_SF.Set(float64(0))
	exporter.M2_AC_Frequency.Set(float64(m.M_AC_Frequency)*math.Pow(10, float64(m.M_AC_Frequency_SF)))
	exporter.M2_AC_Frequency_SF.Set(float64(0))
	exporter.M2_AC_Power.Set(float64(m.M_AC_Power)*math.Pow(10, float64(m.M_AC_Power_SF)))
	exporter.M2_AC_Power_A.Set(float64(m.M_AC_Power_A)*math.Pow(10, float64(m.M_AC_Power_SF)))
	exporter.M2_AC_Power_B.Set(float64(m.M_AC_Power_B)*math.Pow(10, float64(m.M_AC_Power_SF)))
	exporter.M2_AC_Power_C.Set(float64(m.M_AC_Power_C)*math.Pow(10, float64(m.M_AC_Power_SF)))
	exporter.M2_AC_Power_SF.Set(float64(0))
	exporter.M2_AC_VA.Set(float64(m.M_AC_VA)*math.Pow(10, float64(m.M_AC_VA_SF)))
	exporter.M2_AC_VA_A.Set(float64(m.M_AC_VA_A)*math.Pow(10, float64(m.M_AC_VA_SF)))
	exporter.M2_AC_VA_B.Set(float64(m.M_AC_VA_B)*math.Pow(10, float64(m.M_AC_VA_SF)))
	exporter.M2_AC_VA_C.Set(float64(m.M_AC_VA_C)*math.Pow(10, float64(m.M_AC_VA_SF)))
	exporter.M2_AC_VA_SF.Set(float64(0))
	exporter.M2_AC_VAR.Set(float64(m.M_AC_VAR)*math.Pow(10, float64(m.M_AC_VAR_SF)))
	exporter.M2_AC_VAR_A.Set(float64(m.M_AC_VAR_A)*math.Pow(10, float64(m.M_AC_VAR_SF)))
	exporter.M2_AC_VAR_B.Set(float64(m.M_AC_VAR_B)*math.Pow(10, float64(m.M_AC_VAR_SF)))
	exporter.M2_AC_VAR_C.Set(float64(m.M_AC_VAR_C)*math.Pow(10, float64(m.M_AC_VAR_SF)))
	exporter.M2_AC_VAR_SF.Set(float64(0))
	exporter.M2_AC_PF.Set(float64(m.M_AC_PF)*math.Pow(10, float64(m.M_AC_PF_SF)))
	exporter.M2_AC_PF_A.Set(float64(m.M_AC_PF_A)*math.Pow(10, float64(m.M_AC_PF_SF)))
	exporter.M2_AC_PF_B.Set(float64(m.M_AC_PF_B)*math.Pow(10, float64(m.M_AC_PF_SF)))
	exporter.M2_AC_PF_C.Set(float64(m.M_AC_PF_C)*math.Pow(10, float64(m.M_AC_PF_SF)))
	exporter.M2_AC_PF_SF.Set(float64(0))
	exporter.M2_Exported.Set(float64(m.M_Exported)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M2_Exported_A.Set(float64(m.M_Exported_A)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M2_Exported_B.Set(float64(m.M_Exported_B)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M2_Exported_C.Set(float64(m.M_Exported_C)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M2_Imported.Set(float64(m.M_Imported)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M2_Imported_A.Set(float64(m.M_Imported_A)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M2_Imported_B.Set(float64(m.M_Imported_B)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M2_Imported_C.Set(float64(m.M_Imported_C)*math.Pow(10, float64(m.M_Energy_W_SF)))
	exporter.M2_Energy_W_SF.Set(float64(0))
}
