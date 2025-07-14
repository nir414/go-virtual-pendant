# Go Virtual Pendant

🤖 Go로 구현된 로봇 가상 펜던트 시스템

## 📁 프로젝트 구조

```
go-virtual-pendant/
├── cmd/
│   └── server/          # 메인 애플리케이션
│       └── main.go      # 서버 진입점
├── internal/            # 내부 라이브러리
│   ├── robot/          # 로봇 제어 관련
│   │   └── robot.go    # 로봇 통신 및 제어
│   ├── types/          # 타입 정의
│   │   └── types.go    # 공통 데이터 타입
│   └── web/            # 웹 서버 관련
│       └── handlers.go # 웹 핸들러
├── web/                # 웹 리소스
│   ├── static/         # 정적 파일 (CSS, JS)
│   │   ├── style.css
│   │   └── app.js
│   └── templates/      # HTML 템플릿
│       └── index.html
├── docs/               # 문서
│   ├── debug-guide.md
│   └── README_LOGGING.md
├── build/              # 빌드 출력
├── go.mod
├── go.sum
└── README.md
```

## 🚀 빌드 및 실행

### 개발 모드 실행
```bash
go run ./cmd/server
```

### 빌드
```bash
# 현재 플랫폼용 빌드
go build -o build/go-virtual-pendant ./cmd/server

# Windows용 빌드
GOOS=windows GOARCH=amd64 go build -o build/go-virtual-pendant.exe ./cmd/server
```

### VS Code 작업 실행
- `🚀 Go Run`: 개발 서버 실행
- `🏗️ Go Build`: 프로젝트 빌드
- `🧪 Go Test`: 테스트 실행

## 🌐 API 엔드포인트

### JOG 제어
- `POST /api/jog` - JOG 명령 전송
- `GET /api/jog/state` - 로봇 상태 조회
- `POST /api/jog/mode` - JOG 모드 변경
- `POST /api/jog/axis` - 축 선택

### 웹 인터페이스
- `GET /` - 웹 인터페이스
- `GET /static/*` - 정적 파일 (CSS, JS)

## 📖 문서

- [디버그 가이드](docs/debug-guide.md)
- [로깅 설정](docs/README_LOGGING.md)

## 🔧 환경 변수

- `LOG_LEVEL`: 로그 레벨 설정 (`INFO`, `DEBUG`, `VERBOSE`)

## 📦 의존성

- Go 1.24.4+
- 표준 라이브러리만 사용 (외부 의존성 없음)

## 🔌 로봇 연결

- 로봇 IP: `192.168.0.1`
- 포트: `8082` (웹 서버)
- 프로토콜: HTTP
