// ============================================================================
// internal/types/types.go - 멀티 스택 공통 데이터 타입 정의
// ============================================================================
// Virtual Pendant 시스템에서 사용하는 모든 공통 타입을 정의합니다.
// 🎯 목표: Go, JavaScript, Chrome, Node.js 등 멀티 스택 프로젝트 지원
// 🔧 특징: 기존 코드 스타일 보존, 최소한의 변경, 높은 호환성
//
// 주요 기능:
// - JOG 명령 및 로봇 상태 (Go ↔ JavaScript 호환)
// - 웹 API 응답 형식 표준화
// - 크로스 플랫폼 설정 관리
// - 디버깅 및 로깅 시스템 통합
// ============================================================================

package types

// ============================================================================
// 멀티 스택 호환 명령 타입 (Multi-Stack Command Types)
// ============================================================================

// JogCommand JOG 명령 요청 구조체 (Go ↔ JavaScript 호환)
type JogCommand struct {
	Axis string  `json:"axis"` // "joint1", "joint2", ..., "x", "y", "z", "rx", "ry", "rz"
	Dir  string  `json:"dir"`  // "positive", "negative"
	Step float64 `json:"step"` // 이동 거리/각도
	Mode string  `json:"mode"` // "joint", "cartesian"
}

// JogResponse JOG 명령 응답 구조체 (표준 웹 API 응답 형식)
type JogResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Command   string `json:"command_sent"`
	Timestamp string `json:"timestamp,omitempty"`  // ISO 8601 형식 (JavaScript Date 호환)
	ErrorCode string `json:"error_code,omitempty"` // 에러 코드 (디버깅용)
}

// ============================================================================
// 크로스 플랫폼 상태 타입 (Cross-Platform State Types)
// ============================================================================

// JogState 로봇 현재 상태 구조체 (웹 API 표준 응답)
type JogState struct {
	Cartesian []float64 `json:"cartesian"` // X,Y,Z,Rx,Ry,Rz
	Joint     []float64 `json:"joint"`     // Joint1-12
	ToolData  []float64 `json:"tool"`      // 툴 데이터
	Status    JogStatus `json:"status"`    // 상태 정보
	Meta      StateMeta `json:"meta"`      // 메타데이터 (디버깅/로깅용)
}

// JogStatus 로봇 상태 정보 구조체 (JavaScript 친화적)
type JogStatus struct {
	AxisCount        int    `json:"axis_count"`
	AllowJog         bool   `json:"allow_jog"`
	JogMode          int    `json:"jog_mode"`
	JogModeText      string `json:"jog_mode_text"`      // 모드명 (Joint, World, Tool, etc.)
	SelectedAxis     int    `json:"selected_axis"`      // 현재 선택된 축 (1-6)
	SelectedAxisText string `json:"selected_axis_text"` // 축명 (J1, X, etc.)
	PowerState       int    `json:"power_state"`
	ErrorDesc        string `json:"error_desc"`
	IsConnected      bool   `json:"is_connected"` // 연결 상태 (웹 UI용)
}

// StateMeta 상태 메타데이터 (디버깅 및 멀티 스택 지원)
type StateMeta struct {
	Timestamp   string `json:"timestamp"`   // ISO 8601 형식
	Source      string `json:"source"`      // "go-server", "js-client", "chrome-extension"
	Version     string `json:"version"`     // API 버전
	Environment string `json:"environment"` // "development", "production", "test"
	DebugMode   bool   `json:"debug_mode"`  // 디버그 모드 여부
}

// ============================================================================
// 멀티 스택 설정 타입 (Multi-Stack Configuration Types)
// ============================================================================

// LogLevel 로깅 레벨 타입 (Go, JavaScript, Node.js 공통)
type LogLevel int

// 로깅 레벨 상수 (크로스 플랫폼 호환)
const (
	LogLevelInfo LogLevel = iota
	LogLevelDebug
	LogLevelVerbose
)

// Environment 환경 타입 (멀티 스택 지원)
type Environment string

// 환경 상수 (Go, JavaScript, Node.js 공통)
const (
	EnvDevelopment Environment = "development"
	EnvProduction  Environment = "production"
	EnvTest        Environment = "test"
	EnvDebug       Environment = "debug"
)

// Platform 플랫폼 타입 (멀티 스택 지원)
type Platform string

// 플랫폼 상수 (다양한 스택 지원)
const (
	PlatformGoServer   Platform = "go-server"
	PlatformJavaScript Platform = "javascript"
	PlatformChrome     Platform = "chrome"
	PlatformNodeJS     Platform = "nodejs"
	PlatformWebApp     Platform = "webapp"
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

// ============================================================================
// 웹 API 호환 요청 타입 (Web API Compatible Request Types)
// ============================================================================

// SetJogModeRequest JOG 모드 변경 요청 (표준 웹 API 형식)
type SetJogModeRequest struct {
	Mode string      `json:"mode"`           // "computer", "joint", "world", "tool", "free"
	Meta RequestMeta `json:"meta,omitempty"` // 요청 메타데이터
}

// SetAxisRequest 축 선택 요청 (표준 웹 API 형식)
type SetAxisRequest struct {
	Axis  int         `json:"axis"`           // 1-6 for joints, 1-6 for cartesian
	Robot int         `json:"robot"`          // robot number (usually 1)
	Meta  RequestMeta `json:"meta,omitempty"` // 요청 메타데이터
}

// RequestMeta 요청 메타데이터 (디버깅 및 추적용)
type RequestMeta struct {
	ClientID  string   `json:"client_id,omitempty"`  // 클라이언트 식별자
	Platform  Platform `json:"platform,omitempty"`   // 요청 플랫폼
	UserAgent string   `json:"user_agent,omitempty"` // 브라우저 정보
	Timestamp string   `json:"timestamp,omitempty"`  // 요청 시각
	SessionID string   `json:"session_id,omitempty"` // 세션 식별자
	TraceID   string   `json:"trace_id,omitempty"`   // 추적 식별자
}

// ============================================================================
// 크로스 플랫폼 서버 설정 타입 (Cross-Platform Server Configuration)
// ============================================================================

// ServerConfig 서버 설정 (멀티 스택 환경 지원)
type ServerConfig struct {
	Port         string      `json:"port"`
	Host         string      `json:"host"`
	APIBasePath  string      `json:"api_base_path"`
	Environment  Environment `json:"environment"`
	Platform     Platform    `json:"platform"`
	EnableCORS   bool        `json:"enable_cors"`   // 웹 앱 지원
	EnableWSS    bool        `json:"enable_wss"`    // WebSocket 지원
	StaticPath   string      `json:"static_path"`   // 정적 파일 경로
	TemplatePath string      `json:"template_path"` // 템플릿 경로
	LogLevel     LogLevel    `json:"log_level"`
	DebugMode    bool        `json:"debug_mode"`
}

// WebSocketConfig WebSocket 설정 (실시간 통신 지원)
type WebSocketConfig struct {
	Enable            bool   `json:"enable"`
	Endpoint          string `json:"endpoint"`
	MaxConnections    int    `json:"max_connections"`
	HeartbeatInterval int    `json:"heartbeat_interval"` // 초 단위
}

// CORSConfig CORS 설정 (웹 앱 호환성)
type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
	MaxAge         int      `json:"max_age"`
}

// ============================================================================
// 디버깅 및 로깅 타입 (Debugging & Logging Types)
// ============================================================================

// DebugInfo 디버깅 정보 (멀티 스택 디버깅 지원)
type DebugInfo struct {
	GoVersion    string            `json:"go_version"`
	BuildTime    string            `json:"build_time"`
	GitCommit    string            `json:"git_commit"`
	Platform     Platform          `json:"platform"`
	Environment  Environment       `json:"environment"`
	ConfigValues map[string]string `json:"config_values"`
	HealthChecks []HealthCheck     `json:"health_checks"`
}

// HealthCheck 헬스체크 정보
type HealthCheck struct {
	Name      string `json:"name"`
	Status    string `json:"status"` // "ok", "warning", "error"
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// LogEntry 로그 엔트리 (구조화된 로깅)
type LogEntry struct {
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Timestamp string                 `json:"timestamp"`
	Platform  Platform               `json:"platform"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}
