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
package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// All metrics carry an "inverter" label (address of the inverter they were
// read from) so multiple inverters can be polled by a single exporter.
var labels = []string{"inverter"}

func newGauge(name string, help string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)
}

var (
	SunSpec_DID = newGauge("SunSpec_DID",
		"101 = single phase 102 = split phase1 103 = three phase")

	SunSpec_Length = newGauge("SunSpec_Length",
		"Registers 50 = Length of model block")

	AC_Current = newGauge("AC_Current",
		"Amps AC Total Current value")

	AC_CurrentA = newGauge("AC_CurrentA",
		"Amps AC Phase A Current value")

	AC_CurrentB = newGauge("AC_CurrentB",
		"Amps AC Phase B Current value")

	AC_CurrentC = newGauge("AC_CurrentC",
		"Amps AC Phase C Current value")

	AC_Current_SF = newGauge("AC_Current_SF",
		"AC Current scale factor")

	AC_VoltageAB = newGauge("AC_VoltageAB",
		"Volts AC Voltage Phase AB value")

	AC_VoltageBC = newGauge("AC_VoltageBC",
		"Volts AC Voltage Phase BC value")

	AC_VoltageCA = newGauge("AC_VoltageCA",
		"Volts AC Voltage Phase CA value")

	AC_VoltageAN = newGauge("AC_VoltageAN",
		"Volts AC Voltage Phase A to N value")

	AC_VoltageBN = newGauge("AC_VoltageBN",
		"Volts AC Voltage Phase B to N value")

	AC_VoltageCN = newGauge("AC_VoltageCN",
		"Volts AC Voltage Phase C to N value")

	AC_Voltage_SF = newGauge("AC_Voltage_SF",
		"AC Voltage scale factor")

	AC_Power = newGauge("AC_Power",
		"Watts AC Power value")

	AC_Power_SF = newGauge("AC_Power_SF",
		"AC Power scale factor")

	AC_Frequency = newGauge("AC_Frequency",
		"Hertz AC Frequency value")

	AC_Frequency_SF = newGauge("AC_Frequency_SF",
		"Scale factor")

	AC_VA = newGauge("AC_VA",
		"VA Apparent Power")

	AC_VA_SF = newGauge("AC_VA_SF",
		"Scale factor")

	AC_VAR = newGauge("AC_VAR",
		"VAR Reactive Power")

	AC_VAR_SF = newGauge("AC_VAR_SF",
		"Scale factor")

	AC_PF = newGauge("AC_PF",
		"% Power Factor")

	AC_PF_SF = newGauge("AC_PF_SF",
		"Scale factor")

	AC_Energy_WH = newGauge("AC_Energy_WH",
		"WattHours AC Lifetime Energy production")

	AC_Energy_WH_SF = newGauge("AC_Energy_WH_SF",
		"Scale factor")

	DC_Current = newGauge("DC_Current",
		"Amps DC Current value")

	DC_Current_SF = newGauge("DC_Current_SF",
		"Scale factor")

	DC_Voltage = newGauge("DC_Voltage",
		"Volts DC Voltage value")

	DC_Voltage_SF = newGauge("DC_Voltage_SF",
		"Scale factor")

	DC_Power = newGauge("DC_Power",
		"Watts DC Power value")

	DC_Power_SF = newGauge("DC_Power_SF",
		"Scale factor")

	Temp_Sink = newGauge("Temp_Sink",
		"Degrees C Heat Sink Temperature")

	Temp_SF = newGauge("Temp_SF",
		"Scale factor")

	Status = newGauge("Status",
		"Operating State")

	Status_Vendor = newGauge("Status_Vendor",
		"Vendor-defined operating state and error codes. For error description, meaning and troubleshooting, refer to the SolarEdge Installation Guide.")
)

// MeterMetrics holds the full set of gauges for a single meter. Meter 1 uses
// the historical "M_" metric prefix, additional meters use "M2_", "M3_", ...
type MeterMetrics struct {
	SunSpec_DID     *prometheus.GaugeVec
	SunSpec_Length  *prometheus.GaugeVec
	AC_Current      *prometheus.GaugeVec
	AC_CurrentA     *prometheus.GaugeVec
	AC_CurrentB     *prometheus.GaugeVec
	AC_CurrentC     *prometheus.GaugeVec
	AC_Current_SF   *prometheus.GaugeVec
	AC_VoltageLN    *prometheus.GaugeVec
	AC_VoltageAN    *prometheus.GaugeVec
	AC_VoltageBN    *prometheus.GaugeVec
	AC_VoltageCN    *prometheus.GaugeVec
	AC_VoltageLL    *prometheus.GaugeVec
	AC_VoltageAB    *prometheus.GaugeVec
	AC_VoltageBC    *prometheus.GaugeVec
	AC_VoltageCA    *prometheus.GaugeVec
	AC_Voltage_SF   *prometheus.GaugeVec
	AC_Frequency    *prometheus.GaugeVec
	AC_Frequency_SF *prometheus.GaugeVec
	AC_Power        *prometheus.GaugeVec
	AC_Power_A      *prometheus.GaugeVec
	AC_Power_B      *prometheus.GaugeVec
	AC_Power_C      *prometheus.GaugeVec
	AC_Power_SF     *prometheus.GaugeVec
	AC_VA           *prometheus.GaugeVec
	AC_VA_A         *prometheus.GaugeVec
	AC_VA_B         *prometheus.GaugeVec
	AC_VA_C         *prometheus.GaugeVec
	AC_VA_SF        *prometheus.GaugeVec
	AC_VAR          *prometheus.GaugeVec
	AC_VAR_A        *prometheus.GaugeVec
	AC_VAR_B        *prometheus.GaugeVec
	AC_VAR_C        *prometheus.GaugeVec
	AC_VAR_SF       *prometheus.GaugeVec
	AC_PF           *prometheus.GaugeVec
	AC_PF_A         *prometheus.GaugeVec
	AC_PF_B         *prometheus.GaugeVec
	AC_PF_C         *prometheus.GaugeVec
	AC_PF_SF        *prometheus.GaugeVec
	Exported        *prometheus.GaugeVec
	Exported_A      *prometheus.GaugeVec
	Exported_B      *prometheus.GaugeVec
	Exported_C      *prometheus.GaugeVec
	Imported        *prometheus.GaugeVec
	Imported_A      *prometheus.GaugeVec
	Imported_B      *prometheus.GaugeVec
	Imported_C      *prometheus.GaugeVec
	Energy_W_SF     *prometheus.GaugeVec
}

// NewMeterMetrics registers and returns the gauge set for one meter, using
// the given metric name prefix (e.g. "M" for meter 1, "M2" for meter 2).
func NewMeterMetrics(prefix string) *MeterMetrics {
	g := func(name string, help string) *prometheus.GaugeVec {
		return newGauge(prefix+"_"+name, help)
	}

	return &MeterMetrics{
		SunSpec_DID:     g("SunSpec_DID", "201 = single phase 202 = split phase 203 = wye connect three phase 204 = delta connect three phase"),
		SunSpec_Length:  g("SunSpec_Length", "Length of meter model block"),
		AC_Current:      g("AC_Current", "Amps AC Total Current value"),
		AC_CurrentA:     g("AC_CurrentA", "Amps AC Phase A Current value"),
		AC_CurrentB:     g("AC_CurrentB", "Amps AC Phase B Current value"),
		AC_CurrentC:     g("AC_CurrentC", "Amps AC Phase C Current value"),
		AC_Current_SF:   g("AC_Current_SF", "AC Current scale factor"),
		AC_VoltageLN:    g("AC_VoltageLN", "Volts Line to Neutral AC Voltage (average of active phases)"),
		AC_VoltageAN:    g("AC_VoltageAN", "Volts AC Voltage Phase A to N value"),
		AC_VoltageBN:    g("AC_VoltageBN", "Volts AC Voltage Phase B to N value"),
		AC_VoltageCN:    g("AC_VoltageCN", "Volts AC Voltage Phase C to N value"),
		AC_VoltageLL:    g("AC_VoltageLL", "Volts Line to Line AC Voltage (average of active phases)"),
		AC_VoltageAB:    g("AC_VoltageAB", "Volts AC Voltage Phase AB value"),
		AC_VoltageBC:    g("AC_VoltageBC", "Volts AC Voltage Phase BC value"),
		AC_VoltageCA:    g("AC_VoltageCA", "Volts AC Voltage Phase CA value"),
		AC_Voltage_SF:   g("AC_Voltage_SF", "AC Voltage scale factor"),
		AC_Frequency:    g("AC_Frequency", "Hertz AC Frequency value"),
		AC_Frequency_SF: g("AC_Frequency_SF", "Scale factor"),
		AC_Power:        g("AC_Power", "Watts AC Power value"),
		AC_Power_A:      g("AC_Power_A", "Watts AC Power value Phase A"),
		AC_Power_B:      g("AC_Power_B", "Watts AC Power value Phase B"),
		AC_Power_C:      g("AC_Power_C", "Watts AC Power value Phase C"),
		AC_Power_SF:     g("AC_Power_SF", "AC Power scale factor"),
		AC_VA:           g("AC_VA", "VA Apparent Power"),
		AC_VA_A:         g("AC_VA_A", "VA Apparent Power Phase A"),
		AC_VA_B:         g("AC_VA_B", "VA Apparent Power Phase B"),
		AC_VA_C:         g("AC_VA_C", "VA Apparent Power Phase C"),
		AC_VA_SF:        g("AC_VA_SF", "Scale factor"),
		AC_VAR:          g("AC_VAR", "VAR Reactive Power"),
		AC_VAR_A:        g("AC_VAR_A", "VAR Reactive Power Phase A"),
		AC_VAR_B:        g("AC_VAR_B", "VAR Reactive Power Phase B"),
		AC_VAR_C:        g("AC_VAR_C", "VAR Reactive Power Phase C"),
		AC_VAR_SF:       g("AC_VAR_SF", "Scale factor"),
		AC_PF:           g("AC_PF", "% Power Factor"),
		AC_PF_A:         g("AC_PF_A", "% Power Factor Phase A"),
		AC_PF_B:         g("AC_PF_B", "% Power Factor Phase B"),
		AC_PF_C:         g("AC_PF_C", "% Power Factor Phase C"),
		AC_PF_SF:        g("AC_PF_SF", "Scale factor"),
		Exported:        g("Exported", "WattHours AC Exported"),
		Exported_A:      g("Exported_A", "WattHours AC Exported Phase A"),
		Exported_B:      g("Exported_B", "WattHours AC Exported Phase B"),
		Exported_C:      g("Exported_C", "WattHours AC Exported Phase C"),
		Imported:        g("Imported", "WattHours AC Imported"),
		Imported_A:      g("Imported_A", "WattHours AC Imported Phase A"),
		Imported_B:      g("Imported_B", "WattHours AC Imported Phase B"),
		Imported_C:      g("Imported_C", "WattHours AC Imported Phase C"),
		Energy_W_SF:     g("Energy_W_SF", "Real Energy scale factor"),
	}
}
