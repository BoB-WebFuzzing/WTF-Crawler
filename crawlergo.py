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
    cmd = ["bin/crawlergo", "-c", "/usr/bin/google-chrome",
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
