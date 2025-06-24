%%%-------------------------------------------------------------------
%% @doc Client module for grpc service mainflux.UsersService.
%% @end
%%%-------------------------------------------------------------------

%% this module was generated on 2019-10-27T15:11:30+00:00 and should not be modified manually

-module(mainflux_users_service_client).

-compile(export_all).
-compile(nowarn_export_all).

-include_lib("grpcbox/include/grpcbox.hrl").

-define(is_ctx(Ctx), is_tuple(Ctx) andalso element(1, Ctx) =:= ctx).

-define(SERVICE, 'mainflux.UsersService').
-define(PROTO_MODULE, 'internal_pb').
-define(MARSHAL_FUN(T), fun(I) -> ?PROTO_MODULE:encode_msg(I, T) end).
-define(UNMARSHAL_FUN(T), fun(I) -> ?PROTO_MODULE:decode_msg(I, T) end).
-define(DEF(Input, Output, MessageType), #grpcbox_def{service=?SERVICE,
                                                      message_type=MessageType,
                                                      marshal_fun=?MARSHAL_FUN(Input),
                                                      unmarshal_fun=?UNMARSHAL_FUN(Output)}).

%% @doc Unary RPC
-spec identify(internal_pb:token()) ->
    {ok, internal_pb:user_id(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response().
identify(Input) ->
    identify(ctx:new(), Input, #{}).

-spec identify(ctx:t() | internal_pb:token(), internal_pb:token() | grpcbox_client:options()) ->
    {ok, internal_pb:user_id(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response().
identify(Ctx, Input) when ?is_ctx(Ctx) ->
    identify(Ctx, Input, #{});
identify(Input, Options) ->
    identify(ctx:new(), Input, Options).

-spec identify(ctx:t(), internal_pb:token(), grpcbox_client:options()) ->
    {ok, internal_pb:user_id(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response().
identify(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/mainflux.UsersService/Identify">>, Input, ?DEF(token, user_id, <<"mainflux.Token">>), Options).

