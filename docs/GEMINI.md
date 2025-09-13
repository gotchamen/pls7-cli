# Gemini CLI Action Rules

이 문서에 명시된 규칙은 Gemini CLI 가 해야 하는 것들, 하지 말아야 하는 것들을 담고 있습니다. 이 규칙들은 수시로 추가/삭제/변경될 수 있으며, Gemini CLI의 동작을 정의하는 중요한 기준이 됩니다. Gemini CLI 는 모든 행동을 하기 전 아래 명시된 규칙을 반드시 확인하고 따릅니다.

## 해야 하는 것들

- docs/GEMINI.md 문서에 명시된 규칙을 항상 확인하고, 그에 따라 행동해.
- 한국어로 사용자와 대화해. 이는 사용자와의 소통을 원활하게 하기 위함이야.
- "프로젝트 컨텍스트를 리로드해." 라는 명령에는 README.md > docs/development_plan.md 파일을 읽고, 이후 **/*.go 파일을 읽어서 프로젝트의 현재 상태를 이해하고, 필요한 경우 추가적인 질문을 통해 명확한 이해를 돕도록 해.
- 모든 기능 추가는 TDD 방식으로 할거야. 따라서 기능 추가 전에 반드시 테스트 케이스를 작성하고, 그 테스트가 실패하는 것을 확인한 후에 기능을 구현해. 기능 구현 후 테스트가 성공하는지 확인해.
- 리팩토링을 진행할 때에도 리팩토링 대상 로직에 대해 최대한 다양한 테스트 케이스를 작성하고, 리팩토링 후에도 모든 테스트가 성공하는지 확인해.
- 코드 내에 모든 커멘트나 문서화는 영어로 작성해. 이는 코드의 가독성을 높이고, 국제적인 협업을 용이하게 하기 위함이야.
- for 문이 중첩되거나 복잡한 로직이 있는 경우, 사용자가 눈으로 여러 변수들의 상태 변화를 확인할 수 있도록 logrus 을 사용해 중간 상태를 로그로 남겨. 이는 디버깅과 이해를 돕기 위한 것이야.
- 앞으로 문제가 발생하면 로그 기반으로 문제 분석을 요청할테니 코드를 추가/수정할 때 로그의 형식은 최대한 너가 분석하기 용이한 형태로 남겨줘.
- 테스트를 추가할 때 플레이어 이름들을 담고 있는 slice 는 `[]string{"YOU", "CPU1", "CPU2"}` 처럼 1번째 플레이어 이름은 무조건 `YOU`, 나머지는 `CPUn` 형식으로 작성해. 현재 비즈로직 상 하드코딩된 이름을 기준으로 로직이 짜여있기 때문이야.
- commit title, description은 영어로 작성해. PR title은 영어로, description은 영어와 한글로 작성해. 이 때, 영어로 먼저 작성하고 이후에 한글로 번역해서 작성해.

## 하지 말아야 하는 것들

- 두 번 이상 실패하는 시도는 세번째에는 중단하고, 어떤 시도를 하다가 어떤 실패를 했는지 사용자에게 자세히 알려준 후 사용자의 명령을 기다려.
- 내가 추가한 모든 로그들은 레벨에 상관 없이 절대 너가 임의로 삭제하거나 수정하지 마. 로그는 너와 내가 문제를 분석하는데 중요한 단서가 될 수 있어. 리팩토링 작업 때에도 debug 로그들을 제거하지 마.
- commit 이나 PR 의 title, description 을 만들어달라고 요청하면 line number 가 없이 내가 편하게 copy & paste 할 수 있는 형태로 만들어줘. 아래 나쁜 예와 좋은 예를 참고해.
    - Bad:
      ```
      Commit Title:

      1 feat(poker): Implement PLO hand evaluation logic

      Commit Description:

      1 Refactor the hand evaluation logic to support poker variants with specific hole card usage rules, such as Pot-Limit Omaha (PLO).
      2
      3 The main `EvaluateHand` function is now generalized. It generates all possible 5-card hands based on the game's `UseConstraint` rule ("exact" or
      "any") and evaluates each combination to find the best possible hand.
      4
      5 Key changes:
      6 - Extracted the core ranking logic for a 5-card hand into a new `evaluateSingleHand` helper function.
      7 - For "exact" constraint games like PLO, the system now correctly generates combinations by taking exactly 2 cards from the hole and 3 from the community.
      8 - For "any" constraint games, the logic was also updated to use the more robust combination-based evaluation, ensuring consistency.
      9 - Added a new `TestPLOHandEvaluation` test case to drive the implementation and verify the correctness of the PLO logic.
      ```
    - Good:
      ```
      Commit Title:
      
      feat(poker): Implement PLO hand evaluation logic

      Commit Description:
      
      Refactor the hand evaluation logic to support poker variants with specific hole card usage rules, such as Pot-Limit Omaha (PLO).

      The main `EvaluateHand` function is now generalized. It generates all possible 5-card hands based on the game's `UseConstraint` rule ("exact" or "any") and evaluates each combination to find the best possible hand.

      Key changes:
      - Extracted the core ranking logic for a 5-card hand into a new `evaluateSingleHand` helper function.
      - For "exact" constraint games like PLO, the system now correctly generates combinations by taking exactly 2 cards from the hole and 3 from the community.
      - For "any" constraint games, the logic was also updated to use the more robust combination-based evaluation, ensuring consistency.
      - Added a new `TestPLOHandEvaluation` test case to drive the implementation and verify the correctness of the PLO logic.
      ```
