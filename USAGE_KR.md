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

메시지에서 `@그룹명`을 입력하면 전체 멤버에게 네이티브 푸시 알림이 전송됩니다:

```
팀 미팅 15분 후 시작합니다! @dev
```

**동작:**
1. 입력한 메시지: `팀 미팅 15분 후 시작합니다! @dev`
2. 자동 확장되어 저장: `팀 미팅 15분 후 시작합니다! @김철수 @이영희 @박민수`
3. dev 그룹의 모든 멤버에게 **실제 멘션 알림** 전송:
   - ✅ 모바일 푸시 알림
   - ✅ PC 데스크톱 알림
   - ✅ 채널 내 하이라이트
4. **DM 없음** - 일반 @멘션과 동일한 UX

**참고:**
- 채널에 있는 그룹 멤버에게만 알림 전송
- 메시지 작성자는 자동으로 제외
- 멘션된 사용자명이 메시지에 직접 표시됨

### 5. 그룹 삭제

```bash
/group delete dev
```

### 6. 도움말

```bash
/group help
```

## 📱 알림 동작 원리

플러그인은 **MessageWillBePosted** 훅을 사용하여 메시지가 저장되기 **전에** `@그룹명`을 개별 `@사용자명`으로 확장합니다.

### 기술적 동작 방식

```
1. 사용자 입력: "회의 시작합니다 @dev"
   ↓
2. MessageWillBePosted 훅 실행
   ↓
3. @dev 그룹 조회 (멤버: alice, bob, charlie)
   ↓
4. 채널 멤버 필터링 (alice, bob만 채널에 있음)
   ↓
5. 메시지 확장: "회의 시작합니다 @alice @bob"
   ↓
6. 확장된 메시지로 저장
   ↓
7. Mattermost 네이티브 시스템이 @alice, @bob에게 자동 푸시 알림
```

### 장점

- ✅ **실제 모바일 푸시 알림** - iOS/Android 푸시 알림 전송
- ✅ **데스크톱 알림** - PC/Mac 데스크톱 알림
- ✅ **DM 스팸 없음** - 별도 DM 메시지 없이 깔끔한 UX
- ✅ **네이티브 UX** - Mattermost 기본 멘션과 동일한 경험
- ✅ **배지 카운트** - 읽지 않은 멘션 수에 포함

### 주요 특징

- **채널 멤버만 알림**: 그룹 멤버여도 채널에 없으면 알림 없음
- **실시간 필터링**: 메시지 전송 시점의 채널 멤버 기준
- **자동 제외**: 메시지 작성자는 알림 대상에서 자동 제외
- **우선순위**: 실제 사용자명이 있으면 그룹보다 우선

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

### 역할별 권한

| 역할 | 그룹 생성 | 그룹 관리 | 그룹 멘션 사용 | 비공개 그룹 조회 |
|------|-----------|-----------|----------------|------------------|
| **시스템 관리자** | ✅ | ✅ 모든 그룹 | ✅ | ✅ |
| **팀 관리자** | ✅ | ✅ 팀 내 그룹 | ✅ | ✅ |
| **그룹 오너** | ⚙️* | ✅ 본인 그룹만 | ✅ | ✅ 본인 그룹 |
| **일반 사용자** | ⚙️* | ❌ | ✅** | ❌ |

*설정에 따라 허용 가능 (`Allow Regular Users to Create Groups`)
**레이트 리밋 및 대규모 채널 제한 적용

### 그룹 오너 vs 팀 관리자

**그룹 오너 (Group Owner)**
- 플러그인 내부 개념
- 특정 그룹을 만든 사람 또는 지정된 사람
- **본인이 만든/관리하는 그룹만** 수정/삭제 가능
- 일반 사용자도 그룹 오너가 될 수 있음 (`AllowUserManagedGroups=true` 시)
- 지정 방법: `/group create dev --owners @사용자1 @사용자2`

**팀 관리자 (Team Admin)**
- Mattermost 기본 역할
- **팀 내 모든 그룹** 수정/삭제 가능
- 항상 그룹 생성 가능 (설정 무관)
- System Console에서 지정

### 그룹 생성 권한 상세

**`Allow Regular Users to Create Groups = false` (기본값)**
```
✅ 시스템 관리자: 모든 팀 그룹 생성 가능
✅ 팀 관리자: 본인 팀 그룹 생성 가능
❌ 일반 사용자: 그룹 생성 불가
```

**`Allow Regular Users to Create Groups = true`**
```
✅ 시스템 관리자: 모든 팀 그룹 생성 가능
✅ 팀 관리자: 본인 팀 그룹 생성 가능
✅ 일반 사용자: 그룹 생성 가능 (자동으로 오너가 됨)
```

### 예시 시나리오

**시나리오 1: 일반 사용자가 그룹 생성**
```bash
# 설정: AllowUserManagedGroups = true
# 사용자 @김철수(일반)가 실행:
/group create myteam --members @이영희 @박민수

# 결과:
# - 그룹 생성 성공
# - @김철수가 자동으로 그룹 오너로 지정됨
# - @김철수만 이 그룹을 수정/삭제 가능 (팀관리자/시스템관리자 제외)
```

**시나리오 2: 팀 관리자가 복수 오너 지정**
```bash
# 사용자 @이팀장(팀 관리자)가 실행:
/group create leadership --private --owners @대표 @부대표 --members @임원1 @임원2

# 결과:
# - 비공개 그룹 생성
# - @대표, @부대표가 그룹 오너
# - @이팀장도 팀 관리자이므로 이 그룹 관리 가능
```

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
   - 예: `frontend-team`, `backend-devs`, `qa-team`

2. **실제 사용자명과 충돌 방지**: 그룹명은 실제 사용자명과 다르게 설정
   - 만약 `@john`이라는 사용자가 있다면, 그룹명을 `john-team`으로 설정
   - 실제 사용자명이 우선순위가 높습니다

3. **정기적인 멤버 관리**: 퇴사자/이동자는 그룹에서 제거
   - `/group remove dev @퇴사자`로 정기적으로 정리

4. **비공개 그룹 활용**: 민감한 팀은 `--private` 옵션 사용
   - 예: `/group create leadership --private --members @임원1 @임원2`

5. **그룹 크기 조절**: 너무 큰 그룹은 여러 개로 분할
   - 50명 제한을 고려하여 `dev-frontend`, `dev-backend`로 나누기

6. **채널 컨텍스트 고려**: 채널 멤버만 알림 받음을 기억
   - 전체 팀에게 알림을 보내려면 해당 채널에 모두 초대해야 함

7. **메시지 확장 예상**: @그룹명이 @멤버들로 확장됨
   - 긴 메시지 작성 시 멤버 수를 고려하세요
   - 예: 10명 그룹 멘션 → 10개의 @사용자명으로 확장

## 🐛 문제 해결

**Q: 그룹 멘션해도 알림이 안 옵니다**
- 그룹이 존재하는지 확인: `/group show 그룹명`
- 채널에 그룹 멤버가 있는지 확인 (채널 멤버만 알림 받음)
- 레이트 리밋 확인 (너무 많이 사용했는지)
- 대규모 채널(1000명+)에서는 관리자 권한 필요

**Q: 그룹을 만들 수 없습니다**
- 관리자가 일반 사용자 그룹 생성을 막았을 수 있음 (`AllowUserManagedGroups=false`)
- 팀 관리자 또는 시스템 관리자에게 문의

**Q: @그룹명을 입력했는데 메시지가 @사용자1 @사용자2로 바뀝니다**
- **정상 동작입니다!** 이것이 실제 푸시 알림을 보내는 방식입니다
- 플러그인이 메시지 저장 전에 자동으로 확장합니다
- 이를 통해 Mattermost의 네이티브 알림 시스템을 활용합니다

**Q: 그룹 멤버인데 알림이 안 와요**
- 해당 채널의 멤버인지 확인하세요
- 채널에 없는 그룹 멤버는 알림을 받지 않습니다
- 이는 스팸 방지를 위한 의도된 동작입니다

**Q: DM으로 알림이 오나요?**
- 아니요, DM은 전송되지 않습니다
- 일반 @멘션과 동일한 방식으로 알림이 전송됩니다 (모바일 푸시, 데스크톱 알림)

**Q: HTML 태그가 보입니다**
- 최신 버전으로 업데이트 필요 (v1.0.0+)
- 플러그인 재활성화

**Q: 그룹 오너인데 그룹을 수정할 수 없습니다**
- `/group show 그룹명`으로 실제 오너 목록 확인
- 팀 관리자 또는 시스템 관리자에게 오너 권한 요청

## 📞 지원

- **이슈 리포팅**: [GitHub Issues](https://github.com/mattermost/mattermost-plugin-group-mention/issues)
- **문서**: README.md
- **라이센스**: Apache 2.0
