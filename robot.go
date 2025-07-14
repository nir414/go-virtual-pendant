package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// buildJogCommand converts JOG command to robot protocol
func buildJogCommand(cmd JogCommand) (url.Values, error) {
	form := url.Values{}
	form.Set("nPID", "1")
	form.Set("Redirect", "/ROMDISK/web/dbfunctions.asp")

	// 방향에 따른 부호 결정
	direction := 1.0
	if cmd.Dir == "negative" {
		direction = -1.0
	}

	step := cmd.Step * direction

	var pidCommand string
	var pvalCommand string

	switch cmd.Mode {
	case "joint":
		// 조인트 모드 JOG 명령
		switch cmd.Axis {
		case "joint1", "j1":
			pidCommand = "623,1,0,0" // Joint1 조깅
			pvalCommand = fmt.Sprintf("%.3f", step)
		case "joint2", "j2":
			pidCommand = "623,2,0,0" // Joint2 조깅
			pvalCommand = fmt.Sprintf("%.3f", step)
		case "joint3", "j3":
			pidCommand = "623,3,0,0" // Joint3 조깅
			pvalCommand = fmt.Sprintf("%.3f", step)
		case "joint4", "j4":
			pidCommand = "623,4,0,0" // Joint4 조깅
			pvalCommand = fmt.Sprintf("%.3f", step)
		case "joint5", "j5":
			pidCommand = "623,5,0,0" // Joint5 조깅
			pvalCommand = fmt.Sprintf("%.3f", step)
		case "joint6", "j6":
			pidCommand = "623,6,0,0" // Joint6 조깅
			pvalCommand = fmt.Sprintf("%.3f", step)
		default:
			return nil, fmt.Errorf("지원하지 않는 조인트: %s", cmd.Axis)
		}
	case "cartesian":
		// 카르테시안 모드 JOG 명령
		switch cmd.Axis {
		case "x":
			pidCommand = "624,1,0,0" // X축 조깅
			pvalCommand = fmt.Sprintf("%.3f", step)
		case "y":
			pidCommand = "624,2,0,0" // Y축 조깅
			pvalCommand = fmt.Sprintf("%.3f", step)
		case "z":
			pidCommand = "624,3,0,0" // Z축 조깅
			pvalCommand = fmt.Sprintf("%.3f", step)
		case "rx":
			pidCommand = "624,4,0,0" // Rx 회전 조깅
			pvalCommand = fmt.Sprintf("%.3f", step)
		case "ry":
			pidCommand = "624,5,0,0" // Ry 회전 조깅
			pvalCommand = fmt.Sprintf("%.3f", step)
		case "rz":
			pidCommand = "624,6,0,0" // Rz 회전 조깅
			pvalCommand = fmt.Sprintf("%.3f", step)
		default:
			return nil, fmt.Errorf("지원하지 않는 카르테시안 축: %s", cmd.Axis)
		}
	default:
		return nil, fmt.Errorf("지원하지 않는 모드: %s", cmd.Mode)
	}

	form.Set("PID1", pidCommand)
	form.Set("PVal1", pvalCommand)

	return form, nil
}

// getRobotData fetches all robot data
func getRobotData() (*JogState, error) {
	res, err := http.Get("http://192.168.0.1/ROMDISK/web/Opr/jog/jogrefresh.asp")
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
	// * 디버깅용 로그 (필요시에만 활성화)
	// log.Printf("🔍 로봇 응답 데이터: %s", response)

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
	status := JogStatus{}
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

	return &JogState{
		Cartesian: cartesian,
		Joint:     joint,
		ToolData:  toolData,
		Status:    status,
	}, nil
}

// parseFloat parses string to float64 with error handling
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0.0, nil
	}
	var v float64
	_, err := fmt.Sscanf(s, "%f", &v)
	return v, err
}

// getRobotCoordinates returns robot coordinates for backward compatibility
func getRobotCoordinates() ([]float64, error) {
	data, err := getRobotData()
	if err != nil {
		return nil, err
	}
	return data.Joint, nil
}

// monitorRobotPosition periodically prints robot position
func monitorRobotPosition() {
	ticker := time.NewTicker(1 * time.Second) // 1초마다 확인
	defer ticker.Stop()

	var prevData *JogState // 이전 상태 저장용

	for range ticker.C {
		data, err := getRobotData()
		if err != nil {
			// * 에러 로그 (시간 포함)
			log.Printf("[%s] ❌ 좌표 읽기 실패: %v",
				time.Now().Format("15:04:05"), err)
			continue
		}

		// * 데이터 변경 감지 - 이전 상태와 비교
		if prevData == nil || hasDataChanged(prevData, data) {
			// * 시간 정보와 함께 상태 출력 (변경된 경우에만)
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] 🤖 JOG=(%.1f°, %.1f°, %.1f°) | XYZ=(%.1f, %.1f, %.1f) | 모드=%s | %s\n",
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

// getSafeValue safely gets array value with bounds checking
func getSafeValue(coords []float64, index int) float64 {
	if index < len(coords) {
		return coords[index]
	}
	return 0.0
}

// getCoordValue returns coordinate value for backward compatibility
func getCoordValue(coords []float64, index int) float64 {
	return getSafeValue(coords, index)
}

// sendJogCommand sends jog command to robot
func sendJogCommand(cmd JogCommand) (*JogResponse, error) {
	// 기본값 설정
	if cmd.Mode == "" {
		cmd.Mode = "joint"
	}
	if cmd.Step == 0 {
		cmd.Step = 1.0 // 기본 스텝
	}

	// * 디버깅용 로그 (명령 추적 + 시간)
	log.Printf("[%s] 🕹️  JOG 명령 수신: 모드=%s, 축=%s, 방향=%s, 스텝=%.3f",
		time.Now().Format("15:04:05.000"), cmd.Mode, cmd.Axis, cmd.Dir, cmd.Step)

	// JOG 명령을 로봇 프로토콜로 변환
	form, err := buildJogCommand(cmd)
	if err != nil {
		return &JogResponse{
			Success: false,
			Message: "명령 생성 실패: " + err.Error(),
			Command: "",
		}, err
	}

	// 로봇에 명령 전송
	resp, err := http.PostForm("http://192.168.0.1/wrtpdb", form)
	if err != nil {
		return &JogResponse{
			Success: false,
			Message: "로봇 통신 실패: " + err.Error(),
			Command: form.Encode(),
		}, err
	}
	defer resp.Body.Close()

	// 성공 응답
	response := &JogResponse{
		Success: true,
		Message: fmt.Sprintf("JOG 명령 성공: %s %s %s %.3f",
			cmd.Mode, cmd.Axis, cmd.Dir, cmd.Step),
		Command: form.Encode(),
	}

	// * 성공 메시지 (시간 포함)
	fmt.Printf("[%s] ✅ JOG 명령 전송 성공: %s\n",
		time.Now().Format("15:04:05.000"), response.Message)
	// * 디버깅용 로그 (명령 추적 + 시간)
	log.Printf("[%s] 🔗 전송된 명령: %s",
		time.Now().Format("15:04:05.000"), response.Command)

	return response, nil
}

// setRobotJogMode sends jog mode change command to robot
func setRobotJogMode(mode string) (*JogResponse, error) {
	form := url.Values{}
	form.Set("nPID", "2")
	form.Set("Redirect", "/ROMDISK/web/dbfunctions.asp")

	// 모드별 PID 설정 (원본 jogscripts.asp 참고)
	switch mode {
	case "computer":
		form.Set("PID1", "215,0,0,0")
		form.Set("PVal1", "0")
		form.Set("PID2", "621,0,0,0")
		form.Set("PVal2", "0")
	case "joint":
		form.Set("PID1", "215,1,0,0")
		form.Set("PVal1", "1")
		form.Set("PID2", "621,1,0,0")
		form.Set("PVal2", "1")
	case "world":
		form.Set("PID1", "215,1,0,0")
		form.Set("PVal1", "1")
		form.Set("PID2", "621,2,0,0")
		form.Set("PVal2", "2")
	case "tool":
		form.Set("PID1", "215,1,0,0")
		form.Set("PVal1", "1")
		form.Set("PID2", "621,3,0,0")
		form.Set("PVal2", "3")
	case "free":
		form.Set("PID1", "215,1,0,0")
		form.Set("PVal1", "1")
		form.Set("PID2", "621,4,0,0")
		form.Set("PVal2", "4")
	default:
		return &JogResponse{
			Success: false,
			Message: "지원하지 않는 모드: " + mode,
			Command: "",
		}, fmt.Errorf("unsupported mode: %s", mode)
	}

	// * 디버깅용 로그 (모드 변경 추적 + 시간)
	log.Printf("[%s] 🎮 JOG 모드 변경: %s",
		time.Now().Format("15:04:05.000"), mode)

	// 로봇에 명령 전송
	resp, err := http.PostForm("http://192.168.0.1/wrtpdb", form)
	if err != nil {
		return &JogResponse{
			Success: false,
			Message: "로봇 통신 실패: " + err.Error(),
			Command: form.Encode(),
		}, err
	}
	defer resp.Body.Close()

	response := &JogResponse{
		Success: true,
		Message: fmt.Sprintf("JOG 모드 변경 성공: %s", mode),
		Command: form.Encode(),
	}

	// * 성공 메시지 (시간 포함)
	fmt.Printf("[%s] ✅ JOG 모드 변경 성공: %s\n",
		time.Now().Format("15:04:05.000"), response.Message)
	return response, nil
}

// setRobotAxis sends axis selection command to robot
func setRobotAxis(axis int, robot int) (*JogResponse, error) {
	form := url.Values{}
	form.Set("nPID", "2")
	form.Set("Redirect", "/ROMDISK/web/dbfunctions.asp")

	// 축 선택 PID 설정 (원본 jogscripts.asp 참고)
	form.Set("PID1", "622,0,0,0")
	form.Set("PVal1", fmt.Sprintf("%d", axis))
	form.Set("PID2", "620,0,0,0")
	form.Set("PVal2", fmt.Sprintf("%d", robot))

	// * 디버깅용 로그 (축 선택 추적 + 시간)
	log.Printf("[%s] 🎯 축 선택: 축=%d, 로봇=%d",
		time.Now().Format("15:04:05.000"), axis, robot)

	// 로봇에 명령 전송
	resp, err := http.PostForm("http://192.168.0.1/wrtpdb", form)
	if err != nil {
		return &JogResponse{
			Success: false,
			Message: "로봇 통신 실패: " + err.Error(),
			Command: form.Encode(),
		}, err
	}
	defer resp.Body.Close()

	response := &JogResponse{
		Success: true,
		Message: fmt.Sprintf("축 선택 성공: 축=%d, 로봇=%d", axis, robot),
		Command: form.Encode(),
	}

	// * 성공 메시지 (시간 포함)
	fmt.Printf("[%s] ✅ 축 선택 성공: %s\n",
		time.Now().Format("15:04:05.000"), response.Message)
	return response, nil
}

// getJogModeText converts jog mode number to text
func getJogModeText(mode int) string {
	switch mode {
	case 0:
		return "Computer"
	case 1:
		return "Joint"
	case 2:
		return "World"
	case 3:
		return "Tool"
	case 4:
		return "Free"
	default:
		return fmt.Sprintf("Mode%d", mode)
	}
}

// getAxisText returns axis name based on mode and axis number
func getAxisText(jogMode int, axisNum int) string {
	if jogMode == 1 { // Joint mode
		switch axisNum {
		case 1:
			return "J1"
		case 2:
			return "J2"
		case 3:
			return "J3"
		case 4:
			return "J4"
		case 5:
			return "J5"
		case 6:
			return "J6"
		default:
			return fmt.Sprintf("J%d", axisNum)
		}
	} else { // Cartesian modes (World, Tool, etc.)
		switch axisNum {
		case 1:
			return "X"
		case 2:
			return "Y"
		case 3:
			return "Z"
		case 4:
			return "Rx"
		case 5:
			return "Ry"
		case 6:
			return "Rz"
		default:
			return fmt.Sprintf("Axis%d", axisNum)
		}
	}
}

// hasDataChanged compares two JogState structs to detect changes
func hasDataChanged(prev, current *JogState) bool {
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

// abs returns absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
