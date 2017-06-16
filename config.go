package main

type Config struct {
	Template    string
	Name        string
	Port        int
	GatewayPort int
	HealthPort  int
	MetricsPort int
	Description string
	Year        string
}
