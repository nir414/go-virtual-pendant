// * SCARA 로봇팔 Virtual Pendant - JavaScript
// * HTML5 Konva.js를 사용한 로봇팔 시각화 및 제어

let currentJogMode = 'joint'; // * 전역 변수로 현재 모드 추적

// * SCARA 로봇팔 시각화 관련 변수
let stage, layer, robotArm;
let joint1Angle = 0;
let joint2Angle = 0;
let joint3Position = 0; // * Z축 위치
let joint4Angle = 0;    // * 엔드 이펙터 회전

// * SCARA 로봇팔 파라미터
// NOTE: 실제 로봇 사양에 맞게 조정 가능
const SCARA_PARAMS = {
	link1Length: 100,        // * 첫 번째 링크 길이
	link2Length: 100,        // * 두 번째 링크 길이
	link3Length: 100,        // * 세 번째 링크 길이
	baseRadius: 20,          // * 베이스 반지름
	jointRadius: 8,          // * 조인트 반지름
	endEffectorSize: 15,     // * 엔드 이펙터 크기
	workspaceRadius: 200,    // * 작업 공간 반지름
	centerX: 200,            // * 캔버스 중심 X
	centerY: 200             // * 캔버스 중심 Y
};

// * 로봇 시각화 초기화
function initRobotVisualization() {
	// * Konva 스테이지 생성
	stage = new Konva.Stage({
		container: 'robot-canvas',
		width: 400,
		height: 400
	});

	layer = new Konva.Layer();
	stage.add(layer);

	// * 작업 공간 원 그리기
	const workspace = new Konva.Circle({
		x: SCARA_PARAMS.centerX,
		y: SCARA_PARAMS.centerY,
		radius: SCARA_PARAMS.workspaceRadius,
		stroke: '#ddd',
		strokeWidth: 2,
		dash: [5, 5]
	});
	layer.add(workspace);

	// * 좌표계 표시
	drawCoordinateSystem();

	// * 로봇팔 초기화
	createRobotArm();
	updateRobotVisualization();
}

// * 좌표계 표시 함수
function drawCoordinateSystem() {
	const centerX = SCARA_PARAMS.centerX;
	const centerY = SCARA_PARAMS.centerY;
	const axisLength = 50;

	// * X축 (빨간색)
	const xAxis = new Konva.Line({
		points: [centerX, centerY, centerX + axisLength, centerY],
		stroke: 'red',
		strokeWidth: 2
	});
	layer.add(xAxis);

	// * Y축 (초록색)
	const yAxis = new Konva.Line({
		points: [centerX, centerY, centerX, centerY - axisLength],
		stroke: 'green',
		strokeWidth: 2
	});
	layer.add(yAxis);

	// * 축 레이블
	const xLabel = new Konva.Text({
		x: centerX + axisLength + 5,
		y: centerY - 10,
		text: 'X',
		fontSize: 14,
		fill: 'red'
	});
	layer.add(xLabel);

	const yLabel = new Konva.Text({
		x: centerX + 5,
		y: centerY - axisLength - 15,
		text: 'Y',
		fontSize: 14,
		fill: 'green'
	});
	layer.add(yLabel);
}

// * SCARA 로봇팔 구성 요소 생성
// NOTE: Konva.Group에 모든 로봇 부품들을 추가
function createRobotArm() {
	robotArm = new Konva.Group();

	// * 베이스 (고정부) - 로봇의 기초 플랫폼
	const base = new Konva.Circle({
		x: SCARA_PARAMS.centerX,
		y: SCARA_PARAMS.centerY,
		radius: SCARA_PARAMS.baseRadius,
		fill: '#333',
		stroke: '#000',
		strokeWidth: 2
	});
	robotArm.add(base); // * 인덱스 0

	// * 링크 1 (첫 번째 팔) - Joint1에서 Joint2까지 연결
	const link1 = new Konva.Line({
		points: [0, 0, SCARA_PARAMS.link1Length, 0], // * 초기 위치 (수평)
		stroke: '#4CAF50',    // * 초록색으로 구분
		strokeWidth: 8,
		lineCap: 'round'
	});
	robotArm.add(link1); // * 인덱스 1

	// * 조인트 1 (첫 번째 관절) - 베이스 중심에서 회전
	const joint1 = new Konva.Circle({
		x: SCARA_PARAMS.centerX,
		y: SCARA_PARAMS.centerY,
		radius: SCARA_PARAMS.jointRadius,
		fill: '#2196F3',      // * 파란색 조인트
		stroke: '#1976D2',
		strokeWidth: 2
	});
	robotArm.add(joint1); // * 인덱스 2

	// * 링크 2 (두 번째 팔) - Joint2에서 엔드 이펙터까지 연결
	const link2 = new Konva.Line({
		points: [0, 0, SCARA_PARAMS.link2Length, 0], // * 초기 위치 (수평)
		stroke: '#FF9800',    // * 주황색으로 구분
		strokeWidth: 6,
		lineCap: 'round'
	});
	robotArm.add(link2); // * 인덱스 3

	// * 조인트 2 (두 번째 관절) - Link1 끝에서 회전
	const joint2 = new Konva.Circle({
		radius: SCARA_PARAMS.jointRadius - 2,
		fill: '#2196F3',      // * 파란색 조인트 (작게)
		stroke: '#1976D2',
		strokeWidth: 2
	});
	robotArm.add(joint2); // * 인덱스 4

	// * 엔드 이펙터 (작업 도구) - 실제 작업을 수행하는 부분
	const endEffector = new Konva.RegularPolygon({
		sides: 3,             // * 삼각형 모양
		radius: SCARA_PARAMS.endEffectorSize,
		fill: '#F44336',      // * 빨간색으로 구분
		stroke: '#D32F2F',
		strokeWidth: 2
	});
	robotArm.add(endEffector); // * 인덱스 5

	layer.add(robotArm);
}

// * 로봇팔 위치 업데이트 - 실제 조인트 각도에 따라 시각화
// NOTE: 이 함수가 조인트 각도를 읽어서 로봇팔을 그리는 핵심 함수
function updateRobotVisualization() {
	if (!robotArm) return;

	// * 로봇 구성 요소 참조 가져오기 (createRobotArm에서 추가한 순서)
	const children = robotArm.children;
	const base = children[0];         // * 베이스 (인덱스 0)
	const link1 = children[1];        // * 링크 1 (인덱스 1)
	const joint1 = children[2];       // * 조인트 1 (인덱스 2)
	const link2 = children[3];        // * 링크 2 (인덱스 3)
	const joint2 = children[4];       // * 조인트 2 (인덱스 4)
	const endEffector = children[5];  // * 엔드 이펙터 (인덱스 5)

	// * === SCARA 운동학 계산 ===
	// NOTE: 조인트 각도로부터 각 링크의 끝점 위치를 계산

	// * 링크 1 끝점 위치 계산 (Joint1 회전에 의해 결정)
	const link1EndX = SCARA_PARAMS.centerX + SCARA_PARAMS.link1Length * Math.cos(joint1Angle);
	const link1EndY = SCARA_PARAMS.centerY - SCARA_PARAMS.link1Length * Math.sin(joint1Angle);  // * Y축 반전 (캔버스→로봇 좌표계)

	// * 링크 2 끝점 위치 계산 (Joint1 + Joint2 회전에 의해 결정)
	const totalAngle = joint1Angle + joint2Angle; // * Joint2는 Joint1에 상대적
	const link2EndX = link1EndX + SCARA_PARAMS.link2Length * Math.cos(totalAngle);
	const link2EndY = link1EndY - SCARA_PARAMS.link2Length * Math.sin(totalAngle);  // * Y축 반전 (캔버스→로봇 좌표계)

	// * === 시각적 요소 업데이트 ===

	// * 링크 1 선분 업데이트 (베이스 중심 → 링크1 끝점)
	link1.points([
		SCARA_PARAMS.centerX, SCARA_PARAMS.centerY,  // * 시작점: 베이스 중심
		link1EndX, link1EndY                         // * 끝점: 링크1 끝
	]);

	// * 링크 2 선분 업데이트 (링크1 끝점 → 링크2 끝점)
	link2.points([
		link1EndX, link1EndY,                        // * 시작점: 링크1 끝
		link2EndX, link2EndY                         // * 끝점: 엔드 이펙터 위치
	]);

	// * 조인트 2 위치 업데이트 (링크1과 링크2 연결점)
	joint2.x(link1EndX);
	joint2.y(link1EndY);

	// * 엔드 이펙터 위치 및 회전 업데이트
	endEffector.x(link2EndX);                        // * X 위치
	endEffector.y(link2EndY);                        // * Y 위치
	endEffector.rotation(joint4Angle * 180 / Math.PI); // * Joint4 회전 (라디안 → 도)

	// * 화면에 변경사항 반영
	layer.draw();

	// * 현재 위치 정보 UI 업데이트
	updateRobotInfo(link2EndX, link2EndY);
}

// * 로봇 정보 UI 업데이트 - 각도와 위치를 화면에 표시
// NOTE: 엔드 이펙터의 실제 좌표를 계산하여 표시
function updateRobotInfo(endX, endY) {
	// * 캔버스 좌표계에서 실제 로봇 좌표계로 변환
	const actualX = endX - SCARA_PARAMS.centerX;      // * 중심점 기준 X 좌표
	const actualY = SCARA_PARAMS.centerY - endY;      // * Y축 반전 (위쪽이 +Y)

	// * 정보 표시 영역 생성 또는 가져오기
	let infoDiv = document.getElementById('robot-info');
	if (!infoDiv) {
		infoDiv = document.createElement('div');
		infoDiv.id = 'robot-info';
		infoDiv.className = 'robot-info';
		document.getElementById('robot-canvas-container').appendChild(infoDiv);
	}

	// * 조인트 각도 및 위치 정보 표시
	infoDiv.innerHTML = `
		<div class="joint-info">J1: ${(joint1Angle * 180 / Math.PI).toFixed(1)}°</div>
		<div class="joint-info">J2: ${(joint2Angle * 180 / Math.PI).toFixed(1)}°</div>
		<div class="joint-info">Z: ${joint3Position.toFixed(1)}mm</div>
		<div class="joint-info">R: ${(joint4Angle * 180 / Math.PI).toFixed(1)}°</div>
		<br>
		<div class="joint-info">X: ${actualX.toFixed(1)}mm</div>
		<div class="joint-info">Y: ${actualY.toFixed(1)}mm</div>
	`;
}

// * 조인트 각도 업데이트 - 서버에서 받은 데이터로 로봇팔 업데이트
// NOTE: 이 함수가 실제 로봇 데이터를 받아서 시각화를 업데이트하는 핵심!
function updateJointAngles(jointValues) {
	if (jointValues && jointValues.length >= 4) {
		// * 서버에서 받은 각도 데이터를 라디안으로 변환 (도 → 라디안)
		joint1Angle = jointValues[0] * Math.PI / 180;  // * Joint 1 각도
		joint2Angle = jointValues[1] * Math.PI / 180;  // * Joint 2 각도
		joint3Position = jointValues[2] || 0;          // * Z축 위치 (직선)
		joint4Angle = jointValues[3] * Math.PI / 180;  // * Joint 4 회전 (엔드 이펙터)

		// * 새로운 각도로 로봇팔 시각화 업데이트
		updateRobotVisualization();
	}
}


// Get current jog mode from select
function getSelectedMode() {
	return document.getElementById('modeSelect').value;
}

// Handle mode change from select
function setJogModeButton(mode) {
	// Update current mode
	currentJogMode = mode;
	// Send mode change to robot
	setJogMode(mode);
	// Update axis options
	updateAxisOptions();
}

function updateAxisOptions() {
	const axisSelect = document.getElementById('axisSelect');
	const mode = getSelectedMode();

	if (mode === 'joint') {
		axisSelect.innerHTML =
			'<option value="joint1">Joint 1</option>' +
			'<option value="joint2">Joint 2</option>' +
			'<option value="joint3">Joint 3</option>' +
			'<option value="joint4">Joint 4</option>' +
			'<option value="joint5">Joint 5</option>' +
			'<option value="joint6">Joint 6</option>';
	} else {
		axisSelect.innerHTML =
			'<option value="x">X axis</option>' +
			'<option value="y">Y axis</option>' +
			'<option value="z">Z axis</option>' +
			'<option value="rx">Rx rotation</option>' +
			'<option value="ry">Ry rotation</option>' +
			'<option value="rz">Rz rotation</option>';
	}

	// Send initial axis selection after updating options
	jogListChanged();
}

function setJogSpeedValue(speed) {
	document.getElementById('jogSpeed').value = speed;
	setJogSpeed();
}

function setJogSpeed() {
	const speed = document.getElementById('jogSpeed').value;
	console.log('조깅 속도 설정:', speed + '%');
	// 실제 로봇 속도 설정 구현 가능
}

function getSelectedAxis() {
	const select = document.getElementById('axisSelect');
	return select.value;
}

function jogListChanged() {
	const selectedAxis = getSelectedAxis();
	// Remove deprecated selectedAxisSpan update
	// const selectedAxisSpan = document.getElementById('selectedAxis');
	const mode = getSelectedMode();

	// 선택된 축 이름 표시 업데이트
	const axisNames = {
		'joint1': 'Joint 1', 'joint2': 'Joint 2', 'joint3': 'Joint 3',
		'joint4': 'Joint 4', 'joint5': 'Joint 5', 'joint6': 'Joint 6',
		'x': 'X축', 'y': 'Y축', 'z': 'Z축',
		'rx': 'Rx 회전', 'ry': 'Ry 회전', 'rz': 'Rz 회전'
	};

	// No UI update here; position panel shows current-axis

	// 축 번호 계산
	let axisNumber = 1;
	if (mode === 'joint') {
		axisNumber = parseInt(selectedAxis.replace('joint', ''));
	} else {
		const cartesianMap = { 'x': 1, 'y': 2, 'z': 3, 'rx': 4, 'ry': 5, 'rz': 6 };
		axisNumber = cartesianMap[selectedAxis] || 1;
	}

	// 로봇에 축 선택 전송
	fetch('/api/jog/axis', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			axis: axisNumber,
			robot: 1
		})
	})
		.then(response => response.json())
		.then(data => {
			console.log('축 선택 응답:', data);
			if (!data.success) {
				document.getElementById('status').textContent = '❌ 축 선택 실패: ' + data.message;
				document.getElementById('status').style.background = '#f8d7da';
			}
		})
		.catch(error => {
			console.error('축 선택 오류:', error);
		});

	console.log('선택된 축:', selectedAxis, '축 번호:', axisNumber);
}

function setJogMode(mode) {
	// 로봇에 모드 변경 전송
	fetch('/api/jog/mode', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ mode: mode })
	})
		.then(response => response.json())
		.then(data => {
			console.log('모드 변경 응답:', data);
			if (data.success) {
				document.getElementById('status').textContent = '✅ ' + data.message;
				document.getElementById('status').style.background = '#d4edda';
			} else {
				document.getElementById('status').textContent = '❌ 모드 변경 실패: ' + data.message;
				document.getElementById('status').style.background = '#f8d7da';
			}
		})
		.catch(error => {
			console.error('모드 변경 오류:', error);
			document.getElementById('status').textContent = '❌ 모드 변경 통신 오류';
			document.getElementById('status').style.background = '#f8d7da';
		});
}

function sendSelectedAxisJog(direction) {
	const mode = getSelectedMode();
	const axis = getSelectedAxis();
	const step = parseFloat(document.getElementById('stepSize').value);

	sendJog(axis, direction);
}

function sendJog(axis, direction) {
	const currentTime = performance.now();
	const timeSinceLastJog = lastJogTime > 0 ? currentTime - lastJogTime : 0;
	lastJogTime = currentTime;
	jogCommandCount++;

	const mode = getSelectedMode();
	const step = parseFloat(document.getElementById('stepSize').value);

	// Validate step size
	const stepInput = parseFloat(document.getElementById('stepSize').value);
	if (isNaN(stepInput) || stepInput < 0.1 || stepInput > 10) {
		document.getElementById('status').textContent = '❌ 잘못된 스텝 크기: ' + stepInput;
		document.getElementById('status').style.background = '#f8d7da';
		return;
	}
	const command = {
		axis: axis,
		dir: direction,
		step: stepInput,
		mode: mode
	};

	// 디버깅: 상세한 명령 전송 로그
	console.log('📡 JOG 명령 전송:', {
		command: command,
		timestamp: new Date().toLocaleTimeString() + '.' + (currentTime % 1000).toFixed(0).padStart(3, '0'),
		intervalSinceLastJog: timeSinceLastJog.toFixed(1) + 'ms',
		commandNumber: jogCommandCount,
		isJogging: isJogging
	});

	// 상태 표시는 연속 조깅 중에는 스킵 (성능 향상)
	if (!isJogging) {
		document.getElementById('status').textContent = '명령 전송 중... (' + mode + ' 모드, ' + axis + ', ' + direction + ')';
	}

	// 서버 응답 시간 측정
	const fetchStartTime = performance.now();

	// 서버 응답을 기다리지 않고 즉시 전송 (Fire and Forget 방식)
	fetch('/api/jog', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(command)
	})
		.then(response => {
			if (!response.ok) {
				throw new Error('HTTP ' + response.status + ' ' + response.statusText);
			}
			return response.json();
		})
		.then(data => {
			const responseTime = performance.now() - fetchStartTime;

			console.log('✅ JOG 응답:', {
				response: data,
				responseTime: responseTime.toFixed(1) + 'ms',
				commandNumber: jogCommandCount,
				timestamp: new Date().toLocaleTimeString() + '.' + (performance.now() % 1000).toFixed(0).padStart(3, '0')
			});

			// 연속 조깅 중이 아닐 때만 상태 업데이트
			if (!isJogging) {
				if (data.success) {
					document.getElementById('status').textContent = '✅ ' + data.message;
					document.getElementById('status').style.background = '#d4edda';
				} else {
					document.getElementById('status').textContent = '❌ ' + data.message;
					document.getElementById('status').style.background = '#f8d7da';
				}
			}
		})
		.catch(error => {
			const responseTime = performance.now() - fetchStartTime;

			console.error('❌ JOG 통신 오류:', {
				error: error,
				responseTime: responseTime.toFixed(1) + 'ms',
				commandNumber: jogCommandCount,
				timestamp: new Date().toLocaleTimeString() + '.' + (performance.now() % 1000).toFixed(0).padStart(3, '0')
			});

			// 연속 조깅 중이 아닐 때만 에러 표시
			if (!isJogging) {
				document.getElementById('status').textContent = '❌ 통신 오류: ' + error;
				document.getElementById('status').style.background = '#f8d7da';
			}
		});
}

function updatePosition() {
	fetch('/api/jog/state')
		.then(response => response.json())
		.then(data => {
			// 위치 정보 업데이트
			let coordsText = '';
			// Extract properties from JSON response
			const joints = data.joint;
			const carts = data.cartesian;
			coordsText += '🦾 조인트: ' + joints.map((v, i) => 'J' + (i + 1) + '=' + v.toFixed(3) + '°').join(', ') + '\n';
			coordsText += '📐 카르테시안: X=' + carts[0].toFixed(3) + ', Y=' + carts[1].toFixed(3) + ', Z=' + carts[2].toFixed(3) + '\n';
			coordsText += '🔄 회전: Rx=' + carts[3].toFixed(3) + '°, Ry=' + carts[4].toFixed(3) + '°, Rz=' + carts[5].toFixed(3) + '°\n';
			const stat = data.status;
			coordsText += '⚙️  상태: 축수=' + stat.axis_count + ', 조깅=' + stat.allow_jog + ', 모드=' + stat.jog_mode;

			document.getElementById('coordinates').textContent = coordsText;

			// 로봇팔 시각화 업데이트
			updateJointAngles(data.joint);

			// 실시간 상태 정보 업데이트
			document.getElementById('current-jog-mode').textContent = data.status.jog_mode_text + ' (' + data.status.jog_mode + ')';
			document.getElementById('current-axis').textContent = data.status.selected_axis_text + ' (축' + data.status.selected_axis + ')';
			document.getElementById('power-state').textContent = data.status.power_state;
			document.getElementById('axis-count').textContent = data.status.axis_count;
			document.getElementById('allow-jog').textContent = data.status.allow_jog ? '허용' : '금지';
			document.getElementById('error-desc').textContent = data.status.error_desc || '없음';

			// 상태에 따른 색상 변경
			const jogModeElement = document.getElementById('current-jog-mode');
			const allowJogElement = document.getElementById('allow-jog');

			if (data.status.allow_jog) {
				allowJogElement.style.color = '#28a745';
				allowJogElement.style.fontWeight = 'bold';
			} else {
				allowJogElement.style.color = '#dc3545';
				allowJogElement.style.fontWeight = 'bold';
			}

			// JOG 모드에 따른 색상
			switch (data.status.jog_mode) {
				case 1:
					jogModeElement.style.color = '#007bff'; // Joint - 파란색
					break;
				case 2:
					jogModeElement.style.color = '#28a745'; // World - 초록색
					break;
				case 3:
					jogModeElement.style.color = '#fd7e14'; // Tool - 주황색
					break;
				default:
					jogModeElement.style.color = '#6c757d'; // 기본 - 회색
			}
		})
		.catch(error => {
			console.error('위치 정보 업데이트 실패:', error);
			document.getElementById('coordinates').textContent = '❌ 위치 정보 로딩 실패: ' + error;
			document.getElementById('current-jog-mode').textContent = '연결 오류';
			document.getElementById('current-axis').textContent = '연결 오류';
		});
}

// 🔍 네트워크 신호 캡처용 Fetch 인터셉터 추가
(function () {
	const originalFetch = window.fetch;
	window.fetch = async function (input, init) {
		console.log('[Intercepted Request]', input, init);
		const startTime = performance.now();
		try {
			const response = await originalFetch(input, init);
			const elapsed = (performance.now() - startTime).toFixed(1);
			let cloned = response.clone();
			let payload;
			try {
				payload = await cloned.json();
			} catch (_) {
				payload = await cloned.text();
			}
			console.log('[Intercepted Response]', input, payload, `(${elapsed}ms)`);
			return response;
		} catch (error) {
			console.error('[Fetch Error]', input, error);
			throw error;
		}
	};
})();

// * 연속 조깅을 위한 변수들
let continuousJogInterval = null;
let isJogging = false;
let keyBusy = false;  // 키 중복 방지를 위한 플래그 (원본 방식)

// 성능 측정 변수들
let jogStartTime = 0;
let jogCommandCount = 0;
let lastJogTime = 0;

// * 연속 조깅 시작 함수 - 원본 방식 개선
function startContinuousJog(direction) {
	const currentTime = performance.now();
	jogStartTime = currentTime;
	jogCommandCount = 0;

	console.log('🚀 연속 조깅 시작:', {
		direction: direction,
		startTime: new Date().toLocaleTimeString() + '.' + (currentTime % 1000).toFixed(0).padStart(3, '0'),
		timestamp: currentTime
	});

	// 이미 조깅 중이면 먼저 중단
	if (isJogging) {
		console.log('⚠️  이미 조깅 중 - 기존 조깅 중단');
		stopContinuousJog();
	}

	isJogging = true;

	// 즉시 첫 번째 조깅 실행 (딜레이 없이)
	sendSelectedAxisJog(direction);

	// 연속 조깅을 위한 인터벌 시작 (30ms 간격으로 더 빠른 반응)
	continuousJogInterval = setInterval(() => {
		if (isJogging) {
			sendSelectedAxisJog(direction);
		}
	}, 30);

	console.log('⏱️  연속 조깅 인터벌 시작 (30ms)');
}

// * 연속 조깅 중단 함수 - 원본 방식 개선
function stopContinuousJog() {
	const currentTime = performance.now();
	const duration = currentTime - jogStartTime;
	const avgInterval = jogCommandCount > 0 ? duration / jogCommandCount : 0;

	console.log('🛑 연속 조깅 중단:', {
		duration: duration.toFixed(1) + 'ms',
		commandCount: jogCommandCount,
		avgInterval: avgInterval.toFixed(1) + 'ms',
		expectedInterval: '30ms',
		performance: (avgInterval / 30 * 100).toFixed(1) + '%'
	});

	isJogging = false;
	keyBusy = false;  // 키 잠금 해제

	if (continuousJogInterval) {
		clearInterval(continuousJogInterval);
		continuousJogInterval = null;
		console.log('⏹️  연속 조깅 인터벌 정리 완료');
	}

	// 조깅 중단 명령을 서버에 전송 (원본 방식과 유사)
	sendJogStop();

	// 연속 조깅이 끝났을 때 상태 업데이트
	document.getElementById('status').textContent = '대기 중...';
	document.getElementById('status').style.background = '';
}

// * 조깅 중단 명령 전송 함수 (원본의 jog(0) 방식)
function sendJogStop() {
	const currentTime = performance.now();
	const mode = getSelectedMode();
	const axis = getSelectedAxis();

	const stopCommand = {
		axis: axis,
		dir: 'stop',      // 중단 신호
		step: 0,          // 스텝 0
		mode: mode
	};

	console.log('🛑 조깅 중단 명령 전송:', {
		command: stopCommand,
		timestamp: new Date().toLocaleTimeString() + '.' + (currentTime % 1000).toFixed(0).padStart(3, '0'),
		totalJogDuration: (currentTime - jogStartTime).toFixed(1) + 'ms'
	});

	const fetchStartTime = performance.now();

	// 중단 명령은 즉시 전송 (우선순위 높음)
	fetch('/api/jog', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(stopCommand)
	})
		.then(response => response.json())
		.then(data => {
			const responseTime = performance.now() - fetchStartTime;

			console.log('✅ 조깅 중단 응답:', {
				response: data,
				responseTime: responseTime.toFixed(1) + 'ms',
				timestamp: new Date().toLocaleTimeString() + '.' + (performance.now() % 1000).toFixed(0).padStart(3, '0')
			});
		})
		.catch(error => {
			const responseTime = performance.now() - fetchStartTime;

			console.error('❌ 조깅 중단 명령 오류:', {
				error: error,
				responseTime: responseTime.toFixed(1) + 'ms',
				timestamp: new Date().toLocaleTimeString() + '.' + (performance.now() % 1000).toFixed(0).padStart(3, '0')
			});
		});
}

// ...existing code...
// (함수 simulateJointMove 등 나머지 함수 및 이벤트 핸들러 포함)

// 키보드 단축키 지원 - 원본 방식 개선
document.addEventListener('keydown', function (event) {
	if (event.ctrlKey) return; // Ctrl 키가 눌려있으면 무시

	// 텍스트 입력 필드에 포커스가 있는지 확인
	const activeElement = document.activeElement;
	const isInputFocused = activeElement && (
		activeElement.tagName === 'INPUT' ||
		activeElement.tagName === 'TEXTAREA' ||
		activeElement.contentEditable === 'true'
	);

	// 텍스트 입력 중일 때는 숫자키 단축키 비활성화
	if (isInputFocused && /^[0-9]$/.test(event.key)) {
		return; // 숫자키는 텍스트 입력에 우선권 부여
	}

	switch (event.key) {
		case 'ArrowLeft':
		case '-':
			// 텍스트 입력 중이 아닐 때만 조깅 명령 실행
			if (!isInputFocused && keyBusy === false) {
				keyBusy = true;  // 키 잠금 (원본 방식)
				event.preventDefault();
				startContinuousJog('negative');
			}
			break;
		case 'ArrowRight':
		case '+':
		case '=':
			// 텍스트 입력 중이 아닐 때만 조깅 명령 실행
			if (!isInputFocused && keyBusy === false) {
				keyBusy = true;  // 키 잠금 (원본 방식)
				event.preventDefault();
				startContinuousJog('positive');
			}
			break;
		case '1':
		case '2':
		case '3':
		case '4':
		case '5':
		case '6':
			// 텍스트 입력 중이 아닐 때만 조인트 선택 실행
			if (!isInputFocused) {
				event.preventDefault();
				const jointNum = parseInt(event.key);
				document.getElementById('axisSelect').value = 'joint' + jointNum;
				jogListChanged();
			}
			break;
	}
});

// 키보드 키를 뗄 때 연속 조깅 중단 - 원본 방식
document.addEventListener('keyup', function (event) {
	switch (event.key) {
		case 'ArrowLeft':
		case '-':
		case 'ArrowRight':
		case '+':
		case '=':
			stopContinuousJog();  // 즉시 중단
			break;
	}
});

// 마우스 휠을 이용한 조인트 제어 - 연속 조깅으로 변경
let wheelTimeout = null;
document.getElementById('robot-canvas').addEventListener('wheel', function (event) {
	event.preventDefault();

	const direction = event.deltaY > 0 ? 'negative' : 'positive';

	// 기존 휠 타이머 제거
	if (wheelTimeout) {
		clearTimeout(wheelTimeout);
	}

	// 짧은 연속 조깅 시작
	startContinuousJog(direction);

	// 짧은 시간 후 자동으로 중단 (150ms로 단축)
	wheelTimeout = setTimeout(() => {
		stopContinuousJog();
	}, 150);
});
