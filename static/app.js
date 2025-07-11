let currentJogMode = 'joint'; // 전역 변수로 현재 모드 추적

function getSelectedMode() {
	return currentJogMode;
}

function setJogModeButton(mode) {
	// 모든 버튼에서 active 클래스 제거
	document.querySelectorAll('.mode-btn').forEach(btn => btn.classList.remove('active'));

	// 선택된 버튼에 active 클래스 추가
	document.getElementById('btn-' + mode).classList.add('active');

	// 현재 모드 업데이트
	currentJogMode = mode;

	// 로봇에 모드 변경 전송
	setJogMode(mode);

	// UI 업데이트
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

	// 선택된 축 표시 업데이트 및 로봇에 전송
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
	const selectedAxisSpan = document.getElementById('selectedAxis');
	const mode = getSelectedMode();

	// 선택된 축 이름 표시 업데이트
	const axisNames = {
		'joint1': 'Joint 1', 'joint2': 'Joint 2', 'joint3': 'Joint 3',
		'joint4': 'Joint 4', 'joint5': 'Joint 5', 'joint6': 'Joint 6',
		'x': 'X축', 'y': 'Y축', 'z': 'Z축',
		'rx': 'Rx 회전', 'ry': 'Ry 회전', 'rz': 'Rz 회전'
	};

	selectedAxisSpan.textContent = axisNames[selectedAxis] || selectedAxis;

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
	const mode = getSelectedMode();
	const step = parseFloat(document.getElementById('stepSize').value);

	const command = {
		axis: axis,
		dir: direction,
		step: step,
		mode: mode
	};

	document.getElementById('status').textContent = '명령 전송 중... (' + mode + ' 모드, ' + axis + ', ' + direction + ')';

	fetch('/api/jog', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(command)
	})
		.then(response => response.json())
		.then(data => {
			if (data.success) {
				document.getElementById('status').textContent = '✅ ' + data.message;
				document.getElementById('status').style.background = '#d4edda';
			} else {
				document.getElementById('status').textContent = '❌ ' + data.message;
				document.getElementById('status').style.background = '#f8d7da';
			}
			setTimeout(updatePosition, 500); // 0.5초 후 위치 업데이트
		})
		.catch(error => {
			document.getElementById('status').textContent = '❌ 통신 오류: ' + error;
			document.getElementById('status').style.background = '#f8d7da';
		});
}

function updatePosition() {
	fetch('/api/jog/state')
		.then(response => response.json())
		.then(data => {
			// 위치 정보 업데이트
			let coordsText = '';
			coordsText += '🦾 조인트: ' + data.joint.map((v, i) => 'J' + (i + 1) + '=' + v.toFixed(3) + '°').join(', ') + '\n';
			coordsText += '📐 카르테시안: X=' + data.cartesian[0].toFixed(3) + ', Y=' + data.cartesian[1].toFixed(3) + ', Z=' + data.cartesian[2].toFixed(3) + '\n';
			coordsText += '🔄 회전: Rx=' + data.cartesian[3].toFixed(3) + '°, Ry=' + data.cartesian[4].toFixed(3) + '°, Rz=' + data.cartesian[5].toFixed(3) + '°\n';
			coordsText += '⚙️  상태: 축수=' + data.status.axis_count + ', 조깅=' + data.status.allow_jog + ', 모드=' + data.status.jog_mode;

			document.getElementById('coordinates').textContent = coordsText;

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
			document.getElementById('coordinates').textContent = '❌ 위치 정보 로딩 실패: ' + error;
			document.getElementById('current-jog-mode').textContent = '연결 오류';
			document.getElementById('current-axis').textContent = '연결 오류';
		});
}

// 페이지 로드 시 초기화
document.addEventListener('DOMContentLoaded', function () {
	// 초기 UI 설정
	updateAxisOptions();

	// 페이지 로드 시 위치 정보 업데이트
	updatePosition();

	// 2초마다 자동 업데이트 (실시간 모니터링)
	setInterval(updatePosition, 2000);
});
