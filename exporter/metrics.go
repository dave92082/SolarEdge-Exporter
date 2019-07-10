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
package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SunSpec_DID = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "SunSpec_DID",
		Help: "101 = single phase 102 = split phase1 103 = three phase",
	})

	SunSpec_Length = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "SunSpec_Length",
		Help: "Registers 50 = Length of model block",
	})

	AC_Current = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_Current",
		Help: "Amps AC Total Current value",
	})

	AC_CurrentA = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_CurrentA",
		Help: "Amps AC Phase A Current value",
	})

	AC_CurrentB = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_CurrentB",
		Help: "Amps AC Phase B Current value",
	})

	AC_CurrentC = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_CurrentC",
		Help: "Amps AC Phase C Current value",
	})

	AC_Current_SF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_Current_SF",
		Help: "AC Current scale factor",
	})

	AC_VoltageAB = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_VoltageAB",
		Help: "Volts AC Voltage Phase AB value",
	})

	AC_VoltageBC = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_VoltageBC",
		Help: "Volts AC Voltage Phase BC value",
	})

	AC_VoltageCA = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_VoltageCA",
		Help: "Volts AC Voltage Phase CA value",
	})

	AC_VoltageAN = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_VoltageAN",
		Help: "Volts AC Voltage Phase A to N value",
	})

	AC_VoltageBN = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_VoltageBN",
		Help: "Volts AC Voltage Phase B to N value",
	})

	AC_VoltageCN = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_VoltageCN",
		Help: "Volts AC Voltage Phase C to N value",
	})

	AC_Voltage_SF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_Voltage_SF",
		Help: "AC Voltage scale factor",
	})

	AC_Power = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_Power",
		Help: "Watts AC Power value",
	})

	AC_Power_SF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_Power_SF",
		Help: "AC Power scale factor",
	})

	AC_Frequency = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_Frequency",
		Help: "Hertz AC Frequency value",
	})

	AC_Frequency_SF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_Frequency_SF",
		Help: "Scale factor",
	})

	AC_VA = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_VA",
		Help: "VA Apparent Power",
	})

	AC_VA_SF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_VA_SF",
		Help: "Scale factor",
	})

	AC_VAR = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_VAR",
		Help: "VAR Reactive Power",
	})

	AC_VAR_SF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_VAR_SF",
		Help: "Scale factor",
	})

	AC_PF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_PF",
		Help: "% Power Factor",
	})

	AC_PF_SF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_PF_SF",
		Help: "Scale factor",
	})

	AC_Energy_WH = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_Energy_WH",
		Help: "WattHours AC Lifetime Energy production",
	})

	AC_Energy_WH_SF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "AC_Energy_WH_SF",
		Help: "Scale factor",
	})

	DC_Current = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "DC_Current",
		Help: "Amps DC Current value",
	})

	DC_Current_SF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "DC_Current_SF",
		Help: "Scale factor",
	})

	DC_Voltage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "DC_Voltage",
		Help: "Volts DC Voltage value",
	})

	DC_Voltage_SF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "DC_Voltage_SF",
		Help: "Scale factor",
	})

	DC_Power = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "DC_Power",
		Help: "Watts DC Power value",
	})

	DC_Power_SF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "DC_Power_SF",
		Help: "Scale factor",
	})

	Temp_Sink = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Temp_Sink",
		Help: "Degrees C Heat Sink Temperature",
	})

	Temp_SF = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Temp_SF",
		Help: "Scale factor",
	})

	Status = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Status",
		Help: "Operating State",
	})

	Status_Vendor = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Status_Vendor",
		Help: "Vendor-defined operating state and error codes. For error description, meaning and troubleshooting, refer to the SolarEdge Installation Guide.",
	})

)