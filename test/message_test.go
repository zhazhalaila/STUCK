package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	pb "github.com/stuck/message/messagepb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Tests whether the any type can be marshal and unmarshal properly
func TestAnyWork(t *testing.T) {
	op := &pb.Operation{}
	op.Op = pb.OpType_HandleRelationship
	payloads := &pb.AnyTest{}
	payloads.Message = "test"

	// create any instance
	any, err := anypb.New(payloads)
	assert.Nil(t, err)
	op.Payloads = any

	// marshal
	out, err := proto.Marshal(op)
	assert.Nil(t, err)

	// unmarshal
	inOp := &pb.Operation{}
	err = proto.Unmarshal(out, inOp)
	assert.Nil(t, err)

	// unmarshal any instance
	inPayloads := &pb.AnyTest{}
	err = inOp.Payloads.UnmarshalTo(inPayloads)
	assert.Nil(t, err)
	assert.Equal(t, "test", inPayloads.Message)
}
