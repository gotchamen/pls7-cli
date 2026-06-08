## 질문

게임을 플레이해보니 프리플랍에서 2명의 플레이어가 남은 경우 플랍, 턴, 리버 세 스트릿에서 플레이어의 턴이 건너뛰어지는 현상이 발생하는데, 원인을 찾아봐. 아래는 full logs 이니 참고해.

```text
DEBU[2025-08-27T10:38:58+09:00] Logger initialized in DEBUG mode.            
======== No-Limit Texas Hold'em ========
DEBU[2025-08-27T10:38:58+09:00] Player YOU: Status: Playing, Current Bet: 0, Last Action: , actionInfo: [, Current Bet: 0     ] 
DEBU[2025-08-27T10:38:58+09:00] Player CPU 1: Status: Playing, Current Bet: 0, Last Action: , actionInfo: [, Current Bet: 0     ] 
DEBU[2025-08-27T10:38:58+09:00] Player CPU 2: Status: Playing, Current Bet: 0, Last Action: , actionInfo: [, Current Bet: 0     ] 
DEBU[2025-08-27T10:38:58+09:00] Player CPU 3: Status: Playing, Current Bet: 0, Last Action: , actionInfo: [, Current Bet: 0     ] 
DEBU[2025-08-27T10:38:58+09:00] Player CPU 4: Status: Playing, Current Bet: 0, Last Action: , actionInfo: [, Current Bet: 0     ] 
DEBU[2025-08-27T10:38:58+09:00] Player CPU 5: Status: Playing, Current Bet: 0, Last Action: , actionInfo: [, Current Bet: 0     ] 
DEBU[2025-08-27T10:38:58+09:00] Total chips match expected value: 1,800,000  


--- NLH (Easy) | HAND #0 | PHASE: PRE-FLOP | POT: 0 | BLINDS: 500/1,000 ---
Board: 

Players:
> YOU                         : Chips: 300,000  , Current Bet: 0       | Hand:
CPU 1 (Loose-Passive)       : Chips: 300,000  , Current Bet: 0       | Hand:
CPU 2 (Loose-Passive)       : Chips: 300,000  , Current Bet: 0       | Hand:
CPU 3 (Loose-Passive)       : Chips: 300,000  , Current Bet: 0       | Hand:
CPU 4 (Loose-Passive)       : Chips: 300,000  , Current Bet: 0       | Hand:
CPU 5 (Loose-Passive)       : Chips: 300,000  , Current Bet: 0       | Hand:
-------------------------------------------------
CPU 3 folds.
CPU 4 folds.
CPU 5 calls 1,000.
DEBU[2025-08-27T10:38:59+09:00] Player YOU: Status: Playing, Current Bet: 0, Last Action: , actionInfo: [, Current Bet: 0     ] 
DEBU[2025-08-27T10:38:59+09:00] Player CPU 1: Status: Playing, Current Bet: 500, Last Action: , actionInfo: [, Current Bet: 500   ] 
DEBU[2025-08-27T10:38:59+09:00] Player CPU 2: Status: Playing, Current Bet: 1,000, Last Action: , actionInfo: [, Current Bet: 1,000 ] 
DEBU[2025-08-27T10:38:59+09:00] Player CPU 3: Status: Folded, Current Bet: 0, Last Action: Fold, actionInfo: [, Current Bet: 0      - Fold] 
DEBU[2025-08-27T10:38:59+09:00] Player CPU 4: Status: Folded, Current Bet: 0, Last Action: Fold, actionInfo: [, Current Bet: 0      - Fold] 
DEBU[2025-08-27T10:38:59+09:00] Player CPU 5: Status: Playing, Current Bet: 1,000, Last Action: Call 1000, actionInfo: [, Current Bet: 1,000  - Call 1000] 
DEBU[2025-08-27T10:38:59+09:00] Total chips match expected value: 1,800,000  


--- NLH (Easy) | HAND #1 | PHASE: PRE-FLOP | POT: 2,500 | BLINDS: 500/1,000 ---
Board: 

Players:
> YOU                         : Chips: 300,000  , Current Bet: 0       | Hand: A♠️️  A♥️️
CPU 1 (Loose-Passive)       : Chips: 299,500  , Current Bet: 500     | Hand: 9♠️️  8♣️️
CPU 2 (Loose-Passive)       : Chips: 299,000  , Current Bet: 1,000   | Hand: 9♥️️  J♦️
CPU 3 (Loose-Passive)       : Chips: 300,000  , Current Bet: 0      - Fold (Folded) | Hand: 6♦️  2♠️️
CPU 4 (Loose-Passive)       : Chips: 300,000  , Current Bet: 0      - Fold (Folded) | Hand: 6♥️️  2♥️️
CPU 5 (Loose-Passive)       : Chips: 299,000  , Current Bet: 1,000  - Call 1000  | Hand: 10♦️  K♥️️
-------------------------------------------------
Choose your action: (c)all 1,000, (r)aise, (f)old > r
Enter amount to raise to (min: 2,000, max: 300,000): 100000
YOU raises to 100,000.
CPU 1 folds.
CPU 2 folds.
CPU 5 calls 99,000.
CPU 5 checks.
CPU 5 checks.
CPU 5 checks.
DEBU[2025-08-27T10:39:05+09:00] DistributePot: Initial Pot: 201500, All Contributors: [YOU CPU 1 CPU 2 CPU 5], Bet Tiers: [500 1000 100000] 
DEBU[2025-08-27T10:39:05+09:00]   New PotTier created: Amount: 2000, MaxBet: 500, Players: [YOU CPU 5] 
DEBU[2025-08-27T10:39:05+09:00]   New PotTier created: Amount: 1500, MaxBet: 1000, Players: [YOU CPU 5] 
DEBU[2025-08-27T10:39:05+09:00]   New PotTier created: Amount: 198000, MaxBet: 100000, Players: [YOU CPU 5] 
DEBU[2025-08-27T10:39:05+09:00] Distributing PotTier: Amount: 2000, MaxBet: 500, Eligible Players: [YOU CPU 5] 
DEBU[2025-08-27T10:39:05+09:00] DistributePot: High Winners: [YOU], Best High Hand: One Pair, A♠️️ -A♥️️ -10♣️️ -9♣️️ -6♣️️  
DEBU[2025-08-27T10:39:05+09:00] DistributePot: Low Winners: [], Best Low Hand: N/A 
DEBU[2025-08-27T10:39:05+09:00]     YOU scoops 2000 from pot                 
DEBU[2025-08-27T10:39:05+09:00] Distributing PotTier: Amount: 1500, MaxBet: 1000, Eligible Players: [YOU CPU 5] 
DEBU[2025-08-27T10:39:05+09:00] DistributePot: High Winners: [YOU], Best High Hand: One Pair, A♠️️ -A♥️️ -10♣️️ -9♣️️ -6♣️️  
DEBU[2025-08-27T10:39:05+09:00] DistributePot: Low Winners: [], Best Low Hand: N/A 
DEBU[2025-08-27T10:39:05+09:00]     YOU scoops 1500 from pot                 
DEBU[2025-08-27T10:39:05+09:00] Distributing PotTier: Amount: 198000, MaxBet: 100000, Eligible Players: [YOU CPU 5] 
DEBU[2025-08-27T10:39:05+09:00] DistributePot: High Winners: [YOU], Best High Hand: One Pair, A♠️️ -A♥️️ -10♣️️ -9♣️️ -6♣️️  
DEBU[2025-08-27T10:39:05+09:00] DistributePot: Low Winners: [], Best Low Hand: N/A 
DEBU[2025-08-27T10:39:05+09:00]     YOU scoops 198000 from pot               
DEBU[2025-08-27T10:39:05+09:00] DistributePot: Final results: [{PlayerName:YOU AmountWon:201500 HandDesc:High: One Pair, A♠️️ -A♥️️ -10♣️️ -9♣️️ -6♣️️  (Scoop)}] 

--- SHOWDOWN ---
Community Cards: [10♣️️  4♦️  9♣️️  6♣️️  5♠️️ ]
- YOU    : [A♠️️  A♥️️ ] -> One Pair, A♠️️ -A♥️️ -10♣️️ -9♣️️ -6♣️️  (High Winner)
- CPU 5  : [10♦️  K♥️️ ] -> One Pair, 10♦️ -10♣️️ -K♥️️ -9♣️️ -6♣️️ 

--- POT DISTRIBUTION ---
YOU wins 201,500 chips with High: One Pair, A♠️️ -A♥️️ -10♣️️ -9♣️️ -6♣️️  (Scoop)
------------------------

--- End of Hand ---
```

## 답변

알겠습니다. 로그를 자세히 살펴보니 원인이 파악됩니다.

### 문제 원인:

현재 베팅 라운드가 종료되는지를 판단하는 IsBettingRoundOver() 함수의 로직에 허점이 있습니다. 이 함수는 마지막으로 베팅/레이즈한 플레이어(Aggressor)에게 차례가 돌아오거나, 아무도 베팅하지 않았을 경우 특정 플레이어(ActionCloserPos)에게 차례가 돌아오면 라운드가 종료된다고 판단합니다.

하지만 2명만 남은 헤즈업(Heads-up) 상황에서는, 포스트-플랍의 첫 액션이 체크로 시작되면, 두 번째 플레이어에게 턴이 갔을 때 바로 라운드가 종료되어 버립니다. 두 번째 플레이어가 액션(체크 또는 베팅)을 할 기회를 갖지 못하는 것입니다. 로그에서 CPU 5만 계속 체크하고 YOU에게는 턴이 오지 않는 것이 바로 이 때문입니다.

### 해결 방안:

IsBettingRoundOver() 함수의 로직을 더 명확하고 단순하게 변경해야 합니다. 베팅 라운드는 아래 두 조건이 모두 만족될 때 종료되어야 합니다.

1. 이번 라운드에서 액션을 할 수 있는 모든 플레이어가 최소 한 번씩 자신의 턴을 가졌을 때
2. 모든 플레이어의 베팅 금액이 동일할 때 (즉, 모든 베팅이 정리되었을 때)

이 로직을 구현하기 위해 다음과 같이 코드를 수정하겠습니다.

1. game.Game 구조체에 ActionsTakenThisRound int 필드를 추가하여, 현재 베팅 라운드에서 몇 번의 액션이 있었는지 카운트합니다.
2. 베팅 라운드가 새로 시작될 때마다(PrepareNewBettingRound 함수) 이 카운트를 0으로 초기화합니다.
3. 플레이어가 액션을 할 때마다(ProcessAction 함수) 이 카운트를 1씩 증가시킵니다.
4. IsBettingRoundOver() 함수는 "액션 수"와 "베팅 금액 일치 여부" 두 가지만을 확인하도록 로직을 대폭 단순화합니다.

이 방식은 플레이어의 위치나 마지막 공격자 변수(Aggressor)에 의존하지 않기 때문에, 헤즈업을 포함한 모든 상황에서 더 안정적으로 동작할 것입니다.
