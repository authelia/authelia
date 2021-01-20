--
-- HTTP 1.1 library for HAProxy Lua modules
--
-- The library is loosely modeled after Python's Requests Library
-- using the same field names and very similar calling conventions for
-- "HTTP verb" methods (where we use Lua specific named parameter support)
--
-- In addition to client side, the library also supports server side request
-- parsing, where we utilize HAProxy Lua API for all heavy lifting.
--
--
-- Copyright (c) 2017-2020. Adis Nezirović <anezirovic@haproxy.com>
-- Copyright (c) 2017-2020. HAProxy Technologies, LLC.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--    http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
--
-- SPDX-License-Identifier: Apache-2.0

local _author = "Adis Nezirovic <anezirovic@haproxy.com>"
local _copyright = "Copyright 2017-2020. HAProxy Technologies, LLC."
local _version = "1.0.0"

local json = require "json"

-- Utility functions

-- HTTP headers fetch helper
--
-- Returns a header value(s) according to strategy (fold by default):
--  - single/string value for "fold", "first" and "last" strategies
--  - table for "all" strategy (for single value, a table with single element)
--
-- @param hdrs table Headers table as received by http.get and friends
-- @param name string Header name
-- @param strategy string "multiple header values" handling strategy
-- @return header value (string or table) or nil
local function get_header(hdrs, name, strategy)
    if hdrs == nil or name == nil then return nil end

    local v = hdrs[name:lower()]
    if type(v) ~= "table" and strategy ~= "all" then return v end

    if strategy == nil or strategy == "fold" then
        return table.concat(v, ",")
    elseif strategy == "first" then
        return v[1]
    elseif strategy == "last" then
        return v[#v]
    elseif strategy == "all" then
        if type(v) ~= "table" then
            return {v}
        else
            return v
        end
    end
end



-- HTTP headers iterator helper
--
-- Returns key/value pairs for all header, making sure that returned values
-- are always of string type (if necessary, it folds multiple headers with
-- the same name)
--
-- @param hdrs table Headers table as received by http.get and friends
-- @return header name/value iterator (suitable for use in "for" loops)
local function get_headers_folded(hdrs)
    if hdrs == nil then
        return function() end
    end

    local function iter(t, k)
        local v
        k, v = next(t, k)

        if v ~= nil then
            if type(v) ~= "table" then
                return k, v
            else
                return k, table.concat(v, ",")
            end
        end
    end

    return iter, hdrs, nil
end

-- HTTP headers iterator
--
-- Returns key/value pairs for all headers, for multiple headers with same name
-- it will return every name/value pair
-- (i.e. you can safely use it to process responses with 'Set-Cookie' header)
--
-- @param hdrs table Headers table as received by http.get and friends
-- @return header name/value iterator (suitable for use in "for" loops)
local function get_headers_flattened(hdrs)
    if hdrs == nil then
        return function() end
    end

    local k  -- top level key (string)
    local k_sub = 0  -- sub table key (integer), 0 if item not a table,
                     -- nil after last sub table iteration
    local v_sub  -- sub table

    return function ()
        local v
        if k_sub == 0 then
            k, v = next(hdrs, k)
            if k == nil then return end
        else
            k_sub, v = next(v_sub, k_sub)

            if k_sub == nil then
                k_sub = 0
                k, v = next(hdrs, k)
            end
        end

        if k == nil then return end

        if type(v) ~= "table" then
            return k, v
        else
            v_sub = v
            k_sub = k_sub + 1
            return k, v[k_sub]
        end
    end
end


--- Parse request cookies from string
--
-- @param s Lua string with value of cookie header (can be nil)
--
-- @return Table with parsed cookies or nil
local function parse_request_cookies(s)
    if s == nil then return nil end
    idx = 1
    cookies = {}

    while idx < s:len() do
        i, j = s:find("; ", idx)

        if i == nil then
            k, v = string.match(s:sub(idx), "^(.-)=(.*)$")
            if k then cookies[k] = v end
            break
        end

        k, v = string.match(s:sub(idx, i-1), "^(.-)=(.*)$")
        if k then cookies[k] = v end
        idx = j + 1
    end

    if next(cookies) == nil then
        return nil
    else
        return cookies
    end
end


--- Namespace object which hosts HTTP verb methods and request/response classes
local M = {}


--- HTTP response class
M.response = {}
M.response.__index = M.response

local _reason = {
    [200] = "OK",
    [201] = "Created",
    [204] = "No Content",
    [301] = "Moved Permanently",
    [302] = "Found",
    [400] = "Bad Request",
    [403] = "Forbidden",
    [404] = "Not Found",
    [405] = "Method Not Allowed",
    [408] = "Request Timeout",
    [413] = "Payload Too Large",
    [429] = "Too many requests",
    [500] = "Internal Server Error",
    [501] = "Not Implemented",
    [502] = "Bad Gateway",
    [503] = "Service Unavailable",
    [504] = "Gateway Timeout"
}

--- Creates HTTP response from scratch
--
-- @param status_code HTTP status code
-- @param reason HTTP status code text (e.g. "OK" for 200 response)
-- @param headers HTTP response headers
-- @param request The HTTP request which triggered the response
-- @param encoding Default encoding for response or conversions
--
-- @return response object
function M.response.create(t)
    local self = setmetatable({}, M.response)

    if not t then
        t =  {}
    end

    self.status_code = t.status_code or nil
    self.reason = t.reason or _reason[self.status_code] or ""
    self.headers = t.headers or {}
    self.content = t.content or ""
    self.request = t.request or nil
    self.encoding = t.encoding or "utf-8"

    return self
end

function M.response.send(self, applet)
    applet:set_status(tonumber(self.status_code), self.reason)

    for k, v in pairs(self.headers) do
        if type(v) == "table" then
            for _, hdr_val in pairs(v) do
                applet:add_header(k, hdr_val)
            end
        else
            applet:add_header(k, v)
        end
    end

    if not self.headers["content-type"] then
        if type(self.content) == "table" then
            applet:add_header("content-type", "application/json; charset=" ..
                              self.encoding)
            if next(self.content) == nil then
                -- Return empty JSON object for empty Lua tables
                -- (that makes more sense then returning [])
                self.content = "{}"
            else
                self.content = json.encode(self.content)
            end
        else
            applet:add_header("content-type", "text/plain; charset=" ..
                              self.encoding)
        end
    end

    if not self.headers["content-length"] then
        applet:add_header("content-length", #tostring(self.content))
    end

    applet:start_response()
    applet:send(tostring(self.content))
end

--- Convert response content to JSON
--
-- @return Lua table (decoded json)
function M.response.json(self)
    return json.decode(self.content)
end

-- Response headers getter
--
-- Returns a header value(s) according to strategy (fold by default):
--  - single/string value for "fold", "first" and "last" strategies
--  - table for "all" strategy (for single value, a table with single element)
--
-- @param name string Header name
-- @param strategy string "multiple header values" handling strategy
-- @return header value (string or table) or nil
function M.response.get_header(self, name, strategy)
    return get_header(self.headers, name, strategy)
end

-- Response headers iterator
--
-- Yields key/value pairs for all headers, making sure that returned values
-- are always of string type
--
-- @param folded boolean Specifies whether to fold headers with same name
-- @return header name/value iterator (suitable for use in "for" loops)
function M.response.get_headers(self, folded)
    if folded == true then
        return get_headers_folded(self.headers)
    else
        return get_headers_flattened(self.headers)
    end
end


--- HTTP request class (client or server side, depending on the constructor)
M.request = {}
M.request.__index = M.request

--- HTTP request constructor
--
-- Parses client HTTP request (as forwarded by HAProxy)
--
-- @param applet HAProxy AppletHTTP Lua object
--
-- @return Request object
function M.request.parse(applet)
    local self = setmetatable({}, M.request)
    self.method = applet.method

    if (applet.method == "POST" or applet.method == "PUT") and
            applet.length > 0 then
        self.data = applet:receive()
        if self.data == "" then self.data = nil end
    end

    self.headers = {}
    for k, v in pairs(applet.headers) do
        if (v[1]) then  -- (non folded header with multiple values)
            self.headers[k] = {}
            for _, val in pairs(v) do
                table.insert(self.headers[k], val)
            end
        else
            self.headers[k] = v[0]
        end
    end

    if not self.headers["host"] then
        return nil, "Bad request, no Host header specified"
    end

    self.cookies = parse_request_cookies(self.headers["cookie"])

    -- TODO: Patch ApletHTTP and add schema of request
    local schema = applet.schema or "http"
    local url = {schema, "://", self.headers["host"], applet.path}

    self.params = {}
    if applet.qs:len() > 0 then
        for _, arg in ipairs(core.tokenize(applet.qs, "&", true)) do
            kv = core.tokenize(arg, "=", true)
            self.params[kv[1]] = kv[2]
        end
        url[#url+1] = "?"
        url[#url+1] = applet.qs
    end

    self.url = table.concat(url)

    return self
end

--- Parse HTTP POST data
--
-- @return Table with submitted form data
function M.request.parse_multipart(self)
    local result ={}
    local ct = self.headers['content-type']
    local body = self.data

    if ct:match('^multipart/form[-]data;') then
        local boundary = ct:match('^multipart/form[-]data; boundary=(.+)$')
        if boundary == nil then
            return nil, 'Could not parse boundary from Content-Type'
        end

        local i = 1
        local j
        local old_i

        while true do
            i, j = body:find(boundary, i)

            if i == nil then break end

            if old_i then
                local part = body:sub(old_i, i - 1)
                local k, fn, t, v = part:match('^\r\n[cC]ontent[-][dD]isposition: form[-]data; name[=]"(.+)"; filename="(.+)"\r\n[cC]ontent[-][tT]ype: (.+)\r\n\r\n(.+)\r\n$')

                if k then
                    result[k] = {
                        filename = fn,
                        content_type = t,
                        data = v
                    }
                else
                    k, v = part:match('^\r\n[cC]ontent[-][dD]isposition: form[-]data; name[=]"(.+)"\r\n\r\n(.+)\r\n$')

                    if k then
                        result[k] = v
                    end
                end

            end

            i = j + 1
            old_i = i
        end
    elseif ct == 'application/x-www-form-urlencoded' then
        local i = 1
        local j
        while true do
            j = body:find('&', i)
            if j == nil then break end

            local part = body:sub(i, j-1)
            local k, v = part:match('^(.+)=(.+)$')
            if k then
                result[k] = v
            end
            i = j + 1
        end
    else
        return nil, 'Unsupported Content-Type: ' .. ct
    end

    if not next(result) then
        return nil, 'Could not parse form data'
    end

    return result
end

--- Reads (all) chunks from a HTTP response
--
-- @param socket socket object (with already established tcp connection)
-- @param get_all boolean (true by default), collect all chunks at once
--                or yield every chunk separately.
--
-- @return Full response payload or nil and an error message
function M.receive_chunked(socket, get_all)
    if socket == nil then
        return nil, "http.receive_chunked: Socket is nil"
    end
    local data = {}

    while true do
        local chunk, err = socket:receive("*l")

        if chunk == nil then
            return nil, "http.receive_chunked(): Receive error (chunk length): " .. tostring(err)
        end

        local chunk_len = tonumber(chunk, 16)
        if chunk_len == nil then
            return nil, "http.receive_chunked(): Could not parse chunk length"
        end

        if chunk_len == 0 then
            -- TODO: support trailers
            break
        end

        -- Consume next chunk (including the \r\n)
        chunk, err = socket:receive(chunk_len+2)
        if chunk == nil then
            return nil, "http.receive_chunked(): Receive error (chunk data): " .. tostring(err)
        end

        -- Strip the \r\n before collection
        local chunk_data = string.sub(chunk, 1, -3)

        if get_all == false then
            return chunk_data
        end

        table.insert(data, chunk_data)
    end

    return table.concat(data)
end


-- Request headers getter
--
-- Returns a header value(s) according to strategy (fold by default):
--  - single/string value for "fold", "first" and "last" strategies
--  - table for "all" strategy (for single value, a table with single element)
--
-- @param name string Header name
-- @param strategy string "multiple header values" handling strategy
-- @return header value (string or table) or nil
function M.request.get_header(self, name, strategy)
    return get_header(self.headers, name, strategy)
end

-- Request headers iterator
--
-- Yields key/value pairs for all headers, making sure that returned values
-- are always of string type
--
-- @param hdrs table Headers table as received by http.get and friends
-- @param folded boolean Specifies whether to fold headers with same name
-- @return header name/value iterator (suitable for use in "for" loops)
function M.request.get_headers(self, folded)
    if folded == true then
        return get_headers_folded(self.headers)
    else
        return get_headers_flattened(self.headers)
    end
end

--- Creates HTTP request from scratch
--
-- @param method HTTP method
-- @param url Valid HTTP url
-- @param headers Lua table with request headers
-- @param data Request content
-- @param params Lua table with request url arguments
-- @param auth (username, password) tuple for HTTP auth
--
-- @return request object
function M.request.create(t)
    local self = setmetatable({}, M.request)

    if t.method then
        self.method = t.method:lower()
    else
        self.method = "get"
    end
    self.url = t.url or nil
    self.headers = t.headers or {}
    self.data = t.data or nil
    self.params = t.params or {}
    self.auth = t.auth or {}

    return self
end

--- HTTP HEAD request
function M.head(t)
    return M.send("HEAD", t)
end

--- HTTP GET request
function M.get(t)
    return M.send("GET", t)
end

--- HTTP PUT request
function M.put(t)
    return M.send("PUT", t)
end

--- HTTP POST request
function M.post(t)
    return M.send("POST", t)
end

--- HTTP DELETE request
function M.delete(t)
    return M.send("DELETE", t)
end


--- Send HTTP request
--
-- @param method HTTP method
-- @param url Valid HTTP url (mandatory)
-- @param headers Lua table with request headers
-- @param data Request content
-- @param params Lua table with request url arguments
-- @param auth (username, password) tuple for HTTP auth
-- @param timeout Optional timeout for socket operations (5s by default)
--
-- @return Response object or tuple (nil, msg) on errors

-- Note that the prefered way to call this method is via Lua
-- "keyword arguments" convention, e.g.
--   http.get{uri="http://example.net"}
function M.send(method, t)
    if type(t) ~= "table" then
        return nil, "http." .. method:lower() ..
            ": expecting Request object for named parameters"
    end

    if type(t.url) ~= "string" then
        return nil, "http." .. method:lower() .. ": 'url' parameter missing"
    end

    local socket = core.tcp()
    socket:settimeout(t.timeout or 5)
    local connect
    if t.url:sub(1, 7) ~= "http://" and t.url:sub(1, 8) ~= "https://" then
        t.url = "http://" .. t.url
    end
    local schema, host, req_uri = t.url:match("^(.*)://(.-)(/.*)$")

    if not schema then
        -- maybe path (request uri) is missing
        schema, host = t.url:match("^(.*)://(.-)$")
        if not schema then
            return nil, "http." .. method:lower() .. ": Could not parse URL: " .. t.url
        end
        req_uri = "/"
    end

    local addr, port = host:match("(.*):(%d+)")

    if schema == "http" then
        connect = socket.connect
        if not port then
            addr = host
            port = 80
        end
    elseif schema == "https" then
        connect = socket.connect_ssl
        if not port then
            addr = host
            port = 443
        end
    else
        return nil, "http." .. method:lower() .. ": Invalid URL schema " .. tostring(schema)
    end

    local c, err = connect(socket, addr, port)

    if c then
        local req = {}
        local hdr_tbl = {}

        if t.headers then
            for k, v in pairs(t.headers) do
                if type(v) == "table" then
                    table.insert(hdr_tbl, k .. ": " .. table.concat(v, ","))
                else
                    table.insert(hdr_tbl, k .. ": " .. tostring(v))
                end
            end
        else
            t.headers = {}  -- dummy table
        end

        if not t.headers.host then
            -- 'Host' header must be provided for HTTP/1.1
            table.insert(hdr_tbl, "host: " .. host)
        end

        if not t.headers["accept"] then
            table.insert(hdr_tbl, "accept: */*")
        end

        if not t.headers["user-agent"] then
            table.insert(hdr_tbl, "user-agent: haproxy-lua-http/1.0")
        end

        if not t.headers.connection then
            table.insert(hdr_tbl, "connection: close")
        end

        if t.data then
            req[4] = t.data
            if not t.headers or not t.headers["content-length"] then
                table.insert(hdr_tbl, "content-length: " .. tostring(#t.data))
            end
        end

        req[1] = method .. " " .. req_uri .. " HTTP/1.1\r\n"
        req[2] = table.concat(hdr_tbl, "\r\n")
        req[3] = "\r\n\r\n"

        local r, e = socket:send(table.concat(req))

        if not r then
            socket:close()
            return nil, "http." .. method:lower() .. ": " .. tostring(e)
        end

        local line
        r = M.response.create()

        while true do
            line, err = socket:receive("*l")

            if not line then
                socket:close()
                return nil, "http." .. method:lower() ..
                       ": Receive error (headers): "  .. err
            end

            if line == "" then break end

            if not r.status_code then
                _, r.status_code, r.reason =
                    line:match("(HTTP/1.[01]) (%d%d%d)(.*)")
                if not _ then
                    socket:close()
                    return nil, "http." .. method:lower() ..
                        ": Could not parse request line"
                end
                r.status_code = tonumber(r.status_code)
            else
                local sep = line:find(":")
                local hdr_name = line:sub(1, sep-1):lower()
                local hdr_val = line:sub(sep+1):match("^%s*(.*%S)%s*$") or ""

                if r.headers[hdr_name] == nil then
                    r.headers[hdr_name] = hdr_val
                elseif type(r.headers[hdr_name]) == "table" then
                    table.insert(r.headers[hdr_name], hdr_val)
                else
                    r.headers[hdr_name] = {
                        r.headers[hdr_name],
                        hdr_val
                    }
                end
            end
        end

        if method:lower() == "head" then
            r.content = nil
            socket:close()
            return r
        end

        if r.headers["content-length"] and tonumber(r.headers["content-length"]) > 0 then
            r.content, err = socket:receive("*a")

            if not r.content then
                socket:close()
                return nil, "http." .. method:lower() ..
                       ": Receive error (content): " .. err
            end
        end

        if r.headers["transfer-encoding"] and r.headers["transfer-encoding"] == "chunked" then
            r.content, err = M.receive_chunked(socket)
            if r.content == nil then
                socket:close()
                return nil, err
            end
        end

        socket:close()
        return r
    else
        return nil, "http." .. method:lower() .. ": Connection error: " .. tostring(err)
    end
end

M.base64 = {}

--- URL safe base64 encoder
--
-- Padding ('=') is omited, as permited per RFC
--   https://tools.ietf.org/html/rfc4648
-- in order to follow JSON Web Signature RFC
--   https://tools.ietf.org/html/rfc7515
--
-- @param s String (can be binary data) to encode
-- @param enc Function which implements base64 encoder (e.g. HAProxy base64 fetch)
-- @return Encoded string
function M.base64.encode(s, enc)
    if not s then return nil end
    local u = enc(s)

    if not u then
        return nil
    end

    local pad_len = 2 - ((#s-1) % 3)

    if pad_len > 0 then
        return u:sub(1, - pad_len - 1):gsub('[+]', '-'):gsub('[/]', '_')
    else
        return u:gsub('[+]', '-'):gsub('[/]', '_')
    end
end

--- URLsafe base64 decoder
--
-- @param s Base64 string to decode
-- @param dec Function which implements base64 decoder (e.g. HAProxy b64dec fetch)
-- @return Decoded string (can be binary data)
function M.base64.decode(s, dec)
    if not s then return nil end

    local e = s:gsub('[-]', '+'):gsub('[_]', '/')
    return dec(e .. string.rep('=', 3 - ((#s - 1) % 4)))
end

return M
