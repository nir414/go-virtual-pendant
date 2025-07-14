// ============================================================================
// internal/types/types.go - 공통 데이터 타입 정의
// ============================================================================
// Virtual Pendant 시스템에서 사용하는 모든 공통 타입을 정의합니다.
// JOG 명령, 로봇 상태, 응답 등의 구조체가 포함됩니다.
// ============================================================================

package types

// ============================================================================
// 명령 관련 타입 (Command Types)
// ============================================================================

// JogCommand JOG 명령 요청 구조체
type JogCommand struct {
	Axis string  `json:"axis"` // "joint1", "joint2", ..., "x", "y", "z", "rx", "ry", "rz"
	Dir  string  `json:"dir"`  // "positive", "negative"
	Step float64 `json:"step"` // 이동 거리/각도
	Mode string  `json:"mode"` // "joint", "cartesian"
}

// JogResponse JOG 명령 응답 구조체
type JogResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Command string `json:"command_sent"`
}

// ============================================================================
// 상태 관련 타입 (State Types)
// ============================================================================

// JogState 로봇 현재 상태 구조체
type JogState struct {
	Cartesian []float64 `json:"cartesian"` // X,Y,Z,Rx,Ry,Rz
	Joint     []float64 `json:"joint"`     // Joint1-12
	ToolData  []float64 `json:"tool"`      // 툴 데이터
	Status    JogStatus `json:"status"`    // 상태 정보
}

// JogStatus 로봇 상태 정보 구조체
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

// ============================================================================
// 설정 관련 타입 (Configuration Types)
// ============================================================================

// LogLevel 로깅 레벨 타입
type LogLevel int

// 로깅 레벨 상수
const (
	LogLevelInfo LogLevel = iota
	LogLevelDebug
	LogLevelVerbose
)

// AxisConfig 축 설정 구조체
type AxisConfig struct {
	PID  string
	Axis int
}

// AxisInfo 축 정보 구조체 (이름과 표시명 포함)
type AxisInfo struct {
	Config      AxisConfig
	DisplayName string
	Aliases     []string // 별칭들 (j1, joint1 등)
}

// ModeConfig JOG 모드 설정 구조체
type ModeConfig struct {
	Enable  string
	JogMode string
}

// ModeInfo JOG 모드 정보 구조체 (설정과 표시명 포함)
type ModeInfo struct {
	Config      ModeConfig
	DisplayName string
	ModeNumber  int
}
