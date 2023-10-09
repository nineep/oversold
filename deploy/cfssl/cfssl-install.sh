#!/bin/bash

cd /usr/local/src

wget https://github.com/cloudflare/cfssl/releases/download/v1.6.1/cfssl_1.6.1_linux_amd64

wget https://github.com/cloudflare/cfssl/releases/download/v1.6.1/cfssljson_1.6.1_linux_amd64

wget https://github.com/cloudflare/cfssl/releases/download/v1.6.1/cfssl-certinfo_1.6.1_linux_amd64

#拷贝文件到/usr/local/bin/目录下

cp cfssl_1.6.1_linux_amd64 /usr/local/bin/cfssl

cp cfssljson_1.6.1_linux_amd64 /usr/local/bin/cfssljson

cp cfssl-certinfo_1.6.1_linux_amd64 /usr/local/bin/cfssl-certinfo

#添加执行权限

chmod +x /usr/local/bin/cfssl

chmod +x /usr/local/bin/cfssljson

chmod +x /usr/local/bin/cfssl-certinfo