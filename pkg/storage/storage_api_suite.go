package storage

import (
	"bytes"
	"math/rand"
	"strconv"

	. "gopkg.in/check.v1"
)

func APITestSuite(c *C, create func() Storage) {
	testCreateBucket(c, create)
	testMultipleObjectCreation(c, create)
	testPaging(c, create)
	testObjectOverwriteFails(c, create)
	testNonExistantBucketOperations(c, create)
	testBucketRecreateFails(c, create)
}

func testCreateBucket(c *C, create func() Storage) {
	// TODO
}

func testMultipleObjectCreation(c *C, create func() Storage) {
	objects := make(map[string][]byte)
	storage := create()
	err := storage.StoreBucket("bucket")
	c.Assert(err, IsNil)
	for i := 0; i < 10; i++ {
		randomPerm := rand.Perm(10)
		randomString := ""
		for _, num := range randomPerm {
			randomString = randomString + strconv.Itoa(num)
		}
		key := "obj" + strconv.Itoa(i)
		objects[key] = []byte(randomString)
		err := storage.StoreObject("bucket", key, bytes.NewBufferString(randomString))
		c.Assert(err, IsNil)
	}

	// ensure no duplicate etags
	etags := make(map[string]string)
	for key, value := range objects {
		var byteBuffer bytes.Buffer
		storage.CopyObjectToWriter(&byteBuffer, "bucket", key)
		c.Assert(bytes.Equal(value, byteBuffer.Bytes()), Equals, true)

		metadata, err := storage.GetObjectMetadata("bucket", key)
		c.Assert(err, IsNil)
		c.Assert(metadata.Size, Equals, int64(len(value)))

		_, ok := etags[metadata.ETag]
		c.Assert(ok, Equals, false)
		etags[metadata.ETag] = metadata.ETag
	}
}

func testPaging(c *C, create func() Storage) {
	storage := create()
	storage.StoreBucket("bucket")
	storage.ListObjects("bucket", "", 5)
	objects, isTruncated, err := storage.ListObjects("bucket", "", 5)
	c.Assert(len(objects), Equals, 0)
	c.Assert(isTruncated, Equals, false)
	c.Assert(err, IsNil)
	// check before paging occurs
	for i := 0; i < 5; i++ {
		key := "obj" + strconv.Itoa(i)
		storage.StoreObject("bucket", key, bytes.NewBufferString(key))
		objects, isTruncated, err = storage.ListObjects("bucket", "", 5)
		c.Assert(len(objects), Equals, i+1)
		c.Assert(isTruncated, Equals, false)
		c.Assert(err, IsNil)
	}
	// check after paging occurs pages work
	for i := 6; i <= 10; i++ {
		key := "obj" + strconv.Itoa(i)
		storage.StoreObject("bucket", key, bytes.NewBufferString(key))
		objects, isTruncated, err = storage.ListObjects("bucket", "", 5)
		c.Assert(len(objects), Equals, 5)
		c.Assert(isTruncated, Equals, true)
		c.Assert(err, IsNil)
	}
	// check paging with prefix at end returns less objects
	{
		storage.StoreObject("bucket", "newPrefix", bytes.NewBufferString("prefix1"))
		storage.StoreObject("bucket", "newPrefix2", bytes.NewBufferString("prefix2"))
		objects, isTruncated, err = storage.ListObjects("bucket", "new", 5)
		c.Assert(len(objects), Equals, 2)
	}

	// check ordering of pages
	{
		objects, isTruncated, err = storage.ListObjects("bucket", "", 5)
		c.Assert(objects[0].Key, Equals, "newPrefix")
		c.Assert(objects[1].Key, Equals, "newPrefix2")
		c.Assert(objects[2].Key, Equals, "obj0")
		c.Assert(objects[3].Key, Equals, "obj1")
		c.Assert(objects[4].Key, Equals, "obj10")
	}
	// check ordering of results with prefix
	{
		objects, isTruncated, err = storage.ListObjects("bucket", "obj", 5)
		c.Assert(objects[0].Key, Equals, "obj0")
		c.Assert(objects[1].Key, Equals, "obj1")
		c.Assert(objects[2].Key, Equals, "obj10")
		c.Assert(objects[3].Key, Equals, "obj2")
		c.Assert(objects[4].Key, Equals, "obj3")
	}
	// check ordering of results with prefix and no paging
	{
		objects, isTruncated, err = storage.ListObjects("bucket", "new", 5)
		c.Assert(objects[0].Key, Equals, "newPrefix")
		c.Assert(objects[1].Key, Equals, "newPrefix2")
	}
}

func testObjectOverwriteFails(c *C, create func() Storage) {
	storage := create()
	storage.StoreBucket("bucket")
	err := storage.StoreObject("bucket", "object", bytes.NewBufferString("one"))
	c.Assert(err, IsNil)
	err = storage.StoreObject("bucket", "object", bytes.NewBufferString("three"))
	c.Assert(err, Not(IsNil))
	var bytesBuffer bytes.Buffer
	length, err := storage.CopyObjectToWriter(&bytesBuffer, "bucket", "object")
	c.Assert(length, Equals, int64(len("one")))
	c.Assert(err, IsNil)
	c.Assert(string(bytesBuffer.Bytes()), Equals, "one")
}

func testNonExistantBucketOperations(c *C, create func() Storage) {
	storage := create()
	err := storage.StoreObject("bucket", "object", bytes.NewBufferString("one"))
	c.Assert(err, Not(IsNil))
}

func testBucketRecreateFails(c *C, create func() Storage) {
	storage := create()
	err := storage.StoreBucket("string")
	c.Assert(err, IsNil)
	err = storage.StoreBucket("string")
	c.Assert(err, Not(IsNil))
}