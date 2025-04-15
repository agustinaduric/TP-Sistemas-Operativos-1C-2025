package config

type IOConfig struct{
	PortIo int `json:"port_io"`
	IpKernel string `json:"ip_kernel"`
	PortKernel int `json:"port_kernel"`
	LogLevel string `json:"log_level"`
}