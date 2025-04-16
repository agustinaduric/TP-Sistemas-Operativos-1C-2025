package config

type MemoriaConfig struct {
	PortMemory    int    `json:"port_memory"`
	IpMemory string `json:"ip_memory"`
	MemorySize       int    `json:"memory_size"`
	PageSize         int    `json:"page_size"`
	EntriesPerPage   int    `json:"entries_per_page"`
	NumberOfLevels   int    `json:"number_of_levels"`
	MemoryDelay      int    `json:"memory_delay"`
	SwapPath         string `json:"swapfile_path"`
	SwapDelay        int    `json:"swap_delay"`
	LogLevel         string `json:"log_level"`
	DumpPath         string `json:"dump_path"`
}