// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: feed/feed.proto

package feed

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type InvalidateFeedCacheRequest_InvalidationType int32

const (
	InvalidateFeedCacheRequest_ALL  InvalidateFeedCacheRequest_InvalidationType = 0
	InvalidateFeedCacheRequest_POST InvalidateFeedCacheRequest_InvalidationType = 1
	InvalidateFeedCacheRequest_USER InvalidateFeedCacheRequest_InvalidationType = 2
)

// Enum value maps for InvalidateFeedCacheRequest_InvalidationType.
var (
	InvalidateFeedCacheRequest_InvalidationType_name = map[int32]string{
		0: "ALL",
		1: "POST",
		2: "USER",
	}
	InvalidateFeedCacheRequest_InvalidationType_value = map[string]int32{
		"ALL":  0,
		"POST": 1,
		"USER": 2,
	}
)

func (x InvalidateFeedCacheRequest_InvalidationType) Enum() *InvalidateFeedCacheRequest_InvalidationType {
	p := new(InvalidateFeedCacheRequest_InvalidationType)
	*p = x
	return p
}

func (x InvalidateFeedCacheRequest_InvalidationType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (InvalidateFeedCacheRequest_InvalidationType) Descriptor() protoreflect.EnumDescriptor {
	return file_feed_feed_proto_enumTypes[0].Descriptor()
}

func (InvalidateFeedCacheRequest_InvalidationType) Type() protoreflect.EnumType {
	return &file_feed_feed_proto_enumTypes[0]
}

func (x InvalidateFeedCacheRequest_InvalidationType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use InvalidateFeedCacheRequest_InvalidationType.Descriptor instead.
func (InvalidateFeedCacheRequest_InvalidationType) EnumDescriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{7, 0}
}

type HealthCheckResponse_Status int32

const (
	HealthCheckResponse_UNKNOWN     HealthCheckResponse_Status = 0
	HealthCheckResponse_SERVING     HealthCheckResponse_Status = 1
	HealthCheckResponse_NOT_SERVING HealthCheckResponse_Status = 2
)

// Enum value maps for HealthCheckResponse_Status.
var (
	HealthCheckResponse_Status_name = map[int32]string{
		0: "UNKNOWN",
		1: "SERVING",
		2: "NOT_SERVING",
	}
	HealthCheckResponse_Status_value = map[string]int32{
		"UNKNOWN":     0,
		"SERVING":     1,
		"NOT_SERVING": 2,
	}
)

func (x HealthCheckResponse_Status) Enum() *HealthCheckResponse_Status {
	p := new(HealthCheckResponse_Status)
	*p = x
	return p
}

func (x HealthCheckResponse_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (HealthCheckResponse_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_feed_feed_proto_enumTypes[1].Descriptor()
}

func (HealthCheckResponse_Status) Type() protoreflect.EnumType {
	return &file_feed_feed_proto_enumTypes[1]
}

func (x HealthCheckResponse_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use HealthCheckResponse_Status.Descriptor instead.
func (HealthCheckResponse_Status) EnumDescriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{10, 0}
}

type GetGlobalFeedRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserId        string                 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"` // ID пользователя, запрашивающего ленту (может быть пустым)
	Limit         int32                  `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
	Cursor        string                 `protobuf:"bytes,3,opt,name=cursor,proto3" json:"cursor,omitempty"`                  // курсор для пагинации
	WorldId       string                 `protobuf:"bytes,4,opt,name=world_id,json=worldId,proto3" json:"world_id,omitempty"` // ID мира, для которого запрашивается лента
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetGlobalFeedRequest) Reset() {
	*x = GetGlobalFeedRequest{}
	mi := &file_feed_feed_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetGlobalFeedRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetGlobalFeedRequest) ProtoMessage() {}

func (x *GetGlobalFeedRequest) ProtoReflect() protoreflect.Message {
	mi := &file_feed_feed_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetGlobalFeedRequest.ProtoReflect.Descriptor instead.
func (*GetGlobalFeedRequest) Descriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{0}
}

func (x *GetGlobalFeedRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *GetGlobalFeedRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *GetGlobalFeedRequest) GetCursor() string {
	if x != nil {
		return x.Cursor
	}
	return ""
}

func (x *GetGlobalFeedRequest) GetWorldId() string {
	if x != nil {
		return x.WorldId
	}
	return ""
}

type GetGlobalFeedResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Posts         []*PostInfo            `protobuf:"bytes,1,rep,name=posts,proto3" json:"posts,omitempty"`
	NextCursor    string                 `protobuf:"bytes,2,opt,name=next_cursor,json=nextCursor,proto3" json:"next_cursor,omitempty"`
	HasMore       bool                   `protobuf:"varint,3,opt,name=has_more,json=hasMore,proto3" json:"has_more,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetGlobalFeedResponse) Reset() {
	*x = GetGlobalFeedResponse{}
	mi := &file_feed_feed_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetGlobalFeedResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetGlobalFeedResponse) ProtoMessage() {}

func (x *GetGlobalFeedResponse) ProtoReflect() protoreflect.Message {
	mi := &file_feed_feed_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetGlobalFeedResponse.ProtoReflect.Descriptor instead.
func (*GetGlobalFeedResponse) Descriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{1}
}

func (x *GetGlobalFeedResponse) GetPosts() []*PostInfo {
	if x != nil {
		return x.Posts
	}
	return nil
}

func (x *GetGlobalFeedResponse) GetNextCursor() string {
	if x != nil {
		return x.NextCursor
	}
	return ""
}

func (x *GetGlobalFeedResponse) GetHasMore() bool {
	if x != nil {
		return x.HasMore
	}
	return false
}

type GetUserFeedRequest struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	UserId           string                 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`                                 // ID пользователя, ленту которого запрашиваем
	RequestingUserId string                 `protobuf:"bytes,2,opt,name=requesting_user_id,json=requestingUserId,proto3" json:"requesting_user_id,omitempty"` // ID пользователя, который делает запрос (может быть пустым)
	Limit            int32                  `protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
	Cursor           string                 `protobuf:"bytes,4,opt,name=cursor,proto3" json:"cursor,omitempty"` // курсор для пагинации
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *GetUserFeedRequest) Reset() {
	*x = GetUserFeedRequest{}
	mi := &file_feed_feed_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetUserFeedRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserFeedRequest) ProtoMessage() {}

func (x *GetUserFeedRequest) ProtoReflect() protoreflect.Message {
	mi := &file_feed_feed_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUserFeedRequest.ProtoReflect.Descriptor instead.
func (*GetUserFeedRequest) Descriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{2}
}

func (x *GetUserFeedRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *GetUserFeedRequest) GetRequestingUserId() string {
	if x != nil {
		return x.RequestingUserId
	}
	return ""
}

func (x *GetUserFeedRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *GetUserFeedRequest) GetCursor() string {
	if x != nil {
		return x.Cursor
	}
	return ""
}

type GetUserFeedResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Posts         []*PostInfo            `protobuf:"bytes,1,rep,name=posts,proto3" json:"posts,omitempty"`
	NextCursor    string                 `protobuf:"bytes,2,opt,name=next_cursor,json=nextCursor,proto3" json:"next_cursor,omitempty"`
	HasMore       bool                   `protobuf:"varint,3,opt,name=has_more,json=hasMore,proto3" json:"has_more,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetUserFeedResponse) Reset() {
	*x = GetUserFeedResponse{}
	mi := &file_feed_feed_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetUserFeedResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserFeedResponse) ProtoMessage() {}

func (x *GetUserFeedResponse) ProtoReflect() protoreflect.Message {
	mi := &file_feed_feed_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUserFeedResponse.ProtoReflect.Descriptor instead.
func (*GetUserFeedResponse) Descriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{3}
}

func (x *GetUserFeedResponse) GetPosts() []*PostInfo {
	if x != nil {
		return x.Posts
	}
	return nil
}

func (x *GetUserFeedResponse) GetNextCursor() string {
	if x != nil {
		return x.NextCursor
	}
	return ""
}

func (x *GetUserFeedResponse) GetHasMore() bool {
	if x != nil {
		return x.HasMore
	}
	return false
}

type PostInfo struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	UserId        string                 `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Caption       string                 `protobuf:"bytes,3,opt,name=caption,proto3" json:"caption,omitempty"`
	MediaId       string                 `protobuf:"bytes,4,opt,name=media_id,json=mediaId,proto3" json:"media_id,omitempty"`
	CreatedAt     int64                  `protobuf:"varint,5,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"` // Unix timestamp
	User          *UserInfo              `protobuf:"bytes,6,opt,name=user,proto3" json:"user,omitempty"`
	Stats         *PostStats             `protobuf:"bytes,7,opt,name=stats,proto3" json:"stats,omitempty"`
	MediaUrl      string                 `protobuf:"bytes,8,opt,name=media_url,json=mediaUrl,proto3" json:"media_url,omitempty"` // URL для доступа к медиафайлу
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PostInfo) Reset() {
	*x = PostInfo{}
	mi := &file_feed_feed_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PostInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostInfo) ProtoMessage() {}

func (x *PostInfo) ProtoReflect() protoreflect.Message {
	mi := &file_feed_feed_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostInfo.ProtoReflect.Descriptor instead.
func (*PostInfo) Descriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{4}
}

func (x *PostInfo) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *PostInfo) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *PostInfo) GetCaption() string {
	if x != nil {
		return x.Caption
	}
	return ""
}

func (x *PostInfo) GetMediaId() string {
	if x != nil {
		return x.MediaId
	}
	return ""
}

func (x *PostInfo) GetCreatedAt() int64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *PostInfo) GetUser() *UserInfo {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *PostInfo) GetStats() *PostStats {
	if x != nil {
		return x.Stats
	}
	return nil
}

func (x *PostInfo) GetMediaUrl() string {
	if x != nil {
		return x.MediaUrl
	}
	return ""
}

type UserInfo struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	Id                string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Username          string                 `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	ProfilePictureUrl string                 `protobuf:"bytes,3,opt,name=profile_picture_url,json=profilePictureUrl,proto3" json:"profile_picture_url,omitempty"` // Опционально
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *UserInfo) Reset() {
	*x = UserInfo{}
	mi := &file_feed_feed_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UserInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserInfo) ProtoMessage() {}

func (x *UserInfo) ProtoReflect() protoreflect.Message {
	mi := &file_feed_feed_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserInfo.ProtoReflect.Descriptor instead.
func (*UserInfo) Descriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{5}
}

func (x *UserInfo) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *UserInfo) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *UserInfo) GetProfilePictureUrl() string {
	if x != nil {
		return x.ProfilePictureUrl
	}
	return ""
}

type PostStats struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	LikesCount    int32                  `protobuf:"varint,1,opt,name=likes_count,json=likesCount,proto3" json:"likes_count,omitempty"`
	CommentsCount int32                  `protobuf:"varint,2,opt,name=comments_count,json=commentsCount,proto3" json:"comments_count,omitempty"`
	UserLiked     bool                   `protobuf:"varint,3,opt,name=user_liked,json=userLiked,proto3" json:"user_liked,omitempty"` // Лайкнул ли текущий пользователь этот пост
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PostStats) Reset() {
	*x = PostStats{}
	mi := &file_feed_feed_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PostStats) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostStats) ProtoMessage() {}

func (x *PostStats) ProtoReflect() protoreflect.Message {
	mi := &file_feed_feed_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostStats.ProtoReflect.Descriptor instead.
func (*PostStats) Descriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{6}
}

func (x *PostStats) GetLikesCount() int32 {
	if x != nil {
		return x.LikesCount
	}
	return 0
}

func (x *PostStats) GetCommentsCount() int32 {
	if x != nil {
		return x.CommentsCount
	}
	return 0
}

func (x *PostStats) GetUserLiked() bool {
	if x != nil {
		return x.UserLiked
	}
	return false
}

type InvalidateFeedCacheRequest struct {
	state         protoimpl.MessageState                      `protogen:"open.v1"`
	Type          InvalidateFeedCacheRequest_InvalidationType `protobuf:"varint,1,opt,name=type,proto3,enum=feed.InvalidateFeedCacheRequest_InvalidationType" json:"type,omitempty"`
	Id            string                                      `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"` // post_id или user_id в зависимости от типа
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InvalidateFeedCacheRequest) Reset() {
	*x = InvalidateFeedCacheRequest{}
	mi := &file_feed_feed_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InvalidateFeedCacheRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InvalidateFeedCacheRequest) ProtoMessage() {}

func (x *InvalidateFeedCacheRequest) ProtoReflect() protoreflect.Message {
	mi := &file_feed_feed_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InvalidateFeedCacheRequest.ProtoReflect.Descriptor instead.
func (*InvalidateFeedCacheRequest) Descriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{7}
}

func (x *InvalidateFeedCacheRequest) GetType() InvalidateFeedCacheRequest_InvalidationType {
	if x != nil {
		return x.Type
	}
	return InvalidateFeedCacheRequest_ALL
}

func (x *InvalidateFeedCacheRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type InvalidateFeedCacheResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InvalidateFeedCacheResponse) Reset() {
	*x = InvalidateFeedCacheResponse{}
	mi := &file_feed_feed_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InvalidateFeedCacheResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InvalidateFeedCacheResponse) ProtoMessage() {}

func (x *InvalidateFeedCacheResponse) ProtoReflect() protoreflect.Message {
	mi := &file_feed_feed_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InvalidateFeedCacheResponse.ProtoReflect.Descriptor instead.
func (*InvalidateFeedCacheResponse) Descriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{8}
}

func (x *InvalidateFeedCacheResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type HealthCheckRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HealthCheckRequest) Reset() {
	*x = HealthCheckRequest{}
	mi := &file_feed_feed_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HealthCheckRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthCheckRequest) ProtoMessage() {}

func (x *HealthCheckRequest) ProtoReflect() protoreflect.Message {
	mi := &file_feed_feed_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthCheckRequest.ProtoReflect.Descriptor instead.
func (*HealthCheckRequest) Descriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{9}
}

type HealthCheckResponse struct {
	state         protoimpl.MessageState     `protogen:"open.v1"`
	Status        HealthCheckResponse_Status `protobuf:"varint,1,opt,name=status,proto3,enum=feed.HealthCheckResponse_Status" json:"status,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HealthCheckResponse) Reset() {
	*x = HealthCheckResponse{}
	mi := &file_feed_feed_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HealthCheckResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthCheckResponse) ProtoMessage() {}

func (x *HealthCheckResponse) ProtoReflect() protoreflect.Message {
	mi := &file_feed_feed_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthCheckResponse.ProtoReflect.Descriptor instead.
func (*HealthCheckResponse) Descriptor() ([]byte, []int) {
	return file_feed_feed_proto_rawDescGZIP(), []int{10}
}

func (x *HealthCheckResponse) GetStatus() HealthCheckResponse_Status {
	if x != nil {
		return x.Status
	}
	return HealthCheckResponse_UNKNOWN
}

var File_feed_feed_proto protoreflect.FileDescriptor

const file_feed_feed_proto_rawDesc = "" +
	"\n" +
	"\x0ffeed/feed.proto\x12\x04feed\"x\n" +
	"\x14GetGlobalFeedRequest\x12\x17\n" +
	"\auser_id\x18\x01 \x01(\tR\x06userId\x12\x14\n" +
	"\x05limit\x18\x02 \x01(\x05R\x05limit\x12\x16\n" +
	"\x06cursor\x18\x03 \x01(\tR\x06cursor\x12\x19\n" +
	"\bworld_id\x18\x04 \x01(\tR\aworldId\"y\n" +
	"\x15GetGlobalFeedResponse\x12$\n" +
	"\x05posts\x18\x01 \x03(\v2\x0e.feed.PostInfoR\x05posts\x12\x1f\n" +
	"\vnext_cursor\x18\x02 \x01(\tR\n" +
	"nextCursor\x12\x19\n" +
	"\bhas_more\x18\x03 \x01(\bR\ahasMore\"\x89\x01\n" +
	"\x12GetUserFeedRequest\x12\x17\n" +
	"\auser_id\x18\x01 \x01(\tR\x06userId\x12,\n" +
	"\x12requesting_user_id\x18\x02 \x01(\tR\x10requestingUserId\x12\x14\n" +
	"\x05limit\x18\x03 \x01(\x05R\x05limit\x12\x16\n" +
	"\x06cursor\x18\x04 \x01(\tR\x06cursor\"w\n" +
	"\x13GetUserFeedResponse\x12$\n" +
	"\x05posts\x18\x01 \x03(\v2\x0e.feed.PostInfoR\x05posts\x12\x1f\n" +
	"\vnext_cursor\x18\x02 \x01(\tR\n" +
	"nextCursor\x12\x19\n" +
	"\bhas_more\x18\x03 \x01(\bR\ahasMore\"\xef\x01\n" +
	"\bPostInfo\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x17\n" +
	"\auser_id\x18\x02 \x01(\tR\x06userId\x12\x18\n" +
	"\acaption\x18\x03 \x01(\tR\acaption\x12\x19\n" +
	"\bmedia_id\x18\x04 \x01(\tR\amediaId\x12\x1d\n" +
	"\n" +
	"created_at\x18\x05 \x01(\x03R\tcreatedAt\x12\"\n" +
	"\x04user\x18\x06 \x01(\v2\x0e.feed.UserInfoR\x04user\x12%\n" +
	"\x05stats\x18\a \x01(\v2\x0f.feed.PostStatsR\x05stats\x12\x1b\n" +
	"\tmedia_url\x18\b \x01(\tR\bmediaUrl\"f\n" +
	"\bUserInfo\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x1a\n" +
	"\busername\x18\x02 \x01(\tR\busername\x12.\n" +
	"\x13profile_picture_url\x18\x03 \x01(\tR\x11profilePictureUrl\"r\n" +
	"\tPostStats\x12\x1f\n" +
	"\vlikes_count\x18\x01 \x01(\x05R\n" +
	"likesCount\x12%\n" +
	"\x0ecomments_count\x18\x02 \x01(\x05R\rcommentsCount\x12\x1d\n" +
	"\n" +
	"user_liked\x18\x03 \x01(\bR\tuserLiked\"\xa4\x01\n" +
	"\x1aInvalidateFeedCacheRequest\x12E\n" +
	"\x04type\x18\x01 \x01(\x0e21.feed.InvalidateFeedCacheRequest.InvalidationTypeR\x04type\x12\x0e\n" +
	"\x02id\x18\x02 \x01(\tR\x02id\"/\n" +
	"\x10InvalidationType\x12\a\n" +
	"\x03ALL\x10\x00\x12\b\n" +
	"\x04POST\x10\x01\x12\b\n" +
	"\x04USER\x10\x02\"7\n" +
	"\x1bInvalidateFeedCacheResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\"\x14\n" +
	"\x12HealthCheckRequest\"\x84\x01\n" +
	"\x13HealthCheckResponse\x128\n" +
	"\x06status\x18\x01 \x01(\x0e2 .feed.HealthCheckResponse.StatusR\x06status\"3\n" +
	"\x06Status\x12\v\n" +
	"\aUNKNOWN\x10\x00\x12\v\n" +
	"\aSERVING\x10\x01\x12\x0f\n" +
	"\vNOT_SERVING\x10\x022\xbb\x02\n" +
	"\vFeedService\x12H\n" +
	"\rGetGlobalFeed\x12\x1a.feed.GetGlobalFeedRequest\x1a\x1b.feed.GetGlobalFeedResponse\x12B\n" +
	"\vGetUserFeed\x12\x18.feed.GetUserFeedRequest\x1a\x19.feed.GetUserFeedResponse\x12Z\n" +
	"\x13InvalidateFeedCache\x12 .feed.InvalidateFeedCacheRequest\x1a!.feed.InvalidateFeedCacheResponse\x12B\n" +
	"\vHealthCheck\x12\x18.feed.HealthCheckRequest\x1a\x19.feed.HealthCheckResponseB,Z*github.com/sdshorin/generia/api/proto/feedb\x06proto3"

var (
	file_feed_feed_proto_rawDescOnce sync.Once
	file_feed_feed_proto_rawDescData []byte
)

func file_feed_feed_proto_rawDescGZIP() []byte {
	file_feed_feed_proto_rawDescOnce.Do(func() {
		file_feed_feed_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_feed_feed_proto_rawDesc), len(file_feed_feed_proto_rawDesc)))
	})
	return file_feed_feed_proto_rawDescData
}

var file_feed_feed_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_feed_feed_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_feed_feed_proto_goTypes = []any{
	(InvalidateFeedCacheRequest_InvalidationType)(0), // 0: feed.InvalidateFeedCacheRequest.InvalidationType
	(HealthCheckResponse_Status)(0),                  // 1: feed.HealthCheckResponse.Status
	(*GetGlobalFeedRequest)(nil),                     // 2: feed.GetGlobalFeedRequest
	(*GetGlobalFeedResponse)(nil),                    // 3: feed.GetGlobalFeedResponse
	(*GetUserFeedRequest)(nil),                       // 4: feed.GetUserFeedRequest
	(*GetUserFeedResponse)(nil),                      // 5: feed.GetUserFeedResponse
	(*PostInfo)(nil),                                 // 6: feed.PostInfo
	(*UserInfo)(nil),                                 // 7: feed.UserInfo
	(*PostStats)(nil),                                // 8: feed.PostStats
	(*InvalidateFeedCacheRequest)(nil),               // 9: feed.InvalidateFeedCacheRequest
	(*InvalidateFeedCacheResponse)(nil),              // 10: feed.InvalidateFeedCacheResponse
	(*HealthCheckRequest)(nil),                       // 11: feed.HealthCheckRequest
	(*HealthCheckResponse)(nil),                      // 12: feed.HealthCheckResponse
}
var file_feed_feed_proto_depIdxs = []int32{
	6,  // 0: feed.GetGlobalFeedResponse.posts:type_name -> feed.PostInfo
	6,  // 1: feed.GetUserFeedResponse.posts:type_name -> feed.PostInfo
	7,  // 2: feed.PostInfo.user:type_name -> feed.UserInfo
	8,  // 3: feed.PostInfo.stats:type_name -> feed.PostStats
	0,  // 4: feed.InvalidateFeedCacheRequest.type:type_name -> feed.InvalidateFeedCacheRequest.InvalidationType
	1,  // 5: feed.HealthCheckResponse.status:type_name -> feed.HealthCheckResponse.Status
	2,  // 6: feed.FeedService.GetGlobalFeed:input_type -> feed.GetGlobalFeedRequest
	4,  // 7: feed.FeedService.GetUserFeed:input_type -> feed.GetUserFeedRequest
	9,  // 8: feed.FeedService.InvalidateFeedCache:input_type -> feed.InvalidateFeedCacheRequest
	11, // 9: feed.FeedService.HealthCheck:input_type -> feed.HealthCheckRequest
	3,  // 10: feed.FeedService.GetGlobalFeed:output_type -> feed.GetGlobalFeedResponse
	5,  // 11: feed.FeedService.GetUserFeed:output_type -> feed.GetUserFeedResponse
	10, // 12: feed.FeedService.InvalidateFeedCache:output_type -> feed.InvalidateFeedCacheResponse
	12, // 13: feed.FeedService.HealthCheck:output_type -> feed.HealthCheckResponse
	10, // [10:14] is the sub-list for method output_type
	6,  // [6:10] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_feed_feed_proto_init() }
func file_feed_feed_proto_init() {
	if File_feed_feed_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_feed_feed_proto_rawDesc), len(file_feed_feed_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_feed_feed_proto_goTypes,
		DependencyIndexes: file_feed_feed_proto_depIdxs,
		EnumInfos:         file_feed_feed_proto_enumTypes,
		MessageInfos:      file_feed_feed_proto_msgTypes,
	}.Build()
	File_feed_feed_proto = out.File
	file_feed_feed_proto_goTypes = nil
	file_feed_feed_proto_depIdxs = nil
}
