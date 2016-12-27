// Autogenerated by Thrift Compiler (0.9.1)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package tutorial

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"math"
)

// (needed to ensure safety because of naive import list construction.)
var _ = math.MinInt32
var _ = thrift.ZERO
var _ = fmt.Printf

type MultiplicationService interface {
	// Parameters:
	//  - N1
	//  - N2
	Multiply(n1 Int, n2 Int) (r Int, err error)
}

type MultiplicationServiceClient struct {
	Transport       thrift.TTransport
	ProtocolFactory thrift.TProtocolFactory
	InputProtocol   thrift.TProtocol
	OutputProtocol  thrift.TProtocol
	SeqId           int32
}

func NewMultiplicationServiceClientFactory(t thrift.TTransport, f thrift.TProtocolFactory) *MultiplicationServiceClient {
	return &MultiplicationServiceClient{Transport: t,
		ProtocolFactory: f,
		InputProtocol:   f.GetProtocol(t),
		OutputProtocol:  f.GetProtocol(t),
		SeqId:           0,
	}
}

func NewMultiplicationServiceClientProtocol(t thrift.TTransport, iprot thrift.TProtocol, oprot thrift.TProtocol) *MultiplicationServiceClient {
	return &MultiplicationServiceClient{Transport: t,
		ProtocolFactory: nil,
		InputProtocol:   iprot,
		OutputProtocol:  oprot,
		SeqId:           0,
	}
}

// Parameters:
//  - N1
//  - N2
func (p *MultiplicationServiceClient) Multiply(n1 Int, n2 Int) (r Int, err error) {
	if err = p.sendMultiply(n1, n2); err != nil {
		return
	}
	return p.recvMultiply()
}

func (p *MultiplicationServiceClient) sendMultiply(n1 Int, n2 Int) (err error) {
	oprot := p.OutputProtocol
	if oprot == nil {
		oprot = p.ProtocolFactory.GetProtocol(p.Transport)
		p.OutputProtocol = oprot
	}
	p.SeqId++
	oprot.WriteMessageBegin("multiply", thrift.CALL, p.SeqId)
	args0 := NewMultiplyArgs()
	args0.N1 = n1
	args0.N2 = n2
	err = args0.Write(oprot)
	oprot.WriteMessageEnd()
	oprot.Flush()
	return
}

func (p *MultiplicationServiceClient) recvMultiply() (value Int, err error) {
	iprot := p.InputProtocol
	if iprot == nil {
		iprot = p.ProtocolFactory.GetProtocol(p.Transport)
		p.InputProtocol = iprot
	}
	_, mTypeId, seqId, err := iprot.ReadMessageBegin()
	if err != nil {
		return
	}
	if mTypeId == thrift.EXCEPTION {
		error2 := thrift.NewTApplicationException(thrift.UNKNOWN_APPLICATION_EXCEPTION, "Unknown Exception")
		var error3 error
		error3, err = error2.Read(iprot)
		if err != nil {
			return
		}
		if err = iprot.ReadMessageEnd(); err != nil {
			return
		}
		err = error3
		return
	}
	if p.SeqId != seqId {
		err = thrift.NewTApplicationException(thrift.BAD_SEQUENCE_ID, "ping failed: out of sequence response")
		return
	}
	result1 := NewMultiplyResult()
	err = result1.Read(iprot)
	iprot.ReadMessageEnd()
	value = result1.Success
	return
}

type MultiplicationServiceProcessor struct {
	processorMap map[string]thrift.TProcessorFunction
	handler      MultiplicationService
}

func (p *MultiplicationServiceProcessor) AddToProcessorMap(key string, processor thrift.TProcessorFunction) {
	p.processorMap[key] = processor
}

func (p *MultiplicationServiceProcessor) GetProcessorFunction(key string) (processor thrift.TProcessorFunction, ok bool) {
	processor, ok = p.processorMap[key]
	return processor, ok
}

func (p *MultiplicationServiceProcessor) ProcessorMap() map[string]thrift.TProcessorFunction {
	return p.processorMap
}

func NewMultiplicationServiceProcessor(handler MultiplicationService) *MultiplicationServiceProcessor {

	self4 := &MultiplicationServiceProcessor{handler: handler, processorMap: make(map[string]thrift.TProcessorFunction)}
	self4.processorMap["multiply"] = &multiplicationServiceProcessorMultiply{handler: handler}
	return self4
}

func (p *MultiplicationServiceProcessor) Process(iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
	name, _, seqId, err := iprot.ReadMessageBegin()
	if err != nil {
		return false, err
	}
	if processor, ok := p.GetProcessorFunction(name); ok {
		return processor.Process(seqId, iprot, oprot)
	}
	iprot.Skip(thrift.STRUCT)
	iprot.ReadMessageEnd()
	x5 := thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+name)
	oprot.WriteMessageBegin(name, thrift.EXCEPTION, seqId)
	x5.Write(oprot)
	oprot.WriteMessageEnd()
	oprot.Flush()
	return false, x5

}

type multiplicationServiceProcessorMultiply struct {
	handler MultiplicationService
}

func (p *multiplicationServiceProcessorMultiply) Process(seqId int32, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
	args := NewMultiplyArgs()
	if err = args.Read(iprot); err != nil {
		iprot.ReadMessageEnd()
		x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err.Error())
		oprot.WriteMessageBegin("multiply", thrift.EXCEPTION, seqId)
		x.Write(oprot)
		oprot.WriteMessageEnd()
		oprot.Flush()
		return
	}
	iprot.ReadMessageEnd()
	result := NewMultiplyResult()
	if result.Success, err = p.handler.Multiply(args.N1, args.N2); err != nil {
		x := thrift.NewTApplicationException(thrift.INTERNAL_ERROR, "Internal error processing multiply: "+err.Error())
		oprot.WriteMessageBegin("multiply", thrift.EXCEPTION, seqId)
		x.Write(oprot)
		oprot.WriteMessageEnd()
		oprot.Flush()
		return
	}
	if err2 := oprot.WriteMessageBegin("multiply", thrift.REPLY, seqId); err2 != nil {
		err = err2
	}
	if err2 := result.Write(oprot); err == nil && err2 != nil {
		err = err2
	}
	if err2 := oprot.WriteMessageEnd(); err == nil && err2 != nil {
		err = err2
	}
	if err2 := oprot.Flush(); err == nil && err2 != nil {
		err = err2
	}
	if err != nil {
		return
	}
	return true, err
}

// HELPER FUNCTIONS AND STRUCTURES

type MultiplyArgs struct {
	N1 Int `thrift:"n1,1"`
	N2 Int `thrift:"n2,2"`
}

func NewMultiplyArgs() *MultiplyArgs {
	return &MultiplyArgs{}
}

func (p *MultiplyArgs) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return fmt.Errorf("%T read error", p)
	}
	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return fmt.Errorf("%T field %d read error: %s", p, fieldId, err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if err := p.readField1(iprot); err != nil {
				return err
			}
		case 2:
			if err := p.readField2(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return fmt.Errorf("%T read struct end error: %s", p, err)
	}
	return nil
}

func (p *MultiplyArgs) readField1(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI32(); err != nil {
		return fmt.Errorf("error reading field 1: %s")
	} else {
		p.N1 = Int(v)
	}
	return nil
}

func (p *MultiplyArgs) readField2(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI32(); err != nil {
		return fmt.Errorf("error reading field 2: %s")
	} else {
		p.N2 = Int(v)
	}
	return nil
}

func (p *MultiplyArgs) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("multiply_args"); err != nil {
		return fmt.Errorf("%T write struct begin error: %s", p, err)
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := p.writeField2(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return fmt.Errorf("%T write field stop error: %s", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return fmt.Errorf("%T write struct stop error: %s", err)
	}
	return nil
}

func (p *MultiplyArgs) writeField1(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("n1", thrift.I32, 1); err != nil {
		return fmt.Errorf("%T write field begin error 1:n1: %s", p, err)
	}
	if err := oprot.WriteI32(int32(p.N1)); err != nil {
		return fmt.Errorf("%T.n1 (1) field write error: %s", p)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return fmt.Errorf("%T write field end error 1:n1: %s", p, err)
	}
	return err
}

func (p *MultiplyArgs) writeField2(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("n2", thrift.I32, 2); err != nil {
		return fmt.Errorf("%T write field begin error 2:n2: %s", p, err)
	}
	if err := oprot.WriteI32(int32(p.N2)); err != nil {
		return fmt.Errorf("%T.n2 (2) field write error: %s", p)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return fmt.Errorf("%T write field end error 2:n2: %s", p, err)
	}
	return err
}

func (p *MultiplyArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("MultiplyArgs(%+v)", *p)
}

type MultiplyResult struct {
	Success Int `thrift:"success,0"`
}

func NewMultiplyResult() *MultiplyResult {
	return &MultiplyResult{}
}

func (p *MultiplyResult) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return fmt.Errorf("%T read error", p)
	}
	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return fmt.Errorf("%T field %d read error: %s", p, fieldId, err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 0:
			if err := p.readField0(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return fmt.Errorf("%T read struct end error: %s", p, err)
	}
	return nil
}

func (p *MultiplyResult) readField0(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI32(); err != nil {
		return fmt.Errorf("error reading field 0: %s")
	} else {
		p.Success = Int(v)
	}
	return nil
}

func (p *MultiplyResult) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("multiply_result"); err != nil {
		return fmt.Errorf("%T write struct begin error: %s", p, err)
	}
	switch {
	default:
		if err := p.writeField0(oprot); err != nil {
			return err
		}
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return fmt.Errorf("%T write field stop error: %s", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return fmt.Errorf("%T write struct stop error: %s", err)
	}
	return nil
}

func (p *MultiplyResult) writeField0(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("success", thrift.I32, 0); err != nil {
		return fmt.Errorf("%T write field begin error 0:success: %s", p, err)
	}
	if err := oprot.WriteI32(int32(p.Success)); err != nil {
		return fmt.Errorf("%T.success (0) field write error: %s", p)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return fmt.Errorf("%T write field end error 0:success: %s", p, err)
	}
	return err
}

func (p *MultiplyResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("MultiplyResult(%+v)", *p)
}