# 웹 애플리케이션 구축 전 구조 리뷰 및 리팩토링

- **작성일**: 2026-03-20
- **브랜치**: `feature/prepare-for-web-app`
- **관련 로드맵**: `docs/roadmap_v20260211.md` — Stage 1 (웹 애플리케이션 기반 구축) 이전 사전 정비

---

## 배경

로드맵의 다음 과제인 GOTH Stack 기반 웹 애플리케이션 구축에 앞서, 기존 코드베이스의 품질과 구조를 점검·개선한다. 현재 3계층 아키텍처(`CLI → Engine → Poker`)가 `docs/architecture.md` 문서와 일치하는지 확인하고, 웹 계층 추가에 걸리는 부분을 조정하며, 테스트 안전망을 강화한 뒤 코드를 개선한다.

## 현재 상태 분석 결과

### 패키지 구조

- **`docs/architecture.md`와 실제 코드 구조 100% 일치** 확인 완료
- 의존성 방향 위반 없음: `pkg/poker`는 내부 의존성 0, `pkg/engine`은 `pkg/poker`만 의존
- `ActionProvider` 인터페이스가 이미 존재하여 웹 핸들러용 구현체 추가 가능

### 웹앱 통합 시 걸리는 부분

- `cmd/root.go`의 게임 루프(약 120줄)가 `bufio.Reader`, `cli.DisplayGameState()` 등 CLI에 직접 결합
- 웹 핸들러가 이 루프를 재사용할 수 없는 구조

### 테스트 현황

- 테스트 함수명: 전반적으로 양호 (의도가 잘 드러남)
- setup 코드 중복: `betting_test.go`, `save_test.go`에서 확인
- 커버리지 부족 영역: AI 결정 엣지 케이스, 3인+ 올인 캐스케이드, 블라인드 업 전환, 타이브레이킹 키커

### 코드 복잡도

| 파일 | 함수 | 줄 수 | 이슈 |
|------|------|-------|------|
| `pkg/poker/evaluation.go` | `evaluateSingleHand` | ~113 | 12-case switch문, 각 case를 헬퍼로 분리 가능 |
| `pkg/poker/odds.go` | `CalculateOuts` | ~75 | 내부 단계를 헬퍼로 분리하여 가독성 개선 가능 |
| `pkg/engine/pot.go` | `DistributePot` | ~160 | 사이드팟+Hi-Lo+타이 처리가 한 함수에 집중 |
| `pkg/engine/ai.go` | `evaluateHandStrength` | ~77 | 매직 넘버 다수, 중첩 조건문 |
| `pkg/engine/ai.go` | `GetCPUAction` | ~64 | 포스트플롭 분기 중첩 |

---

## 작업 계획

### Phase 1: 패키지 구조 — 웹앱 대비 조정

**목표**: 웹 계층을 CLI와 병렬로 추가할 수 있도록 게임 루프를 분리한다.

**작업 내용**:

1. **게임 루프 분리 (`cmd/root.go`)**
   - 한 핸드 처리 로직을 engine 레벨 메서드로 추출
   - 예: `Game.PlaySingleHand(actionProvider)` → 한 핸드를 처리하고 결과 반환
   - 핸드 간 제어(다음 핸드 시작, 저장, 종료)는 호출자(CLI 또는 웹)가 결정
   - CLI 기존 동작은 그대로 유지

2. **구조 변경 없는 영역 확인**
   - `pkg/engine`, `pkg/poker`, `internal/` 패키지 배치: 변경 없음
   - `ActionProvider` 인터페이스: 변경 불필요 (이미 웹 호환)

**체크리스트**:

- [x] `cmd/root.go`의 게임 루프에서 한 핸드 처리 로직 식별 및 경계 정의
- [x] engine 레벨에 `PlaySingleHand(actionProvider, observer)` 메서드 추출 (`pkg/engine/run.go`)
- [x] 핸드 결과를 반환하는 `HandResult` 구조체 정의 + `GameObserver` 인터페이스 (`pkg/engine/event.go`)
- [x] `cmd/root.go`가 `PlaySingleHand`을 호출하도록 수정, `CLIObserver` 구현
- [ ] CLI 기존 동작이 변경 없이 동작하는지 수동 확인
- [x] `go build -v ./...` 성공 확인
- [x] `go test -v ./...` 전체 통과 확인

### Phase 2: 테스트 가드닝 + 커버리지 보강

**목표**: 코드 개선의 안전망을 확보한다.

**작업 내용**:

1. **테스트 setup 코드 정리**
   - `betting_test.go`: `newGameForBettingTests`와 `newGameForBettingTestsWithRules` 통합 (optional rules 파라미터)
   - `save_test.go`: 반복되는 `GameSaveData` 생성을 팩토리 헬퍼로 추출

2. **테스트 가독성 개선**
   - `save_test.go`의 `TestGameSaveDataDeserialization`: 60줄 setup을 헬퍼로 분리
   - `pot_test.go`의 `TestDistributePot_ComplexSidePotAndAllIn`: 핸드 할당 의도를 명시

3. **커버리지 보강**

   | 영역 | 추가할 테스트 |
   |------|--------------|
   | AI 결정 엣지 케이스 | 임계값 근처 핸드의 fold/call/raise 분기 |
   | 3인+ 올인 캐스케이드 | 사이드팟 3개 이상 생성 시나리오 |
   | 블라인드 업 전환 | 인터벌 경계에서 블라인드 증가 검증 |
   | 타이브레이킹 키커 | 동일 핸드랭크 2인+ 키커 비교 |

**체크리스트**:

- [x] `betting_test.go`: 두 헬퍼 함수를 `newGameForBettingTests(..., ruleAbbr ...string)` 하나로 통합
- [x] `save_test.go`: `newTestSaveData()` + `newTestGameRules()` 팩토리 헬퍼 함수 작성
- [x] `save_test.go`: `TestGameSaveDataDeserialization` setup을 `newTestSaveData` 헬퍼로 분리
- [x] `pot_test.go`: `TestDistributePot_ComplexSidePotAndAllIn` 핸드 할당 의도 주석 추가
- [x] 테스트 추가: `TestCPUAction_ThresholdEdgeCases` + `TestCPUAction_PostFlopStrengthBoundaries`
- [x] 테스트 추가: `TestDistributePot_ThreeWayAllIn_ThreeSidePots` (4인 올인, 3개 사이드팟)
- [x] 테스트 추가: `TestBlindUp_IntervalBoundary` (interval=3, interval=0 검증)
- [x] 테스트 추가: `TestDistributePot_KickerTiebreaking` (동일 핸드랭크 키커 비교)
- [x] `go test -v ./...` 전체 통과 확인

### Phase 3: engine/poker 코드 가독성 개선

**목표**: 동작 변경 없이 내부 구조를 개선한다 (순수 리팩토링).

**작업 내용**:

1. **poker 패키지**
   - `evaluateSingleHand`: 12-case switch문 → 핸드 타입별 체커 함수로 분리
   - `CalculateOuts`: 내부 단계를 헬퍼로 분리하여 가독성 개선

2. **engine 패키지**
   - `DistributePot`: 단계별 함수로 분리 (사이드팟 계산, Hi-Lo 분배, 잔여 칩 처리)
   - `evaluateHandStrength`: 매직 넘버를 의미 있는 상수로 추출
   - `GetCPUAction`: 포스트플롭 핸드 강도별 결정 로직을 헬퍼로 분리

3. **cmd 리팩토링**
   - Phase 1에서 식별한 게임 루프 분리 (Phase 1과 동시에 또는 순차 진행)

**개선 원칙**:
- Phase 2에서 확보한 테스트로 동작 검증
- 외부 인터페이스(함수 시그니처, 구조체 필드) 변경 최소화
- 동작 변경 없이 내부 구조만 개선

**체크리스트**:

- [x] `evaluateSingleHand`: `handCheckers` 맵 + 12개 `check*` 함수로 분리
- [x] `CalculateOuts`: `drawChecker` 타입 + `collectDrawOuts` 헬퍼 + `buildSeenCards` 추출
- [x] `DistributePot`: `buildPotTiers` (사이드팟 계산) 별도 함수 추출
- [x] `DistributePot`: `distributeHiLoPot` (Hi-Lo 분배) 별도 함수 추출
- [x] `DistributePot`: `distributeHighOnlyPot` (High-only 분배) 별도 함수 추출
- [x] `evaluateHandStrength`: 매직 넘버를 상수(`pairBonus`, `suitedBonus` 등) + `evaluatePreFlopStrength` 헬퍼로 분리
- [x] `GetCPUAction`: `getPreFlopAction`, `getPostFlopAction`, `bluffAction`, `strongHandAction`, `weakHandAction`으로 분리
- [x] 각 리팩토링 단계마다 `go test -v ./...` 통과 확인
- [x] 최종 `go build -v ./...` 성공 확인
- [ ] 기존 CLI 게임 플레이 동작 변경 없음 수동 확인

---

## 다음 작업 예고: 포커 엔진 핵심 로직 재설계

이번 작업이 완료되면 별도 PR로 진행할 예정:

- `evaluateSingleHand`, `CalculateOuts`, `DistributePot`, `evaluateHandStrength`, `GetCPUAction` 함수들의 **로직 자체를 백지에서 재검토**
- 현재 코드를 참조만 하여 **새로운 포커 엔진 패키지를 설계**하는 방식
- 이번 작업에서 확보한 테스트 안전망으로 동작 정합성 검증
- 범위를 분리하여 각 PR의 리뷰를 명확하게 유지

---

## 검증 기준

- [x] `go build -v ./...` 성공
- [x] `go test -v ./...` 전체 통과
- [ ] 기존 CLI 게임 플레이 동작 변경 없음
- [ ] `docs/architecture.md` 업데이트 (구조 변경 시)
