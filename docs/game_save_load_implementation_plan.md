# 게임 상태 저장 및 로드 기능 구현 계획서

## 1. 개요

### 1.1 목표
- 현재 실행 중인 게임 상태를 JSON 포맷으로 저장
- 저장된 게임 상태를 로드하여 중단된 지점부터 게임 재개
- 멀티플레이 프로토콜과 호환되는 JSON 구조 설계
- CLI 옵션을 통한 저장/로드 기능 제공

### 1.2 범위
- 게임 상태의 완전한 직렬화/역직렬화
- 파일 시스템을 통한 영구 저장
- 명령행 인터페이스 통합
- 기존 게임 로직과의 호환성 유지

## 2. 현재 프로젝트 구조 분석

### 2.1 핵심 컴포넌트
- **Game 구조체** (`pkg/engine/game.go`): 게임의 전체 상태 관리
- **Player 구조체** (`pkg/engine/player.go`): 플레이어 정보 및 상태
- **CLI 인터페이스** (`cmd/root.go`): 사용자 상호작용
- **게임 엔진** (`pkg/engine/`): 게임 로직 및 상태 관리

### 2.2 저장해야 할 주요 상태
- 게임 메타데이터 (핸드 번호, 페이즈, 딜러 위치 등)
- 플레이어 정보 (칩, 상태, 핸드 카드, 베팅 정보)
- 커뮤니티 카드
- 팟 정보 및 베팅 라운드 상태
- 게임 규칙 및 설정

## 3. JSON 저장 포맷 설계

### 3.1 저장 파일 구조
```json
{
  "version": "1.0",
  "timestamp": "2024-01-15T10:30:00Z",
  "game_metadata": {
    "hand_count": 5,
    "phase": "PhaseFlop",
    "dealer_pos": 2,
    "current_turn_pos": 3,
    "pot": 15000,
    "bet_to_call": 2000,
    "last_raise_amount": 1000,
    "small_blind": 500,
    "big_blind": 1000,
    "blind_up_interval": 2,
    "total_initial_chips": 1800000,
    "actions_taken_this_round": 2,
    "action_closer_pos": 1
  },
  "players": [
    {
      "name": "YOU",
      "chips": 285000,
      "is_cpu": false,
      "position": 0,
      "status": "PlayerStatusPlaying",
      "current_bet": 2000,
      "total_bet_in_hand": 5000,
      "last_action_desc": "Call 2000",
      "hand": [
        {"suit": "Spades", "rank": "Ace"},
        {"suit": "Hearts", "rank": "King"}
      ],
      "profile": null
    }
  ],
  "community_cards": [
    {"suit": "Clubs", "rank": "Ace"},
    {"suit": "Diamonds", "rank": "King"},
    {"suit": "Spades", "rank": "Queen"}
  ],
  "deck_state": {
    "remaining_cards": 45,
    "seed": 1234567890
  },
  "game_rules": {
    "name": "Pot Limit Seven Card Stud",
    "abbreviation": "PLS7",
    "hole_cards": {"count": 3},
    "community_cards": {"count": 4},
    "betting_limit": "pot_limit"
  },
  "settings": {
    "difficulty": "DifficultyMedium",
    "dev_mode": false,
    "shows_outs": false
  }
}
```

### 3.2 멀티플레이 호환성 고려사항
- 표준화된 JSON 스키마 사용
- 네트워크 전송을 위한 압축 가능한 구조
- 버전 관리 및 호환성 체크
- 플레이어 식별자 및 세션 정보 포함

## 4. 구현 세부 계획

### 4.1 Phase 1: 데이터 구조 설계 및 직렬화

#### 4.1.1 저장용 구조체 정의
- `GameSaveData` 구조체 생성
- 기존 `Game` 구조체와의 매핑 함수 작성
- JSON 태그 및 검증 로직 추가

#### 4.1.2 직렬화 함수 구현
- `SaveGameState(game *Game, filename string) error`
- `LoadGameState(filename string) (*Game, error)`
- 에러 처리 및 검증 로직

### 4.2 Phase 2: 파일 시스템 관리

#### 4.2.1 저장소 디렉토리 구조
```
saves/
├── manual_save_001.json    # 수동 저장 파일들
├── manual_save_002.json
└── ...
```

#### 4.2.2 파일 관리 기능
- 저장 파일 목록 조회
- 저장 파일 삭제
- 저장 파일 검증

### 4.3 Phase 3: CLI 인터페이스 통합

#### 4.3.1 새로운 명령어 추가
```bash
# 게임 시작 시 저장 파일 로드
pls7 --load saves/manual_save_001.json

# 게임 중 저장 (새로운 서브커맨드)
pls7 save --name "my_game"

# 저장 파일 목록 조회
pls7 saves list

# 저장 파일 삭제
pls7 saves delete saves/manual_save_001.json
```

#### 4.3.2 게임 중 저장 기능
- 게임 루프에 저장 옵션 추가

### 4.4 Phase 4: 테스트 및 검증

#### 4.4.1 단위 테스트
- 직렬화/역직렬화 테스트
- 파일 I/O 테스트
- 에러 케이스 테스트

#### 4.4.2 통합 테스트
- 전체 게임 플로우 테스트
- 저장/로드 후 게임 상태 일치성 검증
- 다양한 게임 시나리오 테스트

## 5. 기술적 고려사항

### 5.1 보안 및 검증
- 저장 파일 무결성 검증
- JSON 스키마 검증
- 버전 호환성 체크
- 악의적인 파일 로드 방지

### 5.2 성능 최적화
- 대용량 게임 상태 처리
- 메모리 사용량 최적화

### 5.3 에러 처리
- 파일 시스템 에러 처리
- JSON 파싱 에러 처리
- 게임 상태 복구 불가능한 경우 처리
- 사용자 친화적 에러 메시지

## 6. 구현 순서

### 6.1 1단계: 기본 구조 설계
1. `GameSaveData` 구조체 정의
2. JSON 직렬화/역직렬화 함수 구현
3. 기본 테스트 케이스 작성

### 6.2 2단계: 파일 시스템 통합
1. 저장소 디렉토리 생성 및 관리
2. 파일 I/O 함수 구현
3. 파일 검증 로직 추가

### 6.3 3단계: CLI 통합
1. 새로운 플래그 및 서브커맨드 추가
2. 게임 루프에 저장 기능 통합
3. 사용자 인터페이스 개선

### 6.5 4단계: 테스트 및 최적화
1. 포괄적인 테스트 작성
2. 성능 최적화
3. 문서화 완료

## 7. 예상 이슈 및 해결 방안

### 7.1 기술적 이슈
- **랜덤 시드 관리**: 게임 재개 시 동일한 랜덤 시드 사용
- **메모리 사용량**: 대용량 게임 상태의 효율적 처리
- **파일 충돌**: 동시 저장 시 파일 충돌 방지

### 7.2 사용자 경험 이슈
- **저장 파일 관리**: 사용자가 쉽게 저장 파일을 관리할 수 있도록
- **에러 복구**: 저장 파일 손상 시 복구 방법 제공
- **성능**: 저장/로드 시간 최소화

## 8. 성공 기준

### 8.1 기능적 요구사항
- [ ] 게임 상태의 완전한 저장/로드
- [ ] JSON 포맷으로 저장
- [ ] CLI를 통한 저장/로드 옵션 제공
- [ ] 멀티플레이 프로토콜 호환성

## 9. 향후 확장 계획

### 9.1 멀티플레이 지원
- 네트워크를 통한 게임 상태 동기화
- 실시간 저장 파일 공유
- 서버 기반 저장소

### 9.2 고급 기능
- 게임 리플레이 기능
- 통계 및 분석 도구
- 클라우드 저장소 연동

이 계획서는 게임 상태 저장 및 로드 기능의 완전한 구현을 위한 로드맵을 제공합니다. 각 단계는 이전 단계의 완료를 전제로 하며, 점진적인 개발을 통해 안정적이고 사용자 친화적인 기능을 구현할 수 있습니다.
