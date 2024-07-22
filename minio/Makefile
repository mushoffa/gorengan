.PHONY: test-bucket test-object

test-all: test-bucket test-object

test-bucket:
	go test -run "^TestBucket_" -v

test-object: test-object-put test-object-get

test-object-put:
	go test -run "^TestPutObject_" -v

test-object-get:
	go test -run "^TestGetObject_" -v