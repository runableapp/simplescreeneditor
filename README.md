# Simple Screen Editor

## 🇰🇷 한국어

Simple Screen Editor는 한글/영문 셀 정렬이 정확하게 유지되도록 설계된
크로스플랫폼 GUI 80x25 텍스트 편집기 프로토타입입니다.

### 기술 스택

- 버퍼/커서/문자폭 로직: Go 코어 (`internal/editor`)
- 데스크톱 셸: Wails (`main.go`) - Linux/macOS/Windows
- 렌더러: HTML 기반 고정 80x25 그리드 (`frontend/dist/index.html`)
- 기본 UI 글꼴: `IyagiGGC` 웹폰트 (로컬 폴백 포함)

### 크레딧

- 글꼴 `IyagiGGC`: PJW48 "이야기 굵은체 복각 프로젝트" (<https://pjw48.net/iyagiggc/>)

### 개발

```bash
task setup
task test
task lint
task run
```

### 라이선스

GNU 일반 공중 사용 허가서(GPL) 버전 3 또는 그 이후 버전(GPL-3.0-or-later)을 따릅니다.
자세한 내용은 <https://www.gnu.org/licenses/>를 참고하십시오.

---

## 🇺🇸 English

Cross-platform GUI 80x25 text editor prototype focused on deterministic
Korean/English cell alignment.

## Stack

- Go core (`internal/editor`) for buffer, cursor, and width logic
- Wails desktop shell (`main.go`) for Linux/macOS/Windows
- HTML renderer (`frontend/dist/index.html`) with fixed 80x25 grid
- Primary UI font: `IyagiGGC` webfont (with local fallbacks)

## Credits

- `IyagiGGC` font: PJW48 "이야기 굵은체 복각 프로젝트" (<https://pjw48.net/iyagiggc/>)


## Development

```bash
task setup
task test
task lint
task run
```

## License

This project is licensed under the GNU General Public License version 3
or later (GPL-3.0-or-later). See <https://www.gnu.org/licenses/> for details.
