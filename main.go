// * go-virtual-pendant/main.go
// * Virtual Pendant API 서버 메인 파일
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
)

// * jogHandler handles jog command requests
// NOTE: POST 방식으로만 JOG 명령을 처리합니다
func jogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var cmd JogCommand
	err := json.NewDecoder(r.Body).Decode(&cmd)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// * 로봇에 JOG 명령 전송
	response, err := sendJogCommand(cmd)
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

// * jogStateHandler handles jog state requests
// NOTE: 로봇의 현재 상태(위치, 각도)를 조회합니다
func jogStateHandler(w http.ResponseWriter, r *http.Request) {
	data, err := getRobotData()
	if err != nil {
		http.Error(w, "Failed to fetch jog state", http.StatusBadGateway)
		return
	}

	// * 로봇 데이터를 서버 로그에 출력 (필요시에만 활성화)
	// log.Printf("🤖 API 요청 - 조인트: %v, 카르테시안: %v", data.Joint, data.Cartesian)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// // webInterfaceHandler는 이제 web.go에서 처리됩니다 (이 함수 삭제)
// // func webInterfaceHandler...

// * setJogModeHandler handles jog mode change requests
// NOTE: 지원 모드: "computer", "joint", "world", "tool", "free"
func setJogModeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Mode string `json:"mode"` // * "computer", "joint", "world", "tool", "free"
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	response, err := setRobotJogMode(req.Mode)
	if err != nil {
		http.Error(w, "Failed to set jog mode", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// * setAxisHandler handles axis selection requests
// NOTE: 1-6 조인트 또는 1-6 카르테시안 축 선택
func setAxisHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Axis  int `json:"axis"`  // * 1-6 for joints, 1-6 for cartesian
		Robot int `json:"robot"` // * robot number (usually 1)
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	response, err := setRobotAxis(req.Axis, req.Robot)
	if err != nil {
		http.Error(w, "Failed to set axis", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

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

// * main function - server entry point
// ! 포트 8082에서 서버 실행
func main() {
	// * 정적 파일 서빙 (CSS, JS)
	http.HandleFunc("/static/", staticFileHandler)

	// * API 엔드포인트 등록
	http.HandleFunc("/api/jog", jogHandler)
	http.HandleFunc("/api/jog/state", jogStateHandler)
	http.HandleFunc("/api/jog/mode", setJogModeHandler)
	http.HandleFunc("/api/jog/axis", setAxisHandler)

	// * 웹 인터페이스 (템플릿 사용)
	http.HandleFunc("/", webInterfaceHandler)

	// * 서버 시작 메시지 (표준 출력으로 깔끔하게)
	fmt.Println("🚀 Virtual Pendant API running on http://localhost:8082")
	fmt.Println("🌐 웹 인터페이스: http://localhost:8082")
	fmt.Println("📍 로봇 위치 모니터링 시작 (1초마다 간격)")

	// * 로봇 위치 모니터링 고루틴 시작
	go monitorRobotPosition()

	// ! 서버 시작 - 포트 충돌 시 자동 해결 방법 안내
	startServerWithErrorHandling("8082")
}
