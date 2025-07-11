package main

// JogCommand represents a jog command request
type JogCommand struct {
	Axis string  `json:"axis"` // "joint1", "joint2", ..., "x", "y", "z", "rx", "ry", "rz"
	Dir  string  `json:"dir"`  // "positive", "negative"
	Step float64 `json:"step"` // 이동 거리/각도
	Mode string  `json:"mode"` // "joint", "cartesian"
}

// JogResponse represents a jog command response
type JogResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Command string `json:"command_sent"`
}

// JogState represents the current robot state
type JogState struct {
	Cartesian []float64 `json:"cartesian"` // X,Y,Z,Rx,Ry,Rz
	Joint     []float64 `json:"joint"`     // Joint1-12
	ToolData  []float64 `json:"tool"`      // 툴 데이터
	Status    JogStatus `json:"status"`    // 상태 정보
}

// JogStatus represents robot status information
type JogStatus struct {
	AxisCount        int    `json:"axis_count"`
	AllowJog         bool   `json:"allow_jog"`
	JogMode          int    `json:"jog_mode"`
	JogModeText      string `json:"jog_mode_text"`      // 모드명 (Joint, World, Tool, etc.)
	SelectedAxis     int    `json:"selected_axis"`      // 현재 선택된 축 (1-6)
	SelectedAxisText string `json:"selected_axis_text"` // 축명 (J1, X, etc.)
	PowerState       int    `json:"power_state"`
	ErrorDesc        string `json:"error_desc"`
}
