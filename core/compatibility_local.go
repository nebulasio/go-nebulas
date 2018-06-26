package core

// CompatibilityLocal ..
type CompatibilityLocal struct {
	transferFromContractEventRecordableHeight uint64

	acceptFuncAvailableHeight uint64

	randomAvailableHeight uint64

	dateAvailableHeight uint64

	recordCallContractResultHeight uint64

	nvmMemoryLimitWithoutInjectHeight uint64

	wsResetRecordDependencyHeight uint64

	v8JSLibVersionControlHeight uint64

	transferFromContractFailureEventRecordableHeight uint64

	newNvmExeTimeoutConsumeGasHeight uint64

	v8JSLibVersionHeightMap *V8JSLibVersionHeightMap

	nvmGasLimitWithoutTimeoutHeight uint64
}

// NewCompatibilityLocal ..
func NewCompatibilityLocal() Compatibility {
	return &CompatibilityLocal{
		transferFromContractEventRecordableHeight:        2,
		acceptFuncAvailableHeight:                        2,
		randomAvailableHeight:                            2,
		dateAvailableHeight:                              2,
		recordCallContractResultHeight:                   2,
		nvmMemoryLimitWithoutInjectHeight:                325666,
		wsResetRecordDependencyHeight:                    2,
		v8JSLibVersionControlHeight:                      2,
		transferFromContractFailureEventRecordableHeight: 2,
		newNvmExeTimeoutConsumeGasHeight:                 2,
		v8JSLibVersionHeightMap: &V8JSLibVersionHeightMap{
			Data: map[string]uint64{
				"1.0.5": 2,
				"1.1.0": 3,
			},
			DescKeys: []string{"1.1.0", "1.0.5"},
		},
		nvmGasLimitWithoutTimeoutHeight: 2,
	}
}

// TransferFromContractEventRecordableHeight ..
func (c *CompatibilityLocal) TransferFromContractEventRecordableHeight() uint64 {
	return c.transferFromContractEventRecordableHeight
}

// AcceptFuncAvailableHeight ..
func (c *CompatibilityLocal) AcceptFuncAvailableHeight() uint64 {
	return c.acceptFuncAvailableHeight
}

// RandomAvailableHeight ..
func (c *CompatibilityLocal) RandomAvailableHeight() uint64 {
	return c.randomAvailableHeight
}

// DateAvailableHeight ..
func (c *CompatibilityLocal) DateAvailableHeight() uint64 {
	return c.dateAvailableHeight
}

// RecordCallContractResultHeight ..
func (c *CompatibilityLocal) RecordCallContractResultHeight() uint64 {
	return c.recordCallContractResultHeight
}

// NvmMemoryLimitWithoutInjectHeight ..
func (c *CompatibilityLocal) NvmMemoryLimitWithoutInjectHeight() uint64 {
	return c.nvmMemoryLimitWithoutInjectHeight
}

// WsResetRecordDependencyHeight ..
func (c *CompatibilityLocal) WsResetRecordDependencyHeight() uint64 {
	return c.wsResetRecordDependencyHeight
}

// V8JSLibVersionControlHeight ..
func (c *CompatibilityLocal) V8JSLibVersionControlHeight() uint64 {
	return c.v8JSLibVersionControlHeight
}

// TransferFromContractFailureEventRecordableHeight ..
func (c *CompatibilityLocal) TransferFromContractFailureEventRecordableHeight() uint64 {
	return c.transferFromContractFailureEventRecordableHeight
}

// NewNvmExeTimeoutConsumeGasHeight ..
func (c *CompatibilityLocal) NewNvmExeTimeoutConsumeGasHeight() uint64 {
	return c.newNvmExeTimeoutConsumeGasHeight
}

// V8JSLibVersionHeightMap ..
func (c *CompatibilityLocal) V8JSLibVersionHeightMap() *V8JSLibVersionHeightMap {
	return c.v8JSLibVersionHeightMap
}

// NvmGasLimitWithoutTimeoutHeight ..
func (c *CompatibilityLocal) NvmGasLimitWithoutTimeoutHeight() uint64 {
	return c.nvmGasLimitWithoutTimeoutHeight
}
