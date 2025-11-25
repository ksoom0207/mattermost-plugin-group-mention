# Mattermost 그룹 멘션 플러그인 사용 가이드

## 📦 설치

1. **플러그인 업로드**
   - Mattermost 시스템 콘솔 접속
   - **System Console > Plugin Management** 이동
   - **Upload Plugin** 클릭
   - `dist/com.mattermost.plugin-group-mention-1.0.0.tar.gz` 선택
   - 플러그인 활성화

## 🚀 기본 사용법

### 1. 그룹 생성

```bash
/group create dev --public --members @alice @bob @charlie
```

**옵션:**
- `--public`: 공개 그룹 (모든 팀원이 볼 수 있음, 기본값)
- `--private`: 비공개 그룹 (관리자와 오너만 멤버 목록 확인 가능)
- `--owners @user1 @user2`: 그룹 오너 지정
- `--members @user1 @user2`: 초기 멤버 추가

**예시:**
```bash
# 개발팀 그룹 만들기
/group create dev --public --members @김철수 @이영희 @박민수

# 리더십 비공개 그룹
/group create leadership --private --owners @대표 --members @임원1 @임원2

# 인프라 팀
/group create infra --members @데브옵스1 @데브옵스2
```

### 2. 멤버 추가/제거

```bash
# 멤버 추가
/group add dev @신입사원

# 멤버 제거
/group remove dev @퇴사자
```

### 3. 그룹 조회

```bash
# 모든 그룹 보기
/group list

# 특정 그룹 상세 정보
/group show dev
```

**출력 예시:**
```
## Group: @dev

Visibility: public
Members: 5
Owners: 1
Created: 2025-11-24 13:00:00

Member list:
- @김철수
- @이영희
- @박민수
- @최지원
- @신입사원
```

### 4. 그룹 멘션 사용

메시지에서 `@그룹명`을 입력하면 전체 멤버에게 알림이 전송됩니다:

```
팀 미팅 15분 후 시작합니다! @dev
```

**동작:**
1. 메시지가 채널에 일반 텍스트로 표시: `팀 미팅 15분 후 시작합니다! @dev`
2. dev 그룹의 모든 멤버가 개인 DM으로 알림 수신:
   ```
   You were mentioned in a group mention by @작성자 in ~채널이름
   [View message](메시지 링크)
   ```

### 5. 그룹 삭제

```bash
/group delete dev
```

### 6. 도움말

```bash
/group help
```

## ⚙️ 주요 설정 (관리자용)

**System Console > Plugins > Group Mention**

| 설정 | 기본값 | 설명 |
|------|--------|------|
| **Maximum Users Per Group Mention** | 50 | 한 번에 알림 가능한 최대 인원 |
| **Maximum Group Mentions Per Message** | 3 | 메시지당 최대 그룹 멘션 수 |
| **Rate Limit: Per User Per Minute** | 10 | 사용자당 분당 멘션 제한 |
| **Allow Regular Users to Create Groups** | false | 일반 사용자 그룹 생성 허용 여부 |
| **Require Channel Admin in Large Channels** | true | 대규모 채널에서 관리자만 사용 가능 |
| **Large Channel Threshold** | 1000 | 대규모 채널 기준 (멤버 수) |

## 🎯 사용 시나리오

### 시나리오 1: 개발팀 코드 리뷰 요청
```bash
# 1. 개발팀 그룹 생성 (최초 1회)
/group create dev --members @김개발 @이개발 @박개발

# 2. PR 리뷰 요청
안녕하세요 @dev, PR #123 리뷰 부탁드립니다!
https://github.com/company/project/pull/123
```

### 시나리오 2: 운영팀 긴급 알림
```bash
# 1. 운영팀 그룹 생성
/group create ops --members @김운영 @이데브옵스 @박인프라

# 2. 긴급 알림
🚨 긴급: 프로덕션 서버 장애 발생 @ops
즉시 대응 부탁드립니다!
```

### 시나리오 3: 부서별 공지
```bash
# 1. 부서 그룹들 생성
/group create marketing --members @마케팅1 @마케팅2 @마케팅3
/group create sales --members @영업1 @영업2 @영업3

# 2. 다중 부서 공지
다음 주 월요일 전사 미팅이 있습니다 @marketing @sales
참석 부탁드립니다!
```

## 🔒 권한 구조

| 역할 | 그룹 생성 | 그룹 관리 | 그룹 멘션 사용 | 비공개 그룹 조회 |
|------|-----------|-----------|----------------|------------------|
| **시스템 관리자** | ✅ | ✅ 모든 그룹 | ✅ | ✅ |
| **팀 관리자** | ✅ | ✅ 팀 내 그룹 | ✅ | ✅ |
| **그룹 오너** | ⚙️* | ✅ 본인 그룹 | ✅ | ✅ 본인 그룹 |
| **일반 사용자** | ⚙️* | ❌ | ✅** | ❌ |

*설정에 따라 허용 가능
**레이트 리밋 및 대규모 채널 제한 적용

## ⚠️ 제한 사항

1. **레이트 리밋**
   - 사용자당: 분당 10회
   - 채널당: 분당 20회
   - 그룹당: 분당 15회

2. **대규모 그룹 제한**
   - 한 그룹 멘션당 최대 50명
   - 초과 시 경고 메시지 표시

3. **대규모 채널 보호**
   - 1000명 이상 채널: 관리자만 그룹 멘션 가능 (기본 설정)

## 💡 팁

1. **명확한 그룹명 사용**: `dev`, `ops`, `marketing` 등 짧고 명확하게
2. **실제 사용자명과 충돌 방지**: 그룹명은 실제 사용자명과 다르게 설정
3. **정기적인 멤버 관리**: 퇴사자/이동자는 그룹에서 제거
4. **비공개 그룹 활용**: 민감한 팀은 `--private` 옵션 사용

## 🐛 문제 해결

**Q: 그룹 멘션해도 알림이 안 옵니다**
- 그룹이 존재하는지 확인: `/group show 그룹명`
- 레이트 리밋 확인 (너무 많이 사용했는지)
- 대규모 채널에서는 관리자 권한 필요

**Q: 그룹을 만들 수 없습니다**
- 관리자가 일반 사용자 그룹 생성을 막았을 수 있음
- 팀 관리자에게 문의

**Q: HTML 태그가 보입니다**
- 최신 버전으로 업데이트 필요 (v1.0.0+)
- 플러그인 재활성화

## 📞 지원

- **이슈 리포팅**: [GitHub Issues](https://github.com/mattermost/mattermost-plugin-group-mention/issues)
- **문서**: README.md
- **라이센스**: Apache 2.0
