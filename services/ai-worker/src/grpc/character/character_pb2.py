# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: character/character.proto
# Protobuf Python Version: 5.29.0
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    29,
    0,
    '',
    'character/character.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x19\x63haracter/character.proto\x12\x11generia.character\"\xba\x01\n\x16\x43reateCharacterRequest\x12\x10\n\x08world_id\x18\x01 \x01(\t\x12\x19\n\x0creal_user_id\x18\x02 \x01(\tH\x00\x88\x01\x01\x12\x14\n\x0c\x64isplay_name\x18\x03 \x01(\t\x12\x1c\n\x0f\x61vatar_media_id\x18\x04 \x01(\tH\x01\x88\x01\x01\x12\x11\n\x04meta\x18\x05 \x01(\tH\x02\x88\x01\x01\x42\x0f\n\r_real_user_idB\x12\n\x10_avatar_media_idB\x07\n\x05_meta\"+\n\x13GetCharacterRequest\x12\x14\n\x0c\x63haracter_id\x18\x01 \x01(\t\"D\n\x1fGetUserCharactersInWorldRequest\x12\x0f\n\x07user_id\x18\x01 \x01(\t\x12\x10\n\x08world_id\x18\x02 \x01(\t\"\xd0\x01\n\x16UpdateCharacterRequest\x12\x14\n\x0c\x63haracter_id\x18\x01 \x01(\t\x12\x19\n\x0c\x64isplay_name\x18\x02 \x01(\tH\x00\x88\x01\x01\x12\x1c\n\x0f\x61vatar_media_id\x18\x03 \x01(\tH\x01\x88\x01\x01\x12\x17\n\navatar_url\x18\x04 \x01(\tH\x02\x88\x01\x01\x12\x11\n\x04meta\x18\x05 \x01(\tH\x03\x88\x01\x01\x42\x0f\n\r_display_nameB\x12\n\x10_avatar_media_idB\r\n\x0b_avatar_urlB\x07\n\x05_meta\"\xf0\x01\n\tCharacter\x12\n\n\x02id\x18\x01 \x01(\t\x12\x10\n\x08world_id\x18\x02 \x01(\t\x12\x19\n\x0creal_user_id\x18\x03 \x01(\tH\x00\x88\x01\x01\x12\r\n\x05is_ai\x18\x04 \x01(\x08\x12\x14\n\x0c\x64isplay_name\x18\x05 \x01(\t\x12\x1c\n\x0f\x61vatar_media_id\x18\x06 \x01(\tH\x01\x88\x01\x01\x12\x12\n\navatar_url\x18\x07 \x01(\t\x12\x11\n\x04meta\x18\x08 \x01(\tH\x02\x88\x01\x01\x12\x12\n\ncreated_at\x18\t \x01(\tB\x0f\n\r_real_user_idB\x12\n\x10_avatar_media_idB\x07\n\x05_meta\"A\n\rCharacterList\x12\x30\n\ncharacters\x18\x01 \x03(\x0b\x32\x1c.generia.character.Character\"\x14\n\x12HealthCheckRequest\"%\n\x13HealthCheckResponse\x12\x0e\n\x06status\x18\x01 \x01(\t2\xf0\x03\n\x10\x43haracterService\x12Z\n\x0f\x43reateCharacter\x12).generia.character.CreateCharacterRequest\x1a\x1c.generia.character.Character\x12Z\n\x0fUpdateCharacter\x12).generia.character.UpdateCharacterRequest\x1a\x1c.generia.character.Character\x12T\n\x0cGetCharacter\x12&.generia.character.GetCharacterRequest\x1a\x1c.generia.character.Character\x12p\n\x18GetUserCharactersInWorld\x12\x32.generia.character.GetUserCharactersInWorldRequest\x1a .generia.character.CharacterList\x12\\\n\x0bHealthCheck\x12%.generia.character.HealthCheckRequest\x1a&.generia.character.HealthCheckResponseB\'Z%github.com/generia/api/grpc/characterb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'character.character_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z%github.com/generia/api/grpc/character'
  _globals['_CREATECHARACTERREQUEST']._serialized_start=49
  _globals['_CREATECHARACTERREQUEST']._serialized_end=235
  _globals['_GETCHARACTERREQUEST']._serialized_start=237
  _globals['_GETCHARACTERREQUEST']._serialized_end=280
  _globals['_GETUSERCHARACTERSINWORLDREQUEST']._serialized_start=282
  _globals['_GETUSERCHARACTERSINWORLDREQUEST']._serialized_end=350
  _globals['_UPDATECHARACTERREQUEST']._serialized_start=353
  _globals['_UPDATECHARACTERREQUEST']._serialized_end=561
  _globals['_CHARACTER']._serialized_start=564
  _globals['_CHARACTER']._serialized_end=804
  _globals['_CHARACTERLIST']._serialized_start=806
  _globals['_CHARACTERLIST']._serialized_end=871
  _globals['_HEALTHCHECKREQUEST']._serialized_start=873
  _globals['_HEALTHCHECKREQUEST']._serialized_end=893
  _globals['_HEALTHCHECKRESPONSE']._serialized_start=895
  _globals['_HEALTHCHECKRESPONSE']._serialized_end=932
  _globals['_CHARACTERSERVICE']._serialized_start=935
  _globals['_CHARACTERSERVICE']._serialized_end=1431
# @@protoc_insertion_point(module_scope)
