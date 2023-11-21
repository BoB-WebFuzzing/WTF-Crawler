# crawlergo
![chromedp](https://img.shields.io/badge/chromedp-v0.5.2-brightgreen.svg) [![BlackHat EU Arsenal](https://img.shields.io/badge/BlackHat%20Europe-2021%20Arsenal-blue.svg)](https://www.blackhat.com/eu-21/arsenal/schedule/index.html#crawlergo-a-powerful-browser-crawler-for-web-vulnerability-scanners-25113)

> A powerful browser crawler for web vulnerability scanners

## Installation

### 0. Go(Golang) 설치
```shell
sudo apt update -y
sudo apt install golang -y
go version # check go version
```

### 1. Git Clone
```shell
git clone https://github.com/BoB-WebFuzzing/WTF-crawlergo.git
cd WTF-crawlergo
```
### 2. Chromium 다운로드
ref : https://www.chromium.org/getting-involved/download-chromium/
```shell
npx @puppeteer/browsers install chrome@stable
```
### 3. Python 모듈 설치
```shell
pip3 install simplejson
```
### 4. 크롤링 시작
```shell
make build
python3 crawlergo.py
```

### Calling crawlergo with python
- target : 크롤링 대상 URL
- headers : 커스텀 헤더 설정 (쿠키 설정)
```python3
#!/usr/bin/python3
# coding: utf-8

import simplejson
import subprocess
def main():
    target = "http://testphp.vulnweb.com/"
    headers = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) "
                      "Chrome/119.0.0.0 Safari/537.36",
        "Cookie": "PHPSESSID=4f5c943a8fc68425a469e5184edabf9b; "
                  "security=low"
    }
    cmd = ["bin/crawlergo", "-c", "chrome/linux-119.0.6045.105/chrome-linux64/chrome",
           "-o", "json", "--output-json", "request_data.json", "--custom-headers", simplejson.dumps(headers),
           target]

    rsp = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output, error = rsp.communicate()
    print(output)

    result = simplejson.loads(output.decode().split("--[Mission Complete]--")[1])
    req_list = result["requestsFound"]
    for each in req_list:
        print(each)

if __name__ == '__main__':
    main()
```