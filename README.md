Aliyun OSS Go SDK
=================

Installation
------------

```bash
go get -u github.com/aliyun/aliyun-oss-go-sdk/oss
go test -v -cover github.com/aliyun/aliyun-oss-go-sdk/oss
```

Highlights
----------
* Complete set of Aliyun OSS API
* 100% test coverage with intuitive table driven tests
* Passed static tests: go fmt, goimports, golint, go vet and race detector
* Idiomatic
  - named options for setting headers & parameters
  - XML is returned as a parsed struct
  - error is returned as an error object
* Easily extensible
  - clean and orthogonal implementation

Documentation
-------------

* [Quick Start](doc/quick-start.md)
* [API Object](doc/api-object.md)
* ...

Differences with Python SDK
---------------------------

* HTTP header User-Agent, e.g. aliyun-sdk-go/0.1.1 (Linux/3.16.0-51-generic/x86_64;go1.5.1)
* Go HTTP client does not support 100-Continue (will be supported after Go 1.6, see https://github.com/golang/go/issues/3665)
* Go handles HTTP better than Python
  - Go GET request does not have redundant "Content-Length: 0" header
  - Parameters will be omitted if the argument is not set
  - HTTP header keys are automatically converted into canonical format, e.g.
    x-oss-acl becomes X-Oss-Acl
  - Go always sends URL parameters and headers in canonical order

Unclear issues in Spec
----------------------

* Resource or ResourceType or both in Error XML?
* CanonicalizedResource should include only supported parameters or any parameters? (Python SDK include any parameters)

Author
------

[Hǎiliàng Wáng](https://github.com/h12w)

License
-------

licensed under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0.html)

References
----------
* [OSS Document (Chinese)](https://docs.aliyun.com/#/pub/oss)
