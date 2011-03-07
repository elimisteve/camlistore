/*
Copyright 2011 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package localdisk

import (
	"camli/blobref"
	"camli/blobserver"
	. "camli/testing"
	"os"
	"testing"
	"time"
)

var defaultPartition blobserver.Partition = nil

func TestEnumerate(t *testing.T) {
	ds := NewStorage(t)
	defer cleanUp(ds)

	// For test simplicity foo, bar, and baz all have ascending
	// sha1s and lengths.
	foo := &testBlob{"foo"}   // 0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33
	bar := &testBlob{"baar"}  // b23361951dde70cb3eca44c0c674181673a129dc
	baz := &testBlob{"bazzz"} // e0eb17003ce1c2812ca8f19089fff44ca32b3710
	foo.ExpectUploadBlob(t, ds)
	bar.ExpectUploadBlob(t, ds)
	baz.ExpectUploadBlob(t, ds)

	limit := uint(5000)
	waitSeconds := 0
	ch := make(chan *blobref.SizedBlobRef)
	errCh := make(chan os.Error)
	go func() {
		errCh <- ds.EnumerateBlobs(ch, defaultPartition, "", limit, waitSeconds)
	}()

	var sb *blobref.SizedBlobRef
	sb = <-ch
	Assert(t, sb != nil, "got 1st blob")
	ExpectInt(t, 3, int(sb.Size), "1st blob size")
	sb = <-ch
	Assert(t, sb != nil, "got 2nd blob")
	ExpectInt(t, 4, int(sb.Size), "2nd blob size")
	sb = <-ch
	Assert(t, sb != nil, "got 3rd blob")
	ExpectInt(t, 5, int(sb.Size), "3rd blob size")
	sb = <-ch
	Assert(t, sb == nil, "got final nil")
	ExpectNil(t, <-errCh, "EnumerateBlobs return value")

	// Now again, but skipping foo's blob
	go func() {
		errCh <- ds.EnumerateBlobs(ch, defaultPartition,
			foo.BlobRef().String(),
			limit, waitSeconds)
	}()
	sb = <-ch
        Assert(t, sb != nil, "got 1st blob, skipping foo")
        ExpectInt(t, 4, int(sb.Size), "blob size")
        sb = <-ch
        Assert(t, sb != nil, "got 2nd blob, skipping foo")
        ExpectInt(t, 5, int(sb.Size), "blob size")
        sb = <-ch
        Assert(t, sb == nil, "got final nil")
        ExpectNil(t, <-errCh, "EnumerateBlobs return value")
}

func TestEnumerateEmpty(t *testing.T) {
	ds := NewStorage(t)
	defer cleanUp(ds)

	limit := uint(5000)
	waitSeconds := 0
	ch := make(chan *blobref.SizedBlobRef)
	errCh := make(chan os.Error)
	go func() {
		errCh <- ds.EnumerateBlobs(ch, defaultPartition,
			"", limit, waitSeconds)
	}()

	Expect(t, (<-ch) == nil, "no first blob")
	ExpectNil(t, <-errCh, "EnumerateBlobs return value")
}

func TestEnumerateEmptyLongPoll(t *testing.T) {
	ds := NewStorage(t)
	defer cleanUp(ds)

	limit := uint(5000)
	waitSeconds := 1
	ch := make(chan *blobref.SizedBlobRef)
	errCh := make(chan os.Error)
	go func() {
		errCh <- ds.EnumerateBlobs(ch, defaultPartition,
			"", limit, waitSeconds)
	}()

	foo := &testBlob{"foo"}   // 0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33
	go func() {
		time.Sleep(100e6)  // 100 ms
		foo.ExpectUploadBlob(t, ds)
	}()

	sb := <-ch
        Assert(t, sb != nil, "got a blob")
        ExpectInt(t, 3, int(sb.Size), "blob size")
	ExpectString(t, "sha1-0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33", sb.BlobRef.String(), "got the right blob")

	Expect(t, (<-ch) == nil, "only one blob returned")
	ExpectNil(t, <-errCh, "EnumerateBlobs return value")
}

