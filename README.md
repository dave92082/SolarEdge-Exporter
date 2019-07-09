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

|			Metric						 | Type  |
|----------------------------------------|-------|
|TotalCurrentAmps                      	 | Guage |
|Phase                                 	 | Guage |
|Length                                	 | Guage |
|TotalCurrentAmps                      	 | Guage |
|PhaseACurrentAmps                     	 | Guage |
|PhaseBCurrentAmps                     	 | Guage |
|PhaseCCurrentAmps                     	 | Guage |
|CurrentScaleFactor                    	 | Guage |
|VoltagePhaseABVolts                   	 | Guage |
|VoltagePhaseBCVolts                   	 | Guage |
|VoltagePhaseCAVolts                   	 | Guage |
|VoltagePhaseANVolts                   	 | Guage |
|VoltagePhaseBNVolts                   	 | Guage |
|VoltagePhaseCNVolts                   	 | Guage |
|VoltageScaleFactor                    	 | Guage |
|ACPowerWatts                          	 | Guage |
|ACPowerScaleFactor                    	 | Guage |
|ACFrequencyHertz                      	 | Guage |
|ACFrequencyScaleFactor                	 | Guage |
|ACApparentPowerVA                     	 | Guage |
|ACApparentPowerScaleFactor            	 | Guage |
|ACReactivePowerVAR                    	 | Guage |
|ACReactivePowerScaleFactor            	 | Guage |
|ACPowerFactorPercent                  	 | Guage |
|ACPowerFactorScaleFactor              	 | Guage |
|ACLifetimeEnergyProductionWH          	 | Guage |
|ACLifetimeEnergyProductionScaleFactor 	 | Guage |
|DCCurrentAmps                         	 | Guage |
|DCCurrentScaleFactor                  	 | Guage |
|DCVoltage                             	 | Guage |
|DCVoltageScaleFactor                  	 | Guage |
|DCPowerWatts                          	 | Guage |
|DCPowerWattsScaleFactor               	 | Guage |
|HeatSinkTemperatureC                  	 | Guage |
|HeatSinkTemperatureScaleFactor        	 | Guage |
|Status                                	 | Guage |
|StatusVendor                          	 | Guage |

