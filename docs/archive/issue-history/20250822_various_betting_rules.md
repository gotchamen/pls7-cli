## 요청 프롬프트

```
╭──────────────────────────────────────────────────────────────────────────────────────────────╮
│  > 13-5 단계 - 베팅 리밋 로직 추상화- 을 진행하자. PLS, PLS7 은 pot-limit 이고, 현재 베팅 로직은 pot limit  │
│    rule 로 고정되어 있어. 하지만 새로 추가된 NLH 는 no limit rule 이 적용되어야 해. 따라서 game rule 에     │
│    따라 pot limit vs. no limit 을 동적으로 사용할 수 있어야 해. 여러 부분의 변경이 필요할텐데, 작은           │
│    여러개의 단위로 쪼개서 하나씩 진행하자. 모든 단계는 TDD 로 실패하는 테스트를 만든 후 구현을 거쳐                │
│    테스트를 통과하게 진행할거야. 어떤 세부작업들을 어떤 순서로 진행할지 계획을 보여줘.                           │
╰──────────────────────────────────────────────────────────────────────────────────────────────╯
```

## 응답

✦ 알겠습니다. 13-5 단계 '베팅 리밋 로직 추상화'를 TDD 방식으로 진행하겠습니다. 전체 작업을 작고 관리 가능한
단위로 나누어, 각 단계마다 실패하는 테스트를 먼저 작성하고 기능을 구현하여 테스트를 통과시키는 과정을
반복하겠습니다.

다음은 제가 제안하는 세부 작업 계획입니다.

  ---

베팅 리밋 로직 추상화 계획 (TDD)

1단계: `BettingLimitCalculator` 인터페이스 정의 및 테스트
* TDD: BettingLimitCalculator 인터페이스를 사용하는 테스트 코드를 먼저 작성합니다. 이 인터페이스는
  CalculateBettingLimits라는 단일 메서드를 가질 것입니다. 이 코드는 인터페이스가 아직 없으므로 컴파일에
  실패합니다.
* 구현: internal/game/betting_limit.go라는 새 파일을 만들고 BettingLimitCalculator 인터페이스를 정의하여
  테스트를 통과시킵니다.

2단계: 기존 `PotLimitCalculator` 구현 및 테스트
* TDD: 기존 팟 리밋 로직을 검증하는 테스트를 betting_test.go에서 새로운 테스트
  파일(betting_limit_test.go)로 옮겨와 PotLimitCalculator를 테스트하도록 수정합니다. 이 테스트는
  PotLimitCalculator 구현체가 없으므로 실패합니다.
* 구현: PotLimitCalculator 구조체를 만들고 BettingLimitCalculator 인터페이스를 구현합니다. 기존 Game
  구조체에 있던 팟 리밋 계산 로직을 이 구조체의 CalculateBettingLimits 메서드로 그대로 옮깁니다.

3단계: `NoLimitCalculator` 구현 및 테스트
* TDD: 노 리밋(No-Limit) 규칙을 위한 NoLimitCalculator 테스트를 작성합니다. 이 테스트는 최대 베팅 가능
  금액이 플레이어의 전체 칩과 동일한지, 최소 레이즈 금액은 규칙대로 계산되는지를 검증해야 합니다. 이
  테스트는 NoLimitCalculator 구현체가 없으므로 실패합니다.
* 구현: NoLimitCalculator 구조체를 만들고 BettingLimitCalculator 인터페이스를 구현합니다.
  CalculateBettingLimits 메서드에는 노 리밋 규칙(최대 베팅 = 플레이어의 모든 칩)에 맞는 로직을 작성합니다.

4단계: `Game` 구조체에 `BettingLimitCalculator` 통합
* TDD: NewGame 함수가 "nlh" 룰을 받았을 때, 생성된 게임 인스턴스가 NoLimitCalculator를 사용하는지 검증하는
  테스트를 작성합니다. 이 테스트는 Game 구조체와 NewGame 생성자가 수정되기 전이므로 실패합니다.
* 구현:
    1. Game 구조체에 BettingLimitCalculator 타입의 필드를 추가합니다.
    2. NewGame 생성자에서 GameRules.BettingLimit 값("pot\_limit" 또는 "no\_limit")에 따라 적절한 Calculator(
       PotLimitCalculator 또는 NoLimitCalculator)를 생성하여 Game 인스턴스에 주입합니다.
    3. 기존 Game.CalculateBettingLimits() 메서드가 Game에 주입된 Calculator를 호출하도록 수정합니다.

5단계: 최종 리팩토링 및 검증
* 리팩토링: internal/cli/input.go 등 CalculateBettingLimits를 호출하는 모든 부분들이 변경된 구조와 잘
  동작하는지 확인하고 필요시 수정합니다.
* 검증: 모든 테스트(go test ./...)를 실행하여 새롭게 추가된 기능이 올바르게 동작하고, 기존 기능에 문제가
  없는지 최종적으로 확인합니다.
