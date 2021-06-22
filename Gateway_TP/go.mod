module gateway_tp/gateway_test

go 1.14

replace gateway_tp/config => ./config

replace gateway_tp/fetcher => ./fetcher

replace gateway_tp/types => ./types

replace gateway_tp/api => ./api

replace gateway_tp/forwarder => ./forwarder

replace gateway_tp/chain => ./_chain

replace gateway_tp/logger => ./logger

replace gateway_tp/keyman => ./keyman

replace gateway_tp/crypto => ./crypto

replace gateway_tp/mondrian => ./mondrian

require (
	gateway_tp/api v0.0.0-00010101000000-000000000000 // indirect
	gateway_tp/chain v0.0.0-00010101000000-000000000000
	gateway_tp/config v0.0.0-00010101000000-000000000000
	gateway_tp/crypto v0.0.0-00010101000000-000000000000 // indirect
	gateway_tp/fetcher v0.0.0-00010101000000-000000000000
	gateway_tp/forwarder v0.0.0-00010101000000-000000000000
	gateway_tp/keyman v0.0.0-00010101000000-000000000000
	gateway_tp/logger v0.0.0-00010101000000-000000000000
	gateway_tp/mondrian v0.0.0-00010101000000-000000000000 // indirect
	gateway_tp/types v0.0.0-00010101000000-000000000000
	github.com/dchest/cmac v0.0.0-20150527144652-62ff55a1048c
)
