/*
 * Minio Go Library for Amazon S3 compatible cloud storage (C) 2015 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package minio_test

import (
	"bytes"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/minio/minio-go"
)

func TestBucketOperations(t *testing.T) {
	bucket := bucketHandler(bucketHandler{
		resource: "/bucket",
	})
	server := httptest.NewServer(bucket)
	defer server.Close()

	a, err := minio.New(minio.Config{Endpoint: server.URL})
	if err != nil {
		t.Errorf("Error")
	}
	err = a.MakeBucket("bucket", "private")
	if err != nil {
		t.Errorf("Error")
	}

	err = a.BucketExists("bucket")
	if err != nil {
		t.Errorf("Error")
	}

	err = a.BucketExists("bucket1")
	if err == nil {
		t.Errorf("Error")
	}
	if err.Error() != "Access Denied" {
		t.Errorf("Error")
	}

	err = a.SetBucketACL("bucket", "public-read-write")
	if err != nil {
		t.Errorf("Error")
	}

	acl, err := a.GetBucketACL("bucket")
	if err != nil {
		t.Errorf("Error")
	}
	if acl != minio.BucketACL("private") {
		t.Fatalf("Error")
	}

	for b := range a.ListBuckets() {
		if b.Err != nil {
			t.Fatalf(b.Err.Error())
		}
		if b.Stat.Name != "bucket" {
			t.Errorf("Error")
		}
	}

	for o := range a.ListObjects("bucket", "", true) {
		if o.Err != nil {
			t.Fatalf(o.Err.Error())
		}
		if o.Stat.Key != "object" {
			t.Errorf("Error")
		}
	}

	err = a.RemoveBucket("bucket")
	if err != nil {
		t.Errorf("Error")
	}

	err = a.RemoveBucket("bucket1")
	if err == nil {
		t.Fatalf("Error")
	}
	if err.Error() != "The specified bucket does not exist." {
		t.Errorf("Error")
	}
}

func TestBucketOperationsFail(t *testing.T) {
	bucket := bucketHandler(bucketHandler{
		resource: "/bucket",
	})
	server := httptest.NewServer(bucket)
	defer server.Close()

	a, err := minio.New(minio.Config{Endpoint: server.URL})
	if err != nil {
		t.Errorf("Error")
	}
	err = a.MakeBucket("bucket$$$", "private")
	if err == nil {
		t.Errorf("Error")
	}

	err = a.BucketExists("bucket.")
	if err == nil {
		t.Errorf("Error")
	}

	err = a.SetBucketACL("bucket-.", "public-read-write")
	if err == nil {
		t.Errorf("Error")
	}

	_, err = a.GetBucketACL("bucket??")
	if err == nil {
		t.Errorf("Error")
	}

	for o := range a.ListObjects("bucket??", "", true) {
		if o.Err == nil {
			t.Fatalf(o.Err.Error())
		}
	}

	err = a.RemoveBucket("bucket??")
	if err == nil {
		t.Errorf("Error")
	}

	if err.Error() != "The specified bucket is not valid." {
		t.Errorf("Error")
	}
}

func TestObjectOperations(t *testing.T) {
	object := objectHandler(objectHandler{
		resource: "/bucket/object",
		data:     []byte("Hello, World"),
	})
	server := httptest.NewServer(object)
	defer server.Close()

	a, err := minio.New(minio.Config{Endpoint: server.URL})
	if err != nil {
		t.Fatalf("Error")
	}
	data := []byte("Hello, World")
	err = a.PutObject("bucket", "object", "", int64(len(data)), bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Error")
	}
	metadata, err := a.StatObject("bucket", "object")
	if err != nil {
		t.Fatalf("Error")
	}
	if metadata.Key != "object" {
		t.Fatalf("Error")
	}
	if metadata.ETag != "9af2f8218b150c351ad802c6f3d66abe" {
		t.Fatalf("Error")
	}

	reader, metadata, err := a.GetObject("bucket", "object")
	if err != nil {
		t.Fatalf("Error")
	}
	if metadata.Key != "object" {
		t.Fatalf("Error")
	}
	if metadata.ETag != "9af2f8218b150c351ad802c6f3d66abe" {
		t.Fatalf("Error")
	}

	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, reader)
	if !bytes.Equal(buffer.Bytes(), data) {
		t.Fatalf("Error")
	}

	err = a.RemoveObject("bucket", "object")
	if err != nil {
		t.Fatalf("Error")
	}
	err = a.RemoveObject("bucket", "object1")
	if err == nil {
		t.Fatalf("Error")
	}
	if err.Error() != "The specified key does not exist." {
		t.Errorf("Error")
	}
}

func TestPresignedURL(t *testing.T) {
	object := objectHandler(objectHandler{
		resource: "/bucket/object",
		data:     []byte("Hello, World"),
	})
	server := httptest.NewServer(object)
	defer server.Close()

	a, err := minio.New(minio.Config{Endpoint: server.URL})
	if err != nil {
		t.Fatalf("Error")
	}
	// should error out for invalid access keys
	_, err = a.PresignedGetObject("bucket", "object", time.Duration(1000)*time.Second)
	if err == nil {
		t.Fatalf("Error")
	}

	a, err = minio.New(minio.Config{
		Endpoint:        server.URL,
		AccessKeyID:     "accessKey",
		SecretAccessKey: "secretKey",
	})
	if err != nil {
		t.Fatalf("Error")
	}
	url, err := a.PresignedGetObject("bucket", "object", time.Duration(1000)*time.Second)
	if err != nil {
		t.Fatalf("Error")
	}
	if url == "" {
		t.Fatalf("Error")
	}
	url, err = a.PresignedGetPartialObject("bucket", "object", time.Duration(1000)*time.Second, 5, 11)
	if err != nil {
		t.Fatalf("Error")
	}
	if url == "" {
		t.Fatalf("Error")
	}
	_, err = a.PresignedGetObject("bucket", "object", time.Duration(0)*time.Second)
	if err == nil {
		t.Fatalf("Error")
	}
	_, err = a.PresignedGetObject("bucket", "object", time.Duration(604801)*time.Second)
	if err == nil {
		t.Fatalf("Error")
	}
}

func TestErrorResponse(t *testing.T) {
	errorResponse := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?><Error><Code>AccessDenied</Code><Message>Access Denied</Message><Resource>/mybucket/myphoto.jpg</Resource><RequestId>F19772218238A85A</RequestId><HostId>GuWkjyviSiGHizehqpmsD1ndz5NClSP19DOT+s2mv7gXGQ8/X1lhbDGiIJEXpGFD</HostId></Error>")
	errorReader := bytes.NewReader(errorResponse)
	err := minio.BodyToErrorResponse(errorReader, "application/xml")
	if err == nil {
		t.Fatal("Error")
	}
	if err.Error() != "Access Denied" {
		t.Fatal("Error")
	}
	resp := minio.ToErrorResponse(err)
	// valid all fields
	if resp == nil {
		t.Fatal("Error")
	}
	if resp.Code != "AccessDenied" {
		t.Fatal("Error")
	}
	if resp.RequestID != "F19772218238A85A" {
		t.Fatal("Error")
	}
	if resp.Message != "Access Denied" {
		t.Fatal("Error")
	}
	if resp.Resource != "/mybucket/myphoto.jpg" {
		t.Fatal("Error")
	}
	if resp.HostID != "GuWkjyviSiGHizehqpmsD1ndz5NClSP19DOT+s2mv7gXGQ8/X1lhbDGiIJEXpGFD" {
		t.Fatal("Error")
	}
	if resp.ToXML() == "" {
		t.Fatal("Error")
	}
	if resp.ToJSON() == "" {
		t.Fatal("Error")
	}
}