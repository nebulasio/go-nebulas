package core

// CompatibilityTestNet ..
type CompatibilityTestNet struct {
	transferFromContractEventRecordableHeight uint64

	transferFromContractFailureEventRecordableHeight2 uint64

	acceptFuncAvailableHeight uint64

	randomAvailableHeight uint64

	dateAvailableHeight uint64

	recordCallContractResultHeight uint64

	nvmMemoryLimitWithoutInjectHeight uint64

	wsResetRecordDependencyHeight uint64

	wsResetRecordDependencyHeight2 uint64

	v8JSLibVersionControlHeight uint64

	transferFromContractFailureEventRecordableHeight uint64

	newNvmExeTimeoutConsumeGasHeight uint64

	nvmExeTimeoutHeight []uint64

	v8JSLibVersionHeightMap *V8JSLibVersionHeightMap

	nvmGasLimitWithoutTimeoutHeight uint64

	nvmValueCheckUpdateHeight uint64

	nbreAvailableHeight uint64
}

// NewCompatibilityTestNet ..
func NewCompatibilityTestNet() Compatibility {
	return &CompatibilityTestNet{
		transferFromContractEventRecordableHeight:        199666,
		acceptFuncAvailableHeight:                        199666,
		randomAvailableHeight:                            199666,
		dateAvailableHeight:                              199666,
		recordCallContractResultHeight:                   199666,
		nvmMemoryLimitWithoutInjectHeight:                281800,
		wsResetRecordDependencyHeight:                    281800,
		v8JSLibVersionControlHeight:                      424400,
		transferFromContractFailureEventRecordableHeight: 424400,
		newNvmExeTimeoutConsumeGasHeight:                 424400,
		nvmExeTimeoutHeight:                              []uint64{309121, 309143, 424583, 424588, 424592, 456365, 456375, 456376, 456714, 456772, 816022, 816025, 816028, 816033, 816036},
		v8JSLibVersionHeightMap: &V8JSLibVersionHeightMap{
			Data: map[string]uint64{
				"1.0.5": 424400, // v8JSLibVersionControlHeight
				"1.1.0": 600600,
			},
			DescKeys: []string{"1.1.0", "1.0.5"},
		},

		nvmGasLimitWithoutTimeoutHeight:                   600600,
		wsResetRecordDependencyHeight2:                    600600,
		transferFromContractFailureEventRecordableHeight2: 600600,
		nvmValueCheckUpdateHeight:                         840000,
		nbreAvailableHeight:                               1550000,
	}
}

// TransferFromContractEventRecordableHeight ..
func (c *CompatibilityTestNet) TransferFromContractEventRecordableHeight() uint64 {
	return c.transferFromContractEventRecordableHeight
}

// AcceptFuncAvailableHeight ..
func (c *CompatibilityTestNet) AcceptFuncAvailableHeight() uint64 {
	return c.acceptFuncAvailableHeight
}

// RandomAvailableHeight ..
func (c *CompatibilityTestNet) RandomAvailableHeight() uint64 {
	return c.randomAvailableHeight
}

// DateAvailableHeight ..
func (c *CompatibilityTestNet) DateAvailableHeight() uint64 {
	return c.dateAvailableHeight
}

// RecordCallContractResultHeight ..
func (c *CompatibilityTestNet) RecordCallContractResultHeight() uint64 {
	return c.recordCallContractResultHeight
}

// NvmMemoryLimitWithoutInjectHeight ..
func (c *CompatibilityTestNet) NvmMemoryLimitWithoutInjectHeight() uint64 {
	return c.nvmMemoryLimitWithoutInjectHeight
}

// WsResetRecordDependencyHeight ..
func (c *CompatibilityTestNet) WsResetRecordDependencyHeight() uint64 {
	return c.wsResetRecordDependencyHeight
}

// WsResetRecordDependencyHeight2 ..
func (c *CompatibilityTestNet) WsResetRecordDependencyHeight2() uint64 {
	return c.wsResetRecordDependencyHeight2
}

// V8JSLibVersionControlHeight ..
func (c *CompatibilityTestNet) V8JSLibVersionControlHeight() uint64 {
	return c.v8JSLibVersionControlHeight
}

// TransferFromContractFailureEventRecordableHeight ..
func (c *CompatibilityTestNet) TransferFromContractFailureEventRecordableHeight() uint64 {
	return c.transferFromContractFailureEventRecordableHeight
}

// TransferFromContractFailureEventRecordableHeight2 ..
func (c *CompatibilityTestNet) TransferFromContractFailureEventRecordableHeight2() uint64 {
	return c.transferFromContractFailureEventRecordableHeight2
}

// NewNvmExeTimeoutConsumeGasHeight ..
func (c *CompatibilityTestNet) NewNvmExeTimeoutConsumeGasHeight() uint64 {
	return c.newNvmExeTimeoutConsumeGasHeight
}

// NvmExeTimeoutHeight ..
func (c *CompatibilityTestNet) NvmExeTimeoutHeight() []uint64 {
	return c.nvmExeTimeoutHeight
}

// V8JSLibVersionHeightMap ..
func (c *CompatibilityTestNet) V8JSLibVersionHeightMap() *V8JSLibVersionHeightMap {
	return c.v8JSLibVersionHeightMap
}

// NvmGasLimitWithoutTimeoutHeight ..
func (c *CompatibilityTestNet) NvmGasLimitWithoutTimeoutHeight() uint64 {
	return c.nvmGasLimitWithoutTimeoutHeight
}

// NvmValueCheckUpdateHeight ..
func (c *CompatibilityTestNet) NvmValueCheckUpdateHeight() uint64 {
	return c.nvmValueCheckUpdateHeight
}

// NbreAvailableHeight ..
func (c *CompatibilityTestNet) NbreAvailableHeight() uint64 {
	return c.nbreAvailableHeight
}
