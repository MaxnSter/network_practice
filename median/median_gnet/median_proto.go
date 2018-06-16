package median_gnet

import (
	"reflect"

	"github.com/MaxnSter/gnet/message_pack/message_meta"
)

var (
	_ message_meta.MetaIdentifier = (*GenerateRequest)(nil)
)

type GenerateRequest struct {
	Id    uint32
	Count int
	Min   int
	Max   int
}

type QueryResponse struct {
	Id    uint32
	Count int
	Sum   int
	Min   int
	Max   int
}

type SearchRequest struct {
	Id    uint32
	Guess int
}

type SearchResponse struct {
	Id      uint32
	Smaller int
	Same    int
}

const (
	IdGenerateRequest = iota
	IdQueryResponse
	IdSearchRequest
	IdSearchResponse
)

func (r *GenerateRequest) GetId() uint32 { return IdGenerateRequest }
func (r *QueryResponse) GetId() uint32   { return IdQueryResponse }
func (r *SearchRequest) GetId() uint32   { return IdSearchRequest }
func (r *SearchResponse) GetId() uint32  { return IdSearchResponse }

func init() {
	message_meta.RegisterMsgMeta(&message_meta.MessageMeta{ID: IdGenerateRequest,
		Type: reflect.TypeOf((*GenerateRequest)(nil)).Elem()})

	message_meta.RegisterMsgMeta(&message_meta.MessageMeta{ID: IdQueryResponse,
		Type: reflect.TypeOf((*QueryResponse)(nil)).Elem()})

	message_meta.RegisterMsgMeta(&message_meta.MessageMeta{ID: IdSearchRequest,
		Type: reflect.TypeOf((*SearchRequest)(nil)).Elem()})

	message_meta.RegisterMsgMeta(&message_meta.MessageMeta{ID: IdSearchResponse,
		Type: reflect.TypeOf((*SearchResponse)(nil)).Elem()})
}
