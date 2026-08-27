-module(durable_pipeline_kernel_ffi).
-export([append_sync/2, read_lines_recover/1, append_torn/2, write_raw/2,
         read_raw/1, delete/1, sha256_hex/1]).

append_sync(Path, Line) ->
    case ensure_parent(Path) of
        ok ->
            case file:open(Path, [append, binary, raw]) of
                {ok, Io} ->
                    Result = case file:write(Io, [Line, <<"\n">>]) of
                        ok -> file:sync(Io);
                        Error -> Error
                    end,
                    _ = file:close(Io),
                    gleam_result(Result);
                {error, Reason} -> {error, reason(Reason)}
            end;
        {error, Reason} -> {error, reason(Reason)}
    end.

read_lines_recover(Path) ->
    case file:read_file(Path) of
        {ok, <<>>} -> {ok, []};
        {ok, Data} -> recover_lines(Path, Data);
        {error, enoent} -> {ok, []};
        {error, Reason} -> {error, reason(Reason)}
    end.

recover_lines(Path, Data) ->
    case binary:last(Data) of
        $\n -> {ok, split_lines(Data)};
        _ -> recover_torn_lines(Path, Data)
    end.

recover_torn_lines(Path, Data) ->
    Matches = binary:matches(Data, <<"\n">>),
    Keep = case Matches of
        [] -> 0;
        _ ->
            {Position, 1} = lists:last(Matches),
            Position + 1
    end,
    case truncate_to(Path, Keep) of
        ok -> {ok, split_lines(binary:part(Data, 0, Keep))};
        {error, Reason} -> {error, reason(Reason)}
    end.

truncate_to(Path, Keep) ->
    case file:open(Path, [read, write, binary, raw]) of
        {ok, Io} ->
            Result = case file:position(Io, Keep) of
                {ok, _} ->
                    case file:truncate(Io) of
                        ok -> file:sync(Io);
                        Error -> Error
                    end;
                Error -> Error
            end,
            _ = file:close(Io),
            Result;
        Error -> Error
    end.

split_lines(<<>>) -> [];
split_lines(Data) ->
    Parts = binary:split(Data, <<"\n">>, [global]),
    case lists:reverse(Parts) of
        [<<>> | Rest] -> lists:reverse(Rest);
        _ -> Parts
    end.

append_torn(Path, Bytes) ->
    case ensure_parent(Path) of
        ok -> gleam_result(file:write_file(Path, Bytes, [append, binary]));
        {error, Reason} -> {error, reason(Reason)}
    end.

write_raw(Path, Bytes) ->
    case ensure_parent(Path) of
        ok -> gleam_result(file:write_file(Path, Bytes, [write, binary, sync]));
        {error, Reason} -> {error, reason(Reason)}
    end.

read_raw(Path) ->
    case file:read_file(Path) of
        {ok, Bytes} -> {ok, Bytes};
        {error, Reason} -> {error, reason(Reason)}
    end.

delete(Path) ->
    case file:delete(Path) of
        ok -> {ok, nil};
        {error, enoent} -> {ok, nil};
        {error, Reason} -> {error, reason(Reason)}
    end.

sha256_hex(Bytes) ->
    binary:encode_hex(crypto:hash(sha256, Bytes), lowercase).

ensure_parent(Path) -> filelib:ensure_dir(binary_to_list(Path)).

gleam_result(ok) -> {ok, nil};
gleam_result({error, Reason}) -> {error, reason(Reason)}.

reason(Reason) -> unicode:characters_to_binary(io_lib:format("~p", [Reason])).
