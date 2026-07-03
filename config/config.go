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
package config

import "github.com/spf13/viper"

// InitConfig initializes the viper configuration
func InitConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/solaredge-exporter")
	viper.AddConfigPath("$HOME/.solaredge-exporter")
	viper.SetDefault("SolarEdge.InverterAddress", "")
	viper.SetDefault("SolarEdge.InverterPort", 0)
	viper.SetDefault("Exporter.Interval", 5)
	viper.SetDefault("Exporter.ListenAddress", "")
	viper.SetDefault("Exporter.ListenPort", 2112)
	viper.BindEnv("SolarEdge.InverterAddress", "INVERTER_ADDRESS")
	viper.BindEnv("SolarEdge.InverterPort", "INVERTER_PORT")
	viper.BindEnv("Exporter.Interval", "EXPORTER_INTERVAL")
	viper.BindEnv("Exporter.ListenAddress", "EXPORTER_ADDRESS")
	viper.BindEnv("Exporter.ListenPort", "EXPORTER_PORT")
	viper.BindEnv("Log.Debug", "DEBUG_LOGGING")
	viper.AutomaticEnv()
	viper.ReadInConfig()
}
