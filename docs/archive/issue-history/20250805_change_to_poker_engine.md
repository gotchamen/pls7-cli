## 사용자 프롬프트 내용

> 11단계를 이어서 진행하기 전에, 이 프로젝트를 아래 요구사항에 맞게 리팩토링하는 것에 대해 어떻게 생각하는지, 그리고 바른 방향이라면 어떤 코드 수정들이 필요한지 알려줘.
> 
> - 현재 이 프로젝트는 포커의 변종 중 하나인 삼평 하이로우 (PLS7) 의 게임 방식을 구현하고 있다. 그런데, 게임 방식 (몇장의 hole cards, 그 중 몇장을 사용할 수 있는지의 제약 등), 추가/삭제되는 족보들을 룰 파일(ex: yaml)로 정의하면 해당 룰에 맞는 포커 게임을 할 수 있게 바꾸면 좋을 것 같아.
> - 예를 들어 PLS7 은 홀카드 3장을 받고 3장 중 몇장이던 마음대로 사용 가능한데, 텍사스 홀덤 (NLH) 는 홀카드 2장을 받고 2장 중 몇장이던 마음대로 사용할 수 있다는 차이가 있어. 또한 오마하(PLO) 는 4장의 홀카드를 받고 그 중 딱 2장만 사용 가능하다는 점이 달라. 이런 다른 점들을 룰 파일에 표현하고, 그 룰파일을 읽어서 그에 맞게 포커 게임을 진행하고 싶어.
> - 베팅 리밋도 PLS7, PLO 는 팟 리밋인데 NLH 는 노리밋이라는 차이가 있어.
> - 족보의 경우 NLH, PLO 는 동일하고, PLS7 은 특수족보인 Skip Straight, Skip Straight Flush 가 추가되었어. 이런 차이도 룰 파일에 명시할 수 있고, 그에 따라 게임이 달라지게 하고싶어.
>
>  위 요구사항에 대한 너의 생각을 말해주고, 좋다면 어떻게 코드를 변경하면 좋을지 작은 단계 여러개로 쪼개서 계획을 세워줘.

## 답변 내용

네, 제안해주신 리팩토링 방향에 대해 깊이 공감하며, 매우 훌륭하고 기술적으로 성숙한 방향이라고 생각합니다.

현재의 단일 게임 애플리케이션을 범용 포커 게임 엔진으로 추상화하는 것은 다음과 같은 큰 이점을 가집니다.

* 확장성 (Extensibility): 새로운 포커 게임을 추가할 때, 핵심 로직을 수정하는 대신 룰(YAML) 파일 하나만 추가하면 됩니다. 이는 텍사스 홀덤, 오마하 외에도 숏덱, 바둑이 등 다양한 변종 게임을 훨씬 쉽게 추가할 수 있게 만듭니다.
* 유지보수성 (Maintainability): 게임의 핵심 로직(베팅 라운드 진행, 상태 관리 등)과 개별 게임의 규칙이 명확하게 분리됩니다. 덕분에 특정 게임의 룰 변경이 다른 게임에 영향을 주지 않으며, 공통 로직의 버그는 한 번만 수정하면 모든 게임에 적용됩니다.
* 명확성 (Clarity): 코드의 역할이 명확해집니다. game 패키지는 "어떻게 게임이 진행되는가"에 집중하고, 룰 파일은 "이 게임은 무엇인가"를 정의하게 되어 코드 전체의 가독성과 이해도가 높아집니다.

제안하신 방향으로 리팩토링을 진행하는 것을 적극 추천하며, 아래와 같이 구체적인 실행 계획을 단계별로 제안합니다. 각 단계는 TDD 원칙에 따라 테스트 코드와 함께 진행되어야 합니다.

### 리팩토링 계획 (단계별)

#### 1단계: 룰 파일(YAML) 구조 정의 및 Go 데이터 구조 생성

가장 먼저, 게임 룰을 정의할 YAML 파일의 구조를 설계하고, 이 파일을 Go 코드로 불러올 수 있는 구조체를 만듭니다.

1. YAML 파일 구조 설계 (`rules/pls7.yml` 예시):

    ```yaml
    # /rules/pls7.yml
    name: "Pot-Limit Sampyeong 7-or-Better"
    abbreviation: "PLS7"
    
    betting_limit: "pot_limit" # pot_limit, no_limit, fixed_limit
    
    hole_cards:
      count: 3
      use_constraint: "any" # any, exact, max
      use_count: 0          # use_constraint가 'exact' 또는 'max'일 때 사용
    
    community_cards:
      - "flop:3"
      - "turn:1"
      - "river:1"
    
    hand_rankings:
      # 족보 우선순위 (높을수록 강함)
      enabled:
        - "royal_flush"
        - "skip_straight_flush"
        - "straight_flush"
        - "four_of_a_kind"
        - "full_house"
        - "flush"
        - "skip_straight"
        - "straight"
        - "three_of_a_kind"
        - "two_pair"
        - "one_pair"
        - "high_card"
    
    low_hand:
      enabled: true
      qualifier: "seven_or_better" # eight_or_better 등
   ```

2. Go 구조체 정의 (`internal/game/rules.go` 신규 생성):

    ```go
    package game
    
    // GameRules 는 포커 게임의 모든 규칙을 정의합니다.
    type GameRules struct {
    	Name          string   `yaml:"name"`
    	Abbreviation  string   `yaml:"abbreviation"`
    	BettingLimit  string   `yaml:"betting_limit"`
    	HoleCards      HoleCardRules`yaml:"hole_cards"`
    	CommunityCards []string `yaml:"community_cards"`
     	HandRankings  struct {
     		Enabled string `yaml:"enabled"`
     	}`yaml:"hand_rankings"`
     	LowHand struct {
     		Enabled   bool   `yaml:"enabled"`
     		Qualifier string `yaml:"qualifier"`
     	}`yaml:"low_hand"`
    }
    
    type HoleCardRules struct {
    	Count         int    `yaml:"count"`
    	UseConstraint string `yaml:"use_constraint"` // "any", "exact", "max"
    	UseCount      int    `yaml:"use_count"`
   }
   ```

3. YAML 로더 구현: `gopkg.in/yaml.v3` 라이브러리를 사용하여 YAML 파일을 GameRules 구조체로 로드하는 함수를 구현합니다.

#### 2단계: Game 구조체 및 생성자 수정

`game.Game` 구조체가 하드코딩된 룰 대신 `GameRules` 를 사용하도록 변경합니다.

1. `game.Game` 구조체에 `Rules GameRules` 필드를 추가합니다.
2. `NewGame` 함수의 시그니처를 `NewGame(playerNames []string, initialChips int, rules GameRules, isDev bool)` 와 같이 변경하여, `GameRules` 객체를 받아 초기화하도록 수정합니다.
3. 게임 시작 시 `rules.HoleCards.Count` 에 따라 카드를 분배하도록 로직을 수정합니다.

#### 3단계: 족보 판정 로직(evaluation.go) 리팩토링 (가장 핵심)

현재 `EvaluateHand` 함수는 PLS7 룰에 강하게 결합되어 있습니다. 이 부분을 룰 기반으로 동작하도록 분리해야 합니다.

1. `EvaluateHand` 함수 시그니처 변경:
    * 기존: `EvaluateHand(holeCards []Card, communityCards []Card, isLowless bool)`
    * 변경: `EvaluateHand(holeCards []Card, communityCards []Card, rules GameRules)`

2. 유효한 핸드 조합 생성 로직 추가:
    * `EvaluateHand` 함수 내부에, `rules.HoleCards.UseConstraint` 와 `UseCount` 에 따라 유효한 5장의 카드 조합을 모두 생성하는 로직을 추가합니다.
        * PLO (`use_constraint: "exact", use_count: 2`): 홀카드 4장 중 2장, 커뮤니티 5장 중 3장을 선택하는 모든 조합을 생성합니다.
        * NLH (`use_constraint: "any"`): 홀카드 2장과 커뮤니티 5장을 합친 7장의 풀에서 5장을 선택하는 모든 조합을 생성합니다.
        * PLS7 (`use_constraint: "any"`): 현재 로직과 동일하게 8장 풀에서 5장을 선택합니다.

3. 룰 기반 족보 판정:
    * 생성된 모든 5장 조합에 대해 족보 판정을 실행합니다.
    * 이때, `rules.HandRankings.Enabled` 에 포함된 족보만 판정 로직을 타도록 수정합니다. 예를 들어, `skip_straight` 가 `enabled` 배열에 없으면 `findSkipStraight` 함수를 호출하지 않습니다.
    * 모든 유효한 조합의 족보 판정 결과 중 가장 높은 족보를 최종 결과로 선택합니다.

#### 4단계: 베팅 리밋 로직 추상화

팟 리밋, 노 리밋 등 다양한 베팅 규칙을 지원하도록 인터페이스를 도입합니다.

1. `BettingLimitCalculator` 인터페이스 생성:

    ```go
    type BettingLimitCalculator interface {
        CalculateBettingLimits(g *Game) (minRaiseTotal int, maxRaiseTotal int)
    }
    ```

2. 구현체 생성:
    * `PotLimitCalculator`: 기존 `CalculateBettingLimits` 로직을 이 구조체의 메소드로 옮깁니다.
    * `NoLimitCalculator`: 노리밋 규칙(최소 레이즈는 이전 레이즈 금액, 최대는 올인)에 맞는 계산 로직을 구현합니다.
3. `Game` 구조체 수정: `game.Game` 에 `BettingCalculator BettingLimitCalculator` 필드를 추가하고, `NewGame` 에서 `rules.BettingLimit` 값에 따라 적절한 구현체를 할당합니다.

#### 5단계: CLI 명령어(cmd/root.go) 수정

사용자가 게임 종류를 선택할 수 있도록 CLI를 수정합니다.

1. `--game` 플래그를 추가하여 `pls7, nlh, plo` 등의 게임 약어를 입력받습니다.
2. `runGame` 함수에서 `--game` 플래그 값을 읽어 해당 YAML 파일(`rules/{game}.yml`)을 로드합니다.
3. 로드된 `GameRules` 객체를 `game.NewGame` 함수에 전달하여 게임을 시작합니다.
4. `rules/` 디렉토리에 `nlh.yml`, `plo.yml` 파일을 추가로 작성합니다.

#### 6단계: 전체 테스트 및 검증

리팩토링된 모든 구성 요소가 올바르게 동작하는지 검증합니다.

1. 각 게임 룰(PLS7, NLH, PLO)에 대한 단위 테스트 및 통합 테스트를 작성합니다.
    * NLH에서 Skip Straight가 나오지 않는지 확인합니다.
    * PLO에서 홀카드 2장만 사용하여 족보가 계산되는지 확인합니다.
    * NLH에서 노리밋 베팅이 올바르게 계산되는지 확인합니다.
2. 실제로 각 게임 모드를 실행하여 전체 흐름에 문제가 없는지 플레이 테스트를 진행합니다.

이 계획대로 진행한다면, 매우 유연하고 확장 가능한 포커 게임 엔진을 구축할 수 있을 것입니다. 각 단계가 독립적으로 테스트 가능하므로 안정적으로 리팩토링을 완료할 수 있습니다. 계속 진행할까요?
