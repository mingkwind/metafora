filename=test.txt && curl -v --upload-file $filename -H "Digest: SHA-256=$(sha256sum $filename|awk '{print $1}')" http://localhost:8888/objects/$filename