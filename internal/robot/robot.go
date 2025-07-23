// ============================================================================
// internal/robot/robot.go - 로봇 제어 및 통신 관리
// ============================================================================
// 이 파일은 로봇과의 통신, 명령 전송, 데이터 파싱, 모니터링 등의
// 모든 로봇 관련 기능을 담당합니다.
//
// 주요 기능:
// - JOG 명령 처리 및 전송
// - 로봇 상태 데이터 조회 및 파싱
// - 실시간 위치 모니터링
// - 축/모드 설정 관리
// - 로봇 통신 프로토콜 처리
// ============================================================================

package robot

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/nir414/go-virtual-pendant/internal/types"
)

// ============================================================================
// 상수 정의 (Constants)
// ============================================================================

// 로봇 명령 PID 상수
const (
	JointModePID     = "623"
	CartesianModePID = "624"
	PID_JOG_ENABLE   = "215" // JOG 활성화 PID
	PID_JOG_MODE     = "621" // JOG 모드 설정 PID
	PID_AXIS_SELECT  = "622" // 축 선택 PID
	PID_ROBOT_SELECT = "620" // 로봇 선택 PID
)

// 로봇 통신 URL 상수
const (
	ROBOT_BASE_URL    = "http://192.168.0.1"
	ROBOT_DATA_URL    = ROBOT_BASE_URL + "/ROMDISK/web/Opr/jog/jogrefresh.asp"
	ROBOT_COMMAND_URL = ROBOT_BASE_URL + "/wrtpdb"
	ROBOT_REDIRECT    = "/ROMDISK/web/dbfunctions.asp"
)

// ============================================================================
// 전역 변수 (Global Variables)
// ============================================================================

// 로깅 레벨 전역 변수
var currentLogLevel types.LogLevel

// HTTP 클라이언트 재사용으로 연결 풀링 최적화
var httpClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 2,
		IdleConnTimeout:     30 * time.Second,
	},
}

// 축 정보 정의
var (
	jointAxisInfos = []types.AxisInfo{
		{DisplayName: "J1", Aliases: []string{"joint1", "j1"}},
		{DisplayName: "J2", Aliases: []string{"joint2", "j2"}},
		{DisplayName: "J3", Aliases: []string{"joint3", "j3"}},
		{DisplayName: "J4", Aliases: []string{"joint4", "j4"}},
		{DisplayName: "J5", Aliases: []string{"joint5", "j5"}},
		{DisplayName: "J6", Aliases: []string{"joint6", "j6"}},
	}

	cartesianAxisInfos = []types.AxisInfo{
		{DisplayName: "X", Aliases: []string{"x"}},
		{DisplayName: "Y", Aliases: []string{"y"}},
		{DisplayName: "Z", Aliases: []string{"z"}},
		{DisplayName: "Rx", Aliases: []string{"rx"}},
		{DisplayName: "Ry", Aliases: []string{"ry"}},
		{DisplayName: "Rz", Aliases: []string{"rz"}},
	}

	// 동적으로 생성된 축 맵들
	jointAxisMap     = generateAxisMap(JointModePID, jointAxisInfos)
	cartesianAxisMap = generateAxisMap(CartesianModePID, cartesianAxisInfos)
)

// 모드 정보 정의
var (
	jogModeInfos = []types.ModeInfo{
		{DisplayName: "Computer", ModeNumber: 0},
		{DisplayName: "Joint", ModeNumber: 1},
		{DisplayName: "World", ModeNumber: 2},
		{DisplayName: "Tool", ModeNumber: 3},
		{DisplayName: "Free", ModeNumber: 4},
	}

	// 동적으로 생성된 모드 맵
	jogModeConfigMap = generateModeMap(jogModeInfos)
)

// ============================================================================
// 초기화 (Initialization)
// ============================================================================

func init() {
	// 환경변수로 로그 레벨 설정
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		currentLogLevel = types.LogLevelDebug
	case "VERBOSE":
		currentLogLevel = types.LogLevelVerbose
	default:
		currentLogLevel = types.LogLevelInfo
	}
}

// ============================================================================
// 로깅 유틸리티 (Logging Utilities)
// ============================================================================

// logInfo 정보 레벨 로그 출력
func logInfo(format string, args ...interface{}) {
	if currentLogLevel >= types.LogLevelInfo {
		log.Printf("ℹ️  "+format, args...)
	}
}

// logDebug 디버그 레벨 로그 출력
func logDebug(format string, args ...interface{}) {
	if currentLogLevel >= types.LogLevelDebug {
		log.Printf("🔍 "+format, args...)
	}
}

// logVerbose 상세 레벨 로그 출력
func logVerbose(format string, args ...interface{}) {
	if currentLogLevel >= types.LogLevelVerbose {
		log.Printf("🔧 "+format, args...)
	}
}

// ============================================================================
// 축 및 모드 생성 함수 (Generator Functions)
// ============================================================================

// generateAxisMap 축 맵을 동적으로 생성하는 함수
func generateAxisMap(pidBase string, axisInfos []types.AxisInfo) map[string]types.AxisConfig {
	axisMap := make(map[string]types.AxisConfig)
	for i, info := range axisInfos {
		config := types.AxisConfig{PID: pidBase, Axis: i + 1}
		for _, alias := range info.Aliases {
			axisMap[alias] = config
		}
	}
	return axisMap
}

// generateModeMap 모드 맵을 동적으로 생성하는 함수
func generateModeMap(modeInfos []types.ModeInfo) map[string]types.ModeConfig {
	modeMap := make(map[string]types.ModeConfig)
	for i, info := range modeInfos {
		var enable string
		if i == 0 { // computer 모드만 "0"
			enable = "0"
		} else {
			enable = "1"
		}
		config := types.ModeConfig{
			Enable:  enable,
			JogMode: fmt.Sprintf("%d", info.ModeNumber),
		}
		modeMap[strings.ToLower(info.DisplayName)] = config
	}
	return modeMap
}

// ============================================================================
// 명령 빌더 함수 (Command Builder Functions)
// ============================================================================

// buildAxisCommand 축별 명령 생성 헬퍼 함수
func buildAxisCommand(axisMap map[string]types.AxisConfig, axis string, step float64) (string, string, error) {
	if config, exists := axisMap[axis]; exists {
		pidCommand := fmt.Sprintf("%s,%d,0,0", config.PID, config.Axis)
		pvalCommand := fmt.Sprintf("%.3f", step)
		return pidCommand, pvalCommand, nil
	}
	return "", "", fmt.Errorf("지원하지 않는 축: %s", axis)
}

// buildJogCommand JOG 명령을 로봇 프로토콜로 변환
func buildJogCommand(cmd types.JogCommand) (url.Values, error) {
	form := url.Values{}
	// Send two PIDs: movement command (PID1) and jog start trigger (PID2)
	form.Set("nPID", "2")
	form.Set("Redirect", ROBOT_REDIRECT)

	// 조깅 중단 명령 처리 (원본 jogright.asp의 jog(0) 방식)
	if cmd.Dir == "stop" {
		// 조깅 중단은 PID 값 0으로 전송
		form.Set("PID1", "0,0,0,0")
		form.Set("PVal1", "0")
		logInfo("조깅 중단 명령 전송")
		return form, nil
	}

	// 방향에 따른 부호 결정
	direction := 1.0
	if cmd.Dir == "negative" {
		direction = -1.0
	}

	step := cmd.Step * direction

	var pidCommand, pvalCommand string
	var err error

	switch cmd.Mode {
	case "joint":
		pidCommand, pvalCommand, err = buildAxisCommand(jointAxisMap, cmd.Axis, step)
		if err != nil {
			return nil, fmt.Errorf("지원하지 않는 조인트: %s", cmd.Axis)
		}
	case "cartesian":
		pidCommand, pvalCommand, err = buildAxisCommand(cartesianAxisMap, cmd.Axis, step)
		if err != nil {
			return nil, fmt.Errorf("지원하지 않는 카르테시안 축: %s", cmd.Axis)
		}
	default:
		return nil, fmt.Errorf("지원하지 않는 모드: %s", cmd.Mode)
	}

	// Movement command: PID1 = speed/axis command
	form.Set("PID1", pidCommand)
	form.Set("PVal1", pvalCommand)
	// Jog heartbeat trigger: always PID2=CartesianModePID with trigger value 1
	form.Set("PID2", fmt.Sprintf("%s,1,0,0", CartesianModePID))
	form.Set("PVal2", "1")

	return form, nil
}

// ============================================================================
// 로봇 통신 함수 (Robot Communication Functions)
// ============================================================================

// sendRobotCommand 로봇에 명령 전송
func sendRobotCommand(form url.Values, successMsg string) (*types.JogResponse, error) {
	resp, err := httpClient.PostForm(ROBOT_COMMAND_URL, form)
	if err != nil {
		return &types.JogResponse{
			Success: false,
			Message: "로봇 통신 실패: " + err.Error(),
			Command: form.Encode(),
		}, err
	}
	defer resp.Body.Close()

	response := &types.JogResponse{
		Success: true,
		Message: successMsg,
		Command: form.Encode(),
	}

	// 성공 메시지 로그
	logInfo("%s", successMsg)

	return response, nil
}

// SendJogCommand JOG 명령을 로봇에 전송 (외부 호출용)
func SendJogCommand(cmd types.JogCommand) (*types.JogResponse, error) {
	// 조깅 중단 명령 처리
	if cmd.Dir == "stop" {
		logInfo("JOG 중단 명령 수신")

		// 중단 명령을 로봇 프로토콜로 변환
		form, err := buildJogCommand(cmd)
		if err != nil {
			return &types.JogResponse{
				Success: false,
				Message: "중단 명령 생성 실패: " + err.Error(),
				Command: "",
			}, err
		}

		// 로봇에 중단 명령 전송
		response, err := sendRobotCommand(form, "JOG 중단 명령 전송 완료")
		if err != nil {
			return response, err
		}

		logDebug("전송된 중단 명령: %s", response.Command)
		return response, nil
	}

	// 기본값 설정
	if cmd.Mode == "" {
		cmd.Mode = "joint"
	}
	if cmd.Step == 0 {
		cmd.Step = 1.0 // 기본 스텝
	}

	// 명령 수신 로그
	logInfo("JOG 명령 수신: 모드=%s, 축=%s, 방향=%s, 스텝=%.3f", cmd.Mode, cmd.Axis, cmd.Dir, cmd.Step)

	// JOG 명령을 로봇 프로토콜로 변환
	form, err := buildJogCommand(cmd)
	if err != nil {
		return &types.JogResponse{
			Success: false,
			Message: "명령 생성 실패: " + err.Error(),
			Command: "",
		}, err
	}

	// 로봇에 명령 전송
	successMsg := fmt.Sprintf("JOG 명령 성공: %s %s %s %.3f", cmd.Mode, cmd.Axis, cmd.Dir, cmd.Step)
	response, err := sendRobotCommand(form, successMsg)
	if err != nil {
		return response, err
	}

	// 명령 전송 로그
	logDebug("전송된 명령: %s", response.Command)

	return response, nil
}

// SetRobotJogMode 로봇 JOG 모드 변경 (외부 호출용)
func SetRobotJogMode(mode string) (*types.JogResponse, error) {
	config, exists := jogModeConfigMap[mode]
	if !exists {
		return &types.JogResponse{
			Success: false,
			Message: "지원하지 않는 모드: " + mode,
			Command: "",
		}, fmt.Errorf("unsupported mode: %s", mode)
	}

	form := url.Values{}
	form.Set("nPID", "2")
	form.Set("Redirect", ROBOT_REDIRECT)
	form.Set("PID1", fmt.Sprintf("%s,%s,0,0", PID_JOG_ENABLE, config.Enable))
	form.Set("PVal1", config.Enable)
	form.Set("PID2", fmt.Sprintf("%s,%s,0,0", PID_JOG_MODE, config.JogMode))
	form.Set("PVal2", config.JogMode)

	// 모드 변경 로그
	logInfo("JOG 모드 변경: %s", mode)

	// 로봇에 명령 전송
	successMsg := fmt.Sprintf("JOG 모드 변경 성공: %s", mode)
	return sendRobotCommand(form, successMsg)
}

// SetRobotAxis 로봇 축 선택 (외부 호출용)
func SetRobotAxis(axis int, robot int) (*types.JogResponse, error) {
	form := url.Values{}
	form.Set("nPID", "2")
	form.Set("Redirect", ROBOT_REDIRECT)

	// 축 선택 PID 설정 (원본 jogscripts.asp 참고)
	form.Set("PID1", fmt.Sprintf("%s,0,0,0", PID_AXIS_SELECT))
	form.Set("PVal1", fmt.Sprintf("%d", axis))
	form.Set("PID2", fmt.Sprintf("%s,0,0,0", PID_ROBOT_SELECT))
	form.Set("PVal2", fmt.Sprintf("%d", robot))

	// 축 선택 로그
	logInfo("축 선택: 축=%d, 로봇=%d", axis, robot)

	// 로봇에 명령 전송
	successMsg := fmt.Sprintf("축 선택 성공: 축=%d, 로봇=%d", axis, robot)
	return sendRobotCommand(form, successMsg)
}

// ============================================================================
// 데이터 파싱 및 조회 함수 (Data Parsing & Retrieval Functions)
// ============================================================================

// GetRobotData 로봇의 모든 데이터 조회 (외부 호출용)
func GetRobotData() (*types.JogState, error) {
	res, err := httpClient.Get(ROBOT_DATA_URL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// 응답 내용을 텍스트로 읽기
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	response := strings.TrimSpace(string(body))

	// 파이프(|)로 구분된 데이터 파싱
	parts := strings.Split(response, "|")

	if len(parts) < 25 {
		return nil, fmt.Errorf("응답 데이터가 부족합니다: %d개 항목", len(parts))
	}

	// 카르테시안 좌표 (X,Y,Z,Rx,Ry,Rz) - jData[0-5]
	cartesian := make([]float64, 6)
	for i := 0; i < 6 && i < len(parts); i++ {
		if v, err := parseFloat(parts[i]); err == nil {
			cartesian[i] = v
		}
	}

	// 조인트 각도 (Joint1-12) - jData[6-17]
	joint := make([]float64, 12)
	for i := 0; i < 12 && (i+6) < len(parts); i++ {
		if v, err := parseFloat(parts[i+6]); err == nil {
			joint[i] = v
		}
	}

	// 툴 데이터 파싱 - jData[24] (콤마로 구분)
	toolData := make([]float64, 6)
	if len(parts) > 24 && parts[24] != "" {
		toolParts := strings.Split(parts[24], ",")
		for i := 0; i < 6 && i < len(toolParts); i++ {
			if v, err := parseFloat(toolParts[i]); err == nil {
				toolData[i] = v
			}
		}
	}

	// 상태 정보 파싱
	status := types.JogStatus{}
	if len(parts) > 19 {
		if v, err := parseFloat(parts[19]); err == nil {
			status.AxisCount = int(v)
		}
	}
	if len(parts) > 20 {
		if v, err := parseFloat(parts[20]); err == nil {
			status.AllowJog = v > 0
		}
	}
	if len(parts) > 21 {
		if v, err := parseFloat(parts[21]); err == nil {
			status.JogMode = int(v)
			status.JogModeText = getJogModeText(int(v))
		}
	}
	if len(parts) > 22 {
		if v, err := parseFloat(parts[22]); err == nil {
			status.PowerState = int(v)
		}
	}
	if len(parts) > 23 {
		status.ErrorDesc = strings.TrimSpace(parts[23])
	}

	// 현재 선택된 축 정보 (임시로 1로 설정, 실제로는 별도 API에서 가져와야 함)
	status.SelectedAxis = 1
	status.SelectedAxisText = getAxisText(status.JogMode, status.SelectedAxis)

	return &types.JogState{
		Cartesian: cartesian,
		Joint:     joint,
		ToolData:  toolData,
		Status:    status,
	}, nil
}

// ============================================================================
// 모니터링 함수 (Monitoring Functions)
// ============================================================================

// MonitorRobotPosition 로봇 위치를 주기적으로 모니터링 (외부 호출용)
func MonitorRobotPosition() {
	ticker := time.NewTicker(1 * time.Second) // 1초마다 확인
	defer ticker.Stop()

	var prevData *types.JogState // 이전 상태 저장용

	for range ticker.C {
		data, err := GetRobotData()
		if err != nil {
			logDebug("좌표 읽기 실패: %v", err)
			continue
		}

		// 데이터 변경 감지 - 이전 상태와 비교
		if prevData == nil || hasDataChanged(prevData, data) {
			// 시간 정보와 함께 상태 출력 (변경된 경우에만)
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] 🤖 JOG=(%.1f, %.1f, %.1f) | XYZ=(%.1f, %.1f, %.1f) | 모드=%s | %s\n",
				timestamp,
				getSafeValue(data.Joint, 0), getSafeValue(data.Joint, 1), getSafeValue(data.Joint, 2),
				getSafeValue(data.Cartesian, 0), getSafeValue(data.Cartesian, 1), getSafeValue(data.Cartesian, 2),
				data.Status.JogModeText,
				func() string {
					if data.Status.ErrorDesc != "" {
						return "⚠️ " + data.Status.ErrorDesc
					}
					return "✅ 정상"
				}())

			// 현재 상태를 이전 상태로 저장
			prevData = data
		}
	}
}

// hasDataChanged 두 JogState 구조체를 비교하여 변경 사항을 감지
func hasDataChanged(prev, current *types.JogState) bool {
	// 조인트 각도 변경 확인 (0.1도 이상 차이)
	for i := 0; i < 3 && i < len(prev.Joint) && i < len(current.Joint); i++ {
		if abs(prev.Joint[i]-current.Joint[i]) > 0.1 {
			return true
		}
	}

	// 카르테시안 좌표 변경 확인 (0.1mm 이상 차이)
	for i := 0; i < 3 && i < len(prev.Cartesian) && i < len(current.Cartesian); i++ {
		if abs(prev.Cartesian[i]-current.Cartesian[i]) > 0.1 {
			return true
		}
	}

	// 모드 변경 확인
	if prev.Status.JogMode != current.Status.JogMode {
		return true
	}

	// 에러 상태 변경 확인
	if prev.Status.ErrorDesc != current.Status.ErrorDesc {
		return true
	}

	return false
}

// ============================================================================
// 유틸리티 함수 (Utility Functions)
// ============================================================================

// parseFloat 문자열을 float64로 안전하게 파싱
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0.0, nil
	}
	var v float64
	_, err := fmt.Sscanf(s, "%f", &v)
	return v, err
}

// getSafeValue 배열 경계 검사와 함께 안전하게 값 조회
func getSafeValue(coords []float64, index int) float64 {
	if index < len(coords) {
		return coords[index]
	}
	return 0.0
}

// getJogModeText 모드 번호를 텍스트로 변환
func getJogModeText(mode int) string {
	// 모드 번호가 유효한 범위 내에 있는지 확인
	if mode >= 0 && mode < len(jogModeInfos) {
		return jogModeInfos[mode].DisplayName
	}

	// 범위를 벗어난 경우 기본 형식으로 반환
	return fmt.Sprintf("Mode%d", mode)
}

// getAxisText 모드와 축 번호에 따른 축 이름 반환
func getAxisText(jogMode int, axisNum int) string {
	var axisInfos []types.AxisInfo

	if jogMode == 1 { // Joint mode
		axisInfos = jointAxisInfos
	} else { // Cartesian modes (World, Tool, etc.)
		axisInfos = cartesianAxisInfos
	}

	// 축 번호가 유효한 범위 내에 있는지 확인
	if axisNum >= 1 && axisNum <= len(axisInfos) {
		return axisInfos[axisNum-1].DisplayName
	}

	// 범위를 벗어난 경우 기본 형식으로 반환
	if jogMode == 1 {
		return fmt.Sprintf("J%d", axisNum)
	}
	return fmt.Sprintf("Axis%d", axisNum)
}

// abs float64의 절댓값 반환
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
