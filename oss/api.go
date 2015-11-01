package oss

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

// API is the entry object for all OSS methods
type API struct {
	endPoint        string
	accessKeyID     string
	accessKeySecret string
	now             func() time.Time
	client          *http.Client
}

// New creates an API object
func New(endPoint, accessKeyID, accessKeySecret string) *API {
	return &API{
		endPoint:        endPoint,
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		now:             time.Now,
		client:          http.DefaultClient,
	}
}

// GetService list all buckets
func (a *API) GetService() (res *ListAllMyBucketsResult, _ error) {
	return res, a.do("GET", "", &res)
}

// PutBucket creates a new bucket
func (a *API) PutBucket(name string, acl ACLType) error {
	return a.do("PUT", name+"/", nil, ACL(acl))
}

func (a *API) GetBucket(name string, options ...Option) (res *ListBucketResult, _ error) {
	return res, a.do("GET", name+"/", &res, options...)
}

func (a *API) GetBucketACL(name string) (res *AccessControlPolicy, _ error) {
	return res, a.do("GET", name+"/?acl", &res)
}

func (a *API) GetBucketLocation(name string) (res *LocationConstraint, _ error) {
	return res, a.do("GET", name+"/?location", &res)
}

func (a *API) DeleteBucket(name string) error {
	return a.do("DELETE", name+"/", nil)
}

func (a *API) GetObject(bucket, object string, w io.Writer, options ...Option) error {
	return a.do("GET", bucket+"/"+object, &writerResult{w})
}

func (a *API) GetObjectToFile(bucket, object, fileName string, options ...Option) error {
	w, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer w.Close()
	return a.GetObject(bucket, object, w, options...)
}

func (a *API) PutObject(bucket, object string, rd io.Reader, options ...Option) error {
	return a.do("PUT", bucket+"/"+object, nil, append([]Option{httpBody(rd)}, options...)...)
}

func (a *API) PutObjectFromString(bucket, object, str string, options ...Option) error {
	return a.PutObject(bucket, object, strings.NewReader(str), options...)
}

func (a *API) PutObjectFromFile(bucket, object, file string, options ...Option) error {
	rd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer rd.Close()
	return a.PutObject(bucket, object, rd, options...)
}

func (a *API) AppendObject(bucket, object string, rd io.Reader, position AppendPosition, options ...Option) (res AppendPosition, _ error) {
	return res, a.do("POST", fmt.Sprintf("%s/%s?append&position=%d", bucket, object, position), &res, append([]Option{httpBody(rd)}, options...)...)
}

func (a *API) AppendObjectFromFile(bucket, object, file string, position AppendPosition, options ...Option) (res AppendPosition, _ error) {
	rd, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	defer rd.Close()
	return a.AppendObject(bucket, object, rd, position, options...)
}

func (a *API) HeadObject(bucket, object string) (res Header, _ error) {
	return res, a.do("HEAD", bucket+"/"+object, &res)
}

func (a *API) DeleteObject(bucket, object string) error {
	return a.do("DELETE", bucket+"/"+object, nil)
}

func (a *API) DeleteObjects(bucket string, quiet bool, objects ...string) (res *DeleteResult, _ error) {
	return res, a.do("POST", bucket+"/?delete", &res, xmlBody(newDelete(objects, quiet)), ContentMD5)
}

func (a *API) CopyObject(sourceBucket, sourceObject, targetBucket, targetObject string, options ...Option) (res *CopyObjectResult, _ error) {
	return res, a.do("PUT", targetBucket+"/"+targetObject, &res, append(options, CopySource(sourceBucket, sourceObject))...)
}

func (a *API) InitUpload(bucket, object string, options ...Option) (res *InitiateMultipartUploadResult, _ error) {
	return res, a.do("POST", bucket+"/"+object+"?uploads", &res, append(options, ContentType("application/octet-stream"))...)
}

func (a *API) UploadPart(bucket, object string, uploadID string, partNumber int, rd io.Reader, size int64) (res *UploadPartResult, _ error) {
	return res, a.do("PUT", fmt.Sprintf("%s/%s?partNumber=%d&uploadId=%s", bucket, object, partNumber, uploadID), &res, httpBody(&io.LimitedReader{R: rd, N: size}), ContentLength(size))
}

func (a *API) CompleteUpload(bucket, object string, uploadID string, list *CompleteMultipartUpload) (res *CompleteMultipartUploadResult, _ error) {
	return res, a.do("POST", fmt.Sprintf("%s/%s?uploadId=%s", bucket, object, uploadID), &res, xmlBody(list), ContentMD5)
}

func (a *API) CancelUpload(bucket, object string, uploadID string) error {
	return a.do("DELETE", fmt.Sprintf("%s/%s?uploadId=%s", bucket, object, uploadID), nil)
}

func (a *API) ListUploads(bucket, object string) (res *ListMultipartUploadsResult, _ error) {
	return res, a.do("GET", bucket+"/"+"?uploads", &res)
}

func (a *API) ListParts(bucket, object, uploadID string) (res *ListPartsResult, _ error) {
	return res, a.do("GET", fmt.Sprintf("%s/%s?uploadId=%s", bucket, object, uploadID), &res)
}

func (a *API) PutCORS(bucket string, cors *CORSConfiguration) error {
	return a.do("PUT", bucket+"/?cors", nil, xmlBody(cors), ContentMD5)
}

func (a *API) GetCORS(bucket string) (res *CORSConfiguration, _ error) {
	return res, a.do("GET", bucket+"/?cors", &res)
}

func (a *API) DeleteCORS(bucket string) error {
	return a.do("DELETE", bucket+"/?cors", nil)
}

func (a *API) PutLifecycle(bucket string, lifecycle *LifecycleConfiguration) error {
	return a.do("PUT", bucket+"/?lifecycle", nil, xmlBody(lifecycle))
}

func (a *API) GetLifecycle(bucket string) (res *LifecycleConfiguration, _ error) {
	return res, a.do("GET", bucket+"/?lifecycle", &res)
}

func (a *API) do(method, resource string, result interface{}, options ...Option) error {
	req, err := a.newRequest(method, resource, options)
	if err != nil {
		return err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return a.handleResponse(resp, result)
}

func (a *API) newRequest(method, resource string, options []Option) (*http.Request, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("http://%s/%s", a.endPoint, resource), nil)
	if err != nil {
		return nil, err
	}
	for _, option := range options {
		if err := option(req); err != nil {
			return nil, err
		}
	}
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Date", a.now().UTC().Format(gmtTime))
	req.Header.Set("User-Agent", userAgent)
	auth := authorization{req: req, secret: []byte(a.accessKeySecret)}
	req.Header.Set("Authorization", "OSS "+a.accessKeyID+":"+auth.value())
	return req, nil
}

func (a *API) handleResponse(resp *http.Response, result interface{}) error {
	if resp.StatusCode/100 > 2 {
		return parseError(resp)
	}
	if result == nil {
		return nil
	}
	if v := reflect.ValueOf(result); v.Kind() == reflect.Ptr {
		if v.Elem().Kind() == reflect.Ptr {
			v = v.Elem()
			v.Set(reflect.New(v.Type().Elem()))
			result = v.Interface()
		}
	}
	if respParser, ok := result.(responseParser); ok {
		return respParser.parse(resp)
	}
	panic(fmt.Sprintf("result %#v should implement responseParser interface", result))
}
