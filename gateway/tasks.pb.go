// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v4.24.4
// source: tasks.proto

package gateway

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type TaskStatus int32

const (
	TaskStatus_NA     TaskStatus = 0 // invalid
	TaskStatus_OK     TaskStatus = 1 // OKAY
	TaskStatus_ERR    TaskStatus = 2 // Something went wrong.
	TaskStatus_QUEUED TaskStatus = 3 // Try again later
	TaskStatus_POKE   TaskStatus = 6 // Needs a poke to wake up
	TaskStatus_CANCEL TaskStatus = 7 // Cancel requested
	TaskStatus_KO     TaskStatus = 4 // Imposible to complete
	TaskStatus_SKIP   TaskStatus = 5 // Not processed, see message
)

// Enum value maps for TaskStatus.
var (
	TaskStatus_name = map[int32]string{
		0: "NA",
		1: "OK",
		2: "ERR",
		3: "QUEUED",
		6: "POKE",
		7: "CANCEL",
		4: "KO",
		5: "SKIP",
	}
	TaskStatus_value = map[string]int32{
		"NA":     0,
		"OK":     1,
		"ERR":    2,
		"QUEUED": 3,
		"POKE":   6,
		"CANCEL": 7,
		"KO":     4,
		"SKIP":   5,
	}
)

func (x TaskStatus) Enum() *TaskStatus {
	p := new(TaskStatus)
	*p = x
	return p
}

func (x TaskStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TaskStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_tasks_proto_enumTypes[0].Descriptor()
}

func (TaskStatus) Type() protoreflect.EnumType {
	return &file_tasks_proto_enumTypes[0]
}

func (x TaskStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TaskStatus.Descriptor instead.
func (TaskStatus) EnumDescriptor() ([]byte, []int) {
	return file_tasks_proto_rawDescGZIP(), []int{0}
}

// Task keeps together the request and base data to be used by a processor
// to provide a result.
type Task struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Basic details
	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	JobId       string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	SiloEntryId string `protobuf:"bytes,3,opt,name=silo_entry_id,json=siloEntryId,proto3" json:"silo_entry_id,omitempty"` // was previously envelope_id
	OwnerId     string `protobuf:"bytes,9,opt,name=owner_id,json=ownerId,proto3" json:"owner_id,omitempty"`               // May be used for authentication
	Ref         string `protobuf:"bytes,10,opt,name=ref,proto3" json:"ref,omitempty"`                                     // If provided in a previous request, will be here too
	Action      string `protobuf:"bytes,12,opt,name=action,proto3" json:"action,omitempty"`                               // Action to be performed by the provider.
	// Token proves that the request came from the gateway and can be used
	// by the provider to make additional requests to the gateway on behalf
	// of the end-user.
	Token string `protobuf:"bytes,11,opt,name=token,proto3" json:"token,omitempty"`
	// Quick access to data
	State             string `protobuf:"bytes,13,opt,name=state,proto3" json:"state,omitempty"` // state of the silo entry
	Envelope          []byte `protobuf:"bytes,4,opt,name=envelope,proto3" json:"envelope,omitempty"`
	Config            []byte `protobuf:"bytes,5,opt,name=config,proto3" json:"config,omitempty"`
	EnvelopePublicUrl string `protobuf:"bytes,7,opt,name=envelope_public_url,json=envelopePublicUrl,proto3" json:"envelope_public_url,omitempty"`
	// Rerefence to any existing attachments
	Files []*File `protobuf:"bytes,6,rep,name=files,proto3" json:"files,omitempty"`
	// Tracking timestamp, issued by the gateway service. Includes nano seconds.
	Ts float64 `protobuf:"fixed64,8,opt,name=ts,proto3" json:"ts,omitempty"`
}

func (x *Task) Reset() {
	*x = Task{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tasks_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Task) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Task) ProtoMessage() {}

func (x *Task) ProtoReflect() protoreflect.Message {
	mi := &file_tasks_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Task.ProtoReflect.Descriptor instead.
func (*Task) Descriptor() ([]byte, []int) {
	return file_tasks_proto_rawDescGZIP(), []int{0}
}

func (x *Task) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Task) GetJobId() string {
	if x != nil {
		return x.JobId
	}
	return ""
}

func (x *Task) GetSiloEntryId() string {
	if x != nil {
		return x.SiloEntryId
	}
	return ""
}

func (x *Task) GetOwnerId() string {
	if x != nil {
		return x.OwnerId
	}
	return ""
}

func (x *Task) GetRef() string {
	if x != nil {
		return x.Ref
	}
	return ""
}

func (x *Task) GetAction() string {
	if x != nil {
		return x.Action
	}
	return ""
}

func (x *Task) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *Task) GetState() string {
	if x != nil {
		return x.State
	}
	return ""
}

func (x *Task) GetEnvelope() []byte {
	if x != nil {
		return x.Envelope
	}
	return nil
}

func (x *Task) GetConfig() []byte {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *Task) GetEnvelopePublicUrl() string {
	if x != nil {
		return x.EnvelopePublicUrl
	}
	return ""
}

func (x *Task) GetFiles() []*File {
	if x != nil {
		return x.Files
	}
	return nil
}

func (x *Task) GetTs() float64 {
	if x != nil {
		return x.Ts
	}
	return 0
}

// TaskResult says what we expect from a provider after attempting to complete
// a task.
type TaskResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status      TaskStatus `protobuf:"varint,1,opt,name=status,proto3,enum=invopop.provider.v1.TaskStatus" json:"status,omitempty"`
	Code        string     `protobuf:"bytes,2,opt,name=code,proto3" json:"code,omitempty"`                                   // optional provider response code
	Message     string     `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`                             // human message
	Ref         string     `protobuf:"bytes,11,opt,name=ref,proto3" json:"ref,omitempty"`                                    // optional reference to identify the result later
	Data        []byte     `protobuf:"bytes,9,opt,name=data,proto3" json:"data,omitempty"`                                   // New Envelope either complete or patched data
	ContentType string     `protobuf:"bytes,10,opt,name=content_type,json=contentType,proto3" json:"content_type,omitempty"` // Data content type to send to silo
	RetryIn     int32      `protobuf:"varint,6,opt,name=retry_in,json=retryIn,proto3" json:"retry_in,omitempty"`             // For QUEUED or ERR response, how long to wait to try again (optional).
}

func (x *TaskResult) Reset() {
	*x = TaskResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tasks_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskResult) ProtoMessage() {}

func (x *TaskResult) ProtoReflect() protoreflect.Message {
	mi := &file_tasks_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskResult.ProtoReflect.Descriptor instead.
func (*TaskResult) Descriptor() ([]byte, []int) {
	return file_tasks_proto_rawDescGZIP(), []int{1}
}

func (x *TaskResult) GetStatus() TaskStatus {
	if x != nil {
		return x.Status
	}
	return TaskStatus_NA
}

func (x *TaskResult) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *TaskResult) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *TaskResult) GetRef() string {
	if x != nil {
		return x.Ref
	}
	return ""
}

func (x *TaskResult) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *TaskResult) GetContentType() string {
	if x != nil {
		return x.ContentType
	}
	return ""
}

func (x *TaskResult) GetRetryIn() int32 {
	if x != nil {
		return x.RetryIn
	}
	return 0
}

// TaskPoke is used to wake up a task that is currently QUEUED.
type TaskPoke struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"` // Task (Intent) ID
	JobId   string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	Ref     string `protobuf:"bytes,3,opt,name=ref,proto3" json:"ref,omitempty"` // If id and job_id are not available
	Code    string `protobuf:"bytes,4,opt,name=code,proto3" json:"code,omitempty"`
	Message string `protobuf:"bytes,5,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *TaskPoke) Reset() {
	*x = TaskPoke{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tasks_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskPoke) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskPoke) ProtoMessage() {}

func (x *TaskPoke) ProtoReflect() protoreflect.Message {
	mi := &file_tasks_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskPoke.ProtoReflect.Descriptor instead.
func (*TaskPoke) Descriptor() ([]byte, []int) {
	return file_tasks_proto_rawDescGZIP(), []int{2}
}

func (x *TaskPoke) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TaskPoke) GetJobId() string {
	if x != nil {
		return x.JobId
	}
	return ""
}

func (x *TaskPoke) GetRef() string {
	if x != nil {
		return x.Ref
	}
	return ""
}

func (x *TaskPoke) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *TaskPoke) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

// TaskPokeResponse is the response to a TaskPoke request.
type TaskPokeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Err *Error `protobuf:"bytes,1,opt,name=err,proto3" json:"err,omitempty"`
}

func (x *TaskPokeResponse) Reset() {
	*x = TaskPokeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tasks_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskPokeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskPokeResponse) ProtoMessage() {}

func (x *TaskPokeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_tasks_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskPokeResponse.ProtoReflect.Descriptor instead.
func (*TaskPokeResponse) Descriptor() ([]byte, []int) {
	return file_tasks_proto_rawDescGZIP(), []int{3}
}

func (x *TaskPokeResponse) GetErr() *Error {
	if x != nil {
		return x.Err
	}
	return nil
}

var File_tasks_proto protoreflect.FileDescriptor

var file_tasks_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x74, 0x61, 0x73, 0x6b, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x69,
	0x6e, 0x76, 0x6f, 0x70, 0x6f, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x1a, 0x0b, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0c, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe7, 0x02,
	0x0a, 0x04, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x15, 0x0a, 0x06, 0x6a, 0x6f, 0x62, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6a, 0x6f, 0x62, 0x49, 0x64, 0x12, 0x22, 0x0a,
	0x0d, 0x73, 0x69, 0x6c, 0x6f, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x5f, 0x69, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x69, 0x6c, 0x6f, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x49,
	0x64, 0x12, 0x19, 0x0a, 0x08, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x09, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03,
	0x72, 0x65, 0x66, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x72, 0x65, 0x66, 0x12, 0x16,
	0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18,
	0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x14, 0x0a, 0x05,
	0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x65, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x2e, 0x0a, 0x13, 0x65, 0x6e, 0x76, 0x65, 0x6c, 0x6f,
	0x70, 0x65, 0x5f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x11, 0x65, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x50, 0x75, 0x62,
	0x6c, 0x69, 0x63, 0x55, 0x72, 0x6c, 0x12, 0x2f, 0x0a, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x18,
	0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x69, 0x6e, 0x76, 0x6f, 0x70, 0x6f, 0x70, 0x2e,
	0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x69, 0x6c, 0x65,
	0x52, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x73, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x01, 0x52, 0x02, 0x74, 0x73, 0x22, 0xd7, 0x01, 0x0a, 0x0a, 0x54, 0x61, 0x73, 0x6b,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x37, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1f, 0x2e, 0x69, 0x6e, 0x76, 0x6f, 0x70, 0x6f, 0x70,
	0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x61, 0x73,
	0x6b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63,
	0x6f, 0x64, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x72, 0x65, 0x66, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x72, 0x65, 0x66, 0x12,
	0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x72, 0x65, 0x74, 0x72, 0x79, 0x5f,
	0x69, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x72, 0x65, 0x74, 0x72, 0x79, 0x49,
	0x6e, 0x22, 0x71, 0x0a, 0x08, 0x54, 0x61, 0x73, 0x6b, 0x50, 0x6f, 0x6b, 0x65, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x15, 0x0a,
	0x06, 0x6a, 0x6f, 0x62, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6a,
	0x6f, 0x62, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x65, 0x66, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x72, 0x65, 0x66, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x22, 0x40, 0x0a, 0x10, 0x54, 0x61, 0x73, 0x6b, 0x50, 0x6f, 0x6b, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c, 0x0a, 0x03, 0x65, 0x72, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x69, 0x6e, 0x76, 0x6f, 0x70, 0x6f, 0x70, 0x2e,
	0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x52, 0x03, 0x65, 0x72, 0x72, 0x2a, 0x59, 0x0a, 0x0a, 0x54, 0x61, 0x73, 0x6b, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x06, 0x0a, 0x02, 0x4e, 0x41, 0x10, 0x00, 0x12, 0x06, 0x0a, 0x02,
	0x4f, 0x4b, 0x10, 0x01, 0x12, 0x07, 0x0a, 0x03, 0x45, 0x52, 0x52, 0x10, 0x02, 0x12, 0x0a, 0x0a,
	0x06, 0x51, 0x55, 0x45, 0x55, 0x45, 0x44, 0x10, 0x03, 0x12, 0x08, 0x0a, 0x04, 0x50, 0x4f, 0x4b,
	0x45, 0x10, 0x06, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x41, 0x4e, 0x43, 0x45, 0x4c, 0x10, 0x07, 0x12,
	0x06, 0x0a, 0x02, 0x4b, 0x4f, 0x10, 0x04, 0x12, 0x08, 0x0a, 0x04, 0x53, 0x4b, 0x49, 0x50, 0x10,
	0x05, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x3b, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_tasks_proto_rawDescOnce sync.Once
	file_tasks_proto_rawDescData = file_tasks_proto_rawDesc
)

func file_tasks_proto_rawDescGZIP() []byte {
	file_tasks_proto_rawDescOnce.Do(func() {
		file_tasks_proto_rawDescData = protoimpl.X.CompressGZIP(file_tasks_proto_rawDescData)
	})
	return file_tasks_proto_rawDescData
}

var file_tasks_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_tasks_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_tasks_proto_goTypes = []interface{}{
	(TaskStatus)(0),          // 0: invopop.provider.v1.TaskStatus
	(*Task)(nil),             // 1: invopop.provider.v1.Task
	(*TaskResult)(nil),       // 2: invopop.provider.v1.TaskResult
	(*TaskPoke)(nil),         // 3: invopop.provider.v1.TaskPoke
	(*TaskPokeResponse)(nil), // 4: invopop.provider.v1.TaskPokeResponse
	(*File)(nil),             // 5: invopop.provider.v1.File
	(*Error)(nil),            // 6: invopop.provider.v1.Error
}
var file_tasks_proto_depIdxs = []int32{
	5, // 0: invopop.provider.v1.Task.files:type_name -> invopop.provider.v1.File
	0, // 1: invopop.provider.v1.TaskResult.status:type_name -> invopop.provider.v1.TaskStatus
	6, // 2: invopop.provider.v1.TaskPokeResponse.err:type_name -> invopop.provider.v1.Error
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_tasks_proto_init() }
func file_tasks_proto_init() {
	if File_tasks_proto != nil {
		return
	}
	file_files_proto_init()
	file_errors_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_tasks_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Task); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tasks_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskResult); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tasks_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskPoke); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tasks_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskPokeResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_tasks_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_tasks_proto_goTypes,
		DependencyIndexes: file_tasks_proto_depIdxs,
		EnumInfos:         file_tasks_proto_enumTypes,
		MessageInfos:      file_tasks_proto_msgTypes,
	}.Build()
	File_tasks_proto = out.File
	file_tasks_proto_rawDesc = nil
	file_tasks_proto_goTypes = nil
	file_tasks_proto_depIdxs = nil
}
