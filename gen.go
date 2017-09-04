//go:generate protoc -I . -I ../../.. pqstream.proto --go_out=plugins=grpc:pqs

package pqstream
