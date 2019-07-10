# SolarEdge Prometheus Exporter

Having just installed a SolarEdge inverter and not happy with the 15 minute delay and low resolution of the monitoring data
provided by the monitoring service/api, I created this exporter to connects directly to SolarEdge inverter over ModBus TCP 
to export (near) real time data to Prometheus.

## Status
The code could use some clean up but I have had it running for a weeks scraping data from the inverter every 5 seconds without any issues.

## Requirements
* SolarEdge Inverter that supports SunSpec protocol (Tested with SE5000 w. CPU version 3.2221.0)
* ModBus TCP Enabled on the inverter
* Local network connection to the inverter (No ZigBee/GSM support)

Modbus TCP is a local network connection only and *does not* interfere or replace your connection to the SolarEdge monitoring 
service. As per the SolarEdge documentation, the two monitoring methods can be used in parallel without impacting each other.

More information on how to enable ModBus TCP can be found in the SolarEdge Documentation [here](https://www.solaredge.com/sites/default/files/sunspec-implementation-technical-note.pdf)

## TODO
* Implement consumption meter output.
	* This may already be working however my consumption meter is not installed yet so I cannot test

## Quick Start

1. Download the binary from the Releases section for your platform
2. Configure the exporter using *one* of the two methods available.
	
	*Replace the IP address in these samples with the address of your inverter*
	* Environment Variables:
	``` 
		INVERTER_ADDRESS=192.168.1.189
		EXPORTER_INTERVAL=5
		INVERTER_PORT=502
	``` 
	* config.yaml:
	Create a config file named `config.yaml` in the same location that you downloaded the executable with the following contents:
	```yaml
	SolarEdge:
	  InverterAddress: "192.168.1.189"
	  InverterPort: 502
	Exporter:
	  # Update Interval in seconds
	  Interval: 5	
	```
3. Add the target to your prometheus server with port `2112`

## Metrics

|		Metric	 	 |	 Type	 |	Description/Help																																	 |
|--------------------|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
|SunSpec_DID     	 | 	 Guage 	 | 	 101 = single phase 102 = split phase1 103 = three phase                                                                                        	 |
|SunSpec_Length  	 | 	 Guage 	 | 	 Registers 50 = Length of model block                                                                                                           	 |
|AC_Current      	 | 	 Guage 	 | 	 Amps AC Total Current value                                                                                                                    	 |
|AC_CurrentA     	 | 	 Guage 	 | 	 Amps AC Phase A Current value                                                                                                                  	 |
|AC_CurrentB     	 | 	 Guage 	 | 	 Amps AC Phase B Current value                                                                                                                  	 |
|AC_CurrentC     	 | 	 Guage 	 | 	 Amps AC Phase C Current value                                                                                                                  	 |
|AC_Current_SF   	 | 	 Guage 	 | 	 AC Current scale factor                                                                                                                        	 |
|AC_VoltageAB    	 | 	 Guage 	 | 	 Volts AC Voltage Phase AB value                                                                                                                	 |
|AC_VoltageBC    	 | 	 Guage 	 | 	 Volts AC Voltage Phase BC value                                                                                                                	 |
|AC_VoltageCA    	 | 	 Guage 	 | 	 Volts AC Voltage Phase CA value                                                                                                                	 |
|AC_VoltageAN    	 | 	 Guage 	 | 	 Volts AC Voltage Phase A to N value                                                                                                            	 |
|AC_VoltageBN    	 | 	 Guage 	 | 	 Volts AC Voltage Phase B to N value                                                                                                            	 |
|AC_VoltageCN    	 | 	 Guage 	 | 	 Volts AC Voltage Phase C to N value                                                                                                            	 |
|AC_Voltage_SF   	 | 	 Guage 	 | 	 AC Voltage scale factor                                                                                                                        	 |
|AC_Power        	 | 	 Guage 	 | 	 Watts AC Power value                                                                                                                           	 |
|AC_Power_SF     	 | 	 Guage 	 | 	 AC Power scale factor                                                                                                                          	 |
|AC_Frequency    	 | 	 Guage 	 | 	 Hertz AC Frequency value                                                                                                                       	 |
|AC_Frequency_SF 	 | 	 Guage 	 | 	 Scale factor                                                                                                                                   	 |
|AC_VA           	 | 	 Guage 	 | 	 VA Apparent Power                                                                                                                              	 |
|AC_VA_SF        	 | 	 Guage 	 | 	 Scale factor                                                                                                                                   	 |
|AC_VAR          	 | 	 Guage 	 | 	 VAR Reactive Power                                                                                                                             	 |
|AC_VAR_SF       	 | 	 Guage 	 | 	 Scale factor                                                                                                                                   	 |
|AC_PF           	 | 	 Guage 	 | 	 % Power Factor                                                                                                                                 	 |
|AC_PF_SF        	 | 	 Guage 	 | 	 Scale factor                                                                                                                                   	 |
|AC_Energy_WH    	 | 	 Guage 	 | 	 WattHours AC Lifetime Energy production                                                                                                        	 |
|AC_Energy_WH_SF 	 | 	 Guage 	 | 	 Scale factor                                                                                                                                   	 |
|DC_Current      	 | 	 Guage 	 | 	 Amps DC Current value                                                                                                                          	 |
|DC_Current_SF   	 | 	 Guage 	 | 	 Scale factor                                                                                                                                   	 |
|DC_Voltage      	 | 	 Guage 	 | 	 Volts DC Voltage value                                                                                                                         	 |
|DC_Voltage_SF   	 | 	 Guage 	 | 	 Scale factor                                                                                                                                   	 |
|DC_Power        	 | 	 Guage 	 | 	 Watts DC Power value                                                                                                                           	 |
|DC_Power_SF     	 | 	 Guage 	 | 	 Scale factor                                                                                                                                   	 |
|Temp_Sink       	 | 	 Guage 	 | 	 Degrees C Heat Sink Temperature                                                                                                                	 |
|Temp_SF         	 | 	 Guage 	 | 	 Scale factor                                                                                                                                   	 |
|Status          	 | 	 Guage 	 | 	 Operating State                                                                                                                                	 |
|Status_Vendor   	 | 	 Guage 	 | 	 Vendor-defined operating state and error codes. For error description, meaning and troubleshooting, refer to the SolarEdge Installation Guide. 	 |

