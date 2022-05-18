// Licensed to ClickHouse, Inc. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. ClickHouse, Inc. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package proto

import (
	"fmt"
	"regexp"
	"router/app/clickhouse/lib/binary"

	"go.opentelemetry.io/otel/trace"
)

type Query struct {
	ID             string
	Span           trace.SpanContext
	Body           string
	QuotaKey       string
	Settings       Settings
	Compression    bool
	InitialUser    string
	InitialAddress string
	TableName      string
}

func (q *Query) Decode(decoder *binary.Decoder /*, revision uint64*/) error {
	var err error
	if q.ID, err = decoder.String(); err != nil {
		return err
	}
	// client_info
	if err = q.decodeClientInfo(decoder /*, revision*/); err != nil {
		return err
	}
	// settings
	//if err := q.Settings.Decode(decoder /*, revision*/); err != nil {
	//	return err
	//}
	//if _, err = decoder.String( /* empty string is a marker of the end of setting */ ); err != nil {
	//	return err
	//}

	//if revision >= DBMS_MIN_REVISION_WITH_INTERSERVER_SECRET {
	//	decoder.String("")
	//}
	//{
	//	if _, err = decoder.ReadByte(); err != nil { // StateComplete
	//		return err
	//	}
	//	if q.Compression, err = decoder.Bool(); err != nil {
	//		return err
	//	}
	//}
	//if q.Body, err = decoder.String(); err != nil {
	//	return err
	//}

	fmt.Printf("%v\n", q.Body)

	return err
}

//func (q *Query) Encode(encoder *binary.Encoder, revision uint64) error {
//	if err := encoder.String(q.ID); err != nil {
//		return err
//	}
//	// client_info
//	if err := q.encodeClientInfo(encoder, revision); err != nil {
//		return err
//	}
//	// settings
//	if err := q.Settings.Encode(encoder, revision); err != nil {
//		return err
//	}
//	encoder.String("" /* empty string is a marker of the end of setting */)
//
//	if revision >= DBMS_MIN_REVISION_WITH_INTERSERVER_SECRET {
//		encoder.String("")
//	}
//	{
//		encoder.Byte(StateComplete)
//		encoder.Bool(q.Compression)
//	}
//	return encoder.String(q.Body)
//}

func (q *Query) decodeClientInfo(decoder *binary.Decoder /*, revision uint64*/) error {
	var osUser string
	var hostname string
	var clientName string
	var varName string
	var varValue uint32

	var err error
	d, _ := decoder.ReadByte() // ClientQueryInitial
	fmt.Printf("%v\n", d)
	if q.InitialUser, err = decoder.String(); err != nil { // initial_user
		return err
	}
	decoder.String() // "" initial_query_id
	if q.InitialAddress, err = decoder.String(); err != nil {
		return err
	}
	//if revision >= DBMS_MIN_PROTOCOL_VERSION_WITH_INITIAL_QUERY_START_TIME {
	//	decoder.Int64(0) // initial_query_start_time_microseconds
	//}
	for i := 0; i < 9; i++ {
		d, _ = decoder.ReadByte()
		fmt.Printf("%v\n", d)
	}
	{
		if osUser, err = decoder.String(); err != nil {
			return err
		}
		if hostname, err = decoder.String(); err != nil { // client name
			return err
		}
		if clientName, err = decoder.String(); err != nil { // client name
			return err
		}
		for i := 0; i < 12; i++ {
			d, _ = decoder.ReadByte()
			fmt.Printf("%v\n", d)
		}
		if varName, err = decoder.String(); err != nil {
			return err
		}
		for i := 0; i < 1; i++ {
			d, _ = decoder.ReadByte()
			fmt.Printf("%v\n", d)
		}
		//if varValue, err = decoder.UInt32(); err != nil {
		//	return err
		//}
		for i := 0; i < 7; i++ {
			d, _ = decoder.ReadByte()
			fmt.Printf("%v\n", d)
		}
		if q.Body, err = decoder.String(); err != nil {
			return err
		}
		if _, err = decoder.Uvarint(); err != nil {
			return err
		}
	}
	fmt.Printf("%v\n", osUser)
	fmt.Printf("%v\n", hostname)
	fmt.Printf("%v\n", clientName)
	fmt.Printf("%v\n", varName)
	fmt.Printf("%v\n", varValue)

	rxp, _ := regexp.Compile(`FROM ([_\-\w0-9]+)`)
	tname := rxp.FindSubmatch([]byte(q.Body))
	fmt.Printf("%v\n", string(tname[1]))
	q.TableName = string(tname[1])
	for i := 0; i < 11; i++ {
		d, _ = decoder.ReadByte()
		fmt.Printf("%x\n", d)
	}
	//if revision >= DBMS_MIN_REVISION_WITH_QUOTA_KEY_IN_CLIENT_INFO {
	//	decoder.String(q.QuotaKey)
	//}
	//if revision >= DBMS_MIN_PROTOCOL_VERSION_WITH_DISTRIBUTED_DEPTH {
	//	decoder.Uvarint(0)
	//}
	//if revision >= DBMS_MIN_REVISION_WITH_VERSION_PATCH {
	//	decoder.Uvarint(0)
	//}
	//if revision >= DBMS_MIN_REVISION_WITH_OPENTELEMETRY {
	//	switch {
	//	case q.Span.IsValid():
	//		decoder.Byte(1)
	//		{
	//			v := q.Span.TraceID()
	//			decoder.Raw(v[:])
	//		}
	//		{
	//			v := q.Span.SpanID()
	//			decoder.Raw(v[:])
	//		}
	//		decoder.String(q.Span.TraceState().String())
	//		decoder.Byte(byte(q.Span.TraceFlags()))
	//
	//	default:
	//		decoder.Byte(0)
	//	}
	//}
	//if revision >= DBMS_MIN_REVISION_WITH_PARALLEL_REPLICAS {
	//	decoder.Uvarint(0) // collaborate_with_initiator
	//	decoder.Uvarint(0) // count_participating_replicas
	//	decoder.Uvarint(0) // number_of_current_replica
	//}
	return nil
}

//func (q *Query) encodeClientInfo(encoder *binary.Encoder, revision uint64) error {
//	encoder.Byte(ClientQueryInitial)
//	encoder.String(q.InitialUser)    // initial_user
//	encoder.String("")               // initial_query_id
//	encoder.String(q.InitialAddress) // initial_address
//	if revision >= DBMS_MIN_PROTOCOL_VERSION_WITH_INITIAL_QUERY_START_TIME {
//		encoder.Int64(0) // initial_query_start_time_microseconds
//	}
//	encoder.Byte(1) // interface [tcp - 1, http - 2]
//	{
//		encoder.String(osUser)
//		encoder.String(hostname)
//		encoder.String("!!!!!")
//		encoder.Uvarint(ClientVersionMajor)
//		encoder.Uvarint(ClientVersionMinor)
//		encoder.Uvarint(ClientTCPProtocolVersion)
//	}
//	if revision >= DBMS_MIN_REVISION_WITH_QUOTA_KEY_IN_CLIENT_INFO {
//		encoder.String(q.QuotaKey)
//	}
//	if revision >= DBMS_MIN_PROTOCOL_VERSION_WITH_DISTRIBUTED_DEPTH {
//		encoder.Uvarint(0)
//	}
//	if revision >= DBMS_MIN_REVISION_WITH_VERSION_PATCH {
//		encoder.Uvarint(0)
//	}
//	if revision >= DBMS_MIN_REVISION_WITH_OPENTELEMETRY {
//		switch {
//		case q.Span.IsValid():
//			encoder.Byte(1)
//			{
//				v := q.Span.TraceID()
//				encoder.Raw(v[:])
//			}
//			{
//				v := q.Span.SpanID()
//				encoder.Raw(v[:])
//			}
//			encoder.String(q.Span.TraceState().String())
//			encoder.Byte(byte(q.Span.TraceFlags()))
//
//		default:
//			encoder.Byte(0)
//		}
//	}
//	if revision >= DBMS_MIN_REVISION_WITH_PARALLEL_REPLICAS {
//		encoder.Uvarint(0) // collaborate_with_initiator
//		encoder.Uvarint(0) // count_participating_replicas
//		encoder.Uvarint(0) // number_of_current_replica
//	}
//	return nil
//}

type Settings []Setting

type Setting struct {
	Key   string
	Value interface{}
}

func (s Settings) Decode(decoder *binary.Decoder /*, revision uint64*/) error {
	for _, s := range s {
		if err := s.decode(decoder /*, revision*/); err != nil {
			return err
		}
	}
	return nil
}

//func (s Settings) Encode(encoder *binary.Encoder, revision uint64) error {
//	for _, s := range s {
//		if err := s.encode(encoder, revision); err != nil {
//			return err
//		}
//	}
//	return nil
//}

func (s *Setting) decode(encoder *binary.Decoder /*, revision uint64*/) error {
	var err error
	if s.Key, err = encoder.String(); err != nil {
		return err
	}
	//if revision <= DBMS_MIN_REVISION_WITH_SETTINGS_SERIALIZED_AS_STRINGS {
	//	var value uint64
	//	switch v := s.Value.(type) {
	//	case int:
	//		value = uint64(v)
	//	case bool:
	//		if value = 0; v {
	//			value = 1
	//		}
	//	default:
	//		return fmt.Errorf("query setting %s has unsupported data type", s.Key)
	//	}
	//	return encoder.Uvarint(value)
	//}
	if _, err = encoder.Bool(); err != nil { // true is_important
		return err
	}
	if s.Value, err = encoder.String(); err != nil {
		return err
	}

	return err
}

//func (s *Setting) encode(encoder *binary.Encoder, revision uint64) error {
//	if err := encoder.String(s.Key); err != nil {
//		return err
//	}
//	if revision <= DBMS_MIN_REVISION_WITH_SETTINGS_SERIALIZED_AS_STRINGS {
//		var value uint64
//		switch v := s.Value.(type) {
//		case int:
//			value = uint64(v)
//		case bool:
//			if value = 0; v {
//				value = 1
//			}
//		default:
//			return fmt.Errorf("query setting %s has unsupported data type", s.Key)
//		}
//		return encoder.Uvarint(value)
//	}
//	if err := encoder.Bool(true); err != nil { // is_important
//		return err
//	}
//	return encoder.String(fmt.Sprint(s.Value))
//}
