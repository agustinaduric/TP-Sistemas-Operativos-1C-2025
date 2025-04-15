package config

type KernelConfig struct{
	PortKernel int `json:"port_kernel"`
	IpMemory string `json:"ip_memory"`
	PortMemory int `json:"port_memory"`
	SchedulerAlgorithm string `json:"scheduler_algorithm"`
	NewAlgorithm string `json:"new_algorithm"`
	Alpha int `json:"alpha"`
	SuspensionTime int `json:"suspension_time"`
	LogLevel string `json:"log_level"`
}