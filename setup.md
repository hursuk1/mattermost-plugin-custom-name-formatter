# Node Docker 설치
1. docker pull node:20
2. docker run -it -e http_proxy=http://10.225.21.115:3128 -e https_proxy=http://10.225.21.115:3128 -v ./mattermost-plugin-starter-template:/app node:20 bash
3. cd /app

## ssl 등록
```
cp ssl_hlcompany.crt /usr/local/share/ca-certificates/
update-ca-certificates
```

# Go 설치
1. wget https://go.dev/dl/go1.21.4.linux-amd64.tar.gz -O go.tar.gz
2. rm -rf /usr/local/go && tar -C /usr/local -xzf go.tar.gz
3. echo export PATH=$HOME/go/bin:/usr/local/go/bin:$PATH >> ~/.profile
4. source ~/.profile
5. go version 통해 확인

# git 설정
1. git config --global --add safe.directory /app
<br>
<br>
<br>

# TroubleShooting(문제 해결)
- proxy 또는 인터넷 연결 체크 필요
```
npm error code ERR_INVALID_URL
npm error Invalid URL
```
- cert file 등록 필요
    - /usr/local/share/ca-certificates 로 파일 복사
    - update-ca-certificates
```
ERROR: The certificate of ~~~
```
- git이 깔려있지 않아서 생기는 문제
```
npm error git dep preparation failed
```

- ## npm install시 아래와 같은 문구가 있어도 무시해도 상관없음
```
75 vulnerabilities (3 low, 36 moderate, 31 high, 5 critical)

To address issues that do not require attention, run:
  npm audit fix

To address all issues (including breaking changes), run:
  npm audit fix --force

```