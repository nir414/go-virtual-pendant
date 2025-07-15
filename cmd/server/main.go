// ============================================================================
// cmd/server/main.go - Virtual Pendant 서버 애플리케이션
// ============================================================================
// Virtual Pendant API 서버의 메인 진입점입니다.
// 웹 서버 시작, API 엔드포인트 등록, 로봇 모니터링을 담당합니다.
// ============================================================================

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/nir414/go-virtual-pendant/internal/robot"
	"github.com/nir414/go-virtual-pendant/internal/types"
	"github.com/nir414/go-virtual-pendant/internal/web"
)

// ============================================================================
// 서버 설정 상수 (Server Configuration Constants)
// ============================================================================

const (
	// 서버 설정
	DEFAULT_PORT  = "8082"
	DEFAULT_HOST  = "localhost"
	API_BASE_PATH = "/api"
	STATIC_PATH   = "/static/"

	// API 엔드포인트
	ENDPOINT_JOG       = "/api/jog"
	ENDPOINT_JOG_STATE = "/api/jog/state"
	ENDPOINT_JOG_MODE  = "/api/jog/mode"
	ENDPOINT_JOG_AXIS  = "/api/jog/axis"

	// 메시지
	MSG_METHOD_NOT_ALLOWED = "Method Not Allowed"
	MSG_BAD_REQUEST        = "Bad Request"
	MSG_FETCH_STATE_FAILED = "Failed to fetch jog state"
	MSG_SET_MODE_FAILED    = "Failed to set jog mode"
	MSG_SET_AXIS_FAILED    = "Failed to set axis"
)

// ============================================================================
// API 핸들러 함수들 (API Handlers)
// ============================================================================

// jogHandler JOG 명령 요청 처리
func jogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var cmd types.JogCommand
	err := json.NewDecoder(r.Body).Decode(&cmd)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 로봇에 JOG 명령 전송
	response, err := robot.SendJogCommand(cmd)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if response.Success {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadGateway)
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// jogStateHandler 로봇 상태 조회 요청 처리
func jogStateHandler(w http.ResponseWriter, r *http.Request) {
	data, err := robot.GetRobotData()
	if err != nil {
		http.Error(w, MSG_FETCH_STATE_FAILED, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// setJogModeHandler JOG 모드 변경 요청 처리
func setJogModeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.SetJogModeRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	response, err := robot.SetRobotJogMode(req.Mode)
	if err != nil {
		http.Error(w, "Failed to set jog mode", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// setAxisHandler 축 선택 요청 처리
func setAxisHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.SetAxisRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	response, err := robot.SetRobotAxis(req.Axis, req.Robot)
	if err != nil {
		http.Error(w, "Failed to set axis", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ============================================================================
// 서버 관리 함수들 (Server Management)
// ============================================================================

// checkPortConflict 포트 사용 중인 프로세스 찾기
func checkPortConflict(port string) {
	fmt.Printf("❌ 포트 %s가 이미 사용 중입니다!\n", port)
	fmt.Println("🔍 포트 충돌 해결 방법:")

	if runtime.GOOS == "windows" {
		// Windows용 명령어
		fmt.Printf("   1️⃣  포트 사용 프로세스 확인: netstat -ano | findstr :%s\n", port)
		fmt.Println("   2️⃣  프로세스 종료: taskkill /PID <PID번호> /F")
		fmt.Println("   3️⃣  또는 다른 포트 사용을 원한다면 코드에서 포트 번호 변경")

		// 실제로 포트 사용 프로세스 찾기 시도
		fmt.Printf("\n🔎 포트 %s 사용 중인 프로세스 자동 검색:\n", port)
		cmd := exec.Command("netstat", "-ano")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			found := false
			for _, line := range lines {
				if strings.Contains(line, ":"+port) && strings.Contains(line, "LISTENING") {
					fmt.Printf("   📍 %s\n", strings.TrimSpace(line))
					// PID 추출
					fields := strings.Fields(line)
					if len(fields) >= 5 {
						pid := fields[len(fields)-1]
						fmt.Printf("   💡 해결 명령어: taskkill /PID %s /F\n", pid)
					}
					found = true
				}
			}
			if !found {
				fmt.Println("   ℹ️  포트 정보를 찾을 수 없습니다. 수동으로 확인해 주세요.")
			}
		}
	} else {
		// Linux/Mac용 명령어
		fmt.Printf("   1️⃣  포트 사용 프로세스 확인: lsof -i :%s\n", port)
		fmt.Println("   2️⃣  프로세스 종료: kill -9 <PID번호>")
		fmt.Println("   3️⃣  또는 다른 포트 사용을 원한다면 코드에서 포트 번호 변경")

		// 실제로 포트 사용 프로세스 찾기 시도
		fmt.Printf("\n🔎 포트 %s 사용 중인 프로세스 자동 검색:\n", port)
		cmd := exec.Command("lsof", "-i", ":"+port)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			fmt.Printf("   📍 %s\n", string(output))
		} else {
			fmt.Println("   ℹ️  포트 정보를 찾을 수 없습니다. 수동으로 확인해 주세요.")
		}
	}

	fmt.Println("\n⚡ 빠른 해결 방법:")
	fmt.Println("   • VS Code 터미널에서 위 명령어를 복사해서 실행하세요")
	fmt.Println("   • 또는 이 프로그램을 다시 실행해 보세요")
	fmt.Println()
}

// startServerWithErrorHandling 서버 시작 및 에러 처리
func startServerWithErrorHandling(port string) {
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		if strings.Contains(err.Error(), "bind") && strings.Contains(err.Error(), "address already in use") ||
			strings.Contains(err.Error(), "Only one usage of each socket address") {
			// 포트 충돌 오류
			checkPortConflict(port)
			os.Exit(1)
		} else {
			// 기타 서버 오류
			log.Fatalf("❌ 서버 시작 실패: %v", err)
		}
	}
}

// ============================================================================
// 메인 함수 (Main Function)
// ============================================================================

// main 서버 진입점 - 포트 8082에서 서버 실행
func main() {
	// 정적 파일 서빙 (CSS, JS)
	http.HandleFunc(STATIC_PATH, web.StaticFileHandler)

	// API 엔드포인트 등록
	http.HandleFunc(ENDPOINT_JOG, jogHandler)
	http.HandleFunc(ENDPOINT_JOG_STATE, jogStateHandler)
	http.HandleFunc(ENDPOINT_JOG_MODE, setJogModeHandler)
	http.HandleFunc(ENDPOINT_JOG_AXIS, setAxisHandler)

	// 웹 인터페이스 (템플릿 사용)
	http.HandleFunc("/", web.InterfaceHandler)

	// 서버 시작 메시지
	fmt.Printf("🚀 Virtual Pendant API running on http://%s:%s\n", DEFAULT_HOST, DEFAULT_PORT)
	fmt.Printf("🌐 웹 인터페이스: http://%s:%s\n", DEFAULT_HOST, DEFAULT_PORT)
	fmt.Println("📍 로봇 위치 모니터링 시작 (1초마다 간격)")

	// 로봇 위치 모니터링 고루틴 시작
	go robot.MonitorRobotPosition()

	// 서버 시작 - 포트 충돌 시 자동 해결 방법 안내
	startServerWithErrorHandling(DEFAULT_PORT)
}
