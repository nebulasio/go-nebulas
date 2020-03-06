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

	nvmExeTimeoutHeight []uint64

	wsResetRecordDependencyHeight2 uint64

	transferFromContractFailureEventRecordableHeight2 uint64

	nvmValueCheckUpdateHeight uint64

	nbreAvailableHeight uint64

	nrc20SecurityCheckHeight uint64

	nbreSplitHeight uint64

	nodeUpdateHeight uint64

	nodeStartSerial        uint64
	nodeAccessContract     *Address
	nodePodContract        *Address
	nodeGovernanceContract *Address
}

// NewCompatibilityLocal ..
func NewCompatibilityLocal() Compatibility {
	nodeAccessContract, _ := AddressParse("n1wRERsLCoGsh2YZu7Qy74iFrraJwnV9gKX")
	nodePodContract, _ := AddressParse("n1xS3BoziPPidb5nXDmGfH9pb4RHQMfyBe9")
	nodeGovernanceContract, _ := AddressParse("n1tD8Lh4vxWpha1K1hmZ7ZuCg86THvFyYwr")

	return &CompatibilityLocal{
		transferFromContractEventRecordableHeight:         2,
		acceptFuncAvailableHeight:                         2,
		randomAvailableHeight:                             2,
		dateAvailableHeight:                               2,
		recordCallContractResultHeight:                    2,
		nvmMemoryLimitWithoutInjectHeight:                 325666,
		wsResetRecordDependencyHeight:                     2,
		v8JSLibVersionControlHeight:                       2,
		transferFromContractFailureEventRecordableHeight:  2,
		newNvmExeTimeoutConsumeGasHeight:                  2,
		nvmExeTimeoutHeight:                               []uint64{2},
		wsResetRecordDependencyHeight2:                    2,
		transferFromContractFailureEventRecordableHeight2: 2,

		v8JSLibVersionHeightMap: &V8JSLibVersionHeightMap{
			Data: map[string]uint64{
				"1.0.5": 2,
				"1.1.0": 3,
			},
			DescKeys: []string{"1.1.0", "1.0.5"},
		},
		nvmGasLimitWithoutTimeoutHeight: 2,
		nvmValueCheckUpdateHeight:       2,
		nbreAvailableHeight:             2,
		nrc20SecurityCheckHeight:        2,
		nbreSplitHeight:                 3,
		nodeUpdateHeight:                3,

		nodeStartSerial:        0,
		nodeAccessContract:     nodeAccessContract,
		nodePodContract:        nodePodContract,
		nodeGovernanceContract: nodeGovernanceContract,
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

// WsResetRecordDependencyHeight2 ..
func (c *CompatibilityLocal) WsResetRecordDependencyHeight2() uint64 {
	return c.wsResetRecordDependencyHeight2
}

// V8JSLibVersionControlHeight ..
func (c *CompatibilityLocal) V8JSLibVersionControlHeight() uint64 {
	return c.v8JSLibVersionControlHeight
}

// TransferFromContractFailureEventRecordableHeight ..
func (c *CompatibilityLocal) TransferFromContractFailureEventRecordableHeight() uint64 {
	return c.transferFromContractFailureEventRecordableHeight
}

// TransferFromContractFailureEventRecordableHeight2 ..
func (c *CompatibilityLocal) TransferFromContractFailureEventRecordableHeight2() uint64 {
	return c.transferFromContractFailureEventRecordableHeight2
}

// NewNvmExeTimeoutConsumeGasHeight ..
func (c *CompatibilityLocal) NewNvmExeTimeoutConsumeGasHeight() uint64 {
	return c.newNvmExeTimeoutConsumeGasHeight
}

// NvmExeTimeoutHeight ..
func (c *CompatibilityLocal) NvmExeTimeoutHeight() []uint64 {
	return c.nvmExeTimeoutHeight
}

// V8JSLibVersionHeightMap ..
func (c *CompatibilityLocal) V8JSLibVersionHeightMap() *V8JSLibVersionHeightMap {
	return c.v8JSLibVersionHeightMap
}

// NvmGasLimitWithoutTimeoutHeight ..
func (c *CompatibilityLocal) NvmGasLimitWithoutTimeoutHeight() uint64 {
	return c.nvmGasLimitWithoutTimeoutHeight
}

// NvmValueCheckUpdateHeight ..
func (c *CompatibilityLocal) NvmValueCheckUpdateHeight() uint64 {
	return c.nvmValueCheckUpdateHeight
}

// NbreAvailableHeight ..
func (c *CompatibilityLocal) NbreAvailableHeight() uint64 {
	return c.nbreAvailableHeight
}

// Nrc20SecurityCheckHeight ..
func (c *CompatibilityLocal) Nrc20SecurityCheckHeight() uint64 {
	return c.nrc20SecurityCheckHeight
}

// NbreSplitHeight ..
func (c *CompatibilityLocal) NbreSplitHeight() uint64 {
	return c.nbreSplitHeight
}

// NodeUpdateHeight ..
func (c *CompatibilityLocal) NodeUpdateHeight() uint64 {
	return c.nodeUpdateHeight
}

// NodeStartSerial ..
func (c *CompatibilityLocal) NodeStartSerial() uint64 {
	return c.nodeStartSerial
}

// NodeAccessContract ..
func (c *CompatibilityLocal) NodeAccessContract() *Address {
	return c.nodeAccessContract
}

// NodePodContract ..
func (c *CompatibilityLocal) NodePodContract() *Address {
	return c.nodePodContract
}

// NodeGovernanceContract ..
func (c *CompatibilityLocal) NodeGovernanceContract() *Address {
	return c.nodeGovernanceContract
}
