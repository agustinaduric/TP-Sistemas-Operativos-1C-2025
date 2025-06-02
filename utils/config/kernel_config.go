package config

type KernelConfig struct {
	IpMemory              string  `json:"ip_memory"`
	PortMemory            int     `json:"port_memory"`
	SchedulerAlgorithm    string  `json:"scheduler_algorithm"`
	ReadyIngressAlgorithm string  `json:"ready_ingress_algorithm"`
	Alpha                 int     `json:"alpha"`
	InitialEstimate       float64 `json:"initial_estimate"`
	SuspensionTime        int     `json:"suspension_time"`
	LogLevel              string  `json:"log_level"`
	PortKernel            int     `json:"port_kernel"`
	IpKernel              string  `json:"ip_kernel"`
	Mensaje               string  `json:"mensaje"`
}
