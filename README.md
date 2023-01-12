# simple, toy, unsecure ssh certificate CA


usage: `ssh_cert_ca -c /path/to/config.json`

- To sign a host key: 
```
curl -X POST -H "Authorization: Bearer <token>" -F 'pubkey=@/path/to/ssh_host_key.pub' "http://<ca server address>/ca/sign/host?signto=<list of hosts>"&ttl=<expire time in seconds>
```

- To sign a user key: 
```
curl -X POST -H "Authorization: Bearer <token>" -F 'pubkey=@/path/to/ssh_user_key.pub' "http://<ca server address>/ca/sign/user?signto=<user>"&ttl=<expire time in seconds>
```

- To get host CA public key
```
curl -X GET -H "Authorization: Bearer <token>" "http://<ca server address>/ca/capubkey/host" 
``` 

- To get user CA public key
```
curl -X GET -H "Authorization: Bearer <token>" "http://<ca server address>/ca/capubkey/user" 
``` 
 

### Notes:
- The default TTL of host and user public key is 8 hours.

### Quick Start

- server side
1. get the user CA public key with the token and server address found in `config.json`
for example:
```
$ curl -X GET -H "Authorization: Bearer 7741a1348cfee150506faf96017f6723c3c114eb" "http://127.0.0.1:8077/ca/capubkey/user" | jq

{
  "code": 0,
  "errMsg": "OK",
  "data": "ecdsa-sha2-nistp521 AAAAE2VjZHNhLXNoYTItbmlzdHA1MjEAAAAIbmlzdHA1MjEAAACFBAGjbXDuSF/xkIxpN5UqdHXIpIdynFdA+X5RpIO/YiSiujBZUAvmxrBVupLWfOlC3kq8uUAcQXBQLvJGxhr0tNscRAGfHAzNV5Bsk0gJl9b10hdDMwFKhWsvQ/aDLUu/7xNdk48YB+dR+aFVJ3aS3l4nRNgLb3U+owgghd9OcZVeAauWrA=="
}
```

2. save the content of `data` as file and put the following line on `/etc/ssh/sshd_config`:
```
TrustedUserCAKeys /path/to/userca.pub
``` 

3. sign your server host key and save it to `/etc/ssh/your_host_key-cert.pub`, an example to sign `/etc/ssh/ssh_host_rsa_key.pub` 
with list of host `a,b,c`:
```
$ curl -X POST -H "Authorization: Bearer 7741a1348cfee150506faf96017f6723c3c114eb" -F 'pubkey=@/etc/ssh/ssh_host_rsa_key.pub' "http://127.0.0.1:8077/ca/sign/host?signto=a,b,c" | jq -r '.data.cert_content' | sudo tee /etc/ssh/ssh_host_rsa_key-cert.pub

```

4. put the following line on `/etc/ssh/sshd_config`, assume your host key certificate signed by CA saved to `/path/to/your_host_key-cert.pub`:
```
HostCertificate /path/to/your_host_key-cert.pub
```

5. repeat step 3 and 4 if you have more host key to sign

- client side
1. get the host CA public key with your token, for example
```
$ curl -X GET -H "Authorization: Bearer 7741a1348cfee150506faf96017f6723c3c114eb" "http://127.0.0.1:8077/ca/capubkey/host" | jq

{
  "code": 0,
  "errMsg": "OK",
  "data": "ecdsa-sha2-nistp521 AAAAE2VjZHNhLXNoYTItbmlzdHA1MjEAAAAIbmlzdHA1MjEAAACFBAFK6RtiAgVCGLX1XC2KJxJ0p8FzhaqmakyCWzxiFOoN+7mZyWlwbr3zqCWpdWkw6ZFMann8LRMFA1QDYghFpNY52QB3UvOF3y3xUdFNgk9zyyPwYGH4ln2Xoes90qYPd7ckXur5C2/72PFd+GylR0Bu/aIN3RiogWGCqy4SfkhgRXN8Rg=="
}
```

2. store the public key to `~/.ssh/known_hosts`, for example, the hosts you signed were `a,b,c`:
```
@cert-authority a,b,c ecdsa-sha2-nistp521 AAAAE2VjZHNhLXNoYTItbmlzdHA1MjEAAAAIbmlzdHA1MjEAAACFBAFK6RtiAgVCGLX1XC2KJxJ0p8FzhaqmakyCWzxiFOoN+7mZyWlwbr3zqCWpdWkw6ZFMann8LRMFA1QDYghFpNY52QB3UvOF3y3xUdFNgk9zyyPwYGH4ln2Xoes90qYPd7ckXur5C2/72PFd+GylR0Bu/aIN3RiogWGCqy4SfkhgRXN8Rg==
```