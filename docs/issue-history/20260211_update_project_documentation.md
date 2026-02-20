# 프로젝트 문서 업데이트

## 요청 1: 프로젝트 관련 markdown 문서 전면 업데이트

프로젝트의 모든 markdown 문서를 최신화하고, 새로운 로드맵 문서를 작성하며, README를 완전히 재작성하는 작업.

## 작업 내용

### 코드베이스 분석

- CLAUDE.md 기반으로 전체 Go 소스 파일(44개) 분석 완료
- `pkg/poker` (8 파일), `pkg/engine` (11 파일), `cmd/root.go`, `internal/cli` (3 파일), `internal/config`, `internal/util` 전체 구조 파악
- `rules/*.yml` 5개 변형(PLS7, PLS, NLH, PLO, PLO8) 규칙 파일 확인

### 삭제된 파일

| 파일 | 삭제 사유 |
|------|-----------|
| `docs/GEMINI.md` | CLAUDE.md + pls7-dev-rules 스킬로 대체됨 |
| `docs/GEMINI_en.md` | CLAUDE.md + pls7-dev-rules 스킬로 대체됨 |
| `docs/development_plan.md` | 18단계 모두 완료, 새 로드맵으로 대체 |
| `docs/game_save_load_implementation_plan.md` | 핵심 기능 구현 완료 |

### 재작성된 파일

| 파일 | 주요 변경 |
|------|-----------|
| `README.md` | 완전 재작성. 배지, 아키텍처 다이어그램 이미지, 5개 변형 비교표, Quick Start, 사용법, 게임 컨트롤, Save/Load, 빌드/테스트, 문서 링크 업데이트 |
| `docs/architecture.md` | 완전 재작성. 7개 Mermaid 다이어그램 포함 (패키지 의존성, 상태 머신, 실행 흐름, 시퀀스, 데이터 흐름, AI 의사결정, 클래스 다이어그램). 디자인 패턴, AI 시스템, 변형 비교, Save/Load 상세 문서화 |
| `docs/architecture_ko.md` | 한국어 버전 완전 재작성. 영어 버전과 동일한 구조 및 Mermaid 다이어그램 |
| `docs/directory_structure.md` | 완전 재작성. 실제 현재 파일 구조 반영 (.claude/, PLO/PLO8 룰, saves/ 등), 패키지별 상세 테이블, 의존성 다이어그램 |
| `docs/directory_structure_ko.md` | 한국어 버전 완전 재작성 |

### 신규 생성 파일

| 파일 | 내용 |
|------|------|
| `docs/roadmap_v20260211.md` | 새 로드맵. 완료 항목 제거, 우선순위 재정의 (웹앱 > CLI 멀티플레이), GOTH Stack (Go/Gin + Templ + HTMX + TailwindCSS + daisyUI + Air) 기반 SSR 방식으로 전환. 5단계 개발 계획 및 Mermaid 아키텍처 다이어그램 포함 |

### 결정 사항

- GEMINI.md / GEMINI_en.md: 삭제 (CLAUDE.md + pls7-dev-rules 스킬로 대체됨)
- development_plan.md: 삭제 (새 로드맵으로 대체)
- game_save_load_implementation_plan.md: 삭제 (기능 구현 완료)
- 문서 언어 전략: EN/KO 별도 파일 유지
- GOTH Stack 추가 도구: TailwindCSS + daisyUI (스타일링) + Air (핫 리로드)

---
